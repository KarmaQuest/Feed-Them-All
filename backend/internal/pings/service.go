// Package pings — service.go contient toute la logique métier des pings.
//
// Le Service orchestre les règles business entre le Handler (HTTP) et le Store (DB).
// Il ne contient aucun SQL, et ne connaît pas le format JSON.
//
// Règles métier appliquées ici :
//   - Type d'un ping : uniquement "animal" ou "food"
//   - Coordonnées : lat doit être entre -90 et 90, lon entre -180 et 180
//   - Rayon de recherche : minimum 10 m, maximum 10 000 m, défaut 500 m
//   - Les coordonnées retournées publiquement sont arrondies à 3 décimales (~100 m)
//     pour protéger la vie privée des Feeders
//   - Seul le créateur d'un ping peut le désactiver (ErrNotOwner sinon)
//   - Upload : validation du type MIME (lecture des 512 premiers octets),
//     taille max 10 Mo, seuls JPEG et PNG sont acceptés
//   - Signalement : reason doit être dans l'enum défini (ErrInvalidReason sinon)
//     Un utilisateur ne peut signaler qu'une fois le même ping (ErrAlreadyReported)
//   - Vote : value doit être "up" ou "down" (ErrInvalidVote sinon)
//     Un utilisateur ne peut voter qu'une fois par signalement (ErrAlreadyVoted)
//     Le signalement doit exister (ErrNotFound sinon)
//
// Broadcaster (WebSocket) :
//   Après chaque mutation réussie (Create, Confirm, MarkFed, Deactivate), le Service
//   notifie optionnellement un Broadcaster (le Hub WebSocket) pour diffuser l'événement
//   en temps réel aux clients connectés. Si aucun Broadcaster n'est injecté (nil),
//   le comportement est identique — la mutation réussit sans broadcast.
//
// XPAwarder (Gamification) :
//   Après chaque action récompensée (Create → signal_animal, Confirm → confirm_presence,
//   MarkFed → feed, SaveMedia → upload_photo), le Service appelle xpAwarder.AwardXP()
//   dans une goroutine. Les erreurs XP sont loggées mais ne bloquent jamais la réponse HTTP.
package pings

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors returned by the Service.
// The Handler maps these to the appropriate HTTP status codes.
var (
	// ErrInvalidType is returned when the ping type is not "animal" or "food".
	ErrInvalidType = errors.New("type must be 'animal' or 'food'")

	// ErrInvalidCoords is returned when lat or lon are out of valid GPS range.
	ErrInvalidCoords = errors.New("invalid coordinates: lat must be [-90,90], lon must be [-180,180]")

	// ErrNotFound is returned when the requested ping does not exist or is inactive.
	ErrNotFound = errors.New("ping not found")

	// ErrNotOwner is returned when a user tries to delete a ping they did not create.
	ErrNotOwner = errors.New("you are not the owner of this ping")

	// ErrInvalidMedia is returned when the uploaded file is not a JPEG or PNG.
	ErrInvalidMedia = errors.New("only JPEG and PNG files are accepted (max 10 MB)")

	// ErrInvalidReason is returned when the report reason is not in the allowed enum.
	ErrInvalidReason = errors.New("reason must be one of: wrong_location, animal_gone, duplicate, inappropriate")

	// ErrAlreadyReported is returned when the user already filed a report on this ping.
	ErrAlreadyReported = errors.New("you already reported this ping")

	// ErrInvalidVote is returned when the vote value is not "up" or "down".
	ErrInvalidVote = errors.New("value must be 'up' or 'down'")
)

const (
	defaultRadius = 500.0    // metres
	maxRadius     = 10000.0  // metres
	minRadius     = 10.0     // metres
	maxUploadSize = 10 << 20 // 10 MB in bytes
)

// Broadcaster is the interface implemented by the WebSocket hub.
// The pings package depends on this interface (not on the websocket package directly)
// to avoid circular imports. The websocket package imports pings.Ping — pings must not
// import websocket.
type Broadcaster interface {
	BroadcastPingCreated(p Ping)
	BroadcastPingUpdated(p Ping)
}

// XPAwarder is the interface implemented by the gamification service.
// Called after each XP-eligible action. Defined here to avoid circular imports:
// gamification imports nothing from pings; pings imports nothing from gamification.
type XPAwarder interface {
	AwardXP(ctx context.Context, userID, action string) error
}

// Service holds the business logic for pings.
type Service struct {
	store       Store
	uploadDir   string      // local directory where uploaded files are saved
	broadcaster Broadcaster // optional — nil if WebSocket is not wired
	xpAwarder   XPAwarder   // optional — nil if gamification is not wired
}

// NewService creates a Service. uploadDir is read from the UPLOAD_DIR env var,
// defaulting to "./uploads" if not set.
func NewService(store Store) *Service {
	dir := os.Getenv("UPLOAD_DIR")
	if dir == "" {
		dir = "./uploads"
	}
	return &Service{store: store, uploadDir: dir}
}

// SetBroadcaster injects the WebSocket hub into the service.
// Call this after NewService, before the server starts.
// Passing nil disables broadcasts (useful in tests).
func (s *Service) SetBroadcaster(b Broadcaster) {
	s.broadcaster = b
}

// SetXPAwarder injects the gamification service into the pings service.
// Call this after NewService, before the server starts.
// Passing nil disables XP awards (useful in tests).
func (s *Service) SetXPAwarder(a XPAwarder) {
	s.xpAwarder = a
}

// tryAwardXP calls xpAwarder.AwardXP in a goroutine so it never blocks the HTTP response.
// Errors are logged but never propagated.
func (s *Service) tryAwardXP(userID, action string) {
	if s.xpAwarder == nil {
		return
	}
	go func() {
		if err := s.xpAwarder.AwardXP(context.Background(), userID, action); err != nil {
			slog.Warn("award_xp failed", "user_id", userID, "action", action, "err", err)
		}
	}()
}

// Create validates the request and creates a new ping.
// Returns the created Ping with coordinates rounded for public display.
func (s *Service) Create(ctx context.Context, userID string, req CreatePingRequest) (Ping, error) {
	// Validate type
	t := strings.ToLower(req.Type)
	if t != "animal" && t != "food" {
		return Ping{}, ErrInvalidType
	}

	// Validate coordinates
	if req.Lat < -90 || req.Lat > 90 || req.Lon < -180 || req.Lon > 180 {
		return Ping{}, ErrInvalidCoords
	}

	ping, err := s.store.Create(ctx, userID, t, req.Lat, req.Lon)
	if err != nil {
		return Ping{}, fmt.Errorf("pings.Service.Create: %w", err)
	}

	// Round coordinates before returning publicly
	ping.Lat = roundCoord(ping.Lat)
	ping.Lon = roundCoord(ping.Lon)

	// Broadcast to WebSocket clients in the ping's area.
	if s.broadcaster != nil {
		s.broadcaster.BroadcastPingCreated(ping)
	}
	// Award XP: signal_animal (fire-and-forget)
	s.tryAwardXP(userID, "signal_animal")
	return ping, nil
}

// ListNearby validates query parameters and returns nearby active pings.
// Radius is clamped to [minRadius, maxRadius]. Coordinates are rounded.
func (s *Service) ListNearby(ctx context.Context, q NearbyQuery) ([]Ping, error) {
	if q.Lat < -90 || q.Lat > 90 || q.Lon < -180 || q.Lon > 180 {
		return nil, ErrInvalidCoords
	}

	// Clamp radius
	if q.Radius <= 0 {
		q.Radius = defaultRadius
	}
	if q.Radius > maxRadius {
		q.Radius = maxRadius
	}
	if q.Radius < minRadius {
		q.Radius = minRadius
	}

	// Validate optional type filter
	if q.Type != "" && q.Type != "animal" && q.Type != "food" {
		return nil, ErrInvalidType
	}

	pings, err := s.store.ListNearby(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("pings.Service.ListNearby: %w", err)
	}

	// Round all coordinates before returning
	for i := range pings {
		pings[i].Lat = roundCoord(pings[i].Lat)
		pings[i].Lon = roundCoord(pings[i].Lon)
	}
	return pings, nil
}

// Confirm marks a ping as "animal still present".
// Anyone can confirm — it just touches updated_at.
func (s *Service) Confirm(ctx context.Context, pingID, userID string) error {
	if err := s.store.Confirm(ctx, pingID); err != nil {
		return fmt.Errorf("pings.Service.Confirm: %w", err)
	}
	s.broadcastUpdated(ctx, pingID)
	s.tryAwardXP(userID, "confirm_presence")
	return nil
}

// MarkFed records that the animal at this ping has been fed (sets fed_at).
func (s *Service) MarkFed(ctx context.Context, pingID, userID string) error {
	if err := s.store.MarkFed(ctx, pingID); err != nil {
		return fmt.Errorf("pings.Service.MarkFed: %w", err)
	}
	s.broadcastUpdated(ctx, pingID)
	s.tryAwardXP(userID, "feed")
	return nil
}

// Deactivate soft-deletes a ping. Only the original creator may do this.
// Returns ErrNotOwner if userID is not the creator, ErrNotFound if ping is gone.
func (s *Service) Deactivate(ctx context.Context, pingID, userID string) error {
	err := s.store.Deactivate(ctx, pingID, userID)
	if err != nil {
		return err // already a sentinel error from the store
	}
	s.broadcastUpdated(ctx, pingID)
	return nil
}

// broadcastUpdated fetches the current state of a ping and broadcasts a "ping_updated" event.
// Errors are logged but never propagate — a broadcast failure must not fail the operation.
func (s *Service) broadcastUpdated(ctx context.Context, pingID string) {
	if s.broadcaster == nil {
		return
	}
	ping, err := s.store.GetByID(ctx, pingID)
	if err != nil {
		// Ping may have been deactivated — GetByID returns error for inactive pings.
		// That's fine: the broadcast is best-effort.
		return
	}
	ping.Lat = roundCoord(ping.Lat)
	ping.Lon = roundCoord(ping.Lon)
	s.broadcaster.BroadcastPingUpdated(ping)
}

// SaveMedia validates and saves an uploaded file for a given ping.
// Accepted formats: JPEG, PNG. Max size: 10 MB.
// The file is saved to uploadDir/<pingID>/<uuid>.<ext> and the path is stored in DB.
// Returns the public file path (relative to uploadDir) on success.
func (s *Service) SaveMedia(ctx context.Context, pingID, userID string, data io.Reader, size int64) (string, error) {
	if size > maxUploadSize {
		return "", ErrInvalidMedia
	}

	// Read first 512 bytes to detect MIME type — never trust the Content-Type header
	buf := make([]byte, 512)
	n, err := io.ReadFull(data, buf[:])
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("pings.SaveMedia read header: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])

	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	default:
		return "", ErrInvalidMedia
	}

	// Build destination path: uploads/<pingID>/<uuid><ext>
	dir := filepath.Join(s.uploadDir, pingID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("pings.SaveMedia mkdir: %w", err)
	}

	filename := uuid.New().String() + ext
	dest := filepath.Join(dir, filename)

	// Write: prepend the 512 bytes already read + the rest of the stream
	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("pings.SaveMedia create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(buf[:n]); err != nil {
		return "", fmt.Errorf("pings.SaveMedia write header: %w", err)
	}
	if _, err := io.Copy(f, data); err != nil {
		return "", fmt.Errorf("pings.SaveMedia write body: %w", err)
	}

	// Record the path in DB
	relativePath := filepath.Join(pingID, filename)
	if err := s.store.AddMedia(ctx, pingID, relativePath); err != nil {
		// Best-effort cleanup: remove the file if DB insert fails
		_ = os.Remove(dest)
		return "", fmt.Errorf("pings.SaveMedia db: %w", err)
	}

	// Award XP: upload_photo (fire-and-forget)
	s.tryAwardXP(userID, "upload_photo")

	return relativePath, nil
}

// GetMedia returns the list of media paths for a ping.
func (s *Service) GetMedia(ctx context.Context, pingID string) ([]string, error) {
	return s.store.ListMedia(ctx, pingID)
}

// Report validates the reason enum and files a report for the given ping.
// Any authenticated user (including the ping creator) may report.
// Returns ErrAlreadyReported if the user already reported this ping.
func (s *Service) Report(ctx context.Context, pingID, userID string, req CreateReportRequest) (PingReport, error) {
	validReasons := map[string]bool{
		"wrong_location": true,
		"animal_gone":    true,
		"duplicate":      true,
		"inappropriate":  true,
	}
	if !validReasons[req.Reason] {
		return PingReport{}, ErrInvalidReason
	}

	rp, err := s.store.Report(ctx, pingID, userID, req.Reason, req.Comment)
	if err != nil {
		return PingReport{}, err // ErrAlreadyReported propagates as-is
	}
	return rp, nil
}

// ListReports returns all reports for a ping, with scores, ordered by score desc.
func (s *Service) ListReports(ctx context.Context, pingID string) ([]PingReport, error) {
	reports, err := s.store.ListReports(ctx, pingID)
	if err != nil {
		return nil, fmt.Errorf("pings.Service.ListReports: %w", err)
	}
	return reports, nil
}

// VoteReport validates the vote value and casts the vote.
// The report must exist. Any authenticated user (including the report author or ping creator) may vote.
// Returns ErrNotFound if the report does not exist.
// Returns ErrAlreadyVoted if the user already voted on this report.
func (s *Service) VoteReport(ctx context.Context, reportID, userID string, req VoteReportRequest) error {
	if req.Value != "up" && req.Value != "down" {
		return ErrInvalidVote
	}

	// Verify the report exists before voting
	if _, err := s.store.GetReport(ctx, reportID); err != nil {
		return ErrNotFound
	}

	return s.store.VoteReport(ctx, reportID, userID, req.Value)
}

// mimeExtensions is used for validation display only.
var _ = mime.TypeByExtension

// unusedNow suppresses the "declared and not used" error for the now var in repository.go
var _ = time.Now
