package remediation

import (
	"errors"
	"fmt"
	"strings"
)

// ErrSafetyViolation is thrown when a command fails the compliance validation suite.
var ErrSafetyViolation = errors.New("critical safety violation: dangerous command execution detected")

// Hardcoded blocklist containing restricted command fragments and binaries
var blocklist = []string{"rm -rf", "mkfs", "dd", "shutdown"}

// ValidateCommand normalizes the command and checks it against safety constraints.
func ValidateCommand(cmdStr string) error {
	// Lexical normalization: convert to lowercase and strip outer whitespace whitespace
	cleaned := strings.ToLower(strings.TrimSpace(cmdStr))

	// 1. Evaluate against the keyword blocklist
	for _, blocked := range blocklist {
		if strings.Contains(cleaned, blocked) {
			return fmt.Errorf("%w: prohibited term '%s' detected", ErrSafetyViolation, blocked)
		}
	}

	// 2. Reject raw shell operations targeting the root directory '/' directly
	if strings.Contains(cleaned, " / ") || strings.HasSuffix(cleaned, " /") {
		return fmt.Errorf("%w: command targets the root directory '/' directly", ErrSafetyViolation)
	}

	return nil
}
