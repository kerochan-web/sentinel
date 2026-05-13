package models

import "time"

// ServiceNow State Constants
const (
    StateNew        = 1
    StateInProgress = 2
    StateOnHold     = 3
    StateResolved   = 6
    StateClosed     = 7
    StateCanceled   = 8
)

// Incident represents a ServiceNow Incident record
type Incident struct {
    SysID             string    `json:"sys_id"`             // Unique 32-character GUID
    Number            string    `json:"number"`             // e.g., INC0001234
    ShortDescription  string    `json:"short_description"`
    State             int       `json:"state"`              // SN uses integers for states
    Severity          int       `json:"severity"`           // 1-High, 2-Medium, 3-Low
    AssignmentGroup   string    `json:"assignment_group"`
    OpenedAt          time.Time `json:"opened_at"`
    ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
    CloseNotes        string    `json:"close_notes,omitempty"`
}
