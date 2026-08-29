// Package gamification — repository.go implémente l'interface Store avec pgx/v5.
//
// Toutes les requêtes SQL de gamification sont ici.
// Aucune logique métier — uniquement de l'accès base de données.
//
// LogAndAwardXP utilise un CTE avec deux DML (INSERT + UPDATE) exécutés
// atomiquement dans la même transaction PostgreSQL implicite.
package gamification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements Store using a PostgreSQL connection pool.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new gamification Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetActionConfig fetches the XP value and daily limit for an action.
const getActionConfigQuery = `
SELECT action, xp_value, daily_limit FROM xp_actions WHERE action = $1`

func (r *Repository) GetActionConfig(ctx context.Context, action string) (XPAction, error) {
	var a XPAction
	err := r.db.QueryRow(ctx, getActionConfigQuery, action).
		Scan(&a.Action, &a.XPValue, &a.DailyLimit)
	if errors.Is(err, pgx.ErrNoRows) {
		return XPAction{}, fmt.Errorf("gamification: unknown action %q", action)
	}
	if err != nil {
		return XPAction{}, fmt.Errorf("gamification.GetActionConfig: %w", err)
	}
	return a, nil
}

// GetDailyCount counts how many times a user performed an action since the given time.
const getDailyCountQuery = `
SELECT COUNT(*) FROM xp_log
WHERE user_id = $1 AND action = $2 AND created_at >= $3`

func (r *Repository) GetDailyCount(ctx context.Context, userID, action string, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, getDailyCountQuery, userID, action, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("gamification.GetDailyCount: %w", err)
	}
	return count, nil
}

// LogAndAwardXP atomically logs XP and increments the user's total.
// Uses a CTE: INSERT xp_log + UPDATE users.xp in one round-trip.
// Returns the user's new total XP.
const logAndAwardXPQuery = `
WITH log AS (
  INSERT INTO xp_log (user_id, action, xp_earned)
  VALUES ($1, $2, $3)
)
UPDATE users SET xp = xp + $3 WHERE id = $1 RETURNING xp`

func (r *Repository) LogAndAwardXP(ctx context.Context, userID, action string, xpAmount int) (int, error) {
	var newXP int
	err := r.db.QueryRow(ctx, logAndAwardXPQuery, userID, action, xpAmount).Scan(&newXP)
	if err != nil {
		return 0, fmt.Errorf("gamification.LogAndAwardXP: %w", err)
	}
	return newXP, nil
}

// GetUserActionTotal counts all-time occurrences of an action for a user.
const getUserActionTotalQuery = `
SELECT COUNT(*) FROM xp_log WHERE user_id = $1 AND action = $2`

func (r *Repository) GetUserActionTotal(ctx context.Context, userID, action string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, getUserActionTotalQuery, userID, action).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("gamification.GetUserActionTotal: %w", err)
	}
	return count, nil
}

// ListBadges returns all badge definitions.
const listBadgesQuery = `
SELECT id, slug, label, COALESCE(description, ''), condition FROM badges`

func (r *Repository) ListBadges(ctx context.Context) ([]Badge, error) {
	rows, err := r.db.Query(ctx, listBadgesQuery)
	if err != nil {
		return nil, fmt.Errorf("gamification.ListBadges: %w", err)
	}
	defer rows.Close()

	var badges []Badge
	for rows.Next() {
		var b Badge
		var condJSON []byte
		if err := rows.Scan(&b.ID, &b.Slug, &b.Label, &b.Description, &condJSON); err != nil {
			return nil, fmt.Errorf("gamification.ListBadges scan: %w", err)
		}
		if err := json.Unmarshal(condJSON, &b.Condition); err != nil {
			return nil, fmt.Errorf("gamification.ListBadges unmarshal condition: %w", err)
		}
		badges = append(badges, b)
	}
	return badges, rows.Err()
}

// HasBadge returns true if the user already owns the badge.
const hasBadgeQuery = `
SELECT EXISTS(SELECT 1 FROM user_badges WHERE user_id = $1 AND badge_id = $2)`

func (r *Repository) HasBadge(ctx context.Context, userID, badgeID string) (bool, error) {
	var has bool
	err := r.db.QueryRow(ctx, hasBadgeQuery, userID, badgeID).Scan(&has)
	if err != nil {
		return false, fmt.Errorf("gamification.HasBadge: %w", err)
	}
	return has, nil
}

// GrantBadge awards the badge. Idempotent via ON CONFLICT DO NOTHING.
const grantBadgeQuery = `
INSERT INTO user_badges (user_id, badge_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING`

func (r *Repository) GrantBadge(ctx context.Context, userID, badgeID string) error {
	_, err := r.db.Exec(ctx, grantBadgeQuery, userID, badgeID)
	if err != nil {
		return fmt.Errorf("gamification.GrantBadge: %w", err)
	}
	return nil
}
