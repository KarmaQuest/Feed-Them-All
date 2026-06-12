// Package shop — model.go définit les types de la boutique avatar.
//
// AvatarItem     : un item du catalogue (skin/outfit/accessory).
//                  price_cents = 0 → gratuit, débloqué par quête ou par défaut.
//                  price_cents > 0 → payant, acheté via Stripe.
//
// InventoryItem  : un item possédé par un utilisateur, avec sa source d'acquisition.
//
// PurchaseResponse : réponse à POST /shop/items/:id/purchase.
//                    Contient le client_secret Stripe que le frontend utilise
//                    pour compléter le paiement avec Stripe.js.
//                    Pour les items gratuits (price_cents=0), ce champ est vide
//                    et granted=true (octroi immédiat).
//
// UnlockCondition : même format que badges.condition (JSONB).
//   {"type":"xp_threshold","value":500}
//   {"type":"action_count","action":"feed","value":10}
package shop

// UnlockCondition is the condition that must be met to unlock a free item via quests.
// nil = item is free with no condition (base item) or paid item.
type UnlockCondition struct {
	Type   string `json:"type"`             // "xp_threshold" | "action_count"
	Value  int    `json:"value"`
	Action string `json:"action,omitempty"` // only for "action_count"
}

// AvatarItem represents an item in the shop catalogue.
type AvatarItem struct {
	ID              string           `json:"id"`
	Slug            string           `json:"slug"`
	Name            string           `json:"name"`
	Category        string           `json:"category"` // "skin" | "outfit" | "accessory"
	PriceCents      int              `json:"price_cents"`
	Currency        string           `json:"currency"`
	UnlockCondition *UnlockCondition `json:"unlock_condition,omitempty"`
	IsActive        bool             `json:"is_active"`
}

// InventoryItem is an item owned by a user.
type InventoryItem struct {
	Item       AvatarItem `json:"item"`
	AcquiredAt string     `json:"acquired_at"`
	Source     string     `json:"source"` // "default" | "quest" | "purchase"
}

// PurchaseResponse is the response to POST /shop/items/:id/purchase.
type PurchaseResponse struct {
	// For paid items: Stripe client_secret to complete payment on the frontend.
	ClientSecret string `json:"client_secret,omitempty"`
	// For free items (price_cents=0): granted immediately.
	Granted bool `json:"granted,omitempty"`
}
