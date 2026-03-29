package store

import (
	"context"
	"database/sql"
	"fmt"

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

	if _, err := db.ExecContext(ctx, initSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing database: %w", err)
	}

	return &Store{db: db}, nil
}

const initSQL = `
CREATE TABLE IF NOT EXISTS notes (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	title      TEXT NOT NULL,
	content    TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);

INSERT INTO notes (title, content, created_at)
WITH RECURSIVE seq(i) AS (
	SELECT 1 UNION ALL SELECT i + 1 FROM seq WHERE i < 5 + abs(random()) % 6
)
SELECT
	'note-' || i,
	'content-' || i,
	datetime('now', '-' || i || ' hours')
FROM seq;
`

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
