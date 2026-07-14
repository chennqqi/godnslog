package websocket

import (
	"context"
	"sync"
	"sync/atomic"
)

// Hub manages all active WebSocket client connections.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	stop       chan struct{}
	stopped    atomic.Bool
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan struct{}),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	if h.stopped.Load() {
		return
	}
	h.register <- client
}

// Run starts the hub's event loop. Must be called as a goroutine.
// Returns when Shutdown is called.
func (h *Hub) Run() {
	for {
		select {
		case <-h.stop:
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
// After Shutdown, Broadcast drops the message without blocking.
func (h *Hub) Broadcast(msg []byte) {
	if h.stopped.Load() {
		return
	}
	select {
	case h.broadcast <- msg:
	case <-h.stop:
	}
}

// Shutdown stops the hub's event loop and closes all client connections.
// Safe to call multiple times. Waits for write pumps to finish or ctx to expire.
func (h *Hub) Shutdown(ctx context.Context) error {
	if !h.stopped.CompareAndSwap(false, true) {
		return nil // already stopped
	}
	close(h.stop)
	// Wait for Run() to return and write pumps to drain.
	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Len returns the number of connected clients.
func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
