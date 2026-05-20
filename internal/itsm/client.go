package itsm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kerochan-web/sentinel/pkg/models"
)

// SendIncident pushes the generated incident structure to the simulated ServiceNow endpoint
func SendIncident(baseURL string, inc *models.Incident) error {
	payload, err := json.Marshal(inc)
	if err != nil {
		return fmt.Errorf("failed to marshal incident model: %w", err)
	}

	fullURL := fmt.Sprintf("%s/api/now/table/incident", baseURL)

	// Set a defensive timeout to avoid locking up our execution loop
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(fullURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("http post request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected backend HTTP status: %d", resp.StatusCode)
	}

	return nil
}
