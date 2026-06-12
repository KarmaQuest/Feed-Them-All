// Package users — store.go définit l'interface Store du package users.
//
// Méthodes :
//   GetProfile     → retourne les infos publiques d'un utilisateur + ses badges
//   GetLeaderboard → top 20 utilisateurs par XP (pour le cache in-memory du service)
package users

import "context"

// Store is the interface the Service depends on for user DB operations.
type Store interface {
	// GetProfile returns the public profile of a user by their UUID.
	// Returns ErrNotFound if the user does not exist.
	GetProfile(ctx context.Context, userID string) (UserProfile, error)

	// GetLeaderboard returns the top 20 users sorted by XP descending.
	// Called by the service when the in-memory cache is stale (TTL 5 min).
	GetLeaderboard(ctx context.Context) ([]LeaderboardEntry, error)
}
