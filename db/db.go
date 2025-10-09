package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	server_config "mqtt-mochi-server/config"
)

type Message struct {
	Topic     string      `json:"topic"`
	Payload   interface{} `json:"payload"`
	Frequency int         `json:"frequency"`
}

type AppRouter struct {
	Router *mux.Router
	DB     *sql.DB
}

func InitDB(cfg server_config.DB_Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		cfg.Server_Address, cfg.Server_Port, cfg.Username, cfg.Database_Name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS messages (
            id SERIAL PRIMARY KEY,
            topic TEXT NOT NULL,
            payload JSONB,
			frequency INTEGER NOT NULL DEFAULT 0
        );
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}

func (ar *AppRouter) GetDB() *sql.DB {
	return ar.DB
}

func HandlePostRequest(w http.ResponseWriter, r *http.Request) {
	var msg Message

	// function that returns the db connection from DB in AppRouter
	ar := r.Context().Value("db").(*AppRouter)

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		http.Error(w, "Failed to marshal payload", http.StatusInternalServerError)
		return
	}

	_, err = ar.DB.Exec("INSERT INTO messages (topic, payload, frequency) VALUES ($1, $2, $3)", msg.Topic, payloadBytes, msg.Frequency)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert message into database: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Inserted message with Topic=%s, Payload=%s, Frequency=%d\n", msg.Topic, string(payloadBytes), msg.Frequency)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Message added successfully"})
}

func FetchMessages(db *sql.DB) ([]Message, error) {
	rows, err := db.Query("SELECT topic, payload, frequency FROM messages")
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var payloadBytes []byte
		if err := rows.Scan(&msg.Topic, &payloadBytes, &msg.Frequency); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Unmarshal the JSONB payload back into the interface{}
		if err := json.Unmarshal(payloadBytes, &msg.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return messages, nil
}

func GetTopics(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT topic FROM messages")
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			log.Printf("failed to scan row: %v", err)
			continue
		}
		topics = append(topics, topic)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return topics, nil
}

type Ack struct {
	ID     string `json:"id"`
	Token  string `json:"token"`
	Result bool   `json:"result"`
}

type Incoming_Message struct {
	Topic     string `json:"topic"`
	Payload   []byte `json:"payload"`
	Frequency int    `json:"frequency"`
}

func HandleAddToDB(db *sql.DB, topic string, msg Incoming_Message) error {
	_, err := db.Exec("INSERT INTO messages (topic, payload, frequency) VALUES ($1, $2, $3)", msg.Topic, msg.Payload, msg.Frequency)
	if err != nil {
		fmt.Printf("Failed to insert message into database: %v", err)
	}

	log.Printf("Inserted message with Topic=%s, Payload=%s, Frequency=%d\n", msg.Topic, string(msg.Payload), msg.Frequency)

	return nil
}
