// Package users — service.go contient la logique métier du profil et du leaderboard.
//
// GetProfile : calcule le level à partir de l'XP (pas stocké en DB, calculé à la volée).
//
// GetLeaderboard : cache in-memory TTL 5 minutes pour éviter des requêtes DB répétées.
//   Le leaderboard ne change pas à chaque requête — inutile de recalculer en permanence.
//   Le cache est protégé par un sync.RWMutex (goroutine-safe).
//
// Level formula (courbe RPG) :
//   Paliers : [0, 100, 250, 500, 900, 1400, 2100, 3000, 4500, 7000]
//   Level 1 = 0 XP · Level 10 = 7000 XP · extensible
package users

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// levelThresholds defines the minimum XP required for each level.
// Index 0 = Level 1 (0 XP), Index 1 = Level 2 (100 XP), etc.
var levelThresholds = []int{0, 100, 250, 500, 900, 1400, 2100, 3000, 4500, 7000}

// computeLevel returns the level for a given XP amount.
// Level 1 at 0 XP, up to Level len(levelThresholds) at the last threshold.
func computeLevel(xp int) int {
	for i := len(levelThresholds) - 1; i >= 0; i-- {
		if xp >= levelThresholds[i] {
			return i + 1
		}
	}
	return 1
}

// Service holds the business logic for user profiles and the leaderboard.
type Service struct {
	store Store

	// Leaderboard in-memory cache (TTL 5 min).
	cacheMu     sync.RWMutex
	cacheData   []LeaderboardEntry
	cacheExpiry time.Time
}

// NewService creates a users Service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// GetProfile returns the public profile for the given user, with computed level.
// Returns ErrNotFound if the user doesn't exist.
func (s *Service) GetProfile(ctx context.Context, userID string) (UserProfile, error) {
	p, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return UserProfile{}, err // ErrNotFound propagates as-is
	}
	p.Level = computeLevel(p.XP)
	return p, nil
}

// GetLeaderboard returns the top 20 users by XP with ranks and computed levels.
// Results are cached in memory for 5 minutes to avoid DB load.
func (s *Service) GetLeaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	// Fast path: return cached data if still valid.
	s.cacheMu.RLock()
	if time.Now().Before(s.cacheExpiry) {
		cached := s.cacheData
		s.cacheMu.RUnlock()
		return cached, nil
	}
	s.cacheMu.RUnlock()

	// Cache miss: fetch fresh data from DB.
	entries, err := s.store.GetLeaderboard(ctx)
	if err != nil {
		return nil, fmt.Errorf("users.Service.GetLeaderboard: %w", err)
	}

	// Compute level for each entry (rank is already set by the repository).
	for i := range entries {
		entries[i].Level = computeLevel(entries[i].XP)
	}

	// Update cache.
	s.cacheMu.Lock()
	s.cacheData = entries
	s.cacheExpiry = time.Now().Add(5 * time.Minute)
	s.cacheMu.Unlock()

	return entries, nil
}
