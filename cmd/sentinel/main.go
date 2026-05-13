package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kerochan-web/sentinel/internal/config"
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

	// 2. Simple Monitoring Loop
	// In a real scenario, this would run in a Ticker or as Goroutines
	for {
		fmt.Println("\n--- Starting Health Check Round ---")
		for _, svc := range cfg.Services {
			isHealthy := monitor.Check(svc)
			
			status := "UP"
			if !isHealthy {
				status = "DOWN"
			}

			fmt.Printf("[%s] Service: %s | Target: %s | Status: %s\n", 
				time.Now().Format("15:04:05"), svc.Name, svc.Target, status)
			
			// If DOWN, this is where we'll eventually trigger the ITSM/Remediation logic
		}
		
		fmt.Println("Waiting for next check...")
		time.Sleep(10 * time.Second) 
	}
}
