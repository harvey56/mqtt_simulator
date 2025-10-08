package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO check the origin to prevent cross-site WS
		return true
	},
}

type WebSocketHandler struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mutex     sync.Mutex
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

func (h *WebSocketHandler) Run() {
	for {
		msg := <-h.broadcast

		h.mutex.Lock()

		for client := range h.clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mutex.Unlock()
	}
}

func (h *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()

	log.Println("Client Connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		log.Printf("Received client msg: %s", string(msg))
	}

	h.mutex.Lock()
	delete(h.clients, conn)
	h.mutex.Unlock()

	log.Println("Client Disconnected")
}

func (h *WebSocketHandler) BroadcastMessage(topic string, payload interface{}) {
	message := WS_json_Result{Type: topic, Data: payload}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling json in websocket: %s", err.Error())
	}
	h.broadcast <- jsonMessage
}
