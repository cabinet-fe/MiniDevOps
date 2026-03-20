package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	Send    chan []byte
	Channel string
	UserID  uint
}

type Hub struct {
	clients    map[*Client]bool
	channels   map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	quit       chan struct{}
	mu         sync.RWMutex
}

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		channels:   make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		quit:       make(chan struct{}),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	defer func() {
		if r := recover(); r != nil {
			go h.run()
		}
	}()

	for {
		select {
		case <-h.quit:
			h.mu.Lock()
			for client := range h.clients {
				close(client.Send)
				delete(h.clients, client)
			}
			h.channels = make(map[string]map[*Client]bool)
			h.mu.Unlock()
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if _, ok := h.channels[client.Channel]; !ok {
				h.channels[client.Channel] = make(map[*Client]bool)
			}
			h.channels[client.Channel][client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if ch, ok := h.channels[client.Channel]; ok {
					delete(ch, client)
					if len(ch) == 0 {
						delete(h.channels, client.Channel)
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.quit)
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToChannel(channel string, message []byte) {
	h.mu.RLock()
	clients, ok := h.channels[channel]
	h.mu.RUnlock()
	if !ok {
		return
	}
	h.mu.RLock()
	for client := range clients {
		select {
		case client.Send <- message:
		default:
			go h.Unregister(client)
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastToUser(userID uint, message []byte) {
	h.mu.RLock()
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
			default:
				go h.Unregister(client)
			}
		}
	}
	h.mu.RUnlock()
}

// WritePump sends messages from the Send channel to the WebSocket connection
func WritePump(client *Client, hub *Hub) {
	defer func() {
		hub.Unregister(client)
		client.Conn.Close()
	}()
	for message := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
