package service

import (
	"encoding/json"
	"sync"
)

type ChatWSConn interface {
	Send(data []byte) error
	Close()
}

type ChatWSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ChatHub struct {
	mu           sync.RWMutex
	visitorConns map[int64]ChatWSConn // conversationID -> visitor conn
	adminConns   map[int64]ChatWSConn // adminUserID -> admin conn
}

func NewChatHub() *ChatHub {
	return &ChatHub{
		visitorConns: make(map[int64]ChatWSConn),
		adminConns:   make(map[int64]ChatWSConn),
	}
}

func (h *ChatHub) RegisterVisitor(conversationID int64, conn ChatWSConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.visitorConns[conversationID]; ok {
		old.Close()
	}
	h.visitorConns[conversationID] = conn
}

func (h *ChatHub) UnregisterVisitor(conversationID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.visitorConns, conversationID)
}

func (h *ChatHub) RegisterAdmin(adminUserID int64, conn ChatWSConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.adminConns[adminUserID]; ok {
		old.Close()
	}
	h.adminConns[adminUserID] = conn
}

func (h *ChatHub) UnregisterAdmin(adminUserID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.adminConns, adminUserID)
}

func (h *ChatHub) BroadcastToAdmins(msg ChatWSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conn := range h.adminConns {
		_ = conn.Send(data)
	}
}

func (h *ChatHub) SendToVisitor(conversationID int64, msg ChatWSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conn, ok := h.visitorConns[conversationID]; ok {
		_ = conn.Send(data)
	}
}

func (h *ChatHub) AdminCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.adminConns)
}
