// Package auth — Authentification des utilisateurs FeedThemAll.
//
// Ce fichier (model.go) définit les structures de données (les "types") utilisées
// dans tout le package auth. C'est le vocabulaire commun entre les couches :
//
//   User            → représente un utilisateur tel qu'il est stocké en base de données.
//                     Utilisé dans les réponses JSON renvoyées au client (sans le mot de passe).
//
//   RegisterRequest → les données envoyées par le client lors d'une inscription
//                     (email, username, password, role).
//
//   LoginRequest    → les données envoyées lors d'une connexion (email + password).
//
//   TokenResponse   → ce que le serveur renvoie après une inscription ou connexion réussie :
//                     l'access token JWT + les infos de l'utilisateur.
//                     Le refresh token, lui, est envoyé via un cookie HttpOnly (non visible ici).
package auth

import "time"

// User represents a row in the users table.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	IsPremium    bool      `json:"is_premium"`
	XP           int       `json:"xp"`
	AvatarConfig []byte    `json:"avatar_config"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest is the payload for POST /auth/register.
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"` // feeder | giver | association
}

// LoginRequest is the payload for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse is returned after successful register or login.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}
