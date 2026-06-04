package internal

import (
	"context"
	"sync"

	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/ws"
)

type IHub interface {
	Register(userID string, client *ws.Client)
	Unregister(userID string)
	IsOnline(userID string) bool
	DeliverTo(ctx context.Context, userID string, payload []byte) error
	Shutdown(ctx context.Context) error
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*ws.Client
}

func NewHub(_ do.Injector) (IHub, error) {
	return &Hub{clients: make(map[string]*ws.Client)}, nil
}

func (h *Hub) Register(userID string, client *ws.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.clients[userID]; ok {
		old.Close()
	}
	h.clients[userID] = client
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[userID]; ok {
		c.Close()
		delete(h.clients, userID)
	}
}

func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

func (h *Hub) DeliverTo(_ context.Context, userID string, payload []byte) error {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return nil
	}
	select {
	case c.Send <- payload:
	default:
	}
	return nil
}

func (h *Hub) Shutdown(_ context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for id, c := range h.clients {
		c.Close()
		delete(h.clients, id)
	}
	return nil
}
