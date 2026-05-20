package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	// ServiceNow Table API endpoint pattern
	http.HandleFunc("/api/now/table/incident", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Print the intercepted payload for immediate visibility
		fmt.Printf("\n=== [Mock ServiceNow] Received Incident Create Request ===\n")
		fmt.Println(string(body))
		fmt.Printf("=========================================================\n")

		// Respond with a mock ServiceNow JSON structure
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"result": {"status": "created", "sys_id": "mock-sys-99999"}}`))
	})

	fmt.Println("Mock ServiceNow API listening on :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
