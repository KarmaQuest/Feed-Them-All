// Package pings — store.go définit l'interface Store du package pings.
//
// Même principe que dans le package auth : le Service dépend de cette interface
// et non de l'implémentation concrète (Repository). Cela permet de tester le
// Service avec un fakeStore en mémoire, sans base de données réelle.
//
// Méthodes définies :
//   Create         → insère un nouveau ping, retourne l'ID et la date de création
//   ListNearby     → retourne les pings actifs dans un rayon (ST_DWithin PostGIS)
//   GetByID        → récupère un ping par son UUID (pour vérifier l'existence et le propriétaire)
//   Confirm        → met à jour updated_at pour indiquer "l'animal est toujours là"
//   MarkFed        → enregistre fed_at = NOW() sur le ping
//   Deactivate     → soft delete : is_active = false (le ping reste en base pour l'historique)
//   AddMedia       → insère une ligne dans ping_media (chemin du fichier uploadé)
//   ListMedia      → retourne les chemins des médias associés à un ping
package pings

import "context"

// Store is the interface the Service depends on for all DB operations.
type Store interface {
	// Create inserts a new ping and returns the created Ping (with generated ID and timestamps).
	Create(ctx context.Context, userID, pingType string, lat, lon float64) (Ping, error)

	// ListNearby returns all active pings within the given radius (metres) of the given coordinates.
	// Optionally filtered by type ("animal" or "food"). Returns at most 200 results.
	ListNearby(ctx context.Context, q NearbyQuery) ([]Ping, error)

	// GetByID fetches a single ping by its UUID. Returns an error if not found.
	GetByID(ctx context.Context, id string) (Ping, error)

	// Confirm touches updated_at to signal "animal still present here".
	// Returns an error if the ping does not exist or is not active.
	Confirm(ctx context.Context, id string) error

	// MarkFed sets fed_at = NOW() on the ping to signal it has been fed.
	// Returns an error if the ping does not exist or is not active.
	MarkFed(ctx context.Context, id string) error

	// Deactivate performs a soft delete: sets is_active = false.
	// Only the owner (createdBy) may deactivate a ping.
	// Returns ErrNotOwner if userID does not match, ErrNotFound if ping missing.
	Deactivate(ctx context.Context, id, userID string) error

	// AddMedia inserts a media record linked to a ping (file path on disk).
	AddMedia(ctx context.Context, pingID, filePath string) error

	// ListMedia returns the file paths of all media attached to a ping.
	ListMedia(ctx context.Context, pingID string) ([]string, error)
}
