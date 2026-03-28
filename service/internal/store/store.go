package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Note represents a stored note.
type Note struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// Store provides traced SQLite operations.
type Store struct {
	db     *sql.DB
	tracer trace.Tracer
}

// New opens a SQLite database and creates the schema.
func New(ctx context.Context, dsn, tracerName string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Basic connection settings for SQLite.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &Store{
		db:     db,
		tracer: otel.Tracer(tracerName),
	}

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
	ctx, span := s.tracer.Start(ctx, "db.migrate")
	defer span.End()

	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS notes (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			title      TEXT NOT NULL,
			content    TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("running migration: %w", err)
	}
	return nil
}

// Create inserts a new note and returns it with its generated ID.
func (s *Store) Create(ctx context.Context, title, content string) (Note, error) {
	ctx, span := s.tracer.Start(ctx, "db.insert",
		trace.WithAttributes(
			attribute.String("db.system", "sqlite"),
			attribute.String("db.operation", "INSERT"),
			attribute.String("db.table", "notes"),
		),
	)
	defer span.End()

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO notes (title, content, created_at) VALUES (?, ?, ?)`,
		title, content, now,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Note{}, fmt.Errorf("inserting note: %w", err)
	}

	id, _ := result.LastInsertId()
	span.SetAttributes(attribute.Int64("db.row_id", id))

	return Note{ID: id, Title: title, Content: content, CreatedAt: now}, nil
}

// Get retrieves a note by ID.
func (s *Store) Get(ctx context.Context, id int64) (Note, error) {
	ctx, span := s.tracer.Start(ctx, "db.query",
		trace.WithAttributes(
			attribute.String("db.system", "sqlite"),
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.table", "notes"),
			attribute.Int64("db.row_id", id),
		),
	)
	defer span.End()

	var n Note
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, content, created_at FROM notes WHERE id = ?`, id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Note{}, err
	}
	return n, nil
}

// List returns all notes ordered by creation time descending.
func (s *Store) List(ctx context.Context) ([]Note, error) {
	ctx, span := s.tracer.Start(ctx, "db.query",
		trace.WithAttributes(
			attribute.String("db.system", "sqlite"),
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.table", "notes"),
		),
	)
	defer span.End()

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, content, created_at FROM notes ORDER BY created_at DESC`,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("scanning note: %w", err)
		}
		notes = append(notes, n)
	}

	span.SetAttributes(attribute.Int("db.rows_returned", len(notes)))
	return notes, nil
}

// Delete removes a note by ID. Returns true if a row was deleted.
func (s *Store) Delete(ctx context.Context, id int64) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "db.exec",
		trace.WithAttributes(
			attribute.String("db.system", "sqlite"),
			attribute.String("db.operation", "DELETE"),
			attribute.String("db.table", "notes"),
			attribute.Int64("db.row_id", id),
		),
	)
	defer span.End()

	result, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false, fmt.Errorf("deleting note: %w", err)
	}

	n, _ := result.RowsAffected()
	return n > 0, nil
}
