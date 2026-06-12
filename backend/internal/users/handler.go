// Package users — handler.go expose les routes HTTP du package users.
//
// Routes :
//   GET /users/:id/profile  → profil public d'un utilisateur (public, sans JWT)
//   GET /leaderboard        → top 20 utilisateurs par XP (public, sans JWT)
//
// Les deux routes sont publiques — pas de JWT requis pour consulter un profil
// ou le classement. C'est cohérent avec le principe de la carte publique.
package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
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
// Returns the public profile of the requested user (XP, level, badges, avatar).
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

	writeJSON(w, http.StatusOK, profile)
}

// GetLeaderboard handles GET /leaderboard.
// Returns the top 20 users by XP. Cached in-memory for 5 minutes.
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
