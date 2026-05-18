package remediation

import (
	"fmt"
	"os/exec"
	// "runtime"

	"github.com/kerochan-web/sentinel/internal/config"
)

// Perform attempts to fix the service based on its type.
func Perform(svc config.Service) error {
	fmt.Printf("[Remediator] Attempting remediation for %s (Type: %s)...\n", svc.Name, svc.Type)

	switch svc.Type {
	case "systemd":
		// In a real Linux env, this would be: exec.Command("systemctl", "restart", svc.Target).Run()
		fmt.Printf("[Remediator] SIMULATION: systemctl restart %s\n", svc.Target)
		return nil

	case "http":
		// We'll simulate a fix by creating a 'fixed.txt' file
		// In a real scenario, this might be 'docker restart' or 'systemctl reload'
		cmd := exec.Command("touch", "remediation_was_here.txt")
		err := cmd.Run()
		
		if err != nil {
			return fmt.Errorf("failed to execute remediation command: %w", err)
		}
		
		fmt.Printf("[Remediator] [%s] Executed: touch remediation_was_here.txt\n", svc.Name)
		return nil

	default:
		return fmt.Errorf("no remediation path for type: %s", svc.Type)
	}
}
