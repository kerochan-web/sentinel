package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// AuditEntry defines the structured JSON format for all operational actions.
type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"` // Go's time.Time automatically marshals to RFC3339/ISO8601 JSON strings
	Service   string    `json:"service"`
	Host      string    `json:"host"`
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	Ticket    string    `json:"ticket,omitempty"` // Blanks won't clutter the JSON if no incident exists yet
}

const auditLogFile = "sentinel_audit.log"

// Log writes an AuditEntry as a single JSON line to sentinel_audit.log
func Log(entry AuditEntry) error {
	// Set the timestamp to right now if it wasn't already provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// 1. Serialize the struct to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// 2. Open the log file in Append, Create, and Write-Only mode.
	// 0644 gives read/write to the owner and read-only to others.
	f, err := os.OpenFile(auditLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer f.Close()

	// 3. Write the JSON payload followed by a newline character
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write data to audit log: %w", err)
	}

	return nil
}
