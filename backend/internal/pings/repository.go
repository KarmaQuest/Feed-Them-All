// Package pings — repository.go implémente l'interface Store avec PostgreSQL + PostGIS.
//
// Toutes les requêtes SQL du package pings sont ici. Aucune logique métier :
// ce fichier se contente d'exécuter les requêtes et de retourner les résultats bruts.
//
// Points techniques importants :
//
//   ST_MakePoint($lon, $lat)  → LONGITUDE en premier, LATITUDE en second.
//                               C'est le standard PostGIS (axes X=lon, Y=lat).
//                               Inverser les deux est le bug le plus fréquent.
//
//   ST_DWithin(..., radius)   → retourne les points dans un rayon en mètres.
//                               Fonctionne correctement car le type GEOGRAPHY
//                               calcule les distances en mètres sur la sphère terrestre.
//
//   ST_X / ST_Y               → extraient la longitude (X) et la latitude (Y)
//                               depuis un champ GEOGRAPHY pour les retourner en JSON.
//
//   Soft delete               → on met is_active = false au lieu de DELETE,
//                               pour conserver l'historique des pings.
//
//   Violations de contrainte unique (code PgError 23505) :
//     - ping_reports(ping_id, reported_by) → ErrAlreadyReported
//     - ping_report_votes(report_id, user_id) → ErrAlreadyVoted
//
// La table ping_media est ajoutée dans la migration 000003.
// Les tables ping_reports et ping_report_votes sont ajoutées dans les migrations 000004 et 000005.
package pings

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)
// Repository is the PostgreSQL implementation of Store.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a Repository backed by the given connection pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts a new ping and returns the full Ping struct with generated ID and timestamps.
// lon/lat are stored as GEOGRAPHY(POINT, 4326) — note parameter order: ST_MakePoint(lon, lat).
// animalType is NULL for food pings; animalCount defaults to 1.
func (r *Repository) Create(ctx context.Context, userID, pingType string, lat, lon float64, animalType *string, animalCount int) (Ping, error) {
	const q = `
		INSERT INTO pings (type, location, created_by, animal_type, animal_count)
		VALUES ($1, ST_MakePoint($2, $3)::geography, $4, $5, $6)
		RETURNING
			id,
			type,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			created_by,
			is_active,
			fed_at,
			animal_type,
			animal_count,
			created_at,
			updated_at
	`
	// $2 = lon (X axis), $3 = lat (Y axis) — PostGIS convention
	var p Ping
	err := r.db.QueryRow(ctx, q, pingType, lon, lat, userID, animalType, animalCount).Scan(
		&p.ID, &p.Type, &p.Lat, &p.Lon, &p.CreatedBy,
		&p.IsActive, &p.FedAt, &p.AnimalType, &p.AnimalCount, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return Ping{}, fmt.Errorf("pings.Create: %w", err)
	}
	return p, nil
}

// ListNearby returns active pings within q.Radius metres of (q.Lat, q.Lon).
// Results are ordered by most recent first and capped at 200.
// If q.Type is non-empty, only pings of that type are returned.
func (r *Repository) ListNearby(ctx context.Context, q NearbyQuery) ([]Ping, error) {
	// Two variants: with and without type filter — avoids a dynamic query builder
	const qAll = `
		SELECT
			id, type,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			created_by, is_active, fed_at, animal_type, animal_count, created_at, updated_at
		FROM pings
		WHERE is_active = TRUE
		  AND ST_DWithin(location, ST_MakePoint($1, $2)::geography, $3)
		ORDER BY created_at DESC
		LIMIT 200
	`
	const qFiltered = `
		SELECT
			id, type,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			created_by, is_active, fed_at, animal_type, animal_count, created_at, updated_at
		FROM pings
		WHERE is_active = TRUE
		  AND type = $4
		  AND ST_DWithin(location, ST_MakePoint($1, $2)::geography, $3)
		ORDER BY created_at DESC
		LIMIT 200
	`

	var rows interface {
		Scan(...any) error
		Next() bool
		Err() error
	}

	if q.Type == "" {
		r2, err := r.db.Query(ctx, qAll, q.Lon, q.Lat, q.Radius)
		if err != nil {
			return nil, fmt.Errorf("pings.ListNearby: %w", err)
		}
		defer r2.Close()
		rows = r2
	} else {
		r2, err := r.db.Query(ctx, qFiltered, q.Lon, q.Lat, q.Radius, q.Type)
		if err != nil {
			return nil, fmt.Errorf("pings.ListNearby: %w", err)
		}
		defer r2.Close()
		rows = r2
	}

	var pings []Ping
	for rows.Next() {
		var p Ping
		if err := rows.Scan(
			&p.ID, &p.Type, &p.Lat, &p.Lon, &p.CreatedBy,
			&p.IsActive, &p.FedAt, &p.AnimalType, &p.AnimalCount, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pings.ListNearby scan: %w", err)
		}
		pings = append(pings, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pings.ListNearby rows: %w", err)
	}
	return pings, nil
}

// GetByID fetches a single ping by its UUID.
// Returns an error wrapping pgx.ErrNoRows if the ping does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (Ping, error) {
	const q = `
		SELECT
			id, type,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			created_by, is_active, fed_at, animal_type, animal_count, created_at, updated_at
		FROM pings WHERE id = $1
	`
	var p Ping
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Type, &p.Lat, &p.Lon, &p.CreatedBy,
		&p.IsActive, &p.FedAt, &p.AnimalType, &p.AnimalCount, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return Ping{}, fmt.Errorf("pings.GetByID: %w", err)
	}
	return p, nil
}

// Confirm touches updated_at to record "animal still present".
// Does nothing if the ping is inactive (soft-deleted).
func (r *Repository) Confirm(ctx context.Context, id string) error {
	const q = `UPDATE pings SET updated_at = NOW() WHERE id = $1 AND is_active = TRUE`
	_, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("pings.Confirm: %w", err)
	}
	return nil
}

// MarkFed sets fed_at = NOW() to record that the animal at this ping was fed.
// Does nothing if the ping is already inactive.
func (r *Repository) MarkFed(ctx context.Context, id string) error {
	const q = `UPDATE pings SET fed_at = NOW(), updated_at = NOW() WHERE id = $1 AND is_active = TRUE`
	_, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("pings.MarkFed: %w", err)
	}
	return nil
}

// Deactivate performs a soft delete (is_active = false) only if userID is the creator.
// Returns ErrNotOwner if the caller is not the ping owner.
// Returns ErrNotFound if no matching active ping exists.
func (r *Repository) Deactivate(ctx context.Context, id, userID string) error {
	const q = `
		UPDATE pings SET is_active = FALSE, updated_at = NOW()
		WHERE id = $1 AND created_by = $2 AND is_active = TRUE
	`
	tag, err := r.db.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("pings.Deactivate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Either not found, already inactive, or not the owner — check which
		existing, err2 := r.GetByID(ctx, id)
		if err2 != nil {
			return ErrNotFound
		}
		if existing.CreatedBy != userID {
			return ErrNotOwner
		}
		return ErrNotFound
	}
	return nil
}

// AddMedia inserts a media record (file path) linked to a ping.
// The ping_media table is created in migration 000003.
func (r *Repository) AddMedia(ctx context.Context, pingID, filePath string) error {
	const q = `INSERT INTO ping_media (ping_id, file_path) VALUES ($1, $2)`
	_, err := r.db.Exec(ctx, q, pingID, filePath)
	if err != nil {
		return fmt.Errorf("pings.AddMedia: %w", err)
	}
	return nil
}

// ListMedia returns the file paths of all media attached to a ping, ordered by upload time.
func (r *Repository) ListMedia(ctx context.Context, pingID string) ([]string, error) {
	const q = `SELECT file_path FROM ping_media WHERE ping_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, q, pingID)
	if err != nil {
		return nil, fmt.Errorf("pings.ListMedia: %w", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("pings.ListMedia scan: %w", err)
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

// Report inserts a new report for the given ping.
// Returns ErrAlreadyReported if the user already filed a report on this ping (unique constraint).
func (r *Repository) Report(ctx context.Context, pingID, userID, reason string, comment *string) (PingReport, error) {
	const q = `
		INSERT INTO ping_reports (ping_id, reported_by, reason, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, ping_id, reported_by, reason, comment, created_at
	`
	var rp PingReport
	err := r.db.QueryRow(ctx, q, pingID, userID, reason, comment).Scan(
		&rp.ID, &rp.PingID, &rp.ReportedBy, &rp.Reason, &rp.Comment, &rp.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return PingReport{}, ErrAlreadyReported
		}
		return PingReport{}, fmt.Errorf("pings.Report: %w", err)
	}
	return rp, nil
}

// listReportsQuery returns reports with computed score (up - down) for a given ping or single report.
// scoreQuery is used by both ListReports and GetReport.
const scoreQuery = `
	SELECT
		pr.id,
		pr.ping_id,
		pr.reported_by,
		pr.reason,
		pr.comment,
		pr.created_at,
		COALESCE(SUM(CASE WHEN prv.value = 'up' THEN 1 WHEN prv.value = 'down' THEN -1 ELSE 0 END), 0) AS score
	FROM ping_reports pr
	LEFT JOIN ping_report_votes prv ON prv.report_id = pr.id
`

// ListReports returns all reports for a ping with aggregated vote scores.
func (r *Repository) ListReports(ctx context.Context, pingID string) ([]PingReport, error) {
	q := scoreQuery + `WHERE pr.ping_id = $1 GROUP BY pr.id ORDER BY score DESC, pr.created_at DESC`
	rows, err := r.db.Query(ctx, q, pingID)
	if err != nil {
		return nil, fmt.Errorf("pings.ListReports: %w", err)
	}
	defer rows.Close()

	var reports []PingReport
	for rows.Next() {
		var rp PingReport
		if err := rows.Scan(
			&rp.ID, &rp.PingID, &rp.ReportedBy, &rp.Reason,
			&rp.Comment, &rp.CreatedAt, &rp.Score,
		); err != nil {
			return nil, fmt.Errorf("pings.ListReports scan: %w", err)
		}
		reports = append(reports, rp)
	}
	return reports, rows.Err()
}

// GetReport fetches a single report by UUID with its current vote score.
// Returns ErrNotFound if no report with that ID exists.
func (r *Repository) GetReport(ctx context.Context, reportID string) (PingReport, error) {
	q := scoreQuery + `WHERE pr.id = $1 GROUP BY pr.id`
	var rp PingReport
	err := r.db.QueryRow(ctx, q, reportID).Scan(
		&rp.ID, &rp.PingID, &rp.ReportedBy, &rp.Reason,
		&rp.Comment, &rp.CreatedAt, &rp.Score,
	)
	if err != nil {
		return PingReport{}, fmt.Errorf("pings.GetReport: %w", ErrNotFound)
	}
	return rp, nil
}

// VoteReport inserts or updates a vote (up/down) on a report.
// If the user already voted, the value is updated (upsert — allows changing up→down or down→up).
func (r *Repository) VoteReport(ctx context.Context, reportID, userID, value string) error {
	const q = `
		INSERT INTO ping_report_votes (report_id, user_id, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (report_id, user_id) DO UPDATE SET value = EXCLUDED.value
	`
	_, err := r.db.Exec(ctx, q, reportID, userID, value)
	if err != nil {
		return fmt.Errorf("pings.VoteReport: %w", err)
	}
	return nil
}

// AddFeedingEvent records a feeding action and updates pings.fed_at in a single transaction.
func (r *Repository) AddFeedingEvent(ctx context.Context, pingID, userID string, req CreateFeedingEventRequest) (FeedingEvent, error) {
	const q = `
		WITH ev AS (
			INSERT INTO ping_feeding_events (ping_id, fed_by, note, animal_count_seen, event_type)
			VALUES ($1, $2, $3, $4, 'feeding')
			RETURNING id, ping_id, fed_by, fed_at, note, animal_count_seen, event_type
		), upd AS (
			UPDATE pings SET fed_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND is_active = TRUE
		)
		SELECT id, ping_id, fed_by, fed_at, note, animal_count_seen, event_type FROM ev
	`
	var e FeedingEvent
	err := r.db.QueryRow(ctx, q, pingID, userID, req.Note, req.AnimalCountSeen).Scan(
		&e.ID, &e.PingID, &e.FedBy, &e.FedAt, &e.Note, &e.AnimalCountSeen, &e.EventType,
	)
	if err != nil {
		return FeedingEvent{}, fmt.Errorf("pings.AddFeedingEvent: %w", err)
	}
	return e, nil
}

// ListFeedingEvents returns all feeding events for a ping, most recent first.
// Includes the username of the feeder via JOIN.
func (r *Repository) ListFeedingEvents(ctx context.Context, pingID string) ([]FeedingEvent, error) {
	const q = `
		SELECT e.id, e.ping_id, e.fed_by, COALESCE(u.username, ''), e.fed_at, e.note, e.animal_count_seen, e.event_type
		FROM ping_feeding_events e
		LEFT JOIN users u ON u.id = e.fed_by
		WHERE e.ping_id = $1
		ORDER BY e.fed_at DESC
	`
	rows, err := r.db.Query(ctx, q, pingID)
	if err != nil {
		return nil, fmt.Errorf("pings.ListFeedingEvents: %w", err)
	}
	defer rows.Close()

	var events []FeedingEvent
	for rows.Next() {
		var e FeedingEvent
		if err := rows.Scan(&e.ID, &e.PingID, &e.FedBy, &e.Username, &e.FedAt, &e.Note, &e.AnimalCountSeen, &e.EventType); err != nil {
			return nil, fmt.Errorf("pings.ListFeedingEvents scan: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// AddSignalEvent inserts the initial 'signal' event when a ping is first created.
func (r *Repository) AddSignalEvent(ctx context.Context, pingID, userID string) error {
	const q = `
		INSERT INTO ping_feeding_events (ping_id, fed_by, event_type)
		VALUES ($1, $2, 'signal')
	`
	_, err := r.db.Exec(ctx, q, pingID, userID)
	if err != nil {
		return fmt.Errorf("pings.AddSignalEvent: %w", err)
	}
	return nil
}

// UpdatePing updates animal_type and/or animal_count for a ping.
// Only the owner may update.
func (r *Repository) UpdatePing(ctx context.Context, id, userID string, animalType *string, animalCount *int) (Ping, error) {
	// Verify ownership first
	var ownerID string
	var isActive bool
	err := r.db.QueryRow(ctx, `SELECT created_by, is_active FROM pings WHERE id = $1`, id).Scan(&ownerID, &isActive)
	if err != nil {
		return Ping{}, ErrNotFound
	}
	if ownerID != userID {
		return Ping{}, ErrNotOwner
	}
	if !isActive {
		return Ping{}, ErrNotFound
	}

	const q = `
		UPDATE pings
		SET
			animal_type  = COALESCE($2, animal_type),
			animal_count = COALESCE($3, animal_count),
			updated_at   = NOW()
		WHERE id = $1
		RETURNING
			id, type,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			created_by, is_active, fed_at,
			animal_type, animal_count,
			created_at, updated_at
	`
	var p Ping
	err = r.db.QueryRow(ctx, q, id, animalType, animalCount).Scan(
		&p.ID, &p.Type, &p.Lat, &p.Lon, &p.CreatedBy,
		&p.IsActive, &p.FedAt, &p.AnimalType, &p.AnimalCount, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return Ping{}, fmt.Errorf("pings.UpdatePing: %w", err)
	}
	return p, nil
}

// roundCoord rounds a GPS coordinate to 4 decimal places (~11 m precision).
// Used when returning coordinates publicly to protect feeder privacy.
// roundCoord rounds a GPS coordinate to 3 decimal places (~100 m precision)
// to protect feeder privacy when exposing coordinates publicly.
func roundCoord(v float64) float64 {
	return math.Round(v*1000) / 1000
}

// now is a helper to get the current UTC time (used in tests via injection if needed).
var now = func() time.Time { return time.Now().UTC() }
