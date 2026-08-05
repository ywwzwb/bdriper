package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*wsClient]bool
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*wsClient]bool)}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("websocket upgrade failed", "error", err)
		return
	}
	c := &wsClient{conn: conn}
	h.mu.Lock()
	h.clients[c] = true
	clientCount := len(h.clients)
	h.mu.Unlock()
	slog.Info("websocket client connected", "total", clientCount)
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, c)
			remaining := len(h.clients)
			h.mu.Unlock()
			conn.Close()
			slog.Info("websocket client disconnected", "total", remaining)
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

func (h *Hub) Broadcast(evt Event) {
	data, _ := json.Marshal(evt)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		c.mu.Lock()
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			slog.Warn("websocket write failed", "error", err)
		}
		c.mu.Unlock()
	}
}
