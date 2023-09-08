package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	n "github.com/padiazg/notifier/notification"
)

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming JSON payload
	var (
		notification n.Notification
		err          error
	)

	err = json.NewDecoder(r.Body).Decode(&notification)
	if err != nil {
		http.Error(w, "Failed to decode JSON payload", http.StatusBadRequest)
		return
	}

	formated, err := json.MarshalIndent(notification, "", "  ")
	if err != nil {
		log.Printf("Failed to format payload: %v", err)
		return
	}

	// Process the received notification
	fmt.Printf("%v Received notification: %+v\n", r.Method, string(formated))

	// Print the request headers
	var headers []byte
	headers, err = json.MarshalIndent(r.Header, "", "  ")
	fmt.Printf("%v Headers: %s\n", r.Method, string(headers))

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
		Addr:    ":4000",
		Handler: mux,
	}

	serverTLS := &http.Server{
		Addr:    ":4443",
		Handler: mux,
	}

	go func() {
		fmt.Printf("HTTP Webhook server listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	go func() {
		fmt.Printf("HTTPS Webhook server listening on %s\n", serverTLS.Addr)
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
