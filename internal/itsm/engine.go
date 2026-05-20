package itsm

import (
	"fmt"
	"time"

	"github.com/kerochan-web/sentinel/internal/audit"
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
	settings        config.Remediation
	itsmConfig      config.ServiceNowConfig // Added to hold target URLs and credentials
}

// Update the constructor to accept the full network settings
func NewEngine(settings config.Remediation, itsmConfig config.ServiceNowConfig) *Engine {
	return &Engine{
		activeIncidents: make(map[string]*incidentTracker),
		settings:        settings,
		itsmConfig:      itsmConfig,
	}
}

// ProcessCheck takes the result of a health check and manages the ServiceNow record
func (e *Engine) ProcessCheck(svc config.Service, isHealthy bool) {
	tracker, exists := e.activeIncidents[svc.Name]

	if !isHealthy {
		// 1. Maintenance Check
		if svc.Maintenance && time.Now().Before(svc.MaintenanceUntil) {
			fmt.Printf("[Incident Engine] %s is DOWN (Maintenance Active). Skipping.\n", svc.Name)
			
			// Structured Audit Log for skipped remediation due to maintenance
			_ = audit.Log(audit.AuditEntry{
				Service: svc.Name,
				Host:    svc.Target,
				Action:  "maintenance_skip",
				Result:  "ignored",
			})
			return
		}

		// 2. Incident Creation
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
			
			// NEW: Fire payload over the wire to our client implementation
			if err := SendIncident(e.itsmConfig.InstanceURL, newInc); err != nil {
				fmt.Printf("[Incident Engine] API Client Error: %v\n", err)
			}
			
			// Structured Audit Log for incident creation
			_ = audit.Log(audit.AuditEntry{
				Service: svc.Name,
				Host:    svc.Target,
				Action:  "create_incident",
				Result:  "success",
				Ticket:  newInc.Number,
			})
		}

		// 3. Guardrail Check
		canRetry := tracker.attempts < e.settings.MaxRetries 
		cooldownOver := time.Since(tracker.lastAttemptAt) >= e.settings.CooldownPeriod

		if canRetry && (tracker.attempts == 0 || cooldownOver) {
			tracker.attempts++
			tracker.lastAttemptAt = time.Now()
			
			fmt.Printf("[Incident Engine] [%s] Attempting remediation #%d/%d...\n", 
				svc.Name, tracker.attempts, e.settings.MaxRetries)
			
			// Run remediation
			err := remediation.Perform(svc)
			
			// Determine results for the audit entry
			resultStr := "success"
			if err != nil {
				resultStr = "failed"
			}

			// Structured Audit Log for the remediation attempt
			_ = audit.Log(audit.AuditEntry{
				Service: svc.Name,
				Host:    svc.Target,
				Action:  fmt.Sprintf("remediation_attempt_%d", tracker.attempts),
				Result:  resultStr,
				Ticket:  tracker.incident.Number,
			})

		} else if !canRetry {
			fmt.Printf("[Incident Engine] [%s] Max retries (%d) reached. Manual intervention required.\n", 
				svc.Name, e.settings.MaxRetries)

			// Structured Audit Log for the circuit breaker/max retry limit hit
			_ = audit.Log(audit.AuditEntry{
				Service: svc.Name,
				Host:    svc.Target,
				Action:  "circuit_breaker_lockout",
				Result:  "failed",
				Ticket:  tracker.incident.Number,
			})
		}
	} else if isHealthy && exists {
		// 4. Recovery
		inc := tracker.incident

		fmt.Printf("[Incident Engine] <<< RECOVERY: Resolving Ticket %s for %s\n", inc.Number, svc.Name)

		inc.State = models.StateResolved
		now := time.Now()
		inc.ResolvedAt = &now
		inc.CloseNotes = "Service recovered automatically via monitor check."
		
		// Structured Audit Log for automatic service recovery
		_ = audit.Log(audit.AuditEntry{
			Service: svc.Name,
			Host:    svc.Target,
			Action:  "service_recovery",
			Result:  "success",
			Ticket:  inc.Number,
		})
		// Remove from active tracking so a new one can be created if it fails again later
		delete(e.activeIncidents, svc.Name)
	}
}
