package webhook

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/padiazg/notifier/model"
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
	Channel        chan *model.Notification
	client         HTTPClient
	jsonMarshal    func(v any) ([]byte, error)
	httpNewRequest func(method string, url string, body io.Reader) (*http.Request, error)
}

var _ model.Notifier = (*WebhookNotifier)(nil)

func New(config *Config) *WebhookNotifier {
	return (&WebhookNotifier{}).New(config)
}

func (n *WebhookNotifier) New(config *Config) *WebhookNotifier {
	if config == nil {
		config = &Config{}
	}

	if config.Name == "" {
		config.Name = n.Type() + utils.RandomId(utils.ID8)
	}

	if config.Logger == nil {
		config.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	n.Config = config
	n.Channel = make(chan *model.Notification)
	n.jsonMarshal = json.Marshal
	n.httpNewRequest = http.NewRequest

	return n
}

func (n *WebhookNotifier) Type() string {
	return "webhook"
}

func (n *WebhookNotifier) Name() string {
	return n.Config.Name
}

func (n *WebhookNotifier) Connect() error {
	return nil
}

func (n *WebhookNotifier) Close() error {
	return nil
}

// Run starts receiving notifications
func (n *WebhookNotifier) Run() {
	for notification := range n.Channel {
		r := n.Deliver(notification)
		if !r.Success {
			n.Logger.Printf("%s: %+v", n.Name(), r)
		}
	}
}

// GetChannel returns the channel used by the worker
func (n *WebhookNotifier) GetChannel() chan *model.Notification {
	return n.Channel
}

// Notify sends a notification to worker
func (n *WebhookNotifier) Notify(payload *model.Notification) {
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

// Deliver sends a notification to the webhook
func (n *WebhookNotifier) Deliver(message *model.Notification) *model.Result {
	// Serialize the notification data to JSON
	payload, err := n.jsonMarshal(message)
	if err != nil {
		return &model.Result{Success: false, Error: err}
	}

	// Send the POST request to the webhook endpoint
	r, err := n.httpNewRequest(http.MethodPost, n.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return &model.Result{Success: false, Error: err}
	}

	// Ser headers
	r.Header.Set("Content-Type", "application/json")

	for k, v := range n.Headers {
		r.Header.Set(k, v)
	}

	client := n.getClient()

	resp, err := client.Do(r)
	if err != nil {
		return &model.Result{Success: false, Error: err}
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return &model.Result{Success: false, Error: fmt.Errorf("webhook returned non-OK status: %d", resp.StatusCode)}
	}

	// Read the response body if needed
	// responseBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return NotificationResult{Success: false, Error: err}
	// }

	return &model.Result{Success: true}
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
