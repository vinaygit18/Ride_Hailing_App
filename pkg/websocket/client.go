package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gocomet/ride-hailing/pkg/logger"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	UserID        string
	UserType      string // "rider" or "driver"
	Hub           *Hub
	Conn          *websocket.Conn
	Send          chan []byte
	subscriptions map[string]bool // rideIDs this client is subscribed to
	mu            sync.RWMutex
	logger        *logger.Logger
}

// ClientMessage represents a message from the client
type ClientMessage struct {
	Type     string                 `json:"type"`
	EntityID string                 `json:"entity_id,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID, userType string, logger *logger.Logger) *Client {
	return &Client{
		ID:            generateClientID(),
		UserID:        userID,
		UserType:      userType,
		Hub:           hub,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
		logger:        logger,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error",
					logger.Err(err),
					logger.String("client_id", c.ID),
				)
			}
			break
		}

		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.logger.Error("Failed to unmarshal client message",
			logger.Err(err),
			logger.String("client_id", c.ID),
		)
		return
	}

	switch msg.Type {
	case "subscribe":
		c.Subscribe(msg.EntityID)
	case "unsubscribe":
		c.Unsubscribe(msg.EntityID)
	case "ping":
		c.SendMessage(Message{Type: "pong"})
	default:
		c.logger.Warn("Unknown message type",
			logger.String("type", msg.Type),
			logger.String("client_id", c.ID),
		)
	}
}

// Subscribe subscribes the client to a ride
func (c *Client) Subscribe(rideID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscriptions[rideID] = true
	c.logger.Info("Client subscribed to ride",
		logger.String("client_id", c.ID),
		logger.String("ride_id", rideID),
	)
}

// Unsubscribe unsubscribes the client from a ride
func (c *Client) Unsubscribe(rideID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscriptions, rideID)
	c.logger.Info("Client unsubscribed from ride",
		logger.String("client_id", c.ID),
		logger.String("ride_id", rideID),
	)
}

// IsSubscribedToRide checks if client is subscribed to a ride
func (c *Client) IsSubscribedToRide(rideID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscriptions[rideID]
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		c.logger.Error("Failed to marshal message",
			logger.Err(err),
			logger.String("client_id", c.ID),
		)
		return
	}

	select {
	case c.Send <- data:
	default:
		c.logger.Warn("Client send buffer full",
			logger.String("client_id", c.ID),
		)
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
