package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/padiazg/notifier"
)

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming JSON payload
	var notification notifier.Notification
	err := json.NewDecoder(r.Body).Decode(&notification)
	if err != nil {
		http.Error(w, "Failed to decode JSON payload", http.StatusBadRequest)
		return
	}

	// Process the received notification
	fmt.Printf("Received notification: %+v\n", notification)

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Notification received successfully!")
}

func main() {
	// Set up a simple HTTP server to receive webhooks
	http.HandleFunc("/webhook", handleWebhook)
	fmt.Println("Webhook server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
