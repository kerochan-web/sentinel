package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kerochan-web/sentinel/internal/config"
	"github.com/kerochan-web/sentinel/pkg/models"
	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("--- Sentinel: Automated Remediation Platform ---")

	// 1. Load Configuration
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Successfully loaded configuration for %d services.\n", len(cfg.Services))
	fmt.Printf("ServiceNow Target: %s\n", cfg.ServiceNow.InstanceURL)

	// 2. Placeholder for ServiceNow logic
	mockInc := models.Incident{
		SysID:            "8d8e5e9b1b1a4d00e8f6e0c6e14bcb2a", // Standard 32-char sys_id
		Number:           "INC0000001",
		ShortDescription: "Service monitor initialization",
		State:            models.StateNew,
	}

	fmt.Printf("Monitoring active for: %s (State: %d)\n", mockInc.Number, mockInc.State)
}

func loadConfig(path string) (*config.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg config.Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	return &cfg, err
}
