// Package users — model.go définit les types du profil utilisateur et du leaderboard.
//
// UserProfile : données publiques d'un utilisateur (XP, level, badges, avatar).
//               Retourné par GET /users/:id/profile.
//
// LeaderboardEntry : entrée du classement global.
//                    Retourné par GET /leaderboard.
//
// Level formula (calculée côté service, jamais stockée en DB) :
//   On utilise une courbe de progression RPG classique avec des paliers fixes.
//   Voir computeLevel() dans service.go.
package users

// BadgeSummary is the condensed badge info returned in a user profile.
type BadgeSummary struct {
	Slug  string `json:"slug"`
	Label string `json:"label"`
}

// UserProfile is the public view of a user's account and achievements.
type UserProfile struct {
	ID           string                 `json:"id"`
	Username     string                 `json:"username"`
	Role         string                 `json:"role"`
	XP           int                    `json:"xp"`
	Level        int                    `json:"level"`         // computed, not stored in DB
	Badges       []BadgeSummary         `json:"badges"`
	AvatarConfig map[string]interface{} `json:"avatar_config"`
}

// LeaderboardEntry is a single row in the XP leaderboard.
type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	XP       int    `json:"xp"`
	Level    int    `json:"level"`
}
