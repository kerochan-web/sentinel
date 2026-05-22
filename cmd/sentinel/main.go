package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kerochan-web/sentinel/internal/config"
    "github.com/kerochan-web/sentinel/internal/itsm"
	"github.com/kerochan-web/sentinel/internal/monitor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

    // 2. Initialize the Persistent SQL State Store
	stateStore, err := itsm.NewSqlStore("sentinel.db")
	if err != nil {
		log.Fatalf("Failed to initialize persistent state store: %v", err)
	}
	defer stateStore.Close()

	engine := itsm.NewEngine(stateStore, cfg.Remediation, cfg.ServiceNow, cfg.Notifications)

	// Spin up background Prometheus scrapper server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("[Metrics] Serving endpoint at http://localhost:2112/metrics\n")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("Metrics server encountered an error: %v", err)
		}
	}()

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
