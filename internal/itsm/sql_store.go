package itsm

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kerochan-web/sentinel/pkg/models"
	_ "modernc.org/sqlite" // Pure-Go SQLite driver registration
)

// SqlStore implements StateStore backed by an embedded database
type SqlStore struct {
	db *sql.DB
}

// NewSqlStore instantiates the store and ensures the schema exists
func NewSqlStore(dbPath string) (*SqlStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	// Bootstrap schema automatically
	schema := `
	CREATE TABLE IF NOT EXISTS active_incidents (
		service_name TEXT PRIMARY KEY,
		ticket_number TEXT NOT NULL,
		attempts INTEGER NOT NULL,
		is_locked_out BOOLEAN NOT NULL,
		last_attempt_at DATETIME NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &SqlStore{db: db}, nil
}

// Get fetches existing state or returns (nil, nil) if no incident tracks exist
func (s *SqlStore) Get(serviceName string) (*incidentTracker, error) {
	query := `SELECT ticket_number, attempts, is_locked_out, last_attempt_at 
	          FROM active_incidents WHERE service_name = ?`
	
	var ticketNumber string
	var attempts int
	var isLockedOut bool
	var lastAttemptAt time.Time

	err := s.db.QueryRow(query, serviceName).Scan(&ticketNumber, &attempts, &isLockedOut, &lastAttemptAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	// Reconstitute the tracking structure safely
	return &incidentTracker{
		incident: &models.Incident{
			Number: ticketNumber,
		},
		attempts:      attempts,
		lastAttemptAt: lastAttemptAt,
		isLockedOut:   isLockedOut,
	}, nil
}

// Save inserts or updates state tracking atomically via UPSERT semantics
func (s *SqlStore) Save(serviceName string, tracker *incidentTracker) error {
	if tracker.incident == nil {
		return fmt.Errorf("cannot save state with a nil incident structure")
	}

	query := `
	INSERT INTO active_incidents (service_name, ticket_number, attempts, is_locked_out, last_attempt_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(service_name) DO UPDATE SET
		ticket_number   = excluded.ticket_number,
		attempts        = excluded.attempts,
		is_locked_out   = excluded.is_locked_out,
		last_attempt_at = excluded.last_attempt_at;`

	_, err := s.db.Exec(query, 
		serviceName, 
		tracker.incident.Number, 
		tracker.attempts, 
		tracker.isLockedOut, 
		tracker.lastAttemptAt,
	)
	if err != nil {
		return fmt.Errorf("database save failed: %w", err)
	}
	return nil
}

// Delete cleans up tracking lines once a service completely recovers
func (s *SqlStore) Delete(serviceName string) error {
	_, err := s.db.Exec("DELETE FROM active_incidents WHERE service_name = ?", serviceName)
	if err != nil {
		return fmt.Errorf("database deletion failed: %w", err)
	}
	return nil
}

// Close gracefully disconnects underlying database handles
func (s *SqlStore) Close() error {
	return s.db.Close()
}
