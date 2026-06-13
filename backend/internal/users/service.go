// Package users — service.go contient la logique métier du profil et du leaderboard.
//
// GetProfile : calcule le level à partir de l'XP (pas stocké en DB, calculé à la volée).
//
// GetLeaderboard : cache in-memory TTL 5 minutes pour éviter des requêtes DB répétées.
//   Le leaderboard ne change pas à chaque requête — inutile de recalculer en permanence.
//   Le cache est protégé par un sync.RWMutex (goroutine-safe).
//
// LoadThresholds : lit les paliers XP depuis level_thresholds au démarrage du serveur.
//   Les paliers sont stockés dans le Service (protégés par un RWMutex).
//   Fallback hardcodé si la table est vide.
//   Appelé depuis main.go après NewService().
//
// Level formula (courbe RPG) :
//   Paliers : [0, 100, 250, 500, 900, 1400, 2100, 3000, 4500, 7000]
//   Level 1 = 0 XP · Level 10 = 7000 XP · extensible
package users

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// defaultLevelThresholds is the hardcoded fallback used if level_thresholds table is empty.
// Index 0 = Level 1 (0 XP), Index 1 = Level 2 (100 XP), etc.
var defaultLevelThresholds = []int{0, 100, 250, 500, 900, 1400, 2100, 3000, 4500, 7000}

// Service holds the business logic for user profiles and the leaderboard.
type Service struct {
	store Store

	// Level thresholds loaded from DB at startup (configurable via admin dashboard).
	thresholdsMu sync.RWMutex
	thresholds   []int

	// Leaderboard in-memory cache (TTL 5 min).
	cacheMu     sync.RWMutex
	cacheData   []LeaderboardEntry
	cacheExpiry time.Time
}

// NewService creates a users Service with default hardcoded thresholds.
// Call LoadThresholds(ctx) after construction to override with DB values.
func NewService(store Store) *Service {
	thresholds := make([]int, len(defaultLevelThresholds))
	copy(thresholds, defaultLevelThresholds)
	return &Service{store: store, thresholds: thresholds}
}

// LoadThresholds reads level_thresholds from the DB and stores them in the service.
// Falls back to hardcoded defaults if the table is empty or on error.
func (s *Service) LoadThresholds(ctx context.Context) {
	t, err := s.store.GetLevelThresholds(ctx)
	if err != nil {
		slog.Warn("users.LoadThresholds: failed to load from DB, using hardcoded defaults", "err", err)
		return
	}
	if len(t) == 0 {
		slog.Warn("users.LoadThresholds: level_thresholds table is empty, using hardcoded defaults")
		return
	}
	s.thresholdsMu.Lock()
	s.thresholds = t
	s.thresholdsMu.Unlock()
	slog.Info("users.LoadThresholds: loaded from DB", "count", len(t))
}

// ReloadThresholds re-reads the thresholds from DB (called by admin after a PUT).
func (s *Service) ReloadThresholds(ctx context.Context) error {
	t, err := s.store.GetLevelThresholds(ctx)
	if err != nil {
		return fmt.Errorf("users.ReloadThresholds: %w", err)
	}
	if len(t) == 0 {
		return nil
	}
	s.thresholdsMu.Lock()
	s.thresholds = t
	s.thresholdsMu.Unlock()
	return nil
}

// computeLevel returns the level for a given XP amount using the service's thresholds.
func (s *Service) computeLevel(xp int) int {
	s.thresholdsMu.RLock()
	t := s.thresholds
	s.thresholdsMu.RUnlock()

	for i := len(t) - 1; i >= 0; i-- {
		if xp >= t[i] {
			return i + 1
		}
	}
	return 1
}

// GetProfile returns the public profile for the given user, with computed level.
// Returns ErrNotFound if the user doesn't exist.
func (s *Service) GetProfile(ctx context.Context, userID string) (UserProfile, error) {
	p, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return UserProfile{}, err // ErrNotFound propagates as-is
	}
	p.Level = s.computeLevel(p.XP)
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
		entries[i].Level = s.computeLevel(entries[i].XP)
	}

	// Update cache.
	s.cacheMu.Lock()
	s.cacheData = entries
	s.cacheExpiry = time.Now().Add(5 * time.Minute)
	s.cacheMu.Unlock()

	return entries, nil
}

