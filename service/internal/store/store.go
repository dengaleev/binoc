package store

import (
	"context"
	"database/sql"
	"fmt"
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

// New opens a SQLite database with automatic OTel tracing and creates the schema.
func New(ctx context.Context, dsn string) (*Store, error) {
	db, err := otelsql.Open("sqlite", dsn,
		otelsql.WithAttributes(semconv.DBSystemSqlite),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Basic connection settings for SQLite.
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

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
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

// Create inserts a new note and returns it with its generated ID.
func (s *Store) Create(ctx context.Context, title, content string) (Note, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO notes (title, content, created_at) VALUES (?, ?, ?)`,
		title, content, now,
	)
	if err != nil {
		return Note{}, fmt.Errorf("inserting note: %w", err)
	}

	id, _ := result.LastInsertId()
	return Note{ID: id, Title: title, Content: content, CreatedAt: now}, nil
}

// Get retrieves a note by ID.
func (s *Store) Get(ctx context.Context, id int64) (Note, error) {
	var n Note
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, content, created_at FROM notes WHERE id = ?`, id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt)
	if err != nil {
		return Note{}, err
	}
	return n, nil
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
	return notes, nil
}

// Delete removes a note by ID. Returns true if a row was deleted.
func (s *Store) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("deleting note: %w", err)
	}

	n, _ := result.RowsAffected()
	return n > 0, nil
}
