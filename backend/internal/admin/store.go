// Package admin — store.go définit l'interface Store du dashboard administrateur.
//
// Toutes les opérations CRUD admin passent par cette interface :
//   Users        → ListUsers, UpdateUser
//   XP Actions   → ListXPActions, UpdateXPAction
//   Thresholds   → ListLevelThresholds, ReplaceAllThresholds
//   Badges       → ListBadges, CreateBadge, UpdateBadge, DeleteBadge
//   Shop Items   → ListShopItems, CreateShopItem, UpdateShopItem, DeleteShopItem
//   Pings        → ListPingsAdmin, ForceDeactivatePing
package admin

import "context"

// Store is the interface the admin Service depends on for all DB operations.
type Store interface {
	// ── Users ─────────────────────────────────────────────────────────────────

	// ListUsers returns a paginated + filtered list of users.
	// page is 1-based; pageSize is fixed at 20.
	// search filters on username OR email (case-insensitive, partial match).
	ListUsers(ctx context.Context, page int, search string) ([]AdminUser, error)

	// UpdateUser applies a partial update to a user (role and/or is_banned).
	// Returns ErrNotFound if the user does not exist.
	UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) error

	// CreateUser inserts a new user account. passwordHash is the bcrypt hash.
	// Returns the new user UUID.
	CreateUser(ctx context.Context, req CreateUserRequest, passwordHash string) (string, error)

	// DeleteUser permanently removes a user by ID.
	// Returns ErrNotFound if the user does not exist.
	DeleteUser(ctx context.Context, userID string) error

	// ── XP Actions ────────────────────────────────────────────────────────────

	// ListXPActions returns all reward action configs.
	ListXPActions(ctx context.Context) ([]AdminXPAction, error)

	// UpdateXPAction modifies xp_value and/or daily_limit for the given action.
	// Returns ErrNotFound if the action is unknown.
	UpdateXPAction(ctx context.Context, action string, req UpdateXPActionRequest) error

	// CreateXPAction inserts a new reward action config.
	CreateXPAction(ctx context.Context, req CreateXPActionRequest) error

	// ── Level Thresholds ──────────────────────────────────────────────────────

	// ListLevelThresholds returns all (level, min_xp) rows sorted by level ASC.
	ListLevelThresholds(ctx context.Context) ([]LevelThreshold, error)

	// ReplaceAllThresholds deletes all existing rows and inserts the provided ones.
	// Called within a transaction to ensure atomicity.
	ReplaceAllThresholds(ctx context.Context, thresholds []LevelThreshold) error

	// ── Badges ────────────────────────────────────────────────────────────────

	// ListBadges returns all badge definitions.
	ListBadges(ctx context.Context) ([]AdminBadge, error)

	// CreateBadge inserts a new badge and returns the generated UUID.
	CreateBadge(ctx context.Context, req UpsertBadgeRequest) (string, error)

	// UpdateBadge modifies an existing badge. Returns ErrNotFound if missing.
	UpdateBadge(ctx context.Context, badgeID string, req UpsertBadgeRequest) error

	// DeleteBadge removes a badge by ID. Returns ErrNotFound if missing.
	DeleteBadge(ctx context.Context, badgeID string) error

	// ── Shop Items ────────────────────────────────────────────────────────────

	// ListShopItems returns all avatar items (active and inactive).
	ListShopItems(ctx context.Context) ([]AdminShopItem, error)

	// CreateShopItem inserts a new avatar item and returns the generated UUID.
	CreateShopItem(ctx context.Context, req UpsertShopItemRequest) (string, error)

	// UpdateShopItem modifies an existing item. Returns ErrNotFound if missing.
	UpdateShopItem(ctx context.Context, itemID string, req UpsertShopItemRequest) error

	// DeleteShopItem removes an item by ID. Returns ErrNotFound if missing.
	DeleteShopItem(ctx context.Context, itemID string) error

	// ── Pings (Moderation) ────────────────────────────────────────────────────

	// ListPingsAdmin returns pings with their report count.
	// flagged=true → only pings that have at least one report.
	// active=true  → only active pings.
	// Both filters can be combined.
	ListPingsAdmin(ctx context.Context, activeOnly, flaggedOnly bool) ([]AdminPing, error)

	// ForceDeactivatePing sets is_active=false on a ping regardless of ownership.
	// Returns ErrNotFound if the ping does not exist.
	ForceDeactivatePing(ctx context.Context, pingID string) error

	// CreatePingAdmin inserts a new ping on behalf of any user.
	// Returns the new ping UUID.
	CreatePingAdmin(ctx context.Context, req AdminCreatePingRequest) (string, error)
}
