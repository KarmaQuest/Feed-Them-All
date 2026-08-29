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

// GetProfile fetches a user's public data, badge list, and activity counts.
const getUserQuery = `
SELECT id, username, role, roles, xp, avatar_config, is_private,
  (SELECT COUNT(*) FROM pings WHERE created_by = u.id AND is_active = TRUE) AS nb_pings,
  (SELECT COUNT(*) FROM ping_feeding_events WHERE fed_by = u.id) AS nb_feedings
FROM users u
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
		Scan(&p.ID, &p.Username, &p.Role, &p.Roles, &p.XP, &avatarJSON, &p.IsPrivate, &p.NbPings, &p.NbFeedings)
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

// GetLevelThresholds reads the XP thresholds from level_thresholds, sorted by level ASC.
// Returns a slice where index 0 = level 1 min_xp, index 1 = level 2 min_xp, etc.
const getLevelThresholdsQuery = `
SELECT min_xp
FROM level_thresholds
ORDER BY level ASC`

func (r *Repository) GetLevelThresholds(ctx context.Context) ([]int, error) {
	rows, err := r.db.Query(ctx, getLevelThresholdsQuery)
	if err != nil {
		return nil, fmt.Errorf("users.GetLevelThresholds: %w", err)
	}
	defer rows.Close()

	var thresholds []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("users.GetLevelThresholds scan: %w", err)
		}
		thresholds = append(thresholds, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("users.GetLevelThresholds rows: %w", err)
	}
	return thresholds, nil
}

// UpdatePrivacy sets the is_private flag for the given user.
func (r *Repository) UpdatePrivacy(ctx context.Context, userID string, isPrivate bool) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET is_private = $2 WHERE id = $1`, userID, isPrivate)
	if err != nil {
		return fmt.Errorf("users.UpdatePrivacy: %w", err)
	}
	return nil
}

// UpdateAvatarConfig replaces the avatar_config JSONB for the given user.
func (r *Repository) UpdateAvatarConfig(ctx context.Context, userID string, config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("users.UpdateAvatarConfig marshal: %w", err)
	}
	_, err = r.db.Exec(ctx, `UPDATE users SET avatar_config = $2 WHERE id = $1`, userID, configJSON)
	if err != nil {
		return fmt.Errorf("users.UpdateAvatarConfig: %w", err)
	}
	return nil
}
