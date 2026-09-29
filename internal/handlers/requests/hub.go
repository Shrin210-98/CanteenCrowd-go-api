package requests

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"ccms.com/api/internal/constants"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ============ WEBSOCKET HUB ============

type Hub struct {
	// Map of requestID -> map of userID -> connection
	rooms map[uuid.UUID]map[uuid.UUID]*Client

	// Register/Unregister channels
	register   chan *Client
	unregister chan *Client

	// Broadcast channel
	broadcast chan *ChatMessage

	mu sync.RWMutex
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	userID    uuid.UUID
	requestID uuid.UUID
}

type ChatMessage struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"` // 'message', 'typing', 'read'
	RequestID uuid.UUID `json:"requestId"`
	UserID    uuid.UUID `json:"userId"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (configure for production)
	},
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *ChatMessage),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.requestID] == nil {
				h.rooms[client.requestID] = make(map[uuid.UUID]*Client)
			}
			h.rooms[client.requestID][client.userID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.requestID]; ok {
				if _, exists := room[client.userID]; exists {
					delete(room, client.userID)
					close(client.send)
				}
				if len(room) == 0 {
					delete(h.rooms, client.requestID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			room := h.rooms[message.RequestID]
			h.mu.RUnlock()

			data, _ := json.Marshal(message)

			for userID, client := range room {
				if userID != message.UserID { // Don't send to sender
					select {
					case client.send <- data:
					default:
						h.mu.Lock()
						delete(room, userID)
						close(client.send)
						h.mu.Unlock()
					}
				}
			}
		}
	}
}

// ============ WEBSOCKET HANDLER ============

// ChatWebSocket handles WebSocket connection for chat
func (h *Handler) ChatWebSocket(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:       h.hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		userID:    userID,
		requestID: requestID,
	}

	client.hub.register <- client

	// Start goroutines
	go client.writePump()
	go client.readPump()
}

// ============ CLIENT METHODS ============

// readPump reads messages from WebSocket
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(2048)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var payload struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(message, &payload); err != nil {
			continue
		}

		switch payload.Type {
		case "ping":
			// Client-initiated ping → respond with pong
			c.send <- []byte(`{"type":"pong"}`)

		case "pong":
			// Response to server ping — ignore

		default:
			// Ignore everything else.
			// Message writes go through REST, not WebSocket.
			log.Printf("WS: ignoring frame type=%q from user=%s", payload.Type, c.userID)
		}
	}
}

// writePump writes messages to WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// For extra security, add a dedicated endpoint that issues 5-minute tokens:
// GET /api/v1/ws-token
// func (h *Handler) GetWSToken(w http.ResponseWriter, r *http.Request) {
//     userID := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
//     tenantID := r.Context().Value(constants.ContextKeyTenantID).(uuid.UUID)

//     token, err := utils.GenerateShortLivedToken(userID, tenantID, 5*time.Minute)
//     if err != nil {
//         utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
//         return
//     }

//     utils.SuccessResponse(w, http.StatusOK, map[string]any{
//         "wsToken":   token,
//         "expiresIn": 300,
//     }, "WebSocket token generated")
// }
