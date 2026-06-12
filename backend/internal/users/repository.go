// Package users — repository.go implémente l'interface Store avec pgx/v5.
//
// GetProfile : deux requêtes séquentielles (user row + badges JOIN).
// GetLeaderboard : simple SELECT ORDER BY xp DESC LIMIT 20.
// Aucune logique métier ici — uniquement des requêtes SQL.
package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when the requested user does not exist.
var ErrNotFound = errors.New("user not found")

// Repository implements Store using a PostgreSQL connection pool.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new users Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetProfile fetches a user's public data and badge list.
const getUserQuery = `
SELECT id, username, role, xp, avatar_config
FROM users
WHERE id = $1`

const getUserBadgesQuery = `
SELECT b.slug, b.label
FROM user_badges ub
JOIN badges b ON ub.badge_id = b.id
WHERE ub.user_id = $1
ORDER BY ub.earned_at`

func (r *Repository) GetProfile(ctx context.Context, userID string) (UserProfile, error) {
	var p UserProfile
	var avatarJSON []byte

	err := r.db.QueryRow(ctx, getUserQuery, userID).
		Scan(&p.ID, &p.Username, &p.Role, &p.XP, &avatarJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserProfile{}, ErrNotFound
	}
	if err != nil {
		return UserProfile{}, fmt.Errorf("users.GetProfile: %w", err)
	}

	// Deserialize avatar_config JSONB
	if len(avatarJSON) > 0 {
		if err := json.Unmarshal(avatarJSON, &p.AvatarConfig); err != nil {
			return UserProfile{}, fmt.Errorf("users.GetProfile unmarshal avatar: %w", err)
		}
	}
	if p.AvatarConfig == nil {
		p.AvatarConfig = map[string]interface{}{}
	}

	// Fetch badges
	rows, err := r.db.Query(ctx, getUserBadgesQuery, userID)
	if err != nil {
		return UserProfile{}, fmt.Errorf("users.GetProfile badges: %w", err)
	}
	defer rows.Close()

	p.Badges = []BadgeSummary{}
	for rows.Next() {
		var b BadgeSummary
		if err := rows.Scan(&b.Slug, &b.Label); err != nil {
			return UserProfile{}, fmt.Errorf("users.GetProfile badges scan: %w", err)
		}
		p.Badges = append(p.Badges, b)
	}
	if err := rows.Err(); err != nil {
		return UserProfile{}, fmt.Errorf("users.GetProfile badges rows: %w", err)
	}

	return p, nil
}

// GetLeaderboard returns the top 20 users sorted by XP descending.
const getLeaderboardQuery = `
SELECT id, username, xp
FROM users
ORDER BY xp DESC
LIMIT 20`

func (r *Repository) GetLeaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	rows, err := r.db.Query(ctx, getLeaderboardQuery)
	if err != nil {
		return nil, fmt.Errorf("users.GetLeaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.XP); err != nil {
			return nil, fmt.Errorf("users.GetLeaderboard scan: %w", err)
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("users.GetLeaderboard rows: %w", err)
	}
	if entries == nil {
		entries = []LeaderboardEntry{}
	}
	return entries, nil
}
