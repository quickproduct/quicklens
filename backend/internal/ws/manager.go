package ws

import (
	"encoding/json"
	"go.uber.org/zap"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketManager struct {
	mu          sync.RWMutex
	connections map[string]map[*websocket.Conn]bool
}

var Manager = &WebSocketManager{
	connections: make(map[string]map[*websocket.Conn]bool),
}

func (m *WebSocketManager) Connect(channel string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[channel] == nil {
		m.connections[channel] = make(map[*websocket.Conn]bool)
	}
	m.connections[channel][conn] = true
	zap.L().Sugar().Infof("WebSocket connected to channel '%s' (total: %d)", channel, len(m.connections[channel]))
}

func (m *WebSocketManager) Disconnect(channel string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conns, ok := m.connections[channel]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(m.connections, channel)
		}
		zap.L().Sugar().Infof("WebSocket disconnected from channel '%s'", channel)
	}
	conn.Close()
}

func (m *WebSocketManager) Broadcast(channel string, data interface{}) {
	m.mu.RLock()
	conns, ok := m.connections[channel]
	if !ok || len(conns) == 0 {
		m.mu.RUnlock()
		return
	}

	// Copy connections to avoid holding lock during writes
	targets := make([]*websocket.Conn, 0, len(conns))
	for conn := range conns {
		targets = append(targets, conn)
	}
	m.mu.RUnlock()

	msg, err := json.Marshal(data)
	if err != nil {
		zap.L().Sugar().Infof("Failed to marshal WebSocket message: %v", err)
		return
	}

	for _, conn := range targets {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			zap.L().Sugar().Infof("Failed to send WebSocket message: %v", err)
			m.Disconnect(channel, conn)
		}
	}
}

func (m *WebSocketManager) GetConnectionCount(channel string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections[channel])
}

func (m *WebSocketManager) StartHeartbeat() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			m.mu.RLock()
			for channel, conns := range m.connections {
				for conn := range conns {
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						zap.L().Sugar().Infof("Heartbeat failed for channel '%s': %v", channel, err)
						go m.Disconnect(channel, conn)
					}
				}
			}
			m.mu.RUnlock()
		}
	}()
}
