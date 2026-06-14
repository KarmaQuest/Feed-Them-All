// Package auth — repository.go contient toutes les requêtes SQL du package auth.
//
// Repository est l'implémentation concrète de l'interface Store (voir store.go).
// Il communique directement avec PostgreSQL via pgx (un driver Go pour PostgreSQL).
// pgxpool gère un "pool" de connexions : plutôt que d'ouvrir/fermer une connexion
// à chaque requête, il réutilise des connexions existantes pour la performance.
//
// Règle importante : ce fichier ne contient AUCUNE logique métier.
// Il fait uniquement SELECT / INSERT / UPDATE / DELETE et retourne les résultats bruts.
// Toute logique (validation, règles business) appartient à service.go.
//
// Requêtes SQL implémentées :
//   CreateUser         → INSERT INTO users ... RETURNING (retourne l'utilisateur créé)
//   GetUserByEmail     → SELECT ... WHERE email = $1 (inclut le hash du mot de passe)
//   StoreRefreshToken  → INSERT ... ON CONFLICT UPDATE (upsert : crée ou remplace)
//   GetRefreshToken    → SELECT token_hash WHERE user_id = $1
//   DeleteRefreshToken → DELETE WHERE user_id = $1 (appelé au logout)
package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles all DB operations for auth.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user and returns the created row.
func (r *Repository) CreateUser(ctx context.Context, email, username, passwordHash, role string, roles []string) (User, error) {
	const q = `
		INSERT INTO users (email, username, password_hash, role, roles)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, username, role, roles, is_premium, xp, avatar_config, created_at
	`
	var u User
	err := r.db.QueryRow(ctx, q, email, username, passwordHash, role, roles).Scan(
		&u.ID, &u.Email, &u.Username, &u.Role, &u.Roles, &u.IsPremium, &u.XP, &u.AvatarConfig, &u.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("auth.CreateUser: %w", err)
	}
	return u, nil
}

// GetUserByEmail fetches a user and their password hash for login.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (User, string, error) {
	const q = `
		SELECT id, email, username, role, roles, is_premium, xp, avatar_config, created_at, password_hash
		FROM users
		WHERE email = $1
	`
	var u User
	var hash string
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.Username, &u.Role, &u.Roles, &u.IsPremium, &u.XP, &u.AvatarConfig, &u.CreatedAt, &hash,
	)
	if err != nil {
		return User{}, "", fmt.Errorf("auth.GetUserByEmail: %w", err)
	}
	return u, hash, nil
}

// StoreRefreshToken persists a refresh token for a user.
func (r *Repository) StoreRefreshToken(ctx context.Context, userID, tokenHash string) error {
	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET token_hash = $2, created_at = NOW()
	`
	_, err := r.db.Exec(ctx, q, userID, tokenHash)
	if err != nil {
		return fmt.Errorf("auth.StoreRefreshToken: %w", err)
	}
	return nil
}

// GetRefreshToken fetches the stored token hash for a user.
func (r *Repository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	const q = `SELECT token_hash FROM refresh_tokens WHERE user_id = $1`
	var hash string
	err := r.db.QueryRow(ctx, q, userID).Scan(&hash)
	if err != nil {
		return "", fmt.Errorf("auth.GetRefreshToken: %w", err)
	}
	return hash, nil
}

// DeleteRefreshToken removes a refresh token (logout).
func (r *Repository) DeleteRefreshToken(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("auth.DeleteRefreshToken: %w", err)
	}
	return nil
}
