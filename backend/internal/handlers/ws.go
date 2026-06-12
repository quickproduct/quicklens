package handlers

import (
	"go.uber.org/zap"
	"net/http"

	"github.com/gorilla/websocket"

	"quicklens/backend/internal/auth"
	"quicklens/backend/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev/self-hosted
	},
}

func HandleWSChannel(channel string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		_, err := auth.VerifyToken(token, "access")
		if err != nil {
			http.Error(w, "Unauthorized: invalid or expired token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			zap.L().Sugar().Infof("Failed to upgrade WebSocket connection for channel %s: %v", channel, err)
			return
		}

		ws.Manager.Connect(channel, conn)

		// Keep connection open and handle incoming messages (e.g. pings)
		go func() {
			defer ws.Manager.Disconnect(channel, conn)
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}()
	}
}
