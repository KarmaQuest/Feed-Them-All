// Package auth — middleware.go protège les routes qui nécessitent une connexion.
//
// Un middleware HTTP est une fonction qui s'exécute AVANT le handler de la route.
// Ici, il vérifie que chaque requête vers une route protégée contient un JWT valide.
//
// Comment ça fonctionne :
//   1. Le client envoie un header : Authorization: Bearer <access_token>
//   2. Le middleware extrait le token après "Bearer "
//   3. Il le valide (signature correcte + non expiré) via Service.ValidateAccessToken
//   4. Si valide : il injecte l'ID de l'utilisateur dans le contexte de la requête
//      → les handlers suivants peuvent récupérer cet ID via UserIDFromContext(ctx)
//   5. Si invalide ou absent : il répond immédiatement 401 Unauthorized
//      et la requête n'atteint jamais le handler final.
//
// Utilisation :
//   r.Group(func(r chi.Router) {
//       r.Use(authSvc.Middleware)  // toutes les routes dans ce groupe sont protégées
//       r.Post("/auth/logout", authHandler.Logout)
//   })
package auth

import (
	"context"
	"net/http"
	"strings"
)

// Middleware validates the Bearer token and injects the userID into context.
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := s.ValidateAccessToken(tokenStr)
		if err != nil {
			writeError(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalMiddleware tries to extract a Bearer token if present, but does not block the request if absent.
// Use for routes that are public but behave differently when authenticated (e.g. private profiles).
func (s *Service) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if userID, err := s.ValidateAccessToken(tokenStr); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), ctxKeyUserID, userID))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// UserIDFromContext extracts the authenticated user ID from context.
// Returns empty string if not present.
func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyUserID).(string)
	return id
}

// NewContextWithUserID returns a copy of ctx with the given userID injected.
// Intended for use in tests outside the auth package to simulate an authenticated request.
func NewContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}
