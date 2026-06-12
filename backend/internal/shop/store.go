// Package shop — store.go définit l'interface Store du package shop.
package shop

import "context"

// Store is the interface the Service depends on for all shop DB operations.
type Store interface {
	// ListItems returns all active items in the catalogue.
	ListItems(ctx context.Context) ([]AvatarItem, error)

	// GetItem returns a single item by its UUID.
	// Returns ErrItemNotFound if the item does not exist or is inactive.
	GetItem(ctx context.Context, itemID string) (AvatarItem, error)

	// GetUserInventory returns all items owned by the user.
	GetUserInventory(ctx context.Context, userID string) ([]InventoryItem, error)

	// HasItem returns true if the user already owns the item.
	HasItem(ctx context.Context, userID, itemID string) (bool, error)

	// GrantItem adds an item to the user's inventory.
	// Idempotent (ON CONFLICT DO NOTHING) — safe to call multiple times.
	GrantItem(ctx context.Context, userID, itemID, source string) error

	// CreateOrder records a pending Stripe PaymentIntent.
	// Returns ErrOrderExists if a pending order already exists for this payment_intent_id.
	CreateOrder(ctx context.Context, userID, itemID, paymentIntentID string, amountCents int, currency string) error

	// CompleteOrder marks a Stripe order as succeeded and grants the item.
	// Idempotent — safe to call multiple times for the same payment_intent_id.
	CompleteOrder(ctx context.Context, paymentIntentID string) error

	// ListQuestItems returns all free items (price_cents=0) that have an unlock_condition.
	// Used by the gamification service to check quest rewards after each XP award.
	ListQuestItems(ctx context.Context) ([]AvatarItem, error)
}
