// Package pings — fake_store_test.go fournit une implémentation en mémoire de Store
// pour les tests unitaires. Aucune base de données n'est nécessaire.
//
// Comportements simulés :
//   - Create       → génère un faux UUID incrémental, stocke le ping en mémoire
//   - ListNearby   → retourne tous les pings actifs sans filtre spatial (tests hors PostGIS)
//   - GetByID      → cherche par ID, retourne ErrNotFound sinon
//   - Confirm      → touche updated_at
//   - MarkFed      → remplit fed_at
//   - Deactivate   → vérifie le propriétaire, retourne ErrNotOwner / ErrNotFound
//   - AddMedia     → ajoute un chemin dans mediaPaths[pingID]
//   - ListMedia    → retourne les chemins du ping
//   - Report       → insère un report, retourne ErrAlreadyReported si déjà signalé
//   - ListReports  → retourne les reports avec score calculé
//   - GetReport    → cherche par ID, retourne ErrNotFound sinon
//   - VoteReport   → upsert du vote (pas de ErrAlreadyVoted — changement autorisé)
package pings

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// fakeStore is an in-memory Store for unit tests only.
type fakeStore struct {
	mu            sync.Mutex
	pings         map[string]Ping              // id -> Ping
	media         map[string][]string          // pingID -> []filePath
	reports       map[string]PingReport        // reportID -> PingReport
	votes         map[string]map[string]string // reportID -> userID -> value
	feedingEvents map[string][]FeedingEvent    // pingID -> []FeedingEvent
	pingSeq       int
	reportSeq     int
	feedingSeq    int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		pings:         make(map[string]Ping),
		media:         make(map[string][]string),
		reports:       make(map[string]PingReport),
		votes:         make(map[string]map[string]string),
		feedingEvents: make(map[string][]FeedingEvent),
	}
}

func (f *fakeStore) Create(_ context.Context, userID, pingType string, lat, lon float64, animalType *string, animalCount int) (Ping, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.pingSeq++
	id := fmt.Sprintf("ping-%d", f.pingSeq)
	now := time.Now()
	p := Ping{
		ID:          id,
		Type:        pingType,
		Lat:         lat,
		Lon:         lon,
		CreatedBy:   userID,
		IsActive:    true,
		AnimalType:  animalType,
		AnimalCount: animalCount,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	f.pings[id] = p
	return p, nil
}

func (f *fakeStore) ListNearby(_ context.Context, q NearbyQuery) ([]Ping, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []Ping
	for _, p := range f.pings {
		if !p.IsActive {
			continue
		}
		if q.Type != "" && p.Type != q.Type {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func (f *fakeStore) GetByID(_ context.Context, id string) (Ping, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	p, ok := f.pings[id]
	if !ok {
		return Ping{}, ErrNotFound
	}
	return p, nil
}

func (f *fakeStore) Confirm(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	p, ok := f.pings[id]
	if !ok {
		return ErrNotFound
	}
	p.UpdatedAt = time.Now()
	f.pings[id] = p
	return nil
}

func (f *fakeStore) MarkFed(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	p, ok := f.pings[id]
	if !ok {
		return ErrNotFound
	}
	now := time.Now()
	p.FedAt = &now
	p.UpdatedAt = now
	f.pings[id] = p
	return nil
}

func (f *fakeStore) Deactivate(_ context.Context, id, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	p, ok := f.pings[id]
	if !ok {
		return ErrNotFound
	}
	if p.CreatedBy != userID {
		return ErrNotOwner
	}
	p.IsActive = false
	f.pings[id] = p
	return nil
}

func (f *fakeStore) AddMedia(_ context.Context, pingID, filePath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.media[pingID] = append(f.media[pingID], filePath)
	return nil
}

func (f *fakeStore) ListMedia(_ context.Context, pingID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.media[pingID], nil
}

func (f *fakeStore) Report(_ context.Context, pingID, userID, reason string, comment *string) (PingReport, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check unique constraint (ping_id + reported_by)
	for _, r := range f.reports {
		if r.PingID == pingID && r.ReportedBy == userID {
			return PingReport{}, ErrAlreadyReported
		}
	}

	f.reportSeq++
	id := fmt.Sprintf("report-%d", f.reportSeq)
	rp := PingReport{
		ID:         id,
		PingID:     pingID,
		ReportedBy: userID,
		Reason:     reason,
		Comment:    comment,
		Score:      0,
		CreatedAt:  time.Now(),
	}
	f.reports[id] = rp
	return rp, nil
}

func (f *fakeStore) ListReports(_ context.Context, pingID string) ([]PingReport, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []PingReport
	for _, rp := range f.reports {
		if rp.PingID != pingID {
			continue
		}
		// Compute score from votes
		score := 0
		if v, ok := f.votes[rp.ID]; ok {
			for _, val := range v {
				if val == "up" {
					score++
				} else {
					score--
				}
			}
		}
		rp.Score = score
		out = append(out, rp)
	}
	return out, nil
}

func (f *fakeStore) GetReport(_ context.Context, reportID string) (PingReport, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	rp, ok := f.reports[reportID]
	if !ok {
		return PingReport{}, ErrNotFound
	}
	return rp, nil
}

func (f *fakeStore) VoteReport(_ context.Context, reportID, userID, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.reports[reportID]; !ok {
		return ErrNotFound
	}
	if f.votes[reportID] == nil {
		f.votes[reportID] = make(map[string]string)
	}
	f.votes[reportID][userID] = value // upsert
	return nil
}

func (f *fakeStore) AddFeedingEvent(_ context.Context, pingID, userID string, req CreateFeedingEventRequest) (FeedingEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	p, ok := f.pings[pingID]
	if !ok {
		return FeedingEvent{}, ErrNotFound
	}

	// Update fed_at on ping
	now := time.Now()
	p.FedAt = &now
	p.UpdatedAt = now
	f.pings[pingID] = p

	f.feedingSeq++
	e := FeedingEvent{
		ID:              fmt.Sprintf("feeding-%d", f.feedingSeq),
		PingID:          pingID,
		FedBy:           userID,
		FedAt:           now,
		Note:            req.Note,
		AnimalCountSeen: req.AnimalCountSeen,
	}
	f.feedingEvents[pingID] = append(f.feedingEvents[pingID], e)
	return e, nil
}

func (f *fakeStore) ListFeedingEvents(_ context.Context, pingID string) ([]FeedingEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	events := f.feedingEvents[pingID]
	// Return a copy to avoid races in tests
	out := make([]FeedingEvent, len(events))
	copy(out, events)
	return out, nil
}
