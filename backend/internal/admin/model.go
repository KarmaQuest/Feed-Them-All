// Package admin — model.go définit les types utilisés par le dashboard administrateur.
//
// AdminUser       : vue admin d'un compte utilisateur (inclut email, is_banned, created_at).
// LevelThreshold  : un palier XP → niveau.
// AdminXPAction   : configuration d'une action XP (valeur + daily_limit).
// AdminBadge      : définition complète d'un badge (avec condition brute JSON pour l'édition).
// AdminShopItem   : item de la boutique vu par l'admin (tous les champs éditables).
// AdminPing       : ping vu par l'admin pour modération (avec compteur de reports).
//
// Requêtes de mutation (corps JSON envoyé par le frontend admin) :
//   UpdateUserRequest, UpdateXPActionRequest, UpsertBadgeRequest,
//   UpsertShopItemRequest, UpsertLevelThresholdsRequest
package admin

import "encoding/json"

// ─── Users ────────────────────────────────────────────────────────────────────

// AdminUser is the admin view of a user account.
type AdminUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	XP        int    `json:"xp"`
	IsBanned  bool   `json:"is_banned"`
	CreatedAt string `json:"created_at"`
}

// UpdateUserRequest is the body of PATCH /admin/users/:id.
// Only non-nil fields are applied.
type UpdateUserRequest struct {
	Role     *string `json:"role,omitempty"`
	IsBanned *bool   `json:"is_banned,omitempty"`
}

// ─── Level Thresholds ─────────────────────────────────────────────────────────

// LevelThreshold is a single (level, min_xp) row.
type LevelThreshold struct {
	Level int `json:"level"`
	MinXP int `json:"min_xp"`
}

// UpsertLevelThresholdsRequest is the body of PUT /admin/level-thresholds.
// The full list replaces all existing rows.
type UpsertLevelThresholdsRequest struct {
	Thresholds []LevelThreshold `json:"thresholds"`
}

// ─── XP Actions ───────────────────────────────────────────────────────────────

// AdminXPAction is the admin view of a reward action config.
type AdminXPAction struct {
	Action     string `json:"action"`
	XPValue    int    `json:"xp_value"`
	DailyLimit int    `json:"daily_limit"`
}

// UpdateXPActionRequest is the body of PUT /admin/xp-actions/:action.
type UpdateXPActionRequest struct {
	XPValue    *int `json:"xp_value,omitempty"`
	DailyLimit *int `json:"daily_limit,omitempty"`
}

// ─── Badges ───────────────────────────────────────────────────────────────────

// AdminBadge is the full badge definition for admin CRUD.
// Condition is kept as raw JSON to allow flexible editing.
type AdminBadge struct {
	ID          string          `json:"id"`
	Slug        string          `json:"slug"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Condition   json.RawMessage `json:"condition"`
}

// UpsertBadgeRequest is the body for POST /admin/badges and PUT /admin/badges/:id.
type UpsertBadgeRequest struct {
	Slug        string          `json:"slug"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Condition   json.RawMessage `json:"condition"`
}

// ─── Shop Items ───────────────────────────────────────────────────────────────

// AdminShopItem is the full item view for admin CRUD.
type AdminShopItem struct {
	ID              string          `json:"id"`
	Slug            string          `json:"slug"`
	Name            string          `json:"name"`
	Category        string          `json:"category"` // "skin" | "outfit" | "accessory"
	PriceCents      int             `json:"price_cents"`
	Currency        string          `json:"currency"`
	UnlockCondition json.RawMessage `json:"unlock_condition,omitempty"`
	IsActive        bool            `json:"is_active"`
}

// UpsertShopItemRequest is the body for POST /admin/shop-items and PUT /admin/shop-items/:id.
type UpsertShopItemRequest struct {
	Slug            string          `json:"slug"`
	Name            string          `json:"name"`
	Category        string          `json:"category"`
	PriceCents      int             `json:"price_cents"`
	Currency        string          `json:"currency"`
	UnlockCondition json.RawMessage `json:"unlock_condition,omitempty"`
	IsActive        bool            `json:"is_active"`
}

// ─── Pings ────────────────────────────────────────────────────────────────────

// AdminPing is the moderation view of a ping.
type AdminPing struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	CreatedBy   string  `json:"created_by"`
	IsActive    bool    `json:"is_active"`
	ReportCount int     `json:"report_count"`
	CreatedAt   string  `json:"created_at"`
	AnimalType  *string `json:"animal_type,omitempty"`
	AnimalCount int     `json:"animal_count"`
}

// ─── Admin Create Requests ─────────────────────────────────────────────────────

// CreateUserRequest is the body of POST /admin/users.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// CreateXPActionRequest is the body of POST /admin/xp-actions.
type CreateXPActionRequest struct {
	Action     string `json:"action"`
	XPValue    int    `json:"xp_value"`
	DailyLimit int    `json:"daily_limit"`
}

// AdminCreatePingRequest is the body of POST /admin/pings.
type AdminCreatePingRequest struct {
	UserID      string  `json:"user_id"`
	Type        string  `json:"type"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	AnimalType  *string `json:"animal_type,omitempty"`
	AnimalCount *int    `json:"animal_count,omitempty"`
}

// ─── Comments ─────────────────────────────────────────────────────────────────

// AdminComment is the admin view of a ping comment.
type AdminComment struct {
	ID        string `json:"id"`
	PingID    string `json:"ping_id"`
	AuthorID  string `json:"author_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// UpdateCommentRequest is the body of PATCH /admin/comments/:id.
type UpdateCommentRequest struct {
	Content string `json:"content"`
}

// CreateCommentAdminRequest is the body of POST /admin/pings/:id/comments.
type CreateCommentAdminRequest struct {
	Content string `json:"content"`
}

// ─── Feeding Events (admin view) ──────────────────────────────────────────────

// AdminFeedingEvent is the admin view of a ping feeding event.
type AdminFeedingEvent struct {
	ID              string  `json:"id"`
	PingID          string  `json:"ping_id"`
	FedBy           string  `json:"fed_by"`
	Username        string  `json:"username"`
	FedAt           string  `json:"fed_at"`
	Note            *string `json:"note,omitempty"`
	AnimalCountSeen *int    `json:"animal_count_seen,omitempty"`
}

// UpdateFeedingEventRequest is the body of PATCH /admin/feedings/:id.
type UpdateFeedingEventRequest struct {
	Note            *string `json:"note"`
	AnimalCountSeen *int    `json:"animal_count_seen,omitempty"`
}

// CreateFeedingEventAdminRequest is the body of POST /admin/pings/:id/feedings.
type CreateFeedingEventAdminRequest struct {
	Note            *string `json:"note,omitempty"`
	AnimalCountSeen *int    `json:"animal_count_seen,omitempty"`
}
