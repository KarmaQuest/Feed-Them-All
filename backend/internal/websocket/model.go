// Package websocket — model.go définit les types de messages échangés via WebSocket.
//
// Tous les messages sont encodés en JSON. Le champ "type" discrimine le contenu :
//
//   Client → Serveur :
//     "subscribe"   → le client envoie sa bounding box (zone visible sur la carte)
//     "position"    → le client (feeder) envoie sa position GPS actuelle
//
//   Serveur → Client :
//     "ping_created"   → un nouveau ping vient d'être créé dans la zone du client
//     "ping_updated"   → un ping existant a changé d'état (confirmé, nourri, désactivé)
//     "feeder_position" → un feeder actif s'est déplacé dans la zone du client
//     "error"          → message d'erreur (ex: bounding box invalide)
//
// La bounding box est la zone rectangulaire visible sur l'écran de la carte.
// Le serveur ne broadcast que les événements dont la position GPS est dans cette zone.
package websocket

import "github.com/KarmaQuest/feed-them-all/internal/pings"

// BoundingBox représente la zone géographique visible par le client sur sa carte.
// MinLat/MaxLat : latitude sud/nord. MinLon/MaxLon : longitude ouest/est.
type BoundingBox struct {
	MinLat float64 `json:"min_lat"`
	MaxLat float64 `json:"max_lat"`
	MinLon float64 `json:"min_lon"`
	MaxLon float64 `json:"max_lon"`
}

// Contains retourne true si le point (lat, lon) est dans la bounding box.
func (b BoundingBox) Contains(lat, lon float64) bool {
	return lat >= b.MinLat && lat <= b.MaxLat && lon >= b.MinLon && lon <= b.MaxLon
}

// InboundMessage est le message envoyé par le client au serveur.
type InboundMessage struct {
	Type string `json:"type"` // "subscribe" | "position"

	// Pour type = "subscribe"
	BoundingBox *BoundingBox `json:"bounding_box,omitempty"`

	// Pour type = "position"
	Lat *float64 `json:"lat,omitempty"`
	Lon *float64 `json:"lon,omitempty"`
}

// OutboundMessage est le message envoyé par le serveur au client.
type OutboundMessage struct {
	Type string `json:"type"` // "ping_created" | "ping_updated" | "feeder_position" | "error"

	// Pour type = "ping_created" | "ping_updated"
	Ping *pings.Ping `json:"ping,omitempty"`

	// Pour type = "feeder_position"
	FeederID string   `json:"feeder_id,omitempty"`
	Lat      *float64 `json:"lat,omitempty"`
	Lon      *float64 `json:"lon,omitempty"`

	// Pour type = "error"
	Error string `json:"error,omitempty"`
}
