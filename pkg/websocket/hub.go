package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gocomet/ride-hailing/pkg/logger"
)

// Hub maintains active client connections and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *logger.Logger
}

// Message represents a WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NewHub creates a new WebSocket hub
func NewHub(logger *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("Client registered",
				logger.String("client_id", client.ID),
				logger.String("user_type", client.UserType),
			)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.logger.Info("Client unregistered",
					logger.String("client_id", client.ID),
				)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal broadcast message", logger.Err(err))
		return
	}
	h.broadcast <- data
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID, userType string, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", logger.Err(err))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.UserID == userID && client.UserType == userType {
			select {
			case client.Send <- data:
			default:
				h.logger.Warn("Failed to send message to client",
					logger.String("user_id", userID),
					logger.String("client_id", client.ID),
				)
			}
		}
	}
}

// BroadcastToRide sends a message to all participants of a ride
func (h *Hub) BroadcastToRide(rideID string, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal ride message", logger.Err(err))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		// Check if client is subscribed to this ride
		if client.IsSubscribedToRide(rideID) {
			select {
			case client.Send <- data:
			default:
				h.logger.Warn("Failed to send ride message to client",
					logger.String("ride_id", rideID),
					logger.String("client_id", client.ID),
				)
			}
		}
	}
}

// GetActiveConnections returns the number of active connections
func (h *Hub) GetActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetClientsByUserType returns count of clients by user type
func (h *Hub) GetClientsByUserType(userType string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for client := range h.clients {
		if client.UserType == userType {
			count++
		}
	}
	return count
}

// SendToUser sends a message to a specific user by ID (any type)
func (h *Hub) SendToUser(userID string, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", logger.Err(err))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	sent := false
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
				sent = true
				h.logger.Info("Message sent to user",
					logger.String("user_id", userID),
					logger.String("user_type", client.UserType),
				)
			default:
				h.logger.Warn("Failed to send message to client",
					logger.String("user_id", userID),
					logger.String("client_id", client.ID),
				)
			}
		}
	}

	if !sent {
		h.logger.Warn("No client found for user", logger.String("user_id", userID))
	}
}

// BroadcastToType sends a message to all clients of a specific type
func (h *Hub) BroadcastToType(userType string, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", logger.Err(err))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for client := range h.clients {
		if client.UserType == userType {
			select {
			case client.Send <- data:
				count++
			default:
				h.logger.Warn("Failed to send message to client",
					logger.String("user_type", userType),
					logger.String("client_id", client.ID),
				)
			}
		}
	}

	h.logger.Info("Message broadcast to user type",
		logger.String("user_type", userType),
		logger.Int("count", count),
	)
}
