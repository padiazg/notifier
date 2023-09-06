package webhook

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	n "github.com/padiazg/notifier/notification"
)

// WebhookNotifier implements the Notifier interface for webhooks
type WebhookNotifier struct {
	Config
}

func NewWebhookNotifier(config Config) *WebhookNotifier {
	return &WebhookNotifier{
		Config: config,
	}
}

func (wn *WebhookNotifier) SendNotification(notification *n.Notification) n.NotificationResult {
	// Serialize the notification data to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		return n.NotificationResult{Success: false, Error: err}
	}

	// Send the POST request to the webhook endpoint
	r, err := http.NewRequest(http.MethodPost, wn.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return n.NotificationResult{Success: false, Error: err}
	}

	// Ser headers
	r.Header.Set("Content-Type", "application/json")

	for k, v := range wn.Headers {
		r.Header.Set(k, v)
	}

	client := &http.Client{}

	if wn.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Do(r)
	if err != nil {
		return n.NotificationResult{Success: false, Error: err}
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return n.NotificationResult{Success: false, Error: fmt.Errorf("webhook returned non-OK status: %d", resp.StatusCode)}
	}

	// Read the response body if needed
	// responseBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return NotificationResult{Success: false, Error: err}
	// }

	return n.NotificationResult{Success: true}
}

// Close is a no-op for the WebhookNotifier
func (wn *WebhookNotifier) Close() error {
	// You can add any cleanup code here
	return nil
}

// Connect is a no-op for the WebhookNotifier
func (wn *WebhookNotifier) Connect() error {
	return nil
}
