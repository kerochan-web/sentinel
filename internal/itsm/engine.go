package itsm

import (
	"fmt"
	"time"

	"github.com/kerochan-web/sentinel/internal/config"
	"github.com/kerochan-web/sentinel/internal/remediation"
	"github.com/kerochan-web/sentinel/pkg/models"
)

type incidentTracker struct {
	incident      *models.Incident
	attempts      int
	lastAttemptAt time.Time
}

// Engine tracks the status of incidents to prevent duplicates
type Engine struct {
	// activeIncidents now maps Service Name -> Tracker instead of just the Model
	activeIncidents map[string]*incidentTracker
}

func NewEngine() *Engine {
	return &Engine{
		activeIncidents: make(map[string]*incidentTracker),
	}
}

// ProcessCheck takes the result of a health check and manages the ServiceNow record
func (e *Engine) ProcessCheck(svc config.Service, isHealthy bool) {
	tracker, exists := e.activeIncidents[svc.Name]

	if !isHealthy {
		// 1. Maintenance Check
		if svc.Maintenance && time.Now().Before(svc.MaintenanceUntil) {
			fmt.Printf("[Incident Engine] %s is DOWN (Maintenance Active). Skipping.\n", svc.Name)
			return
		}

		// 2. Ticket Creation
		if !exists {
			newInc := &models.Incident{
				SysID:            fmt.Sprintf("mock-sys-%d", time.Now().Unix()),
				Number:           fmt.Sprintf("INC%07d", 1000+len(e.activeIncidents)+1),
				ShortDescription: fmt.Sprintf("CRITICAL: %s check failed for %s", svc.Type, svc.Name),
				State:            models.StateNew,
				Severity:         1,
				OpenedAt:         time.Now(),
			}
			
			tracker = &incidentTracker{
				incident: newInc,
			}
			e.activeIncidents[svc.Name] = tracker
			fmt.Printf("[Incident Engine] >>> ALERT: Creating ServiceNow Ticket %s for %s\n", newInc.Number, svc.Name)
		}

		// 3. Guardrail Check (Retries & Cooldown)
		// For now, we'll use a hardcoded 1-minute cooldown for testing 
		// until we pass the full config.Remediation object into the Engine.
		
		canRetry := tracker.attempts < 3 
		cooldownOver := time.Since(tracker.lastAttemptAt) > 1 * time.Minute

		if canRetry && (tracker.attempts == 0 || cooldownOver) {
			tracker.attempts++
			tracker.lastAttemptAt = time.Now()
			
			fmt.Printf("[Incident Engine] Attempting remediation #%d for %s...\n", tracker.attempts, svc.Name)
			remediation.Perform(svc)
		} else if !canRetry {
			fmt.Printf("[Incident Engine] Max retries reached for %s. Awaiting manual intervention.\n", svc.Name)
		}
	} else if isHealthy && exists {
		// 4. Recovery
		inc := tracker.incident

		fmt.Printf("[Incident Engine] <<< RECOVERY: Resolving Ticket %s for %s\n", inc.Number, svc.Name)

		inc.State = models.StateResolved
		now := time.Now()
		inc.ResolvedAt = &now
		inc.CloseNotes = "Service recovered automatically via monitor check."
		
		// Remove from active tracking so a new one can be created if it fails again later
		delete(e.activeIncidents, svc.Name)
	}
}
