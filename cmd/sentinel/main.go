package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kerochan-web/sentinel/internal/config"
    "github.com/kerochan-web/sentinel/internal/itsm"
	"github.com/kerochan-web/sentinel/internal/monitor"
)

func main() {
	fmt.Println("--- Sentinel: Automated Remediation Platform ---")

	// 1. Load Configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Successfully loaded configuration for %d services.\n", len(cfg.Services))
	fmt.Printf("ServiceNow Target: %s\n", cfg.ServiceNow.InstanceURL)

    // 2. Initialize the Incident Engine
	engine := itsm.NewEngine()

	fmt.Printf("Monitoring %d services...\n", len(cfg.Services))

	// 3. Execution Loop
	for {
		for _, svc := range cfg.Services {
			// Perform health check
			isHealthy := monitor.Check(svc)
			
			// Let the Engine decide what to do with the result
			engine.ProcessCheck(svc, isHealthy)
		}
		
		time.Sleep(10 * time.Second) 
	}
}
