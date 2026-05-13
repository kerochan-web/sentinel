package monitor

import (
	"net/http"
	"time"
	"fmt"

	"github.com/kerochan-web/sentinel/internal/config"
)

// Check returns true if the service is healthy, false otherwise.
func Check(s config.Service) bool {
	switch s.Type {
	case "http":
		return checkHTTP(s.Target)
	case "systemd":
		// For now, we'll just log that we aren't supporting this yet
		// Tiny steps!
		fmt.Printf("[Monitor] systemd check for %s not yet implemented\n", s.Name)
		return true 
	default:
		fmt.Printf("[Monitor] Unknown service type: %s\n", s.Type)
		return false
	}
}

func checkHTTP(url string) bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider 200-299 as healthy
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
