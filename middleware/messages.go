package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Message struct {
	ID        int         `json:"id"`
	Topic     string      `json:"topic"`
	Payload   interface{} `json:"payload"`
	Frequency int         `json:"frequency"`
}

func GetIndex(w http.ResponseWriter, r *http.Request) {
	Respond_With_JSON(w, http.StatusOK, "Welcome! This is the index route for the MQTT server. Nothing much happening here...")
}

func PostMessage(w http.ResponseWriter, r *http.Request) {
	var msg Message

	ar, ok := r.Context().Value("appRouter").(*AppRouter)
	if !ok || ar == nil || ar.DB == nil {
		// http.Error(w, "Database connection not available", http.StatusInternalServerError)
		// return
		Respond_With_JSON(w, http.StatusInternalServerError, "Database connection not available")
	}

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		Respond_With_JSON(w, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
	}
	defer r.Body.Close()

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, "Failed to marshal payload")
	}

	_, err = ar.DB.Exec("INSERT INTO messages (topic, payload, frequency) VALUES ($1, $2, $3)", msg.Topic, payloadBytes, msg.Frequency)
	if err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to insert message into database: %v", err))
	}

	Respond_With_JSON(w, http.StatusOK, "Message added successfully")
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	ar, ok := r.Context().Value("appRouter").(*AppRouter)
	if !ok || ar == nil || ar.DB == nil {
		Respond_With_JSON(w, http.StatusInternalServerError, "Database connection not available")
	}

	rows, err := ar.DB.Query("SELECT topic, payload, frequency FROM messages")
	if err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query messages: %v", err))
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var payloadBytes []byte
		if err := rows.Scan(&msg.Topic, &payloadBytes, &msg.Frequency); err != nil {
			Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to scan row: %v", err))
		}

		// Unmarshal the JSONB payload back into the interface{}
		if err := json.Unmarshal(payloadBytes, &msg.Payload); err != nil {
			Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal payload: %v", err))
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Rows iteration error: %v", err))
	}

	Respond_With_JSON(w, http.StatusOK, messages)
}

func GetMessageByID(w http.ResponseWriter, r *http.Request) {
	ar, ok := r.Context().Value("appRouter").(*AppRouter)
	if !ok || ar == nil || ar.DB == nil {
		Respond_With_JSON(w, http.StatusInternalServerError, "Database connection not available")
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		Respond_With_JSON(w, http.StatusBadRequest, "Missing 'id' parameter")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		Respond_With_JSON(w, http.StatusBadRequest, "Invalid 'id' parameter")
		return
	}

	var msg Message
	var payloadBytes []byte
	err = ar.DB.QueryRow("SELECT id, topic, payload, frequency FROM messages WHERE id = $1", id).Scan(&msg.ID, &msg.Topic, &payloadBytes, &msg.Frequency)
	if err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve message: %v", err))
		return
	}

	if err := json.Unmarshal(payloadBytes, &msg.Payload); err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal payload: %v", err))
		return
	}
	Respond_With_JSON(w, http.StatusOK, msg)
}

func PutMessage(w http.ResponseWriter, r *http.Request) {
	ar, ok := r.Context().Value("appRouter").(*AppRouter)
	if !ok || ar == nil || ar.DB == nil {
		Respond_With_JSON(w, http.StatusInternalServerError, "Database connection not available")
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		Respond_With_JSON(w, http.StatusBadRequest, "Missing 'id' parameter")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		Respond_With_JSON(w, http.StatusBadRequest, "Invalid 'id' parameter")
	}

	var msg Message

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		Respond_With_JSON(w, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
	}
	defer r.Body.Close()

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, "Failed to marshal payload")
	}

	_, err = ar.DB.Exec("UPDATE messages SET topic = $1, payload = $2, frequency = $3 WHERE id = $4", msg.Topic, payloadBytes, msg.Frequency, id)
	if err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update message: %v", err))
	}

	Respond_With_JSON(w, http.StatusOK, msg)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	ar, ok := r.Context().Value("appRouter").(*AppRouter)
	if !ok || ar == nil || ar.DB == nil {
		Respond_With_JSON(w, http.StatusInternalServerError, "Database connection not available")
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		Respond_With_JSON(w, http.StatusBadRequest, "Missing 'id' parameter")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		Respond_With_JSON(w, http.StatusBadRequest, "Invalid 'id' parameter")
	}

	_, err = ar.DB.Exec("DELETE FROM messages WHERE id = $1", id)
	if err != nil {
		Respond_With_JSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update message: %v", err))
	}

	Respond_With_JSON(w, http.StatusOK, fmt.Sprintf("Message with ID %d deleted successfully", id))
}

type json_Result struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type WS_json_Result struct {
	Type string      `json:"topic"`
	Data interface{} `json:"data"`
}

func Respond_With_JSON(w http.ResponseWriter, code int, payload interface{}) {
	respond_With_JSON(w, code, payload, "")
}

func respond_With_JSON(w http.ResponseWriter, code int, payload interface{}, message string) {

	var result_status string
	var result_message string
	if code == http.StatusOK {
		result_status = "success"
		result_message = "ok"
	} else {
		result_status = "error"
		result_message = message
	}

	result := &json_Result{
		Status:  result_status,
		Message: result_message,
		Data:    payload,
	}

	response_body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed to encode a JSON response: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	_, err = w.Write(response_body)
	if err != nil {
		log.Printf("Failed to write the response body: %v", err)
	}
}
