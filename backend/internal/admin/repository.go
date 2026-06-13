// Package admin — repository.go implémente l'interface Store avec pgx/v5.
//
// Toutes les requêtes SQL du dashboard admin sont regroupées ici.
// Aucune logique métier — uniquement lecture/écriture en base.
//
// ReplaceAllThresholds : utilise une transaction explicite (TRUNCATE + INSERT)
//   pour garantir l'atomicité du remplacement complet des paliers.
package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when the target row does not exist.
var ErrNotFound = errors.New("resource not found")

// Repository implements Store using a PostgreSQL connection pool.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new admin Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ─── Users ────────────────────────────────────────────────────────────────────

const listUsersQuery = `
SELECT id, username, email, role, xp, is_banned,
       to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
FROM users
WHERE ($1 = '' OR username ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT 20 OFFSET $2`

func (r *Repository) ListUsers(ctx context.Context, page int, search string) ([]AdminUser, error) {
	offset := (page - 1) * 20
	rows, err := r.db.Query(ctx, listUsersQuery, search, offset)
	if err != nil {
		return nil, fmt.Errorf("admin.ListUsers: %w", err)
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.XP, &u.IsBanned, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("admin.ListUsers scan: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.ListUsers rows: %w", err)
	}
	if users == nil {
		users = []AdminUser{}
	}
	return users, nil
}

func (r *Repository) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) error {
	// Build partial UPDATE dynamically.
	if req.Role == nil && req.IsBanned == nil {
		return nil // nothing to do
	}

	// We always update at least one field; build the SET clauses.
	args := []any{userID}
	set := ""
	if req.Role != nil {
		args = append(args, *req.Role)
		set += fmt.Sprintf("role = $%d, ", len(args))
	}
	if req.IsBanned != nil {
		args = append(args, *req.IsBanned)
		set += fmt.Sprintf("is_banned = $%d, ", len(args))
	}
	// Trim trailing ", "
	set = set[:len(set)-2]

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $1", set)
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("admin.UpdateUser: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── XP Actions ───────────────────────────────────────────────────────────────

const listXPActionsQuery = `
SELECT action, xp_value, daily_limit
FROM xp_actions
ORDER BY action`

func (r *Repository) ListXPActions(ctx context.Context) ([]AdminXPAction, error) {
	rows, err := r.db.Query(ctx, listXPActionsQuery)
	if err != nil {
		return nil, fmt.Errorf("admin.ListXPActions: %w", err)
	}
	defer rows.Close()

	var actions []AdminXPAction
	for rows.Next() {
		var a AdminXPAction
		if err := rows.Scan(&a.Action, &a.XPValue, &a.DailyLimit); err != nil {
			return nil, fmt.Errorf("admin.ListXPActions scan: %w", err)
		}
		actions = append(actions, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.ListXPActions rows: %w", err)
	}
	if actions == nil {
		actions = []AdminXPAction{}
	}
	return actions, nil
}

func (r *Repository) UpdateXPAction(ctx context.Context, action string, req UpdateXPActionRequest) error {
	if req.XPValue == nil && req.DailyLimit == nil {
		return nil
	}

	args := []any{action}
	set := ""
	if req.XPValue != nil {
		args = append(args, *req.XPValue)
		set += fmt.Sprintf("xp_value = $%d, ", len(args))
	}
	if req.DailyLimit != nil {
		args = append(args, *req.DailyLimit)
		set += fmt.Sprintf("daily_limit = $%d, ", len(args))
	}
	set = set[:len(set)-2]

	query := fmt.Sprintf("UPDATE xp_actions SET %s WHERE action = $1", set)
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("admin.UpdateXPAction: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── Level Thresholds ─────────────────────────────────────────────────────────

const listThresholdsQuery = `SELECT level, min_xp FROM level_thresholds ORDER BY level ASC`

func (r *Repository) ListLevelThresholds(ctx context.Context) ([]LevelThreshold, error) {
	rows, err := r.db.Query(ctx, listThresholdsQuery)
	if err != nil {
		return nil, fmt.Errorf("admin.ListLevelThresholds: %w", err)
	}
	defer rows.Close()

	var list []LevelThreshold
	for rows.Next() {
		var t LevelThreshold
		if err := rows.Scan(&t.Level, &t.MinXP); err != nil {
			return nil, fmt.Errorf("admin.ListLevelThresholds scan: %w", err)
		}
		list = append(list, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.ListLevelThresholds rows: %w", err)
	}
	if list == nil {
		list = []LevelThreshold{}
	}
	return list, nil
}

func (r *Repository) ReplaceAllThresholds(ctx context.Context, thresholds []LevelThreshold) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("admin.ReplaceAllThresholds begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, "TRUNCATE TABLE level_thresholds"); err != nil {
		return fmt.Errorf("admin.ReplaceAllThresholds truncate: %w", err)
	}
	for _, t := range thresholds {
		if _, err := tx.Exec(ctx, "INSERT INTO level_thresholds (level, min_xp) VALUES ($1, $2)", t.Level, t.MinXP); err != nil {
			return fmt.Errorf("admin.ReplaceAllThresholds insert level %d: %w", t.Level, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("admin.ReplaceAllThresholds commit: %w", err)
	}
	return nil
}

// ─── Badges ───────────────────────────────────────────────────────────────────

const listBadgesAdminQuery = `
SELECT id, slug, label, description, condition
FROM badges
ORDER BY slug`

func (r *Repository) ListBadges(ctx context.Context) ([]AdminBadge, error) {
	rows, err := r.db.Query(ctx, listBadgesAdminQuery)
	if err != nil {
		return nil, fmt.Errorf("admin.ListBadges: %w", err)
	}
	defer rows.Close()

	var badges []AdminBadge
	for rows.Next() {
		var b AdminBadge
		if err := rows.Scan(&b.ID, &b.Slug, &b.Label, &b.Description, &b.Condition); err != nil {
			return nil, fmt.Errorf("admin.ListBadges scan: %w", err)
		}
		badges = append(badges, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.ListBadges rows: %w", err)
	}
	if badges == nil {
		badges = []AdminBadge{}
	}
	return badges, nil
}

func (r *Repository) CreateBadge(ctx context.Context, req UpsertBadgeRequest) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO badges (slug, label, description, condition)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		req.Slug, req.Label, req.Description, req.Condition,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("admin.CreateBadge: %w", err)
	}
	return id, nil
}

func (r *Repository) UpdateBadge(ctx context.Context, badgeID string, req UpsertBadgeRequest) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE badges SET slug=$2, label=$3, description=$4, condition=$5 WHERE id=$1`,
		badgeID, req.Slug, req.Label, req.Description, req.Condition,
	)
	if err != nil {
		return fmt.Errorf("admin.UpdateBadge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteBadge(ctx context.Context, badgeID string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM badges WHERE id = $1", badgeID)
	if err != nil {
		return fmt.Errorf("admin.DeleteBadge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── Shop Items ───────────────────────────────────────────────────────────────

const listShopItemsAdminQuery = `
SELECT id, slug, name, category, price_cents, currency, unlock_condition, is_active
FROM avatar_items
ORDER BY category, name`

func (r *Repository) ListShopItems(ctx context.Context) ([]AdminShopItem, error) {
	rows, err := r.db.Query(ctx, listShopItemsAdminQuery)
	if err != nil {
		return nil, fmt.Errorf("admin.ListShopItems: %w", err)
	}
	defer rows.Close()

	var items []AdminShopItem
	for rows.Next() {
		var it AdminShopItem
		if err := rows.Scan(&it.ID, &it.Slug, &it.Name, &it.Category,
			&it.PriceCents, &it.Currency, &it.UnlockCondition, &it.IsActive); err != nil {
			return nil, fmt.Errorf("admin.ListShopItems scan: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.ListShopItems rows: %w", err)
	}
	if items == nil {
		items = []AdminShopItem{}
	}
	return items, nil
}

func (r *Repository) CreateShopItem(ctx context.Context, req UpsertShopItemRequest) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO avatar_items (slug, name, category, price_cents, currency, unlock_condition, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		req.Slug, req.Name, req.Category, req.PriceCents,
		req.Currency, nullableJSON(req.UnlockCondition), req.IsActive,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("admin.CreateShopItem: %w", err)
	}
	return id, nil
}

func (r *Repository) UpdateShopItem(ctx context.Context, itemID string, req UpsertShopItemRequest) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE avatar_items
		 SET slug=$2, name=$3, category=$4, price_cents=$5, currency=$6,
		     unlock_condition=$7, is_active=$8
		 WHERE id=$1`,
		itemID, req.Slug, req.Name, req.Category, req.PriceCents,
		req.Currency, nullableJSON(req.UnlockCondition), req.IsActive,
	)
	if err != nil {
		return fmt.Errorf("admin.UpdateShopItem: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteShopItem(ctx context.Context, itemID string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM avatar_items WHERE id = $1", itemID)
	if err != nil {
		return fmt.Errorf("admin.DeleteShopItem: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── Pings (Moderation) ───────────────────────────────────────────────────────

const listPingsAdminQuery = `
SELECT p.id, p.type, p.created_by, p.is_active,
       COUNT(pr.id) AS report_count,
       to_char(p.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
FROM pings p
LEFT JOIN ping_reports pr ON pr.ping_id = p.id
WHERE ($1 = FALSE OR p.is_active = TRUE)
GROUP BY p.id
HAVING ($2 = FALSE OR COUNT(pr.id) > 0)
ORDER BY report_count DESC, p.created_at DESC
LIMIT 100`

func (r *Repository) ListPingsAdmin(ctx context.Context, activeOnly, flaggedOnly bool) ([]AdminPing, error) {
	rows, err := r.db.Query(ctx, listPingsAdminQuery, activeOnly, flaggedOnly)
	if err != nil {
		return nil, fmt.Errorf("admin.ListPingsAdmin: %w", err)
	}
	defer rows.Close()

	var pings []AdminPing
	for rows.Next() {
		var p AdminPing
		if err := rows.Scan(&p.ID, &p.Type, &p.CreatedBy, &p.IsActive, &p.ReportCount, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("admin.ListPingsAdmin scan: %w", err)
		}
		pings = append(pings, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.ListPingsAdmin rows: %w", err)
	}
	if pings == nil {
		pings = []AdminPing{}
	}
	return pings, nil
}

func (r *Repository) ForceDeactivatePing(ctx context.Context, pingID string) error {
	tag, err := r.db.Exec(ctx,
		"UPDATE pings SET is_active = FALSE WHERE id = $1",
		pingID,
	)
	if err != nil {
		return fmt.Errorf("admin.ForceDeactivatePing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CreatePingAdmin(ctx context.Context, req AdminCreatePingRequest) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO pings (type, location, created_by)
		 VALUES ($1, ST_MakePoint($2, $3)::geography, $4)
		 RETURNING id`,
		req.Type, req.Lon, req.Lat, req.UserID,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("admin.CreatePingAdmin: %w", err)
	}
	return id, nil
}

// ─── Users (Create) ───────────────────────────────────────────────────────────

func (r *Repository) CreateUser(ctx context.Context, req CreateUserRequest, passwordHash string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, username, password_hash, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		req.Email, req.Username, passwordHash, req.Role,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("admin.CreateUser: %w", err)
	}
	return id, nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("admin.DeleteUser: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── XP Actions (Create) ──────────────────────────────────────────────────────

func (r *Repository) CreateXPAction(ctx context.Context, req CreateXPActionRequest) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO xp_actions (action, xp_value, daily_limit) VALUES ($1, $2, $3)`,
		req.Action, req.XPValue, req.DailyLimit,
	)
	if err != nil {
		return fmt.Errorf("admin.CreateXPAction: %w", err)
	}
	return nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// nullableJSON returns nil when j is empty, otherwise the raw bytes.
// Used for optional JSONB columns (unlock_condition) to avoid storing "null".
func nullableJSON(j []byte) interface{} {
	if len(j) == 0 || string(j) == "null" {
		return nil
	}
	return j
}

// ensure Repository satisfies the Store interface at compile time.
var _ Store = (*Repository)(nil)

// pgx.ErrNoRows is used to detect missing rows — import kept alive.
var _ = pgx.ErrNoRows
