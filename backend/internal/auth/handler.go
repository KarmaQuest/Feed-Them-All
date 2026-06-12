package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Handler wires HTTP routes to the auth service.
type Handler struct {
	svc             *Service
	registerLimiter *rate.Limiter
	loginLimiter    *rate.Limiter
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:             svc,
		// 3 requests/minute per server (global — per-IP limiting is in middleware)
		registerLimiter: rate.NewLimiter(rate.Every(time.Minute/3), 3),
		loginLimiter:    rate.NewLimiter(rate.Every(time.Second), 5),
	}
}

// Register godoc
// POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.registerLimiter.Allow() {
		writeError(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Basic input validation
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)
	if req.Email == "" || req.Username == "" || req.Password == "" {
		writeError(w, "email, username and password are required", http.StatusBadRequest)
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeError(w, "invalid email address", http.StatusBadRequest)
		return
	}

	resp, refreshToken, err := h.svc.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			writeError(w, "email already in use", http.StatusConflict)
		case errors.Is(err, ErrUsernameTaken):
			writeError(w, "username already in use", http.StatusConflict)
		case errors.Is(err, ErrInvalidRole):
			writeError(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrWeakPassword):
			writeError(w, err.Error(), http.StatusBadRequest)
		default:
			slog.Error("register failed", "err", err)
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	setRefreshCookie(w, refreshToken)
	writeJSON(w, http.StatusCreated, resp)
}

// Login godoc
// POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.loginLimiter.Allow() {
		writeError(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	resp, refreshToken, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCreds) {
			writeError(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		slog.Error("login failed", "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	setRefreshCookie(w, refreshToken)
	writeJSON(w, http.StatusOK, resp)
}

// Refresh godoc
// POST /auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	resp, newRefreshToken, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	setRefreshCookie(w, newRefreshToken)
	writeJSON(w, http.StatusOK, resp)
}

// Logout godoc
// POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxKeyUserID).(string)
	if !ok || userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	_ = h.svc.Logout(r.Context(), userID)

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // set true in production (HTTPS only)
	})
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

type ctxKey string

const ctxKeyUserID ctxKey = "userID"

func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // set true in production
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writeJSON encode failed", "err", err)
	}
}

func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}
