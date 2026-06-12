// Package pings — model.go définit les structures de données du package pings.
//
// Ce fichier contient les types utilisés dans toutes les couches (handler, service, repository) :
//
//   Ping            → représente un signalement en base de données.
//                     Contient les coordonnées GPS (Lat/Lon extraites de GEOGRAPHY),
//                     le type (animal ou food), l'état (actif/inactif), et les dates.
//
//   CreatePingRequest → données envoyées par le client pour créer un ping.
//                       Latitude et longitude sont envoyées en JSON, converties en
//                       GEOGRAPHY(POINT, 4326) côté repository.
//
//   NearbyQuery     → paramètres de la requête GET /pings :
//                       Lat/Lon du centre, rayon en mètres, type optionnel (animal|food).
//
// Règle GPS : longitude toujours en premier dans ST_MakePoint($lon, $lat) — piège classique PostGIS.
package pings

import "time"

// Ping represents a single geolocated report as stored in the database.
type Ping struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`       // "animal" or "food"
	Lat       float64    `json:"lat"`        // latitude, extracted from GEOGRAPHY
	Lon       float64    `json:"lon"`        // longitude, extracted from GEOGRAPHY
	CreatedBy string     `json:"created_by"` // user UUID
	IsActive  bool       `json:"is_active"`
	FedAt     *time.Time `json:"fed_at,omitempty"` // nil until the animal is fed
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreatePingRequest is the JSON body for POST /pings.
// Both lat and lon are required. Type must be "animal" or "food".
type CreatePingRequest struct {
	Type string  `json:"type"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// NearbyQuery holds the parsed query parameters for GET /pings.
// Radius defaults to 500 m if not provided. Max enforced at 10 000 m.
type NearbyQuery struct {
	Lat    float64
	Lon    float64
	Radius float64 // metres
	Type   string  // optional filter: "animal" | "food" | "" (all)
}
