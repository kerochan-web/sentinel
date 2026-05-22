package notifier

import (
	"fmt"
	"net/http"
	"strings"
)

// SendAlert sends a plain-text notification to a specified ntfy.sh topic.
func SendAlert(topic string, message string) error {
	if topic == "" {
		return fmt.Errorf("ntfy topic cannot be empty")
	}

	url := fmt.Sprintf("https://ntfy.sh/%s", topic)

	// ntfy.sh reads the raw body of the POST request as the notification message
	resp, err := http.Post(url, "text/plain", strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("failed to send ntfy post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy server returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
