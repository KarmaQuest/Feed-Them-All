// Package shop — repository.go implémente l'interface Store avec pgx/v5.
package shop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors.
var (
	ErrItemNotFound = errors.New("item not found")
	ErrOrderExists  = errors.New("order already exists for this payment intent")
	ErrAlreadyOwned = errors.New("user already owns this item")
)

// Repository implements Store using a PostgreSQL connection pool.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new shop Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// --- scanItem is a helper that scans an avatar_items row. ---
func scanItem(scan func(...any) error) (AvatarItem, error) {
	var item AvatarItem
	var condJSON []byte
	err := scan(&item.ID, &item.Slug, &item.Name, &item.Category,
		&item.PriceCents, &item.Currency, &condJSON, &item.IsActive)
	if err != nil {
		return AvatarItem{}, err
	}
	if len(condJSON) > 0 && string(condJSON) != "null" {
		var cond UnlockCondition
		if err := json.Unmarshal(condJSON, &cond); err != nil {
			return AvatarItem{}, fmt.Errorf("shop: unmarshal unlock_condition: %w", err)
		}
		item.UnlockCondition = &cond
	}
	return item, nil
}

const listItemsQuery = `
SELECT id, slug, name, category, price_cents, currency, unlock_condition, is_active
FROM avatar_items WHERE is_active = TRUE ORDER BY category, price_cents`

func (r *Repository) ListItems(ctx context.Context) ([]AvatarItem, error) {
	rows, err := r.db.Query(ctx, listItemsQuery)
	if err != nil {
		return nil, fmt.Errorf("shop.ListItems: %w", err)
	}
	defer rows.Close()

	var items []AvatarItem
	for rows.Next() {
		item, err := scanItem(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("shop.ListItems scan: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []AvatarItem{}
	}
	return items, rows.Err()
}

const getItemQuery = `
SELECT id, slug, name, category, price_cents, currency, unlock_condition, is_active
FROM avatar_items WHERE id = $1 AND is_active = TRUE`

func (r *Repository) GetItem(ctx context.Context, itemID string) (AvatarItem, error) {
	item, err := scanItem(r.db.QueryRow(ctx, getItemQuery, itemID).Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return AvatarItem{}, ErrItemNotFound
	}
	if err != nil {
		return AvatarItem{}, fmt.Errorf("shop.GetItem: %w", err)
	}
	return item, nil
}

const getUserInventoryQuery = `
SELECT ai.id, ai.slug, ai.name, ai.category, ai.price_cents, ai.currency,
       ai.unlock_condition, ai.is_active, uai.acquired_at, uai.source
FROM user_avatar_items uai
JOIN avatar_items ai ON uai.item_id = ai.id
WHERE uai.user_id = $1
ORDER BY uai.acquired_at`

func (r *Repository) GetUserInventory(ctx context.Context, userID string) ([]InventoryItem, error) {
	rows, err := r.db.Query(ctx, getUserInventoryQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("shop.GetUserInventory: %w", err)
	}
	defer rows.Close()

	var inventory []InventoryItem
	for rows.Next() {
		var inv InventoryItem
		var condJSON []byte
		err := rows.Scan(
			&inv.Item.ID, &inv.Item.Slug, &inv.Item.Name, &inv.Item.Category,
			&inv.Item.PriceCents, &inv.Item.Currency, &condJSON, &inv.Item.IsActive,
			&inv.AcquiredAt, &inv.Source,
		)
		if err != nil {
			return nil, fmt.Errorf("shop.GetUserInventory scan: %w", err)
		}
		if len(condJSON) > 0 && string(condJSON) != "null" {
			var cond UnlockCondition
			if err := json.Unmarshal(condJSON, &cond); err != nil {
				return nil, fmt.Errorf("shop.GetUserInventory unmarshal: %w", err)
			}
			inv.Item.UnlockCondition = &cond
		}
		inventory = append(inventory, inv)
	}
	if inventory == nil {
		inventory = []InventoryItem{}
	}
	return inventory, rows.Err()
}

const hasItemQuery = `
SELECT EXISTS(SELECT 1 FROM user_avatar_items WHERE user_id = $1 AND item_id = $2)`

func (r *Repository) HasItem(ctx context.Context, userID, itemID string) (bool, error) {
	var has bool
	err := r.db.QueryRow(ctx, hasItemQuery, userID, itemID).Scan(&has)
	if err != nil {
		return false, fmt.Errorf("shop.HasItem: %w", err)
	}
	return has, nil
}

const grantItemQuery = `
INSERT INTO user_avatar_items (user_id, item_id, source)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, item_id) DO NOTHING`

func (r *Repository) GrantItem(ctx context.Context, userID, itemID, source string) error {
	_, err := r.db.Exec(ctx, grantItemQuery, userID, itemID, source)
	if err != nil {
		return fmt.Errorf("shop.GrantItem: %w", err)
	}
	return nil
}

const createOrderQuery = `
INSERT INTO shop_orders (user_id, item_id, stripe_payment_intent_id, amount_cents, currency)
VALUES ($1, $2, $3, $4, $5)`

func (r *Repository) CreateOrder(ctx context.Context, userID, itemID, paymentIntentID string, amountCents int, currency string) error {
	_, err := r.db.Exec(ctx, createOrderQuery, userID, itemID, paymentIntentID, amountCents, currency)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrOrderExists
		}
		return fmt.Errorf("shop.CreateOrder: %w", err)
	}
	return nil
}

// CompleteOrder marks the order as succeeded and grants the item — atomically.
const completeOrderQuery = `
WITH order_update AS (
  UPDATE shop_orders
  SET status = 'succeeded', updated_at = NOW()
  WHERE stripe_payment_intent_id = $1 AND status = 'pending'
  RETURNING user_id, item_id
)
INSERT INTO user_avatar_items (user_id, item_id, source)
SELECT user_id, item_id, 'purchase' FROM order_update
ON CONFLICT (user_id, item_id) DO NOTHING`

func (r *Repository) CompleteOrder(ctx context.Context, paymentIntentID string) error {
	_, err := r.db.Exec(ctx, completeOrderQuery, paymentIntentID)
	if err != nil {
		return fmt.Errorf("shop.CompleteOrder: %w", err)
	}
	return nil
}

const listQuestItemsQuery = `
SELECT id, slug, name, category, price_cents, currency, unlock_condition, is_active
FROM avatar_items
WHERE is_active = TRUE AND price_cents = 0 AND unlock_condition IS NOT NULL`

func (r *Repository) ListQuestItems(ctx context.Context) ([]AvatarItem, error) {
	rows, err := r.db.Query(ctx, listQuestItemsQuery)
	if err != nil {
		return nil, fmt.Errorf("shop.ListQuestItems: %w", err)
	}
	defer rows.Close()

	var items []AvatarItem
	for rows.Next() {
		item, err := scanItem(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("shop.ListQuestItems scan: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
