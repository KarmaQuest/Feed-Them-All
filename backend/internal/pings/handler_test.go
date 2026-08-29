// Package pings — handler_test.go teste les handlers HTTP du package pings.
//
// Chaque test utilise httptest.NewRecorder() + httptest.NewRequest() sans serveur réel.
// Le fakeStore garantit l'isolation — aucune DB n'est requise.
//
// Cas couverts :
//   POST /pings          → 201, 400 type invalide, 401 sans token
//   GET  /pings          → 200 avec résultats, 400 coords manquantes
//   PATCH confirm/fed    → 204
//   DELETE               → 204 propriétaire, 403 non-propriétaire, 401 sans token
//   POST /report         → 201, 400 reason invalide, 409 doublon
//   GET  /reports        → 200
//   POST /vote           → 204, 400 valeur invalide, 404 report inexistant
package pings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/KarmaQuest/feed-them-all/internal/auth"
)

func newTestHandler(t *testing.T) (*Handler, *fakeStore) {
	t.Helper()
	os.Setenv("UPLOAD_DIR", t.TempDir())
	store := newFakeStore()
	svc := NewService(store)
	return NewHandler(svc), store
}

// withUser injects a user ID into the request context (simulates JWT middleware).
func withUser(r *http.Request, userID string) *http.Request {
	ctx := auth.NewContextWithUserID(r.Context(), userID)
	return r.WithContext(ctx)
}

// withChi wraps a handler with a chi router to support URL params like {id}.
func withChi(pattern string, h http.HandlerFunc, r *http.Request) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	router.HandleFunc(pattern, h)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
	return rr
}

func postJSON(t *testing.T, h http.HandlerFunc, path string, body any, userID string) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req = withUser(req, userID)
	}
	rr := httptest.NewRecorder()
	h(rr, req)
	return rr
}

// --- POST /pings ---

func TestHandler_Create_Created(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := postJSON(t, h.Create, "/pings", map[string]any{
		"type": "animal",
		"lat":  48.8566,
		"lon":  2.3522,
	}, "user-1")

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var p Ping
	if err := json.NewDecoder(rr.Body).Decode(&p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if p.ID == "" {
		t.Error("expected non-empty ID in response")
	}
}

func TestHandler_Create_InvalidType(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := postJSON(t, h.Create, "/pings", map[string]any{
		"type": "invalid",
		"lat":  48.8566,
		"lon":  2.3522,
	}, "user-1")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := postJSON(t, h.Create, "/pings", map[string]any{
		"type": "animal",
		"lat":  48.8566,
		"lon":  2.3522,
	}, "") // no user ID

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// --- GET /pings ---

func TestHandler_ListNearby_OK(t *testing.T) {
	h, store := newTestHandler(t)

	// Pre-populate a ping
	store.pings["p1"] = Ping{ID: "p1", Type: "animal", Lat: 48.8566, Lon: 2.3522, CreatedBy: "u1", IsActive: true}

	req := httptest.NewRequest(http.MethodGet, "/pings?lat=48.8566&lon=2.3522&radius=500", nil)
	rr := httptest.NewRecorder()
	h.ListNearby(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_ListNearby_MissingLat(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/pings?lon=2.3522", nil)
	rr := httptest.NewRecorder()
	h.ListNearby(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// --- PATCH confirm ---

func TestHandler_Confirm_NoContent(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", Type: "animal", IsActive: true, CreatedBy: "u1"}

	req := httptest.NewRequest(http.MethodPatch, "/pings/p1/confirm", nil)
	req = withUser(req, "u1")
	rr := withChi("/pings/{id}/confirm", h.Confirm, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

// --- PATCH fed ---

func TestHandler_MarkFed_NoContent(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", Type: "animal", IsActive: true, CreatedBy: "u1"}

	req := httptest.NewRequest(http.MethodPatch, "/pings/p1/fed", nil)
	req = withUser(req, "u1")
	rr := withChi("/pings/{id}/fed", h.MarkFed, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

// --- DELETE ---

func TestHandler_Deactivate_Owner(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", Type: "food", IsActive: true, CreatedBy: "owner"}

	req := httptest.NewRequest(http.MethodDelete, "/pings/p1", nil)
	req = withUser(req, "owner")
	rr := withChi("/pings/{id}", h.Deactivate, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_Deactivate_NotOwner(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", Type: "food", IsActive: true, CreatedBy: "owner"}

	req := httptest.NewRequest(http.MethodDelete, "/pings/p1", nil)
	req = withUser(req, "other-user")
	rr := withChi("/pings/{id}", h.Deactivate, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestHandler_Deactivate_Unauthorized(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", IsActive: true, CreatedBy: "owner"}

	req := httptest.NewRequest(http.MethodDelete, "/pings/p1", nil)
	// no user injected
	rr := withChi("/pings/{id}", h.Deactivate, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// --- POST /report ---

func TestHandler_Report_Created(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", Type: "animal", IsActive: true, CreatedBy: "u1"}

	req := httptest.NewRequest(http.MethodPost, "/pings/p1/report",
		bytes.NewBufferString(`{"reason":"animal_gone"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUser(req, "u2")
	rr := withChi("/pings/{id}/report", h.Report, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_Report_InvalidReason(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", IsActive: true, CreatedBy: "u1"}

	req := httptest.NewRequest(http.MethodPost, "/pings/p1/report",
		bytes.NewBufferString(`{"reason":"bad_reason"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUser(req, "u2")
	rr := withChi("/pings/{id}/report", h.Report, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_Report_Conflict(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", IsActive: true, CreatedBy: "u1"}

	// First report
	req := httptest.NewRequest(http.MethodPost, "/pings/p1/report",
		bytes.NewBufferString(`{"reason":"duplicate"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUser(req, "u2")
	withChi("/pings/{id}/report", h.Report, req)

	// Second report from same user → 409
	req2 := httptest.NewRequest(http.MethodPost, "/pings/p1/report",
		bytes.NewBufferString(`{"reason":"duplicate"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2 = withUser(req2, "u2")
	rr := withChi("/pings/{id}/report", h.Report, req2)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rr.Code)
	}
}

// --- GET /reports ---

func TestHandler_ListReports_OK(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", IsActive: true, CreatedBy: "u1"}
	store.reports["r1"] = PingReport{ID: "r1", PingID: "p1", ReportedBy: "u2", Reason: "wrong_location"}

	req := httptest.NewRequest(http.MethodGet, "/pings/p1/reports", nil)
	rr := withChi("/pings/{id}/reports", h.ListReports, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

// --- POST /vote ---

func TestHandler_VoteReport_NoContent(t *testing.T) {
	h, store := newTestHandler(t)

	store.pings["p1"] = Ping{ID: "p1", IsActive: true}
	store.reports["r1"] = PingReport{ID: "r1", PingID: "p1", Reason: "duplicate"}

	req := httptest.NewRequest(http.MethodPost, "/pings/p1/reports/r1/vote",
		bytes.NewBufferString(`{"value":"up"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUser(req, "u3")

	router := chi.NewRouter()
	router.HandleFunc("/pings/{id}/reports/{reportID}/vote", h.VoteReport)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_VoteReport_InvalidValue(t *testing.T) {
	h, store := newTestHandler(t)

	store.reports["r1"] = PingReport{ID: "r1", PingID: "p1", Reason: "duplicate"}

	req := httptest.NewRequest(http.MethodPost, "/pings/p1/reports/r1/vote",
		bytes.NewBufferString(`{"value":"maybe"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUser(req, "u3")

	router := chi.NewRouter()
	router.HandleFunc("/pings/{id}/reports/{reportID}/vote", h.VoteReport)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_VoteReport_ReportNotFound(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/pings/p1/reports/nonexistent/vote",
		bytes.NewBufferString(`{"value":"up"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUser(req, "u3")

	router := chi.NewRouter()
	router.HandleFunc("/pings/{id}/reports/{reportID}/vote", h.VoteReport)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
