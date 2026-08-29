// Package config centralise la lecture des variables d'environnement.
// Toutes les valeurs de configuration sont lues ici et validées au démarrage.
// Les packages internes ne doivent pas appeler os.Getenv directement —
// ils reçoivent leurs secrets en paramètre de constructeur.
package config

import (
	"log/slog"
	"os"
)

// Config holds all runtime configuration read from environment variables.
type Config struct {
	DatabaseURL         string
	Port                string
	Env                 string
	JWTSecret           string
	JWTRefreshSecret    string
	StripeSecretKey     string
	StripeWebhookSecret string
	UploadDir           string
	SpritesDir          string
}

// IsDev reports whether the server is running in development mode.
func (c Config) IsDev() bool { return c.Env == "development" }

// Load reads all configuration from environment variables.
// It exits the process immediately if a required variable is missing
// or if insecure secrets are detected in production.
func Load() Config {
	cfg := Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		Port:                os.Getenv("PORT"),
		Env:                 os.Getenv("ENV"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTRefreshSecret:    os.Getenv("JWT_REFRESH_SECRET"),
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		UploadDir:           os.Getenv("UPLOAD_DIR"),
		SpritesDir:          os.Getenv("SPRITES_DIR"),
	}

	// Required
	if cfg.DatabaseURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	// Defaults
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.Env == "" {
		cfg.Env = "development"
	}
	if cfg.UploadDir == "" {
		cfg.UploadDir = "./uploads"
	}
	if cfg.SpritesDir == "" {
		cfg.SpritesDir = "./sprites"
	}

	// S7 — Block insecure default secrets in production (OWASP A02)
	if cfg.Env == "production" {
		insecure := map[string]bool{
			"":                             true,
			"dev-secret-change-in-prod":    true,
			"dev-refresh-secret-change-in-prod": true,
		}
		if insecure[cfg.JWTSecret] || insecure[cfg.JWTRefreshSecret] {
			slog.Error("insecure JWT secrets detected — refusing to start in production")
			os.Exit(1)
		}
	}

	return cfg
}
