package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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
	fmt.Printf("%v Received notification: %+v\n", r.Proto, notification)

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Notification received successfully!")
}

func main() {
	var (
		mux  = http.NewServeMux()
		done = make(chan bool)
		c    = make(chan os.Signal, 2)
	)

	mux.HandleFunc("/webhook", handleWebhook)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	serverTLS := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}

	go func() {
		fmt.Println("HTTP Webhook server listening on :8080")
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	go func() {
		fmt.Println("HTTPS Webhook server listening on :8443")
		if err := serverTLS.ListenAndServeTLS("./localhost.crt", "./localhost.key"); err != nil {
			panic(err)
		}
	}()

	for {
		select {
		case <-done:
			fmt.Println("Closing")
			return
		case <-c:
			fmt.Println("Ctrl+C pressed")
			close(done)
		}
	}
}
