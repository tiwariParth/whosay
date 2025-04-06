package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

type WebhookPayload struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	PushedAt int64 `json:"pushed_at"`
}

func main() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Parse the webhook payload
		var payload WebhookPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		
		// Verify it's for our repository
		if payload.Repository.Name != "whosay" {
			fmt.Fprintf(w, "Ignoring webhook for %s", payload.Repository.Name)
			return
		}
		
		// Execute kubectl to restart the deployment
		cmd := exec.Command("kubectl", "rollout", "restart", "deployment/whosay", "-n", "whosay")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error restarting deployment: %v, output: %s", err, string(output))
			http.Error(w, "Failed to restart deployment", http.StatusInternalServerError)
			return
		}
		
		log.Printf("Deployment restarted successfully")
		fmt.Fprintf(w, "Deployment restarted successfully")
	})
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting webhook server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
