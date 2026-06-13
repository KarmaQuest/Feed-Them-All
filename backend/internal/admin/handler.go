// Package admin — handler.go expose les routes HTTP du dashboard admin.
//
// Toutes les routes nécessitent d'être authentifié (auth.Middleware) ET admin (RequireAdmin).
// La validation des entrées est minimale ici — la logique métier est dans service.go.
//
// Routes :
//   GET    /admin/users?page=&search=
//   PATCH  /admin/users/:id
//   GET    /admin/xp-actions
//   PUT    /admin/xp-actions/:action
//   GET    /admin/level-thresholds
//   PUT    /admin/level-thresholds
//   GET    /admin/badges
//   POST   /admin/badges
//   PUT    /admin/badges/:id
//   DELETE /admin/badges/:id
//   GET    /admin/shop-items
//   POST   /admin/shop-items
//   PUT    /admin/shop-items/:id
//   DELETE /admin/shop-items/:id
//   GET    /admin/pings?active=true&flagged=true
//   DELETE /admin/pings/:id
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// adminService is the interface the Handler depends on.
// Matches Service methods — allows unit testing with a fake.
type adminService interface {
	ListUsers(ctx context.Context, page int, search string) ([]AdminUser, error)
	UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) error
	CreateUser(ctx context.Context, req CreateUserRequest) (string, error)
	DeleteUser(ctx context.Context, userID string) error

	ListXPActions(ctx context.Context) ([]AdminXPAction, error)
	UpdateXPAction(ctx context.Context, action string, req UpdateXPActionRequest) error
	CreateXPAction(ctx context.Context, req CreateXPActionRequest) error

	ListLevelThresholds(ctx context.Context) ([]LevelThreshold, error)
	ReplaceAllThresholds(ctx context.Context, req UpsertLevelThresholdsRequest) error

	ListBadges(ctx context.Context) ([]AdminBadge, error)
	CreateBadge(ctx context.Context, req UpsertBadgeRequest) (string, error)
	UpdateBadge(ctx context.Context, badgeID string, req UpsertBadgeRequest) error
	DeleteBadge(ctx context.Context, badgeID string) error

	ListShopItems(ctx context.Context) ([]AdminShopItem, error)
	CreateShopItem(ctx context.Context, req UpsertShopItemRequest) (string, error)
	UpdateShopItem(ctx context.Context, itemID string, req UpsertShopItemRequest) error
	DeleteShopItem(ctx context.Context, itemID string) error

	ListPingsAdmin(ctx context.Context, activeOnly, flaggedOnly bool) ([]AdminPing, error)
	ForceDeactivatePing(ctx context.Context, pingID string) error
	CreatePingAdmin(ctx context.Context, req AdminCreatePingRequest) (string, error)
}

// Handler holds HTTP handlers for admin routes.
type Handler struct {
	svc adminService
}

// NewHandler creates an admin Handler.
func NewHandler(svc adminService) *Handler {
	return &Handler{svc: svc}
}

// ─── Users ────────────────────────────────────────────────────────────────────

// ListUsers handles GET /admin/users?page=&search=
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	search := r.URL.Query().Get("search")

	users, err := h.svc.ListUsers(r.Context(), page, search)
	if err != nil {
		writeError(w, "failed to list users", http.StatusInternalServerError)
		return
	}
	writeJSON(w, users, http.StatusOK)
}

// UpdateUser handles PATCH /admin/users/:id
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateUser(r.Context(), userID, req); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── XP Actions ───────────────────────────────────────────────────────────────

// ListXPActions handles GET /admin/xp-actions
func (h *Handler) ListXPActions(w http.ResponseWriter, r *http.Request) {
	actions, err := h.svc.ListXPActions(r.Context())
	if err != nil {
		writeError(w, "failed to list xp actions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, actions, http.StatusOK)
}

// UpdateXPAction handles PUT /admin/xp-actions/:action
func (h *Handler) UpdateXPAction(w http.ResponseWriter, r *http.Request) {
	action := chi.URLParam(r, "action")
	var req UpdateXPActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateXPAction(r.Context(), action, req); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "action not found", http.StatusNotFound)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Level Thresholds ─────────────────────────────────────────────────────────

// ListLevelThresholds handles GET /admin/level-thresholds
func (h *Handler) ListLevelThresholds(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListLevelThresholds(r.Context())
	if err != nil {
		writeError(w, "failed to list thresholds", http.StatusInternalServerError)
		return
	}
	writeJSON(w, list, http.StatusOK)
}

// ReplaceAllThresholds handles PUT /admin/level-thresholds
func (h *Handler) ReplaceAllThresholds(w http.ResponseWriter, r *http.Request) {
	var req UpsertLevelThresholdsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.ReplaceAllThresholds(r.Context(), req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Badges ───────────────────────────────────────────────────────────────────

// ListBadges handles GET /admin/badges
func (h *Handler) ListBadges(w http.ResponseWriter, r *http.Request) {
	badges, err := h.svc.ListBadges(r.Context())
	if err != nil {
		writeError(w, "failed to list badges", http.StatusInternalServerError)
		return
	}
	writeJSON(w, badges, http.StatusOK)
}

// CreateBadge handles POST /admin/badges
func (h *Handler) CreateBadge(w http.ResponseWriter, r *http.Request) {
	var req UpsertBadgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := h.svc.CreateBadge(r.Context(), req)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"id": id}, http.StatusCreated)
}

// UpdateBadge handles PUT /admin/badges/:id
func (h *Handler) UpdateBadge(w http.ResponseWriter, r *http.Request) {
	badgeID := chi.URLParam(r, "id")
	var req UpsertBadgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateBadge(r.Context(), badgeID, req); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "badge not found", http.StatusNotFound)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteBadge handles DELETE /admin/badges/:id
func (h *Handler) DeleteBadge(w http.ResponseWriter, r *http.Request) {
	badgeID := chi.URLParam(r, "id")
	if err := h.svc.DeleteBadge(r.Context(), badgeID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "badge not found", http.StatusNotFound)
			return
		}
		writeError(w, "failed to delete badge", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Shop Items ───────────────────────────────────────────────────────────────

// ListShopItems handles GET /admin/shop-items
func (h *Handler) ListShopItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListShopItems(r.Context())
	if err != nil {
		writeError(w, "failed to list shop items", http.StatusInternalServerError)
		return
	}
	writeJSON(w, items, http.StatusOK)
}

// CreateShopItem handles POST /admin/shop-items
func (h *Handler) CreateShopItem(w http.ResponseWriter, r *http.Request) {
	var req UpsertShopItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := h.svc.CreateShopItem(r.Context(), req)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"id": id}, http.StatusCreated)
}

// UpdateShopItem handles PUT /admin/shop-items/:id
func (h *Handler) UpdateShopItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	var req UpsertShopItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateShopItem(r.Context(), itemID, req); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "item not found", http.StatusNotFound)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteShopItem handles DELETE /admin/shop-items/:id
func (h *Handler) DeleteShopItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	if err := h.svc.DeleteShopItem(r.Context(), itemID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "item not found", http.StatusNotFound)
			return
		}
		writeError(w, "failed to delete item", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Pings (Moderation) ───────────────────────────────────────────────────────

// ListPingsAdmin handles GET /admin/pings?active=true&flagged=true
func (h *Handler) ListPingsAdmin(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	flaggedOnly := r.URL.Query().Get("flagged") == "true"

	pings, err := h.svc.ListPingsAdmin(r.Context(), activeOnly, flaggedOnly)
	if err != nil {
		writeError(w, "failed to list pings", http.StatusInternalServerError)
		return
	}
	writeJSON(w, pings, http.StatusOK)
}

// ForceDeactivatePing handles DELETE /admin/pings/:id
func (h *Handler) ForceDeactivatePing(w http.ResponseWriter, r *http.Request) {
	pingID := chi.URLParam(r, "id")
	if err := h.svc.ForceDeactivatePing(r.Context(), pingID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "ping not found", http.StatusNotFound)
			return
		}
		writeError(w, "failed to deactivate ping", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreatePingAdmin handles POST /admin/pings
func (h *Handler) CreatePingAdmin(w http.ResponseWriter, r *http.Request) {
	var req AdminCreatePingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := h.svc.CreatePingAdmin(r.Context(), req)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"id": id}, http.StatusCreated)
}

// CreateUser handles POST /admin/users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := h.svc.CreateUser(r.Context(), req)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"id": id}, http.StatusCreated)
}

// DeleteUser handles DELETE /admin/users/:id
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if err := h.svc.DeleteUser(r.Context(), userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		writeError(w, "failed to delete user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateXPAction handles POST /admin/xp-actions
func (h *Handler) CreateXPAction(w http.ResponseWriter, r *http.Request) {
	var req CreateXPActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateXPAction(r.Context(), req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
