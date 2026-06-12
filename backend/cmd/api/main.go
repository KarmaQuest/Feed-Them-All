// Package main — Point d'entrée du serveur FeedThemAll.
//
// Ce fichier démarre tout le backend :
//   - Il lit les variables d'environnement (DATABASE_URL, JWT_SECRET, PORT...)
//   - Il ouvre un pool de connexions vers PostgreSQL
//   - Il instancie les couches auth (repository → service → handler)
//   - Il configure le routeur HTTP (chi) avec les middlewares globaux
//     (logs, récupération de panics, timeout, CORS)
//   - Il expose les routes publiques (/auth/register, /auth/login, /auth/refresh)
//     et protégées (/auth/logout nécessite un JWT valide)
//   - En mode development, il sert les pages de test depuis /tests/
//
// Pour lancer le serveur en local :
//   cd backend
//   $env:DATABASE_URL="postgres://fta:fta@localhost:5432/feedthemall?sslmode=disable"
//   $env:JWT_SECRET="une-clé-secrète"
//   $env:JWT_REFRESH_SECRET="une-autre-clé-secrète"
//   $env:ENV="development"
//   go run ./cmd/api
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KarmaQuest/feed-them-all/internal/auth"
	"github.com/KarmaQuest/feed-them-all/internal/pings"
)

func main() {
	// --- Config from environment ---
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// --- Database connection pool ---
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		slog.Error("database ping failed", "err", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	// --- Wire up auth ---
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	// --- Wire up pings ---
	pingsRepo := pings.NewRepository(db)
	pingsSvc := pings.NewService(pingsRepo)
	pingsHandler := pings.NewHandler(pingsSvc)

	// --- Router ---
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS: allow frontend origins
	r.Use(corsMiddleware)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// Test pages — dev only
	if os.Getenv("ENV") == "development" {
		r.Handle("/tests/*", http.StripPrefix("/tests", http.FileServer(http.Dir("./tests"))))
	}

	// Auth routes (public)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Pings routes (GET is public, mutations require JWT)
	r.Get("/pings", pingsHandler.ListNearby)
	r.Group(func(r chi.Router) {
		r.Use(authSvc.Middleware)
		// Auth logout
		r.Post("/auth/logout", authHandler.Logout)
		// Pings mutations
		r.Post("/pings", pingsHandler.Create)
		r.Patch("/pings/{id}/confirm", pingsHandler.Confirm)
		r.Patch("/pings/{id}/fed", pingsHandler.MarkFed)
		r.Delete("/pings/{id}", pingsHandler.Deactivate)
		r.Post("/pings/{id}/media", pingsHandler.UploadMedia)
	})
	// Media listing is public
	r.Get("/pings/{id}/media", pingsHandler.ListMedia)
	// Serve uploaded files
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	r.Handle("/uploads/*", http.StripPrefix("/uploads", http.FileServer(http.Dir(uploadDir))))

	// --- Start server ---
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("server starting", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

// corsMiddleware allows requests from the local frontend in development.
func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:5173": true, // Vite dev server
		"http://localhost:8080": true, // same-origin (test pages served by Go)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("ENV") == "development" {
			origin := r.Header.Get("Origin")
			if allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
