// Package gamification — service.go contient toute la logique métier de la gamification.
//
// Point d'entrée unique : AwardXP(ctx, userID, action).
// Appelé par le package pings après chaque action récompensée.
//
// Flux d'AwardXP :
//   1. Récupère la config de l'action (valeur XP + daily_limit)
//   2. Vérifie que la limite journalière n'est pas atteinte
//   3. Log l'XP + incrémente users.xp (atomique en DB)
//   4. Lance une goroutine pour vérifier et débloquer les badges
//
// La vérification de badges est asynchrone (goroutine) pour ne pas bloquer
// la réponse HTTP de l'action initiale. Les erreurs sont loggées, jamais propagées.
//
// Le Service implémente l'interface pings.XPAwarder :
//   AwardXP(ctx, userID, action string) error
// Cela évite l'import circulaire pings ↔ gamification.
package gamification

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Service holds the gamification business logic.
type Service struct {
	store Store
}

// NewService creates a gamification Service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// AwardXP grants XP to a user for completing an action.
// It enforces the daily limit and triggers async badge checking.
// Errors are returned to the caller but the caller (pings service) logs and ignores them.
//
// action must match a slug in the xp_actions table (e.g. "feed", "signal_animal").
func (s *Service) AwardXP(ctx context.Context, userID, action string) error {
	// 1. Get action config (XP value + daily limit)
	cfg, err := s.store.GetActionConfig(ctx, action)
	if err != nil {
		// Unknown action — silently ignore (no XP, no error propagated)
		slog.Warn("gamification: unknown action, skipping XP", "action", action)
		return nil
	}

	// 2. Check daily limit (count since today midnight UTC)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	count, err := s.store.GetDailyCount(ctx, userID, action, today)
	if err != nil {
		return fmt.Errorf("gamification.AwardXP: %w", err)
	}
	if count >= cfg.DailyLimit {
		slog.Debug("gamification: daily limit reached", "user_id", userID, "action", action)
		return nil
	}

	// 3. Log XP + increment users.xp atomically
	newXP, err := s.store.LogAndAwardXP(ctx, userID, action, cfg.XPValue)
	if err != nil {
		return fmt.Errorf("gamification.AwardXP LogAndAward: %w", err)
	}

	slog.Info("xp awarded", "user_id", userID, "action", action, "xp", cfg.XPValue, "total", newXP)

	// 4. Async badge check — must not block the HTTP response
	go s.checkBadges(context.Background(), userID, newXP, action)

	return nil
}

// checkBadges checks all badge conditions for a user after an XP award.
// Runs in a goroutine. All errors are logged and swallowed.
func (s *Service) checkBadges(ctx context.Context, userID string, currentXP int, lastAction string) {
	badges, err := s.store.ListBadges(ctx)
	if err != nil {
		slog.Warn("gamification.checkBadges: list failed", "err", err)
		return
	}

	for _, b := range badges {
		// Check if this badge is already owned before evaluating the condition
		has, err := s.store.HasBadge(ctx, userID, b.ID)
		if err != nil || has {
			continue // skip on error or if already owned
		}

		if s.conditionMet(ctx, b.Condition, currentXP, userID) {
			if err := s.store.GrantBadge(ctx, userID, b.ID); err != nil {
				slog.Warn("gamification.checkBadges: grant failed", "badge", b.Slug, "err", err)
				continue
			}
			slog.Info("badge unlocked", "user_id", userID, "badge", b.Slug)
		}
	}
}

// conditionMet evaluates whether a badge condition is satisfied.
func (s *Service) conditionMet(ctx context.Context, cond BadgeCondition, currentXP int, userID string) bool {
	switch cond.Type {
	case "xp_threshold":
		return currentXP >= cond.Value

	case "action_count":
		if cond.Action == "" {
			return false
		}
		total, err := s.store.GetUserActionTotal(ctx, userID, cond.Action)
		if err != nil {
			slog.Warn("gamification.conditionMet: action total failed", "err", err)
			return false
		}
		return total >= cond.Value
	}
	return false
}
