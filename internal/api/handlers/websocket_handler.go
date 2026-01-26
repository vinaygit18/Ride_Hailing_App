package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
	"github.com/gocomet/ride-hailing/pkg/logger"
	"github.com/gocomet/ride-hailing/pkg/websocket"
)

// HandleWebSocket handles GET /v1/ws
func (h *Handlers) HandleWebSocket(c *gin.Context) {
	// Upgrade connection to WebSocket
	upgrader := gorilla.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.Logger.Error("Failed to upgrade to WebSocket", logger.Err(err))
		return
	}

	// Get user info from query params
	userID := c.Query("user_id")
	userType := c.Query("user_type")

	if userID == "" || userType == "" {
		h.Logger.Warn("Missing user_id or user_type in WebSocket connection")
		conn.Close()
		return
	}

	// Create client and register with hub
	if wsHub, ok := h.Hub.(*websocket.Hub); ok {
		client := websocket.NewClient(wsHub, conn, userID, userType, h.Logger)
		wsHub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}
