// Package pings — service_test.go teste la logique métier du Service pings.
//
// Chaque test utilise un fakeStore en mémoire — aucune DB requise.
// Les cas couverts :
//
//   Create   → succès, type invalide, coordonnées hors limites
//   ListNearby → succès, rayon par défaut, filtre type, coordonnées invalides
//   Confirm  → succès
//   MarkFed  → succès, fed_at rempli
//   Deactivate → succès, ErrNotOwner si autre utilisateur
//   Report   → succès, ErrInvalidReason, ErrAlreadyReported
//   ListReports → succès, retourne le score calculé
//   VoteReport → succès, changement de vote (upsert), ErrNotFound si report inexistant
package pings

import (
	"context"
	"os"
	"testing"
)

func newTestService(t *testing.T) (*Service, *fakeStore) {
	t.Helper()
	// SaveMedia needs an upload dir — use a temp dir
	dir := t.TempDir()
	os.Setenv("UPLOAD_DIR", dir)
	store := newFakeStore()
	svc := NewService(store)
	return svc, store
}

// --- Create ---

func TestService_Create_Success(t *testing.T) {
	svc, _ := newTestService(t)

	p, err := svc.Create(context.Background(), "user-1", CreatePingRequest{
		Type: "animal",
		Lat:  48.8566,
		Lon:  2.3522,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.ID == "" {
		t.Error("expected non-empty ID")
	}
	if p.Type != "animal" {
		t.Errorf("expected type animal, got %s", p.Type)
	}
	if !p.IsActive {
		t.Error("expected ping to be active")
	}
}

func TestService_Create_InvalidType(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Create(context.Background(), "user-1", CreatePingRequest{
		Type: "unknown",
		Lat:  48.8566,
		Lon:  2.3522,
	})
	if err != ErrInvalidType {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestService_Create_InvalidCoords(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Create(context.Background(), "user-1", CreatePingRequest{
		Type: "food",
		Lat:  999,
		Lon:  2.3522,
	})
	if err != ErrInvalidCoords {
		t.Errorf("expected ErrInvalidCoords, got %v", err)
	}
}

// --- ListNearby ---

func TestService_ListNearby_Success(t *testing.T) {
	svc, _ := newTestService(t)

	if _, err := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 48.8566, Lon: 2.3522}); err != nil {
		t.Fatalf("setup: create ping: %v", err)
	}

	pings, err := svc.ListNearby(context.Background(), NearbyQuery{Lat: 48.8566, Lon: 2.3522, Radius: 500})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(pings) != 1 {
		t.Errorf("expected 1 ping, got %d", len(pings))
	}
}

func TestService_ListNearby_DefaultRadius(t *testing.T) {
	svc, _ := newTestService(t)

	// Radius 0 should be clamped to defaultRadius (500m) without error
	_, err := svc.ListNearby(context.Background(), NearbyQuery{Lat: 48.8566, Lon: 2.3522, Radius: 0})
	if err != nil {
		t.Fatalf("expected no error with radius 0, got %v", err)
	}
}

func TestService_ListNearby_TypeFilter(t *testing.T) {
	svc, _ := newTestService(t)

	if _, err := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 48.8566, Lon: 2.3522}); err != nil {
		t.Fatalf("setup create animal: %v", err)
	}
	if _, err := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "food", Lat: 48.8566, Lon: 2.3522}); err != nil {
		t.Fatalf("setup create food: %v", err)
	}

	pings, err := svc.ListNearby(context.Background(), NearbyQuery{Lat: 48.8566, Lon: 2.3522, Radius: 500, Type: "animal"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for _, p := range pings {
		if p.Type != "animal" {
			t.Errorf("expected type animal, got %s", p.Type)
		}
	}
}

func TestService_ListNearby_InvalidCoords(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.ListNearby(context.Background(), NearbyQuery{Lat: 200, Lon: 0})
	if err != ErrInvalidCoords {
		t.Errorf("expected ErrInvalidCoords, got %v", err)
	}
}

// --- Confirm ---

func TestService_Confirm_Success(t *testing.T) {
	svc, _ := newTestService(t)

	p, err := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "food", Lat: 1, Lon: 1})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := svc.Confirm(context.Background(), p.ID); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// --- MarkFed ---

func TestService_MarkFed_Success(t *testing.T) {
	svc, store := newTestService(t)

	p, err := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := svc.MarkFed(context.Background(), p.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stored := store.pings[p.ID]
	if stored.FedAt == nil {
		t.Error("expected fed_at to be set")
	}
}

// --- Deactivate ---

func TestService_Deactivate_Success(t *testing.T) {
	svc, store := newTestService(t)

	p, err := svc.Create(context.Background(), "owner-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := svc.Deactivate(context.Background(), p.ID, "owner-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if store.pings[p.ID].IsActive {
		t.Error("expected ping to be inactive after deactivation")
	}
}

func TestService_Deactivate_NotOwner(t *testing.T) {
	svc, _ := newTestService(t)

	p, err := svc.Create(context.Background(), "owner-1", CreatePingRequest{Type: "food", Lat: 1, Lon: 1})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	err = svc.Deactivate(context.Background(), p.ID, "other-user")
	if err != ErrNotOwner {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

// --- Report ---

func TestService_Report_Success(t *testing.T) {
	svc, _ := newTestService(t)

	p, err := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	rp, err := svc.Report(context.Background(), p.ID, "user-2", CreateReportRequest{Reason: "animal_gone"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rp.Reason != "animal_gone" {
		t.Errorf("expected reason animal_gone, got %s", rp.Reason)
	}
	if rp.PingID != p.ID {
		t.Errorf("expected ping_id %s, got %s", p.ID, rp.PingID)
	}
}

func TestService_Report_CreatorCanReport(t *testing.T) {
	svc, _ := newTestService(t)

	// The ping creator should also be able to report their own ping
	p, _ := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "food", Lat: 1, Lon: 1})
	_, err := svc.Report(context.Background(), p.ID, "user-1", CreateReportRequest{Reason: "duplicate"})
	if err != nil {
		t.Errorf("creator should be able to report own ping, got %v", err)
	}
}

func TestService_Report_InvalidReason(t *testing.T) {
	svc, _ := newTestService(t)

	p, _ := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	_, err := svc.Report(context.Background(), p.ID, "user-2", CreateReportRequest{Reason: "not_a_reason"})
	if err != ErrInvalidReason {
		t.Errorf("expected ErrInvalidReason, got %v", err)
	}
}

func TestService_Report_AlreadyReported(t *testing.T) {
	svc, _ := newTestService(t)

	p, _ := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	if _, err := svc.Report(context.Background(), p.ID, "user-2", CreateReportRequest{Reason: "duplicate"}); err != nil {
		t.Fatalf("first report: %v", err)
	}
	_, err := svc.Report(context.Background(), p.ID, "user-2", CreateReportRequest{Reason: "duplicate"})
	if err != ErrAlreadyReported {
		t.Errorf("expected ErrAlreadyReported, got %v", err)
	}
}

// --- ListReports ---

func TestService_ListReports_WithScore(t *testing.T) {
	svc, _ := newTestService(t)

	p, _ := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	rp, _ := svc.Report(context.Background(), p.ID, "user-2", CreateReportRequest{Reason: "wrong_location"})

	// Vote up from user-3, down from user-4
	if err := svc.VoteReport(context.Background(), rp.ID, "user-3", VoteReportRequest{Value: "up"}); err != nil {
		t.Fatalf("vote up: %v", err)
	}
	if err := svc.VoteReport(context.Background(), rp.ID, "user-4", VoteReportRequest{Value: "down"}); err != nil {
		t.Fatalf("vote down: %v", err)
	}

	reports, err := svc.ListReports(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("list reports: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	// score = up(1) - down(1) = 0
	if reports[0].Score != 0 {
		t.Errorf("expected score 0, got %d", reports[0].Score)
	}
}

// --- VoteReport ---

func TestService_VoteReport_Success(t *testing.T) {
	svc, _ := newTestService(t)

	p, _ := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	rp, _ := svc.Report(context.Background(), p.ID, "user-2", CreateReportRequest{Reason: "animal_gone"})

	if err := svc.VoteReport(context.Background(), rp.ID, "user-3", VoteReportRequest{Value: "up"}); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestService_VoteReport_ChangeVote(t *testing.T) {
	svc, store := newTestService(t)

	p, _ := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	rp, _ := svc.Report(context.Background(), p.ID, "user-2", CreateReportRequest{Reason: "animal_gone"})

	// Vote up then change to down — should not error (upsert)
	if err := svc.VoteReport(context.Background(), rp.ID, "user-3", VoteReportRequest{Value: "up"}); err != nil {
		t.Fatalf("first vote: %v", err)
	}
	if err := svc.VoteReport(context.Background(), rp.ID, "user-3", VoteReportRequest{Value: "down"}); err != nil {
		t.Fatalf("change vote: %v", err)
	}

	if store.votes[rp.ID]["user-3"] != "down" {
		t.Errorf("expected vote to be updated to down, got %s", store.votes[rp.ID]["user-3"])
	}
}

func TestService_VoteReport_InvalidValue(t *testing.T) {
	svc, _ := newTestService(t)

	p, _ := svc.Create(context.Background(), "user-1", CreatePingRequest{Type: "animal", Lat: 1, Lon: 1})
	rp, _ := svc.Report(context.Background(), p.ID, "user-2", CreateReportRequest{Reason: "duplicate"})

	err := svc.VoteReport(context.Background(), rp.ID, "user-3", VoteReportRequest{Value: "maybe"})
	if err != ErrInvalidVote {
		t.Errorf("expected ErrInvalidVote, got %v", err)
	}
}

func TestService_VoteReport_ReportNotFound(t *testing.T) {
	svc, _ := newTestService(t)

	err := svc.VoteReport(context.Background(), "nonexistent-report", "user-1", VoteReportRequest{Value: "up"})
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
