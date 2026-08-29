// Package users — handler.go expose les routes HTTP du package users.
//
// Routes :
//   GET   /users/:id/profile  → profil public (si privé + non-propriétaire → profil réduit)
//   PATCH /users/me/privacy   → toggle is_private (JWT requis)
//   GET   /leaderboard        → top 20 utilisateurs par XP (public, sans JWT)
package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/KarmaQuest/feed-them-all/internal/auth"
)

// Handler wires HTTP routes to the users service.
type Handler struct {
	svc *Service
}

// NewHandler creates a Handler backed by the given Service.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetProfile handles GET /users/:id/profile.
// Returns the full public profile, or a redacted version if the profile is private.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		writeError(w, "missing user id", http.StatusBadRequest)
		return
	}

	profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("GetProfile failed", "user_id", userID, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// If the profile is private and the requester is not the owner, return redacted view.
	if profile.IsPrivate {
		requesterID := auth.UserIDFromContext(r.Context())
		if requesterID != userID {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"id":         profile.ID,
				"username":   profile.Username,
				"level":      profile.Level,
				"is_private": true,
			})
			return
		}
	}

	writeJSON(w, http.StatusOK, profile)
}

// UpdatePrivacy handles PATCH /users/me/privacy (JWT required).
func (h *Handler) UpdatePrivacy(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdatePrivacyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdatePrivacy(r.Context(), userID, req.IsPrivate); err != nil {
		slog.Error("UpdatePrivacy failed", "user_id", userID, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateAvatar handles PATCH /users/me/avatar (JWT required).
// Replaces the avatar_config JSONB for the authenticated user.
func (h *Handler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req AvatarConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Config == nil {
		req.Config = map[string]interface{}{}
	}

	if err := h.svc.UpdateAvatarConfig(r.Context(), userID, req.Config); err != nil {
		slog.Error("UpdateAvatar failed", "user_id", userID, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetLeaderboard handles GET /leaderboard.
func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	entries, err := h.svc.GetLeaderboard(r.Context())
	if err != nil {
		slog.Error("GetLeaderboard failed", "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("users: writeJSON encode failed", "err", err)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}
