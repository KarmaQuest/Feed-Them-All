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

	"github.com/KarmaQuest/feed-them-all/internal/admin"
	"github.com/KarmaQuest/feed-them-all/internal/auth"
	"github.com/KarmaQuest/feed-them-all/internal/config"
	"github.com/KarmaQuest/feed-them-all/internal/gamification"
	"github.com/KarmaQuest/feed-them-all/internal/pings"
	"github.com/KarmaQuest/feed-them-all/internal/shop"
	"github.com/KarmaQuest/feed-them-all/internal/users"
	ws "github.com/KarmaQuest/feed-them-all/internal/websocket"
)

func main() {
	// --- Config from environment ---
	cfg := config.Load()

	// --- Database connection pool ---
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
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

	// --- Wire up gamification ---
	gamRepo := gamification.NewRepository(db)
	gamSvc := gamification.NewService(gamRepo)

	// --- Wire up shop ---
	shopRepo := shop.NewRepository(db)
	shopSvc := shop.NewService(shopRepo, cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	shopHandler := shop.NewHandler(shopSvc)
	gamSvc.SetItemGranter(shopSvc) // quest item unlock after each XP award

	// --- Wire up WebSocket hub ---
	hub := ws.NewHub()
	go hub.Run()

	// --- Wire up pings ---
	pingsRepo := pings.NewRepository(db)
	pingsSvc := pings.NewService(pingsRepo)
	pingsSvc.SetBroadcaster(hub) // inject hub for real-time broadcasts
	pingsSvc.SetXPAwarder(gamSvc) // inject gamification for XP awards
	pingsHandler := pings.NewHandler(pingsSvc)

	// --- Wire up users ---
	usersRepo := users.NewRepository(db)
	usersSvc := users.NewService(usersRepo)
	usersSvc.LoadThresholds(ctx) // load level thresholds from DB (fallback to hardcoded)
	usersHandler := users.NewHandler(usersSvc)

	// --- Wire up admin ---
	adminRepo := admin.NewRepository(db)
	adminSvc := admin.NewService(adminRepo)
	adminSvc.SetThresholdReloader(usersSvc) // reload thresholds in memory after admin update
	adminHandler := admin.NewHandler(adminSvc)
	adminMW := admin.NewMiddleware(db)

	// --- Wire up WebSocket handler ---
	wsHandler := ws.NewHandler(hub, cfg.JWTSecret)

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
	if cfg.IsDev() {
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
		r.Patch("/pings/{id}/fed", pingsHandler.MarkFed) // kept for backward compat
		r.Delete("/pings/{id}", pingsHandler.Deactivate)
		r.Post("/pings/{id}/media", pingsHandler.UploadMedia)
		// Feeding history (new — preferred over PATCH /fed)
		r.Post("/pings/{id}/feedings", pingsHandler.AddFeedingEvent)
		// Reports (any authenticated user, including ping creator)
		r.Post("/pings/{id}/report", pingsHandler.Report)
		r.Post("/pings/{id}/reports/{reportID}/vote", pingsHandler.VoteReport)
	})
	// Public read routes
	r.Get("/pings/{id}/media", pingsHandler.ListMedia)
	r.Get("/pings/{id}/reports", pingsHandler.ListReports)
	r.Get("/pings/{id}/feedings", pingsHandler.ListFeedingEvents)
	// WebSocket endpoint (optional JWT via ?token=<JWT>)
	r.Get("/ws", wsHandler.ServeWS)

	// Users & leaderboard (public)
	r.Get("/users/{id}/profile", usersHandler.GetProfile)
	r.Get("/leaderboard", usersHandler.GetLeaderboard)

	// Shop catalogue (public)
	r.Get("/shop/items", shopHandler.GetCatalogue)
	// Stripe webhook (no JWT — Stripe signature verification inside handler)
	r.Post("/shop/webhook", shopHandler.Webhook)

	// Authenticated shop + inventory routes
	r.Group(func(r chi.Router) {
		r.Use(authSvc.Middleware)
		r.Get("/users/me/inventory", shopHandler.GetInventory)
		r.Post("/shop/items/{id}/purchase", shopHandler.Purchase)
	})

	// Admin routes (auth + admin role required)
	r.Group(func(r chi.Router) {
		r.Use(authSvc.Middleware)
		r.Use(adminMW.RequireAdmin)

		// Users
		r.Get("/admin/users", adminHandler.ListUsers)
		r.Post("/admin/users", adminHandler.CreateUser)
		r.Patch("/admin/users/{id}", adminHandler.UpdateUser)
		r.Delete("/admin/users/{id}", adminHandler.DeleteUser)

		// XP Actions
		r.Get("/admin/xp-actions", adminHandler.ListXPActions)
		r.Post("/admin/xp-actions", adminHandler.CreateXPAction)
		r.Put("/admin/xp-actions/{action}", adminHandler.UpdateXPAction)

		// Level thresholds
		r.Get("/admin/level-thresholds", adminHandler.ListLevelThresholds)
		r.Put("/admin/level-thresholds", adminHandler.ReplaceAllThresholds)

		// Badges
		r.Get("/admin/badges", adminHandler.ListBadges)
		r.Post("/admin/badges", adminHandler.CreateBadge)
		r.Put("/admin/badges/{id}", adminHandler.UpdateBadge)
		r.Delete("/admin/badges/{id}", adminHandler.DeleteBadge)

		// Shop items
		r.Get("/admin/shop-items", adminHandler.ListShopItems)
		r.Post("/admin/shop-items", adminHandler.CreateShopItem)
		r.Put("/admin/shop-items/{id}", adminHandler.UpdateShopItem)
		r.Delete("/admin/shop-items/{id}", adminHandler.DeleteShopItem)

		// Pings moderation
		r.Get("/admin/pings", adminHandler.ListPingsAdmin)
		r.Post("/admin/pings", adminHandler.CreatePingAdmin)
		r.Delete("/admin/pings/{id}", adminHandler.ForceDeactivatePing)
		r.Get("/admin/pings/{id}/comments", adminHandler.ListCommentsAdmin)
		r.Post("/admin/pings/{id}/comments", adminHandler.CreateCommentAdmin)
		r.Get("/admin/pings/{id}/feedings", adminHandler.ListFeedingEventsAdmin)
		r.Post("/admin/pings/{id}/feedings", adminHandler.CreateFeedingEventAdmin)

		// Comments moderation
		r.Patch("/admin/comments/{id}", adminHandler.UpdateComment)
		r.Delete("/admin/comments/{id}", adminHandler.DeleteComment)

		// Feeding events moderation
		r.Patch("/admin/feedings/{id}", adminHandler.UpdateFeedingEvent)
		r.Delete("/admin/feedings/{id}", adminHandler.DeleteFeedingEvent)
	})

	// Serve uploaded files
	r.Handle("/uploads/*", http.StripPrefix("/uploads", http.FileServer(http.Dir(cfg.UploadDir))))

	// --- Start server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

// corsMiddleware restricts cross-origin requests to known frontend origins.
func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:5173":       true, // Vite dev server (port par défaut)
		"http://localhost:5174":       true, // Vite dev server (port alternatif)
		"http://localhost:5175":       true, // Vite dev server (port alternatif)
		"http://localhost:5176":       true, // Vite dev server (port alternatif)
		"http://localhost:8080":       true, // same-origin (test pages served by Go)
		"https://feedthemall.org":     true, // production
		"https://www.feedthemall.org": true, // production with www
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
