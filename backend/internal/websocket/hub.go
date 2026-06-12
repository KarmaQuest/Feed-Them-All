// Package websocket — hub.go gère le registre central de toutes les connexions WebSocket actives.
//
// Le Hub est le "chef d'orchestre" du temps réel :
//   - Il tient à jour la liste de tous les clients connectés
//   - Quand un ping est créé ou modifié, le package pings appelle hub.BroadcastPing()
//   - Le Hub envoie le message uniquement aux clients dont la bounding box contient le ping
//   - Quand un feeder se déplace, hub.BroadcastFeederPosition() diffuse sa position
//   - Il gère proprement les connexions et déconnexions (goroutine-safe via channel)
//
// Architecture goroutine-safe :
//   Le Hub tourne dans sa propre goroutine (Run()). Les opérations register/unregister/broadcast
//   passent toutes par des channels — jamais de mutex direct sur la map des clients.
//   Cela évite les race conditions sans complexité de verrouillage.
package websocket

import (
	"log/slog"

	"github.com/KarmaQuest/feed-them-all/internal/pings"
)

// Hub maintains the set of active WebSocket clients and broadcasts messages to them.
type Hub struct {
	// clients is the set of all connected clients.
	clients map[*Client]struct{}

	// register receives clients connecting to the hub.
	register chan *Client

	// unregister receives clients disconnecting from the hub.
	unregister chan *Client

	// broadcast receives outbound messages to be sent to matching clients.
	broadcast chan broadcastMsg
}

// broadcastMsg pairs a message with an optional position filter (lat/lon).
// If filterLat/filterLon are set, only clients whose bounding box contains
// that position receive the message.
type broadcastMsg struct {
	msg       OutboundMessage
	filterLat *float64 // nil = send to all clients
	filterLon *float64
}

// NewHub creates a new Hub. Call Run() in a goroutine after creation.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan broadcastMsg, 256),
	}
}

// Run starts the hub's event loop. Must be called in a dedicated goroutine.
// It blocks until the process exits.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
			slog.Info("ws: client connected", "total", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				slog.Info("ws: client disconnected", "total", len(h.clients))
			}

		case msg := <-h.broadcast:
			for client := range h.clients {
				// If the message has a position filter, only send to clients
				// whose bounding box contains that position.
				if msg.filterLat != nil && msg.filterLon != nil {
					if client.bbox == nil || !client.bbox.Contains(*msg.filterLat, *msg.filterLon) {
						continue
					}
				}

				// Non-blocking send: drop message if client's send buffer is full
				// (avoids blocking the entire hub for a slow client).
				select {
				case client.send <- msg.msg:
				default:
					// Client is too slow — disconnect it
					delete(h.clients, client)
					close(client.send)
					slog.Warn("ws: slow client dropped")
				}
			}
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// BroadcastPingCreated sends a "ping_created" event to all clients whose bounding box
// contains the ping's location.
func (h *Hub) BroadcastPingCreated(p pings.Ping) {
	h.broadcast <- broadcastMsg{
		msg:       OutboundMessage{Type: "ping_created", Ping: &p},
		filterLat: &p.Lat,
		filterLon: &p.Lon,
	}
}

// BroadcastPingUpdated sends a "ping_updated" event to all clients whose bounding box
// contains the ping's location. Used for confirm, fed, deactivate.
func (h *Hub) BroadcastPingUpdated(p pings.Ping) {
	h.broadcast <- broadcastMsg{
		msg:       OutboundMessage{Type: "ping_updated", Ping: &p},
		filterLat: &p.Lat,
		filterLon: &p.Lon,
	}
}

// BroadcastFeederPosition sends a "feeder_position" event to all clients whose bounding box
// contains the feeder's current position. Rate limiting (1/s) is enforced by the client.
func (h *Hub) BroadcastFeederPosition(feederID string, lat, lon float64) {
	h.broadcast <- broadcastMsg{
		msg:       OutboundMessage{Type: "feeder_position", FeederID: feederID, Lat: &lat, Lon: &lon},
		filterLat: &lat,
		filterLon: &lon,
	}
}
