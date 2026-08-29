// Package pings — model.go définit les structures de données du package pings.
//
// Ce fichier contient les types utilisés dans toutes les couches (handler, service, repository) :
//
//   Ping              → représente un signalement en base de données.
//                       Contient les coordonnées GPS (Lat/Lon extraites de GEOGRAPHY),
//                       le type (animal ou food), l'état (actif/inactif), et les dates.
//
//   CreatePingRequest → données envoyées par le client pour créer un ping.
//                       Latitude et longitude sont envoyées en JSON, converties en
//                       GEOGRAPHY(POINT, 4326) côté repository.
//
//   NearbyQuery       → paramètres de la requête GET /pings :
//                         Lat/Lon du centre, rayon en mètres, type optionnel (animal|food).
//
//   PingReport        → représente un signalement d'un ping par un utilisateur.
//                       Reason est un enum : wrong_location | animal_gone | duplicate | inappropriate.
//                       Score = up_votes - down_votes (calculé en DB par LEFT JOIN).
//
//   CreateReportRequest → body JSON de POST /pings/:id/report.
//
//   VoteReportRequest → body JSON de POST /pings/:id/reports/:report_id/vote.
//                       Value : "up" | "down".
//
// Règle GPS : longitude toujours en premier dans ST_MakePoint($lon, $lat) — piège classique PostGIS.
package pings

import "time"

// Ping represents a single geolocated report as stored in the database.
type Ping struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`         // "animal" or "food"
	Lat         float64    `json:"lat"`          // latitude, extracted from GEOGRAPHY
	Lon         float64    `json:"lon"`          // longitude, extracted from GEOGRAPHY
	CreatedBy   string     `json:"created_by"`   // user UUID
	IsActive    bool       `json:"is_active"`
	FedAt       *time.Time `json:"fed_at,omitempty"`  // last feeding time (kept in sync with feeding events)
	AnimalType  *string    `json:"animal_type,omitempty"` // "cat", "dog", "other" — nil for food pings
	AnimalBreed *string    `json:"animal_breed,omitempty"` // specific breed (carlin, labrador...) — nil if unset
	AnimalCount int        `json:"animal_count"`  // number of animals observed (default 1)
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreatePingRequest is the JSON body for POST /pings.
// Both lat and lon are required. Type must be "animal" or "food".
// AnimalType and AnimalCount are required when Type is "animal".
type CreatePingRequest struct {
	Type        string  `json:"type"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	AnimalType  *string `json:"animal_type,omitempty"`  // "cat", "dog", "other"
	AnimalBreed *string `json:"animal_breed,omitempty"` // specific breed
	AnimalCount *int    `json:"animal_count,omitempty"` // defaults to 1 if omitted
}

// FeedingEvent represents a single feeding action recorded by a user.
type FeedingEvent struct {
	ID              string    `json:"id"`
	PingID          string    `json:"ping_id"`
	FedBy           string    `json:"fed_by"`  // user UUID
	Username        string    `json:"username"` // display name (JOIN with users)
	FedAt           time.Time `json:"fed_at"`
	Note            *string   `json:"note,omitempty"`
	AnimalCountSeen *int      `json:"animal_count_seen,omitempty"`
	EventType       string    `json:"event_type"` // "signal" | "feeding"
}

// UpdatePingRequest is the JSON body for PATCH /pings/:id.
// Only the creator may update animal_type and animal_count.
type UpdatePingRequest struct {
	AnimalType  *string `json:"animal_type"`  // "cat", "dog", "other"
	AnimalBreed *string `json:"animal_breed"` // specific breed
	AnimalCount *int    `json:"animal_count"` // must be >= 1
}

// CreateFeedingEventRequest is the JSON body for POST /pings/:id/feedings.
type CreateFeedingEventRequest struct {
	Note            *string `json:"note,omitempty"`
	AnimalCountSeen *int    `json:"animal_count_seen,omitempty"`
}

// NearbyQuery holds the parsed query parameters for GET /pings.
// Radius defaults to 500 m if not provided. Max enforced at 10 000 m.
type NearbyQuery struct {
	Lat    float64
	Lon    float64
	Radius float64 // metres
	Type   string  // optional filter: "animal" | "food" | "" (all)
}

// PingReport represents a report filed by a user about a ping.
// Score is computed as up_votes - down_votes (via LEFT JOIN on ping_report_votes).
type PingReport struct {
	ID         string    `json:"id"`
	PingID     string    `json:"ping_id"`
	ReportedBy string    `json:"reported_by"`
	Reason     string    `json:"reason"`  // wrong_location | animal_gone | duplicate | inappropriate
	Comment    *string   `json:"comment,omitempty"`
	Score      int       `json:"score"`   // up - down votes
	CreatedAt  time.Time `json:"created_at"`
}

// CreateReportRequest is the JSON body for POST /pings/:id/report.
type CreateReportRequest struct {
	Reason  string  `json:"reason"`
	Comment *string `json:"comment,omitempty"`
}

// VoteReportRequest is the JSON body for POST /pings/:id/reports/:report_id/vote.
type VoteReportRequest struct {
	Value string `json:"value"` // "up" or "down"
}
