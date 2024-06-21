package webhook

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/padiazg/notifier/notification"
	"github.com/padiazg/notifier/utils"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Config struct {
	Logger   *log.Logger
	Headers  map[string]string
	Name     string
	Endpoint string
	Insecure bool
}

type WebhookNotifier struct {
	*Config
	Channel chan *notification.Notification
	client  HTTPClient
}

var (
	jsonMarshal    = json.Marshal
	httpNewRequest = http.NewRequest
)

func New(config *Config) *WebhookNotifier {
	return (&WebhookNotifier{}).New(config)
}

func (n *WebhookNotifier) New(config *Config) *WebhookNotifier {
	if config == nil {
		config = &Config{}
	}

	if config.Name == "" {
		config.Name = n.Type() + utils.RandomId8()
	}

	n.Config = config

	n.Channel = make(chan *notification.Notification)

	return n
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
	return nil
}

func (n *WebhookNotifier) Close() error {
	return nil
}

// Notify sends a notification to worker
func (n *WebhookNotifier) Notify(payload *notification.Notification) {
	if n.Channel == nil {
		n.Logger.Print("channel is nil")
		return
	}

	if payload == nil {
		n.Logger.Print("payload is nil")
		return
	}

	n.Channel <- payload
}

// NewWebhookNotifier creates a new notifier for webhooks
func (n *WebhookNotifier) StartWorker() {
	for notification := range n.Channel {
		n.SendNotification(notification)
	}
}

// SendNotification sends a notification to the webhook
func (n *WebhookNotifier) SendNotification(message *notification.Notification) *notification.Result {
	// Serialize the notification data to JSON
	payload, err := jsonMarshal(message)
	if err != nil {
		return &notification.Result{Success: false, Error: err}
	}

	// Send the POST request to the webhook endpoint
	r, err := httpNewRequest(http.MethodPost, n.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return &notification.Result{Success: false, Error: err}
	}

	// Ser headers
	r.Header.Set("Content-Type", "application/json")

	for k, v := range n.Headers {
		r.Header.Set(k, v)
	}

	client := n.getClient()

	resp, err := client.Do(r)
	if err != nil {
		return &notification.Result{Success: false, Error: err}
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return &notification.Result{Success: false, Error: fmt.Errorf("webhook returned non-OK status: %d", resp.StatusCode)}
	}

	// Read the response body if needed
	// responseBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return NotificationResult{Success: false, Error: err}
	// }

	return &notification.Result{Success: true}
}

func (n *WebhookNotifier) getClient() HTTPClient {
	if n.client != nil {
		return n.client
	}

	n.client = &http.Client{}
	if n.Insecure {
		n.client.(*http.Client).Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return n.client
}
