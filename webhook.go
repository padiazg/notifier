package notifier

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

// WebhookNotifier implements the Notifier interface for webhooks
type WebhookNotifier struct {
	Endpoint string // The URL of the webhook endpoint
	Insecure bool   // Whether to skip TLS verification
	// You can add fields specific to the WebhookNotifier
}

func (wn *WebhookNotifier) SendNotification(notification *Notification) NotificationResult {
	// Serialize the notification data to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		return NotificationResult{Success: false, Error: err}
	}

	// Send the POST request to the webhook endpoint
	r, err := http.NewRequest(http.MethodPost, wn.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return NotificationResult{Success: false, Error: err}
	}

	r.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	if wn.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Do(r)
	if err != nil {
		return NotificationResult{Success: false, Error: err}
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return NotificationResult{Success: false, Error: fmt.Errorf("webhook returned non-OK status: %d", resp.StatusCode)}
	}

	// Read the response body if needed
	// responseBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return NotificationResult{Success: false, Error: err}
	// }

	return NotificationResult{Success: true}
}
