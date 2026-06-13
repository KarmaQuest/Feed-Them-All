// Package pings — handler.go expose les routes HTTP du package pings.
//
// Le Handler reçoit les requêtes HTTP, extrait et valide les paramètres basiques,
// appelle le Service, puis formate la réponse JSON. Aucune logique métier ici.
//
// Routes exposées (toutes dans main.go) :
//
//   GET  /pings?lat=&lon=&radius=&type=  → lister les pings proches (public)
//   POST /pings                          → créer un ping (JWT requis)
//   PATCH /pings/:id/confirm             → confirmer "l'animal est toujours là" (JWT requis)
//   PATCH /pings/:id/fed                 → marquer un animal comme nourri — DEPRECATED, utiliser /feedings
//   DELETE /pings/:id                    → désactiver un ping (JWT requis, propriétaire uniquement)
//   POST /pings/:id/media                → uploader une photo de preuve (JWT requis)
//   GET  /pings/:id/media                → lister les médias d'un ping (public)
//
//   POST /pings/:id/feedings             → enregistrer un nourrissage (JWT requis)
//        Body JSON : { "note": "...", "animal_count_seen": 3 }
//        Retourne : 201 Created + FeedingEvent
//
//   GET  /pings/:id/feedings             → historique des nourrissages (public)
//        Retourne : 200 OK + tableau de FeedingEvent, ordre anti-chronologique
//
//   POST /pings/:id/report               → signaler un ping (JWT requis, tout utilisateur)
//        Body JSON : { "reason": "wrong_location|animal_gone|duplicate|inappropriate", "comment": "..." }
//        Retourne : 201 Created + PingReport
//        Erreur 409 si déjà signalé par cet utilisateur
//        Erreur 400 si reason invalide
//
//   GET  /pings/:id/reports              → lister les signalements d'un ping (public)
//        Retourne : 200 OK + tableau de PingReport avec scores
//
//   POST /pings/:id/reports/:report_id/vote → voter sur un signalement (JWT requis, tout utilisateur)
//        Body JSON : { "value": "up"|"down" }
//        Retourne : 204 No Content
//        Erreur 404 si le signalement n'existe pas
//        Erreur 409 si déjà voté
package pings

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/KarmaQuest/feed-them-all/internal/auth"
)

// Handler wires HTTP routes to the pings service.
type Handler struct {
	svc *Service
}

// NewHandler creates a Handler backed by the given Service.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListNearby handles GET /pings
// Reads lat, lon, radius, type from query parameters and returns matching pings.
func (h *Handler) ListNearby(w http.ResponseWriter, r *http.Request) {
	q := NearbyQuery{}

	// lat and lon are required
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		writeError(w, "lat is required and must be a number", http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if err != nil {
		writeError(w, "lon is required and must be a number", http.StatusBadRequest)
		return
	}
	q.Lat = lat
	q.Lon = lon

	// radius is optional — Service will apply default/clamp
	if raw := r.URL.Query().Get("radius"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.Radius = v
		}
	}

	// type filter is optional
	q.Type = r.URL.Query().Get("type")

	pings, err := h.svc.ListNearby(r.Context(), q)
	if err != nil {
		if errors.Is(err, ErrInvalidCoords) || errors.Is(err, ErrInvalidType) {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("ListNearby failed", "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Return empty array instead of null when no pings found
	if pings == nil {
		pings = []Ping{}
	}
	writeJSON(w, http.StatusOK, pings)
}

// Create handles POST /pings (JWT required).
// Reads type, lat, lon from JSON body and creates a new ping.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreatePingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ping, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidType):
			writeError(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrInvalidCoords):
			writeError(w, err.Error(), http.StatusBadRequest)
		default:
			slog.Error("Create ping failed", "err", err)
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, http.StatusCreated, ping)
}

// Confirm handles PATCH /pings/:id/confirm (JWT required).
// Signals "the animal is still here" by touching updated_at.
func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Confirm(r.Context(), id, userID); err != nil {
		slog.Error("Confirm ping failed", "id", id, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MarkFed handles PATCH /pings/:id/fed (JWT required).
// Records that the animal at this ping has been fed.
func (h *Handler) MarkFed(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.MarkFed(r.Context(), id, userID); err != nil {
		slog.Error("MarkFed failed", "id", id, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Deactivate handles DELETE /pings/:id (JWT required, owner only).
// Soft-deletes the ping (is_active = false). Returns 403 if not the owner.
func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	err := h.svc.Deactivate(r.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotOwner):
			writeError(w, "you are not allowed to delete this ping", http.StatusForbidden)
		case errors.Is(err, ErrNotFound):
			writeError(w, "ping not found", http.StatusNotFound)
		default:
			slog.Error("Deactivate ping failed", "id", id, "err", err)
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadMedia handles POST /pings/:id/media (JWT required).
// Accepts a multipart form with a "file" field (JPEG or PNG, max 10 MB).
// Returns the saved file path on success.
func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")

	// Limit total request size to 10 MB + 1 KB for form overhead
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20+1024)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, "file too large or invalid form (max 10 MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, "missing 'file' field in form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	path, err := h.svc.SaveMedia(r.Context(), id, userID, file, header.Size)
	if err != nil {
		if errors.Is(err, ErrInvalidMedia) {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("SaveMedia failed", "ping_id", id, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"path": path})
}

// ListMedia handles GET /pings/:id/media.
// Returns the list of media paths attached to the given ping.
func (h *Handler) ListMedia(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	paths, err := h.svc.GetMedia(r.Context(), id)
	if err != nil {
		slog.Error("ListMedia failed", "ping_id", id, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if paths == nil {
		paths = []string{}
	}
	writeJSON(w, http.StatusOK, paths)
}

// Report handles POST /pings/:id/report (JWT required).
// Any authenticated user (including the ping creator) may file a report.
// Returns 409 Conflict if the user already reported this ping.
func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	pingID := chi.URLParam(r, "id")

	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rp, err := h.svc.Report(r.Context(), pingID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidReason):
			writeError(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrAlreadyReported):
			writeError(w, err.Error(), http.StatusConflict)
		default:
			slog.Error("Report ping failed", "ping_id", pingID, "err", err)
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, http.StatusCreated, rp)
}

// ListReports handles GET /pings/:id/reports (public).
// Returns all reports for a ping with their vote scores.
func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	pingID := chi.URLParam(r, "id")
	reports, err := h.svc.ListReports(r.Context(), pingID)
	if err != nil {
		slog.Error("ListReports failed", "ping_id", pingID, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if reports == nil {
		reports = []PingReport{}
	}
	writeJSON(w, http.StatusOK, reports)
}

// VoteReport handles POST /pings/:id/reports/:report_id/vote (JWT required).
// Any authenticated user (including the report author or ping creator) may vote.
// Returns 404 if the report does not exist, 409 if already voted.
func (h *Handler) VoteReport(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	reportID := chi.URLParam(r, "reportID")

	var req VoteReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.svc.VoteReport(r.Context(), reportID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidVote):
			writeError(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrNotFound):
			writeError(w, "report not found", http.StatusNotFound)
		default:
			slog.Error("VoteReport failed", "report_id", reportID, "err", err)
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

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

// AddFeedingEvent handles POST /pings/:id/feedings (JWT required).
// Records that the calling user fed the animal(s) at this ping.
// Optionally accepts a note and the number of animals seen.
func (h *Handler) AddFeedingEvent(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	pingID := chi.URLParam(r, "id")

	var req CreateFeedingEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.svc.AddFeedingEvent(r.Context(), pingID, userID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, "ping not found", http.StatusNotFound)
			return
		}
		slog.Error("AddFeedingEvent failed", "ping_id", pingID, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

// ListFeedingEvents handles GET /pings/:id/feedings (public).
// Returns the full feeding history for a ping, most recent first.
func (h *Handler) ListFeedingEvents(w http.ResponseWriter, r *http.Request) {
	pingID := chi.URLParam(r, "id")

	events, err := h.svc.ListFeedingEvents(r.Context(), pingID)
	if err != nil {
		slog.Error("ListFeedingEvents failed", "ping_id", pingID, "err", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, events)
}
