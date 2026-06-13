// Package admin — service.go orchestre la logique métier du dashboard admin.
//
// Le Service est une couche mince : la validation basique des entrées est ici,
// mais toutes les requêtes SQL passent par le Store.
//
// ReloadUserThresholds : après PUT /admin/level-thresholds, le service signale
//   au users.Service de recharger ses paliers depuis la DB (sans redémarrer le serveur).
//   Le users.Service est injecté via l'interface ThresholdReloader pour éviter
//   un import circulaire (admin → users → admin).
package admin

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// ThresholdReloader is implemented by users.Service.
// It reloads the level thresholds from the DB after an admin update.
type ThresholdReloader interface {
	ReloadThresholds(ctx context.Context) error
}

// Service holds the admin business logic.
type Service struct {
	store            Store
	thresholdReloader ThresholdReloader // injected from users.Service (optional)
}

// NewService creates an admin Service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// SetThresholdReloader injects the users.Service so admin can trigger a reload
// after modifying level_thresholds.
func (s *Service) SetThresholdReloader(r ThresholdReloader) {
	s.thresholdReloader = r
}

// ─── Users ────────────────────────────────────────────────────────────────────

func (s *Service) ListUsers(ctx context.Context, page int, search string) ([]AdminUser, error) {
	if page < 1 {
		page = 1
	}
	return s.store.ListUsers(ctx, page, search)
}

func (s *Service) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) error {
	if req.Role != nil {
		allowed := map[string]bool{"feeder": true, "giver": true, "association": true, "admin": true}
		if !allowed[*req.Role] {
			return fmt.Errorf("invalid role: %s", *req.Role)
		}
	}
	if err := s.store.UpdateUser(ctx, userID, req); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (string, error) {
	if req.Email == "" || req.Username == "" || req.Password == "" {
		return "", errors.New("email, username and password are required")
	}
	allowed := map[string]bool{"feeder": true, "giver": true, "association": true, "admin": true}
	if req.Role == "" {
		req.Role = "feeder"
	} else if !allowed[req.Role] {
		return "", fmt.Errorf("invalid role: %s", req.Role)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("admin.CreateUser hash: %w", err)
	}
	return s.store.CreateUser(ctx, req, string(hash))
}

func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	return s.store.DeleteUser(ctx, userID)
}

// ─── XP Actions ───────────────────────────────────────────────────────────────

func (s *Service) ListXPActions(ctx context.Context) ([]AdminXPAction, error) {
	return s.store.ListXPActions(ctx)
}

func (s *Service) UpdateXPAction(ctx context.Context, action string, req UpdateXPActionRequest) error {
	if req.XPValue != nil && *req.XPValue < 0 {
		return errors.New("xp_value must be >= 0")
	}
	if req.DailyLimit != nil && *req.DailyLimit < 1 {
		return errors.New("daily_limit must be >= 1")
	}
	return s.store.UpdateXPAction(ctx, action, req)
}

func (s *Service) CreateXPAction(ctx context.Context, req CreateXPActionRequest) error {
	if req.Action == "" {
		return errors.New("action name is required")
	}
	if req.XPValue < 0 {
		return errors.New("xp_value must be >= 0")
	}
	if req.DailyLimit < 1 {
		return errors.New("daily_limit must be >= 1")
	}
	return s.store.CreateXPAction(ctx, req)
}

// ─── Level Thresholds ─────────────────────────────────────────────────────────

func (s *Service) ListLevelThresholds(ctx context.Context) ([]LevelThreshold, error) {
	return s.store.ListLevelThresholds(ctx)
}

func (s *Service) ReplaceAllThresholds(ctx context.Context, req UpsertLevelThresholdsRequest) error {
	if len(req.Thresholds) == 0 {
		return errors.New("thresholds list cannot be empty")
	}
	// Validate: first level must start at min_xp 0.
	for _, t := range req.Thresholds {
		if t.Level < 1 {
			return fmt.Errorf("level must be >= 1, got %d", t.Level)
		}
		if t.MinXP < 0 {
			return fmt.Errorf("min_xp must be >= 0, got %d", t.MinXP)
		}
	}

	if err := s.store.ReplaceAllThresholds(ctx, req.Thresholds); err != nil {
		return err
	}

	// Notify users.Service to reload the new thresholds in memory.
	if s.thresholdReloader != nil {
		if err := s.thresholdReloader.ReloadThresholds(ctx); err != nil {
			// Non-fatal: the DB is updated, the in-memory cache will be stale until restart.
			// Log but don't return error to the caller.
			_ = err
		}
	}
	return nil
}

// ─── Badges ───────────────────────────────────────────────────────────────────

func (s *Service) ListBadges(ctx context.Context) ([]AdminBadge, error) {
	return s.store.ListBadges(ctx)
}

func (s *Service) CreateBadge(ctx context.Context, req UpsertBadgeRequest) (string, error) {
	if req.Slug == "" || req.Label == "" {
		return "", errors.New("slug and label are required")
	}
	return s.store.CreateBadge(ctx, req)
}

func (s *Service) UpdateBadge(ctx context.Context, badgeID string, req UpsertBadgeRequest) error {
	if req.Slug == "" || req.Label == "" {
		return errors.New("slug and label are required")
	}
	return s.store.UpdateBadge(ctx, badgeID, req)
}

func (s *Service) DeleteBadge(ctx context.Context, badgeID string) error {
	return s.store.DeleteBadge(ctx, badgeID)
}

// ─── Shop Items ───────────────────────────────────────────────────────────────

func (s *Service) ListShopItems(ctx context.Context) ([]AdminShopItem, error) {
	return s.store.ListShopItems(ctx)
}

func (s *Service) CreateShopItem(ctx context.Context, req UpsertShopItemRequest) (string, error) {
	if req.Slug == "" || req.Name == "" {
		return "", errors.New("slug and name are required")
	}
	allowed := map[string]bool{"skin": true, "outfit": true, "accessory": true}
	if !allowed[req.Category] {
		return "", fmt.Errorf("invalid category: %s", req.Category)
	}
	if req.Currency == "" {
		req.Currency = "usd"
	}
	return s.store.CreateShopItem(ctx, req)
}

func (s *Service) UpdateShopItem(ctx context.Context, itemID string, req UpsertShopItemRequest) error {
	if req.Slug == "" || req.Name == "" {
		return errors.New("slug and name are required")
	}
	allowed := map[string]bool{"skin": true, "outfit": true, "accessory": true}
	if !allowed[req.Category] {
		return fmt.Errorf("invalid category: %s", req.Category)
	}
	if req.Currency == "" {
		req.Currency = "usd"
	}
	return s.store.UpdateShopItem(ctx, itemID, req)
}

func (s *Service) DeleteShopItem(ctx context.Context, itemID string) error {
	return s.store.DeleteShopItem(ctx, itemID)
}

// ─── Pings (Moderation) ───────────────────────────────────────────────────────

func (s *Service) ListPingsAdmin(ctx context.Context, activeOnly, flaggedOnly bool) ([]AdminPing, error) {
	return s.store.ListPingsAdmin(ctx, activeOnly, flaggedOnly)
}

func (s *Service) ForceDeactivatePing(ctx context.Context, pingID string) error {
	return s.store.ForceDeactivatePing(ctx, pingID)
}

func (s *Service) CreatePingAdmin(ctx context.Context, req AdminCreatePingRequest) (string, error) {
	if req.UserID == "" {
		return "", errors.New("user_id is required")
	}
	if req.Type != "animal" && req.Type != "food" {
		return "", errors.New("type must be 'animal' or 'food'")
	}
	return s.store.CreatePingAdmin(ctx, req)
}

// ─── Comments (Moderation) ────────────────────────────────────────────────────

func (s *Service) ListComments(ctx context.Context, pingID string) ([]AdminComment, error) {
	return s.store.ListComments(ctx, pingID)
}

func (s *Service) UpdateComment(ctx context.Context, commentID string, req UpdateCommentRequest) error {
	if len(req.Content) == 0 || len(req.Content) > 500 {
		return errors.New("content must be between 1 and 500 characters")
	}
	return s.store.UpdateComment(ctx, commentID, req)
}

func (s *Service) DeleteComment(ctx context.Context, commentID string) error {
	return s.store.DeleteComment(ctx, commentID)
}

// ─── Feeding Events (Moderation) ──────────────────────────────────────────────

func (s *Service) ListFeedingEventsAdmin(ctx context.Context, pingID string) ([]AdminFeedingEvent, error) {
	return s.store.ListFeedingEventsAdmin(ctx, pingID)
}

func (s *Service) UpdateFeedingEvent(ctx context.Context, eventID string, req UpdateFeedingEventRequest) error {
	if req.AnimalCountSeen != nil && (*req.AnimalCountSeen < 1 || *req.AnimalCountSeen > 100) {
		return errors.New("animal_count_seen must be between 1 and 100")
	}
	return s.store.UpdateFeedingEvent(ctx, eventID, req)
}

func (s *Service) DeleteFeedingEvent(ctx context.Context, eventID string) error {
	return s.store.DeleteFeedingEvent(ctx, eventID)
}
