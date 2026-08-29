// Package admin — middleware.go fournit le middleware RequireAdmin.
//
// RequireAdmin s'insère APRÈS auth.Middleware :
//   auth.Middleware vérifie le JWT et injecte userID dans le contexte.
//   RequireAdmin lit le userID, interroge la DB pour confirmer role='admin' et is_banned=false.
//   Si l'une ou l'autre condition échoue → 403 Forbidden.
//
// Usage dans main.go :
//   adminMW := admin.NewMiddleware(db)
//   r.Group(func(r chi.Router) {
//       r.Use(authSvc.Middleware)
//       r.Use(adminMW.RequireAdmin)
//       // routes /admin/...
//   })
package admin

import (
	"net/http"

	"github.com/KarmaQuest/feed-them-all/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Middleware holds the DB pool needed to check the admin role.
type Middleware struct {
	db *pgxpool.Pool
}

// NewMiddleware creates an admin Middleware.
func NewMiddleware(db *pgxpool.Pool) *Middleware {
	return &Middleware{db: db}
}

// RequireAdmin verifies that the authenticated user has role='admin' and is not banned.
// Must be used after auth.Middleware (depends on userID being in context).
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserIDFromContext(r.Context())
		if userID == "" {
			writeError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var role string
		var isBanned bool
		err := m.db.QueryRow(r.Context(),
			"SELECT role, is_banned FROM users WHERE id = $1",
			userID,
		).Scan(&role, &isBanned)
		if err != nil || role != "admin" || isBanned {
			writeError(w, "forbidden: admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
