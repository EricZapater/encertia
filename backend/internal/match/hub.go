package match

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client represents a single connected WebSocket client (host or player).
type Client struct {
	Hub        *Hub
	Room       *Room
	Conn       *websocket.Conn
	Send       chan []byte
	UserID     uuid.UUID
	PlayerID   *uuid.UUID
	Nickname   string
	IsHost     bool
	isClosed   bool
	closeMutex sync.Mutex
}

// SendMessage serializes and queues a message for transmission to the client.
func (c *Client) SendMessage(msg interface{}) error {
	c.closeMutex.Lock()
	if c.isClosed {
		c.closeMutex.Unlock()
		return nil
	}
	c.closeMutex.Unlock()

	var data []byte
	var err error

	switch v := msg.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(msg)
		if err != nil {
			return err
		}
	}

	select {
	case c.Send <- data:
	default:
		// Queue full or closed, client lagging
		log.Printf("[Hub] Client %s buffer ple, descartant missatge", c.UserID)
	}
	return nil
}

// Close gracefully closes the client's send channel and connection.
func (c *Client) Close() {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	if !c.isClosed {
		c.isClosed = true
		close(c.Send)
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	}
}

// ReadPump handles incoming WebSocket frames, heartbeats and dispatches to handler.
func (c *Client) ReadPump(onMessage func(client *Client, msg []byte), onDisconnect func(client *Client)) {
	defer func() {
		if onDisconnect != nil {
			onDisconnect(c)
		}
		c.Close()
	}()

	if c.Conn != nil {
		c.Conn.SetReadLimit(maxMessageSize)
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.Conn.SetPongHandler(func(string) error {
			_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
	}

	for {
		if c.Conn == nil {
			break
		}
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Hub] Error inesperat de desconnexió WebSocket per usuari %s: %v", c.UserID, err)
			}
			break
		}

		if onMessage != nil {
			onMessage(c, message)
		}
	}
}

// WritePump pumps messages from the send channel to the WebSocket connection.
func (c *Client) WritePump() {
	if c.Conn == nil {
		return
	}
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte{'\n'}); err != nil {
					return
				}
				if _, err := w.Write(<-c.Send); err != nil {
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Room represents a live match room managed in-memory.
type Room struct {
	PIN     string
	MatchID uuid.UUID
	HostID  uuid.UUID
	mu      sync.RWMutex
	Host    *Client
	Players map[uuid.UUID]*Client   // keyed by PlayerID
	UserMap map[uuid.UUID]uuid.UUID // UserID -> PlayerID
}

// NewRoom creates a new Room instance.
func NewRoom(pin string, matchID, hostID uuid.UUID) *Room {
	return &Room{
		PIN:     pin,
		MatchID: matchID,
		HostID:  hostID,
		Players: make(map[uuid.UUID]*Client),
		UserMap: make(map[uuid.UUID]uuid.UUID),
	}
}

// RegisterClient adds or updates a client in the room.
func (r *Room) RegisterClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client.Room = r
	if client.IsHost {
		if r.Host != nil && r.Host != client {
			r.Host.Close()
		}
		r.Host = client
	} else if client.PlayerID != nil {
		if existing, ok := r.Players[*client.PlayerID]; ok && existing != client {
			existing.Close()
		}
		r.Players[*client.PlayerID] = client
		r.UserMap[client.UserID] = *client.PlayerID
	}
}

// UnregisterClient removes a client from the room.
func (r *Room) UnregisterClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if client.IsHost {
		if r.Host == client {
			r.Host = nil
		}
	} else if client.PlayerID != nil {
		if current, ok := r.Players[*client.PlayerID]; ok && current == client {
			delete(r.Players, *client.PlayerID)
			delete(r.UserMap, client.UserID)
		}
	}
}

// Broadcast sends a message to all connected participants (host and players).
func (r *Room) Broadcast(msg interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.Host != nil {
		_ = r.Host.SendMessage(msg)
	}
	for _, player := range r.Players {
		_ = player.SendMessage(msg)
	}
}

// BroadcastToPlayers sends a message only to connected players.
func (r *Room) BroadcastToPlayers(msg interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, player := range r.Players {
		_ = player.SendMessage(msg)
	}
}

// SendToHost sends a message specifically to the host.
func (r *Room) SendToHost(msg interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.Host != nil {
		_ = r.Host.SendMessage(msg)
	}
}

// SendToPlayer sends a message specifically to a player by PlayerID.
func (r *Room) SendToPlayer(playerID uuid.UUID, msg interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if player, ok := r.Players[playerID]; ok {
		_ = player.SendMessage(msg)
	}
}

// ConnectedPlayerCount returns the number of currently connected players.
func (r *Room) ConnectedPlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Players)
}

// GetPlayerClient returns the Client for a given PlayerID.
func (r *Room) GetPlayerClient(playerID uuid.UUID) *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Players[playerID]
}

// GetPlayerClientByUserID returns the Client for a given UserID.
func (r *Room) GetPlayerClientByUserID(userID uuid.UUID) *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if playerID, ok := r.UserMap[userID]; ok {
		return r.Players[playerID]
	}
	return nil
}

// Hub manages all active match rooms concurrently.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room // keyed by PIN
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

// GetOrCreateRoom retrieves an existing room by PIN or creates a new one.
func (h *Hub) GetOrCreateRoom(pin string, matchID, hostID uuid.UUID) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[pin]; ok {
		return room
	}

	room := NewRoom(pin, matchID, hostID)
	h.rooms[pin] = room
	return room
}

// GetRoom retrieves an existing room by PIN.
func (h *Hub) GetRoom(pin string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[pin]
}

// RemoveRoom deletes a room from the hub and disconnects all its clients.
func (h *Hub) RemoveRoom(pin string) {
	h.mu.Lock()
	room, ok := h.rooms[pin]
	if ok {
		delete(h.rooms, pin)
	}
	h.mu.Unlock()

	if room != nil {
		room.mu.Lock()
		if room.Host != nil {
			room.Host.Close()
		}
		for _, player := range room.Players {
			player.Close()
		}
		room.mu.Unlock()
	}
}
