// Package shop — handler.go expose les routes HTTP de la boutique.
//
// Routes :
//   GET  /shop/items                → catalogue public de tous les items actifs
//   GET  /users/me/inventory        → inventaire de l'utilisateur connecté (JWT requis)
//   POST /shop/items/:id/purchase   → acheter/réclamer un item (JWT requis)
//   POST /shop/webhook              → webhook Stripe (pas de JWT — signature Stripe)
package shop

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/KarmaQuest/feed-them-all/internal/auth"
)

// Handler wires HTTP routes to the shop service.
type Handler struct {
	svc *Service
}

// NewHandler creates a Handler backed by the given Service.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetCatalogue handles GET /shop/items (public).
// Returns all active items in the shop.
func (h *Handler) GetCatalogue(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.GetCatalogue(r.Context())
	if err != nil {
		slog.Error("GetCatalogue failed", "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// GetInventory handles GET /users/me/inventory (JWT required).
// Returns all items owned by the authenticated user.
func (h *Handler) GetInventory(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	inventory, err := h.svc.GetInventory(r.Context(), userID)
	if err != nil {
		slog.Error("GetInventory failed", "user_id", userID, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, inventory)
}

// Purchase handles POST /shop/items/:id/purchase (JWT required).
//
// Free item (no condition)  → 200 {"granted": true}
// Paid item                 → 200 {"client_secret": "pi_..."}
// Already owned             → 409
// Item not found            → 404
// Stripe not configured     → 503
func (h *Handler) Purchase(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	itemID := chi.URLParam(r, "id")

	resp, err := h.svc.Purchase(r.Context(), userID, itemID)
	if err != nil {
		switch {
		case errors.Is(err, ErrItemNotFound):
			writeError(w, "item not found", http.StatusNotFound)
		case errors.Is(err, ErrAlreadyInInventory):
			writeError(w, "you already own this item", http.StatusConflict)
		case errors.Is(err, ErrItemPaid):
			writeError(w, "this item is unlocked through quests, not purchased", http.StatusBadRequest)
		case errors.Is(err, ErrStripeNotConfigured):
			writeError(w, "payment system is not configured", http.StatusServiceUnavailable)
		default:
			slog.Error("Purchase failed", "user_id", userID, "item_id", itemID, "err", err)
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Webhook handles POST /shop/webhook (Stripe webhook — no JWT).
// Stripe sends this after a successful payment_intent.
// Returns 200 immediately (Stripe expects a fast response).
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("Stripe-Signature")
	if sig == "" {
		writeError(w, "missing Stripe-Signature header", http.StatusBadRequest)
		return
	}

	// Limit body to 64 KB (Stripe events are small)
	body := io.LimitReader(r.Body, 65536)

	if err := h.svc.HandleWebhook(r.Context(), body, sig); err != nil {
		if errors.Is(err, ErrStripeNotConfigured) {
			writeError(w, "payment system is not configured", http.StatusServiceUnavailable)
			return
		}
		slog.Warn("Webhook failed", "err", err)
		writeError(w, "webhook processing failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("shop: writeJSON encode failed", "err", err)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}
