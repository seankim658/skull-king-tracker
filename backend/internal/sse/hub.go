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

	if existingChan, exists := h.clients[userID]; exists {
		close(existingChan)
	}

	h.clients[userID] = clientChan
}

// Removes a client for a user
func (h *Hub) RemoveClient(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.clients[userID]; ok {
		delete(h.clients, userID)
		_ = ch
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
			go func() {
				h.RemoveClient(userID)
			}()
		}
	}
}
