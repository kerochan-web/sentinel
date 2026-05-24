package remediation

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kerochan-web/sentinel/internal/audit"
	"github.com/kerochan-web/sentinel/internal/config"
)

// Perform attempts to fix the service using a dynamic, decoupled shell command.
func Perform(svc config.Service) error {
	if svc.RemediationCommand == "" {
		return fmt.Errorf("no remediation command configured for service: %s", svc.Name)
	}

	// Context Token Injection
	// Intercept the command template and swap out our dynamic tokens
	cmdStr := svc.RemediationCommand
	cmdStr = strings.ReplaceAll(cmdStr, "$SERVICE_NAME", svc.Name)
	cmdStr = strings.ReplaceAll(cmdStr, "$SERVICE_TARGET", svc.Target)

	// Run compliance check before initializing the runner context
	if err := ValidateCommand(cmdStr); err != nil {
		// Immediately write out the CRITICAL_SAFETY_VIOLATION block to the audit log
		_ = audit.Log(audit.AuditEntry{
			Service: svc.Name,
			Host:    svc.Target,
			Action:  "CRITICAL_SAFETY_VIOLATION",
			Result:  "aborted",
		})
		return err
	}

	fmt.Printf("[Remediator] Executing dynamic command for %s: %s\n", svc.Name, cmdStr)

	// Safe Exec Runner with a hard 5-second timeout boundary
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Execute through an explicit shell runner wrapper
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", cmdStr)

	// Trap both stdout and stderr pipes to surface execution details or failures cleanly
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Differentiate between an execution error and a strict timeout threshold breach
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("remediation command timed out after 5s: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
		}
		return fmt.Errorf("remediation command failed: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	// Log standard out for traceability if the script echoed anything
	if stdout.Len() > 0 {
		fmt.Printf("[Remediator] [%s] stdout: %s\n", svc.Name, strings.TrimSpace(stdout.String()))
	}

	return nil
}
