package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	server_config "mqtt-mochi-server/config"
	"mqtt-mochi-server/db"
	router "mqtt-mochi-server/web"

	"github.com/gorilla/mux"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Program struct {
	Server *mqtt.Server
	Hook   *mqtt.HookBase
}

type AppConfig struct {
	Router *mux.Router
	DB     *sql.DB
}

func publisherManager(server *mqtt.Server, routes *router.AppRouter, dbConn *sql.DB, restartChan chan struct{}) {
	var tickers []*time.Ticker

	stopPublishing := func() {
		for _, ticker := range tickers {
			ticker.Stop()
		}
		tickers = nil
	}

	startPublishing := func() {
		messages, err := db.FetchMessages(dbConn)
		if err != nil {
			server.Log.Error("Failed to fetch messages from database", "error", err)
			return
		}

		if len(messages) == 0 {
			server.Log.Info("No messages found in the database to publish.")
			return
		}

		server.Log.Info("Fetched messages from the database and starting to publish")

		for _, msg := range messages {
			if msg.Frequency > 0 {
				ticker := time.NewTicker(time.Duration(msg.Frequency) * time.Second)
				tickers = append(tickers, ticker)

				go func(m db.Message, t *time.Ticker) {
					for range t.C {
						publishMessage(server, routes, m)
					}
				}(msg, ticker)
			} else {
				publishMessage(server, routes, msg)
			}
		}
	}

	startPublishing()

	for {
		<-restartChan
		server.Log.Info("Restarting publisher due to new message.")
		stopPublishing()
		startPublishing()
	}
}

func publishMessage(server *mqtt.Server, routes *router.AppRouter, msg db.Message) {
	if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
		if _, ok := payloadMap["ts"]; ok {
			payloadMap["ts"] = time.Now().Unix()
		} else if _, ok := payloadMap["timestamp"]; ok {
			payloadMap["timestamp"] = time.Now().Unix()
		}
		msg.Payload = payloadMap
	}

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		server.Log.Error("Failed to marshal payload for publishing", "topic", msg.Topic, "error", err)
		return
	}

	err = server.Publish(msg.Topic, payload, false, 0)
	if err != nil {
		server.Log.Error("Failed to publish message", "topic", msg.Topic, "error", err)
	} else {
		server.Log.Info("Published message", "topic", msg.Topic)
		routes.WSHub.BroadcastMessage(msg.Topic, msg.Payload)
	}
}

func startPublisher(server *mqtt.Server, routes *router.AppRouter, msg db.Message) {
	if msg.Frequency == 0 {

		// Update "ts" or "timestamp" if present in payload
		if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
			if _, ok := payloadMap["ts"]; ok {
				payloadMap["ts"] = time.Now().Unix()
			} else if _, ok := payloadMap["timestamp"]; ok {
				payloadMap["timestamp"] = time.Now().Unix()
			}
			msg.Payload = payloadMap
		} else {
			server.Log.Warn("Payload is not a map[string]interface{}, cannot update timestamp", "topic", msg.Topic)
		}

		payload, err := json.Marshal(msg.Payload) // Ensure payload is []byte
		if err != nil {
			server.Log.Error("Failed to marshal payload for publishing", "topic", msg.Topic, "error", err)
		}

		err = server.Publish(msg.Topic, payload, false, 0)
		if err != nil {
			server.Log.Error("Failed to publish message", "topic", msg.Topic, "error", err)
		} else {
			server.Log.Info("Published message from database once", "topic", msg.Topic)
			routes.WSHub.BroadcastMessage(msg.Topic, msg.Payload)
			// TODO: Implement removal or marking as published in the database to avoid republishing
		}
	} else {
		ticker := time.NewTicker(time.Duration(msg.Frequency) * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C

			// Update "ts" or "timestamp" if present in payload
			if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
				if _, ok := payloadMap["ts"]; ok {
					payloadMap["ts"] = time.Now().Unix()
				} else if _, ok := payloadMap["timestamp"]; ok {
					payloadMap["timestamp"] = time.Now().Unix()
				}
				msg.Payload = payloadMap
			} else {
				server.Log.Warn("Payload is not a map[string]interface{}, cannot update timestamp", "topic", msg.Topic)
			}

			payload, err := json.Marshal(msg.Payload) // Ensure payload is []byte
			if err != nil {
				server.Log.Error("Failed to marshal payload for publishing", "topic", msg.Topic, "error", err)
				continue
			}

			err = server.Publish(msg.Topic, payload, false, 0)
			if err != nil {
				server.Log.Error("Failed to publish message", "topic", msg.Topic, "error", err)
			} else {
				routes.WSHub.BroadcastMessage(msg.Topic, msg.Payload)
				server.Log.Info("Published message from database with frequency", "topic", msg.Topic, "frequency", msg.Frequency)
			}
			time.Sleep(time.Duration(msg.Frequency) * time.Second)
		}
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Setup configuration
	server_config.Config_Initialization()
	server_config.Main.LoadConfig()
	defer server_config.Main.Close()

	// Create a new Mochi-MQTT server instance with minimal options
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	err := server.AddHook(new(auth.Hook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{ // Auth disallows all by default
				{Username: "peach", Password: "password1", Allow: true},
				{Username: "melon", Password: "password2", Allow: true},
				{Remote: "127.0.0.1:*", Allow: true},
				{Remote: "localhost:*", Allow: true},
			},
			ACL: auth.ACLRules{ // ACL allows all by default
				{Remote: "127.0.0.1:*"}, // local superuser allow all
				{
					// user melon can read and write to their own topic
					Username: "melon", Filters: auth.Filters{
						"melon/#":   auth.ReadWrite,
						"updates/#": auth.WriteOnly, // can write to updates, but can't read updates from others
					},
				},
				{
					// Otherwise, no clients have publishing permissions
					Filters: auth.Filters{
						"#":         auth.ReadOnly,
						"updates/#": auth.Deny,
					},
				},
			},
		},
	})

	// Create a TCP listener on the standard MQTT port
	tcp := listeners.NewTCP(listeners.Config{ID: "t1", Address: ":1883"})

	err = server.AddListener(tcp)
	if err != nil {
		log.Fatalf("Error adding listener: %v", err)
	}

	// Create a WS listener on the standard MQTT port
	ws := listeners.NewWebsocket(listeners.Config{
		ID:      "ws1",
		Address: ":1885",
	})

	err = server.AddListener(ws)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Initialize the database connection
	db_conn, err := db.InitDB(server_config.Main.IoT_Server_DB)

	if err != nil {
		server.Log.Error("Failed to initialize database connection", "error", err)
	} else {
		server.Log.Info("Database connection initialized successfully")
		defer db_conn.Close()

		routes := router.NewRouter()

		// default port definition
		httpPort := ":8100"

		go routes.Run(httpPort)
		log.Printf("MQTT Server is running on port %s...\n", httpPort)

		routes.SetDB(db_conn)

		// Start the publisher manager
		go publisherManager(server, routes, db_conn, routes.RestartChan)

		_ = func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
			server.Log.Info("inline client received message from subscription", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName, "payload", string(pk.Payload))
			err := db.HandleAddToDB(db_conn, pk.TopicName, db.Incoming_Message{
				Topic:     pk.TopicName,
				Payload:   pk.Payload,
				Frequency: 0,
			})

			if err != nil {
				server.Log.Error("Failed to add message to database", "error", err)
			} else {
				server.Log.Info("Successfully added message to database", "topic", pk.TopicName)
				routes.RestartChan <- struct{}{}
			}
		}

		_ = func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
			server.Log.Info("inline client received message from subscription", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName, "payload", string(pk.Payload))
			err := db.HandleAddToDB(db_conn, pk.TopicName, db.Incoming_Message{
				Topic:     pk.TopicName,
				Payload:   pk.Payload,
				Frequency: 0,
			})

			if err != nil {
				server.Log.Error("Failed to add message to database", "error", err)
			} else {
				server.Log.Info("Successfully added message to database", "topic", pk.TopicName)
				routes.RestartChan <- struct{}{}
			}
		}

		// server.Log.Info("named_area subscribing")
		// _ = server.Subscribe("named_area/+/event", 1, callbackFnEvent)
		// _ = server.Subscribe("named_area/+/ack", 1, callbackFnAck)
	}

	<-sigs
	log.Println("Shutting down server...")
	server.Close()
	log.Println("Server gracefully stopped.")
}
