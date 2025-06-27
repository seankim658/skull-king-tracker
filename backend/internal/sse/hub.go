package sse

import "sync"

type Hub struct {
	clients map[string]chan string
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]chan string),
	}
}

// Registers a new client for a user
func (h *Hub) AddClient(userID string, clientChan chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = clientChan
}

// Removes a client for a user
func (h *Hub) RemoveClient(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.clients[userID]; ok {
		close(ch)
		delete(h.clients, userID)
	}
}

// Broadcast sends a message to a specific user if they are connected
func (h *Hub) Broadcast(userID string, message string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clientChan, ok := h.clients[userID]; ok {
		select {
		case clientChan <- message:
		default:
    // TODO : should we log if the client's channel is full
		}
	}
}
