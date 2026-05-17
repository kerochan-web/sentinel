package itsm

import (
	"fmt"
	"time"

	"github.com/kerochan-web/sentinel/internal/config"
	"github.com/kerochan-web/sentinel/internal/remediation"
	"github.com/kerochan-web/sentinel/pkg/models"
)

// Engine tracks the status of incidents to prevent duplicates
type Engine struct {
	// activeIncidents maps Service Name -> Active Incident model
	activeIncidents map[string]*models.Incident
}

func NewEngine() *Engine {
	return &Engine{
		activeIncidents: make(map[string]*models.Incident),
	}
}

// ProcessCheck takes the result of a health check and manages the ServiceNow record
func (e *Engine) ProcessCheck(svc config.Service, isHealthy bool) {
	existingInc, exists := e.activeIncidents[svc.Name]

	if !isHealthy {
		// 1. Maintenance Check
		if svc.Maintenance && time.Now().Before(svc.MaintenanceUntil) {
			fmt.Printf("[Incident Engine] %s is DOWN (Maintenance Active). Skipping.\n", svc.Name)
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
			e.activeIncidents[svc.Name] = newInc
			fmt.Printf("[Incident Engine] >>> ALERT: Creating ServiceNow Ticket %s for %s\n", newInc.Number, svc.Name)

			// 3. Simple Remediation Trigger
			err := remediation.Perform(svc)
			if err != nil {
				fmt.Printf("[Incident Engine] Remediation failed: %v\n", err)
			}
		}
	} else if isHealthy && exists {
		// Case: Service was down but is now back UP
		fmt.Printf("[Incident Engine] <<< RECOVERY: Resolving Ticket %s for %s\n", existingInc.Number, svc.Name)
		
		existingInc.State = models.StateResolved
		now := time.Now()
		existingInc.ResolvedAt = &now
		existingInc.CloseNotes = "Service recovered automatically via monitor check."
		
		// Remove from active tracking so a new one can be created if it fails again later
		delete(e.activeIncidents, svc.Name)
	}
}
