package webhook

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/padiazg/notifier/notification"
)

type WebhookNotifier struct {
	*Config
	Channel chan *notification.Notification
}

func (n *WebhookNotifier) Type() string {
	return "webhook"
}

func (n *WebhookNotifier) Name() string {
	return n.Config.Name
}

// GetChannel returns the channel used by the worker
func (n *WebhookNotifier) GetChannel() chan *notification.Notification {
	return n.Channel
}

func (n *WebhookNotifier) Connect() error {
	fmt.Println("WebhookNotifier.Connect")
	return nil
}

func (n *WebhookNotifier) Close() error {
	fmt.Println("WebhookNotifier.Close")
	return nil
}

// Notify sends a notification to worker
func (n *WebhookNotifier) Notify(payload *notification.Notification) {
	fmt.Printf("WebhookNotifier.Notify: %v\n", payload.ID)
	if n.Channel == nil || payload == nil {
		return
	}
	n.Channel <- payload
}

// NewWebhookNotifier creates a new notifier for webhooks
func (n *WebhookNotifier) StartWorker() {
	fmt.Println("WebhookNotifier.StartWorker")
	n.Channel = make(chan *notification.Notification)
	for notification := range n.Channel {
		n.SendNotification(notification)
	}

	fmt.Printf("Webhook notifier stopped\n")
}

// SendNotification sends a notification to the webhook
func (n *WebhookNotifier) SendNotification(message *notification.Notification) notification.Result {
	fmt.Printf("WebhookNotifier.sendNotification: %v\n", message)

	// Serialize the notification data to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return notification.Result{Success: false, Error: err}
	}

	// Send the POST request to the webhook endpoint
	r, err := http.NewRequest(http.MethodPost, n.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return notification.Result{Success: false, Error: err}
	}

	// Ser headers
	r.Header.Set("Content-Type", "application/json")

	for k, v := range n.Headers {
		r.Header.Set(k, v)
	}

	client := &http.Client{}

	if n.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Do(r)
	if err != nil {
		return notification.Result{Success: false, Error: err}
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return notification.Result{Success: false, Error: fmt.Errorf("webhook returned non-OK status: %d", resp.StatusCode)}
	}

	// Read the response body if needed
	// responseBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return NotificationResult{Success: false, Error: err}
	// }

	return notification.Result{Success: true}

}
