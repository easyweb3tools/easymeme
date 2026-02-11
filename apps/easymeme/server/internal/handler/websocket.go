package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn, 32),
		unregister: make(chan *websocket.Conn, 32),
	}
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					delete(h.clients, client)
					client.Close()
				}
			}
		}
	}
}

func (h *WebSocketHub) Broadcast(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WebSocket] Marshal error: %v", err)
		return
	}
	h.broadcast <- data
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *WebSocketHub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.register <- conn

	go func() {
		defer func() { h.unregister <- conn }()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
