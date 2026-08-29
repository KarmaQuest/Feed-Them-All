// Package websocket — client.go représente une connexion WebSocket individuelle.
//
// Chaque client connecté est une paire de goroutines :
//   - readPump() : lit les messages entrants du client (subscribe, position)
//   - writePump() : écrit les messages sortants vers le client (ping_created, feeder_position…)
//
// Ce pattern "une goroutine lecture + une goroutine écriture" par connexion est le pattern
// standard gorilla/websocket : gorilla déconseille explicitement les accès concurrents
// à la même connexion depuis plusieurs goroutines.
//
// Rate limiting des positions GPS :
//   Un feeder peut se déplacer plusieurs fois par seconde, mais on ne relaie sa position
//   qu'une fois par seconde maximum (token bucket, 1 token/s). Cela évite de flood le hub.
//
// Déconnexion propre :
//   Quand readPump reçoit une erreur (disconnect, timeout), il appelle hub.Unregister()
//   et sort. writePump surveille la fermeture du channel send pour s'arrêter aussi.
package websocket

import (
	"encoding/json"
	"log/slog"
	"math"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	// writeWait is the time allowed to write a message to the client.
	writeWait = 10 * time.Second

	// pongWait is the time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// pingPeriod is how often we ping the client (must be less than pongWait).
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize is the maximum message size allowed from a client.
	maxMessageSize = 512 // bytes
)

// Client represents a single WebSocket connection.
type Client struct {
	hub *Hub

	// conn is the underlying WebSocket connection.
	conn *websocket.Conn

	// send is a buffered channel of outbound messages for this client.
	send chan OutboundMessage

	// bbox is the client's current bounding box (the visible map area).
	// nil until the client sends a "subscribe" message.
	bbox *BoundingBox

	// userID is set if the client authenticated via JWT (optional).
	// Used to attribute feeder position broadcasts.
	userID string

	// positionLimiter enforces max 1 GPS position push per second.
	positionLimiter *rate.Limiter
}

// NewClient creates a new Client for the given connection.
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:             hub,
		conn:            conn,
		send:            make(chan OutboundMessage, 32),
		userID:          userID,
		positionLimiter: rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

// ReadPump reads messages from the WebSocket connection and processes them.
// Must be called in a dedicated goroutine. Exits when the connection closes.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("ws: unexpected close", "err", err)
			}
			break
		}
		c.handleInbound(msg)
	}
}

// WritePump writes messages from the send channel to the WebSocket connection.
// Must be called in a dedicated goroutine. Exits when the send channel is closed.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel — send a close message.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				slog.Warn("ws: write error", "err", err)
				return
			}

		case <-ticker.C:
			// Send a WebSocket ping to keep the connection alive.
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleInbound processes a raw JSON message received from the client.
func (c *Client) handleInbound(raw []byte) {
	var msg InboundMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		c.sendError("invalid JSON")
		return
	}

	switch msg.Type {
	case "subscribe":
		c.handleSubscribe(msg)
	case "position":
		c.handlePosition(msg)
	default:
		c.sendError("unknown message type")
	}
}

// handleSubscribe processes a "subscribe" message — the client sets its bounding box.
func (c *Client) handleSubscribe(msg InboundMessage) {
	if msg.BoundingBox == nil {
		c.sendError("subscribe requires bounding_box")
		return
	}
	bb := msg.BoundingBox
	if !isValidBoundingBox(bb) {
		c.sendError("invalid bounding_box values")
		return
	}
	c.bbox = bb
	slog.Info("ws: client subscribed", "user_id", c.userID, "bbox", bb)
}

// handlePosition processes a "position" message — a feeder broadcasts their GPS position.
// Only authenticated users can send position updates.
// Rate limited to 1 message/second.
func (c *Client) handlePosition(msg InboundMessage) {
	if c.userID == "" {
		c.sendError("position requires authentication")
		return
	}
	if msg.Lat == nil || msg.Lon == nil {
		c.sendError("position requires lat and lon")
		return
	}
	if !isValidCoord(*msg.Lat, *msg.Lon) {
		c.sendError("invalid coordinates")
		return
	}
	// Rate limit: max 1 GPS push per second.
	if !c.positionLimiter.Allow() {
		return // silently drop — don't send error for rate-limited messages
	}
	// Round to 4 decimal places (~11m precision) before broadcasting.
	lat := roundCoord(*msg.Lat)
	lon := roundCoord(*msg.Lon)
	c.hub.BroadcastFeederPosition(c.userID, lat, lon)
}

// sendError sends an "error" message to this client.
func (c *Client) sendError(reason string) {
	select {
	case c.send <- OutboundMessage{Type: "error", Error: reason}:
	default:
	}
}

// isValidBoundingBox returns true if the bounding box has valid geographic coordinates
// and is not degenerate (min < max).
func isValidBoundingBox(b *BoundingBox) bool {
	return b.MinLat >= -90 && b.MaxLat <= 90 && b.MinLat < b.MaxLat &&
		b.MinLon >= -180 && b.MaxLon <= 180 && b.MinLon < b.MaxLon
}

// isValidCoord returns true if the lat/lon are within valid geographic ranges.
func isValidCoord(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

// roundCoord rounds a coordinate to 4 decimal places (~11m precision).
func roundCoord(v float64) float64 {
	return math.Round(v*10000) / 10000
}
