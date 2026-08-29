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
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/KarmaQuest/feed-them-all/internal/auth"
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

	ListComments(ctx context.Context, pingID string) ([]AdminComment, error)
	CreateComment(ctx context.Context, pingID, authorID string, req CreateCommentAdminRequest) (AdminComment, error)
	UpdateComment(ctx context.Context, commentID string, req UpdateCommentRequest) error
	DeleteComment(ctx context.Context, commentID string) error

	ListFeedingEventsAdmin(ctx context.Context, pingID string) ([]AdminFeedingEvent, error)
	CreateFeedingEventAdmin(ctx context.Context, pingID, fedBy string, req CreateFeedingEventAdminRequest) (AdminFeedingEvent, error)
	UpdateFeedingEvent(ctx context.Context, eventID string, req UpdateFeedingEventRequest) error
	DeleteFeedingEvent(ctx context.Context, eventID string) error

	// ── Sprites ────────────────────────────────────────────────────────────────
	ListSprites(ctx context.Context) ([]SpriteEntry, error)
	UploadSprite(ctx context.Context, filePath, destDir string, destName ...string) (UploadSpriteResponse, error)
	DeleteSprite(ctx context.Context, fullPath string) error
	UploadShopItemSprite(ctx context.Context, itemSlug string, filePath string) (UploadSpriteResponse, error)
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
	// S5: cap page to prevent excessive offsets (OWASP A04)
	if page > 100 {
		page = 100
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

// ─── Comments (Moderation) ────────────────────────────────────────────────────

// ListCommentsAdmin handles GET /admin/pings/:id/comments
func (h *Handler) ListCommentsAdmin(w http.ResponseWriter, r *http.Request) {
	pingID := chi.URLParam(r, "id")
	comments, err := h.svc.ListComments(r.Context(), pingID)
	if err != nil {
		writeError(w, "failed to list comments", http.StatusInternalServerError)
		return
	}
	writeJSON(w, comments, http.StatusOK)
}

// CreateCommentAdmin handles POST /admin/pings/:id/comments
func (h *Handler) CreateCommentAdmin(w http.ResponseWriter, r *http.Request) {
	pingID := chi.URLParam(r, "id")
	adminID := auth.UserIDFromContext(r.Context())
	var req CreateCommentAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	comment, err := h.svc.CreateComment(r.Context(), pingID, adminID, req)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, comment, http.StatusCreated)
}

// UpdateComment handles PATCH /admin/comments/:id
func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateComment(r.Context(), commentID, req); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "comment not found", http.StatusNotFound)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteComment handles DELETE /admin/comments/:id
func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	if err := h.svc.DeleteComment(r.Context(), commentID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "comment not found", http.StatusNotFound)
			return
		}
		writeError(w, "failed to delete comment", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Feeding Events (Moderation) ──────────────────────────────────────────────

// ListFeedingEventsAdmin handles GET /admin/pings/:id/feedings
func (h *Handler) ListFeedingEventsAdmin(w http.ResponseWriter, r *http.Request) {
	pingID := chi.URLParam(r, "id")
	events, err := h.svc.ListFeedingEventsAdmin(r.Context(), pingID)
	if err != nil {
		writeError(w, "failed to list feeding events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, events, http.StatusOK)
}

// CreateFeedingEventAdmin handles POST /admin/pings/:id/feedings
func (h *Handler) CreateFeedingEventAdmin(w http.ResponseWriter, r *http.Request) {
	pingID := chi.URLParam(r, "id")
	adminID := auth.UserIDFromContext(r.Context())
	var req CreateFeedingEventAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	event, err := h.svc.CreateFeedingEventAdmin(r.Context(), pingID, adminID, req)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, event, http.StatusCreated)
}

// UpdateFeedingEvent handles PATCH /admin/feedings/:id
func (h *Handler) UpdateFeedingEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	var req UpdateFeedingEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateFeedingEvent(r.Context(), eventID, req); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "feeding event not found", http.StatusNotFound)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteFeedingEvent handles DELETE /admin/feedings/:id
func (h *Handler) DeleteFeedingEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	if err := h.svc.DeleteFeedingEvent(r.Context(), eventID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "feeding event not found", http.StatusNotFound)
			return
		}
		writeError(w, "failed to delete feeding event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

// ─── Sprites ───────────────────────────────────────────────────────────────────

// ListSprites handles GET /admin/sprites
func (h *Handler) ListSprites(w http.ResponseWriter, r *http.Request) {
	entries, err := h.svc.ListSprites(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, entries, http.StatusOK)
}

// UploadSprite handles POST /admin/sprites/upload
// Expects multipart/form-data with fields: file (the PNG), path (destination relative dir).
func (h *Handler) UploadSprite(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB limit
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32 MB memory
		writeError(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := validateSpriteFile(file, header); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	destDir := r.FormValue("path")
	if destDir == "" {
		writeError(w, "missing path field", http.StatusBadRequest)
		return
	}

	destName := r.FormValue("filename") // optional — if empty, preserve original name

	// Save to temp file so the service can move it.
	tmpFile, err := os.CreateTemp("", "sprite-*.png")
	if err != nil {
		writeError(w, "failed to create temp file", http.StatusInternalServerError)
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		writeError(w, "failed to save temp file", http.StatusInternalServerError)
		return
	}
	tmpFile.Close()

	resp, err := h.svc.UploadSprite(r.Context(), tmpPath, destDir, destName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp, http.StatusCreated)
}

// DeleteSprite handles DELETE /admin/sprites?path=relative/path/to/file.png
func (h *Handler) DeleteSprite(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, "missing path query param", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteSprite(r.Context(), path); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadShopItemSprite handles POST /admin/shop-items/{id}/sprite
// Expects multipart/form-data with field: file (the PNG).
func (h *Handler) UploadShopItemSprite(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")

	// Resolve slug from shop item ID.
	items, err := h.svc.ListShopItems(r.Context())
	if err != nil {
		writeError(w, "failed to list shop items", http.StatusInternalServerError)
		return
	}
	var slug string
	for _, item := range items {
		if item.ID == itemID {
			slug = item.Slug
			break
		}
	}
	if slug == "" {
		writeError(w, "shop item not found", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := validateSpriteFile(file, header); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	tmpFile, err := os.CreateTemp("", "sprite-*.png")
	if err != nil {
		writeError(w, "failed to create temp file", http.StatusInternalServerError)
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		writeError(w, "failed to save temp file", http.StatusInternalServerError)
		return
	}
	tmpFile.Close()

	resp, err := h.svc.UploadShopItemSprite(r.Context(), slug, tmpPath)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp, http.StatusCreated)
}

// validateSpriteFile checks MIME type and size.
func validateSpriteFile(file multipart.File, header *multipart.FileHeader) error {
	if header.Size > 5*1024*1024 {
		return errors.New("file too large: max 5 MB")
	}
	buf := make([]byte, 8)
	if _, err := file.Read(buf); err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	// PNG magic bytes
	if len(buf) < 8 || buf[0] != 0x89 || buf[1] != 'P' || buf[2] != 'N' || buf[3] != 'G' {
		return errors.New("only PNG files are accepted")
	}
	return nil
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
