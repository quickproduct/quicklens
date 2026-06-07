package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"quicklens/backend/ws"
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
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade WebSocket connection for channel %s: %v", channel, err)
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
