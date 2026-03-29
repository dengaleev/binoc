package store

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	_ "modernc.org/sqlite"
)

// Note represents a stored note.
type Note struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// Store provides SQLite operations with automatic OTel instrumentation.
type Store struct {
	db *sql.DB
}

// New opens an in-memory SQLite database with OTel tracing, creates the
// schema, and seeds it with sample data.
func New(ctx context.Context) (*Store, error) {
	db, err := otelsql.Open("sqlite", ":memory:",
		otelsql.WithAttributes(semconv.DBSystemSqlite),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemSqlite)); err != nil {
		db.Close()
		return nil, fmt.Errorf("registering db stats metrics: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.seed(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// List returns all notes ordered by creation time descending.
func (s *Store) List(ctx context.Context) ([]Note, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, content, created_at FROM notes ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning note: %w", err)
		}
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating notes: %w", err)
	}
	return notes, nil
}

func (s *Store) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS notes (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			title      TEXT NOT NULL,
			content    TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("running migration: %w", err)
	}
	return nil
}

var sampleNotes = []struct{ title, content string }{
	{"Deploy v2.3.1", "Rolled out new caching layer to production"},
	{"Incident #847", "Latency spike caused by connection pool exhaustion"},
	{"Sprint retro", "Improve alerting thresholds for p99 latency"},
	{"DB migration plan", "Add index on created_at for the events table"},
	{"Load test results", "Sustained 12k rps with p95 under 50ms"},
	{"On-call handoff", "Watch the memory usage on worker-3, trending up"},
	{"Feature flag cleanup", "Remove stale flags from Q3 experiment"},
	{"Cert rotation", "TLS certs expire in 14 days, auto-renew configured"},
	{"Capacity review", "Current headroom at 40%, plan expansion for Q2"},
	{"Runbook update", "Added steps for OTel collector restart procedure"},
}

func (s *Store) seed(ctx context.Context) error {
	n := 3 + rand.IntN(len(sampleNotes)-3)
	base := time.Now().UTC().Add(-time.Duration(n) * time.Hour)

	for i := range n {
		sample := sampleNotes[i]
		ts := base.Add(time.Duration(i) * time.Hour).Format(time.RFC3339)
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO notes (title, content, created_at) VALUES (?, ?, ?)`,
			sample.title, sample.content, ts,
		)
		if err != nil {
			return fmt.Errorf("seeding note: %w", err)
		}
	}
	return nil
}
