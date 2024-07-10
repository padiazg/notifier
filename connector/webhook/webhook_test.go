package webhook

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/padiazg/notifier/model"
	"github.com/stretchr/testify/assert"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func checkName(name string) model.TestCheckNotifierFn {
	return func(t *testing.T, np model.Notifier) {
		t.Helper()
		n, _ := np.(*WebhookNotifier)
		if name != "" {
			assert.Equalf(t, name, n.Name(), "checkName = %s, expected %s", n.Name(), name)
			return
		}

		var (
			test = n.Type() + `[abcdef0-9]{8}`
			re   = regexp.MustCompile(test)
		)

		assert.Regexpf(t, re, n.Name(), "checkName = %v doesn't appply for %v", n.Name(), test)
	}
}

func TestNew(t *testing.T) {
	var (
		checkConfigSet = func() model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*WebhookNotifier)
				assert.NotNilf(t, n.Config, "configSet Config expexted to be not nil")
			}
		}

		checkEndpoint = func(endpoint string) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*WebhookNotifier)
				assert.Equalf(t, endpoint, n.Config.Endpoint, "checkEndpoint = %s, expected %s", n.Config.Endpoint, endpoint)
			}
		}

		checkInsecure = func(insecure bool) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*WebhookNotifier)
				assert.Equalf(t, insecure, n.Config.Insecure, "checkInsecure = %t, expected %t", n.Config.Endpoint, insecure)
			}
		}

		checkHeaderExist = func(headerKey, headerValue string) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*WebhookNotifier)
				value, ok := n.Config.Headers[headerKey]
				if !ok {
					t.Errorf("checkHeaderExist %s header key not registered", headerKey)
				}

				assert.Equalf(t, headerValue, value, "checkInsecure = %t, expected %t", value, headerValue)
			}
		}

		checkChannel = func(wantPanic bool) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*WebhookNotifier)

				defer func() {
					if r := recover(); r != nil && !wantPanic {
						t.Errorf("chechChannel has panicked, expected not to")
					}
				}()

				go func() {
					close(n.Channel)
				}()

				time.Sleep(20 * time.Millisecond)
				select {
				case <-n.Channel:
				default:
					t.Errorf("chechChannel is not closed")
				}
			}
		}

		tests = []struct {
			name   string
			config *Config
			checks []model.TestCheckNotifierFn
		}{
			{
				name:   "success-empty-config",
				config: nil,
				checks: model.CheckNotifier(
					checkConfigSet(),
					checkName(""),
				),
			},
			{
				name: "success-full-config",
				config: &Config{
					Name:     "webhook-01",
					Endpoint: "http://localhost:8081/webhook",
					Insecure: true,
					Headers: map[string]string{
						"x-token-id": "abcdef",
						"y-feature":  "123456",
					},
				},
				checks: model.CheckNotifier(
					checkConfigSet(),
					checkName("webhook-01"),
					checkEndpoint("http://localhost:8081/webhook"),
					checkInsecure(true),
					checkHeaderExist("x-token-id", "abcdef"),
					checkHeaderExist("y-feature", "123456"),
					checkChannel(false),
				),
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(tt.config)
			for _, c := range tt.checks {
				c(t, n)
			}
		})
	}
}

func TestWebhookNotifier_Type(t *testing.T) {
	d := New(&Config{})
	got := d.Type()
	assert.Equalf(t, got, "webhook", "Type() = %v, want webhook", got)
}

func TestWebhookNotifier_Name(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		checks []model.TestCheckNotifierFn
	}{
		{
			name:   "success-no-name-set",
			config: nil,
			checks: model.CheckNotifier(
				checkName(""),
			),
		},
		{
			name:   "success-name-set",
			config: &Config{Name: "webhook-01"},
			checks: model.CheckNotifier(
				checkName("webhook-01"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(tt.config)
			for _, c := range tt.checks {
				c(t, n)
			}
		})
	}
}

func TestWebhookNotifier_Connect(t *testing.T) {
	d := New(&Config{})
	got := d.Connect()
	assert.Nilf(t, got, "Connect() = %v, want nil", got)
}

func TestWebhookNotifier_Close(t *testing.T) {
	d := New(&Config{})
	got := d.Close()
	assert.Nilf(t, got, "Close() = %v, want nil", got)
}

func TestWebhookNotifier_GetChannel(t *testing.T) {
	d := New(&Config{})
	defer func() { close(d.Channel) }()
	got := d.GetChannel()
	assert.NotNilf(t, got, "GetChannel() = nil, want not nil")
	time.Sleep(10 * time.Millisecond)
}

func TestWebhookNotifier_Notify(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "test:", log.LstdFlags)

		tests = []struct {
			name      string
			channel   chan *model.Notification
			payload   *model.Notification
			wantLog   string
			wantPanic bool
			wantValue bool
		}{
			{
				name:      "nil-channel",
				channel:   nil,
				payload:   nil,
				wantLog:   "channel is nil",
				wantPanic: false,
			},
			{
				name:      "nil-payload",
				channel:   make(chan *model.Notification, 1),
				payload:   nil,
				wantLog:   "payload is nil",
				wantPanic: false,
			},
			{
				name:      "valid-payload",
				channel:   make(chan *model.Notification, 1),
				payload:   &model.Notification{},
				wantLog:   "",
				wantPanic: false,
				wantValue: true,
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantPanic {
					t.Errorf("Notify has panicked, expected not to")
				}
			}()

			dn := &WebhookNotifier{
				Config:  &Config{Logger: logger},
				Channel: tt.channel,
			}

			dn.Notify(tt.payload)

			model.CheckLoggerError(&buf, tt.wantLog)

			if tt.wantValue {
				select {
				case value := <-tt.channel:
					assert.Equalf(t, tt.payload, value, "Notify payload = %+v, expected %+v", value, tt.payload)

				default:
					t.Errorf("Notify payload is empty, expected %+v", tt.payload)
				}
			}

			buf.Reset()
		})
	}
}

func TestWebhookNotifier_Deliver(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		message *model.Notification
		before  func(*WebhookNotifier)
		checks  []model.TestCheckResultFn
	}{
		{
			name:   "json-marshal-error",
			config: nil,
			before: func(n *WebhookNotifier) {
				n.jsonMarshal = func(_ any) ([]byte, error) { return nil, fmt.Errorf("error from json.Marshal") }
			},
			checks: model.CheckResult(
				model.CheckResultError("error from json.Marshal"),
			),
		},
		{
			name:   "http-newrequest-error",
			config: nil,
			before: func(n *WebhookNotifier) {
				n.httpNewRequest = func(_, _ string, _ io.Reader) (*http.Request, error) {
					return nil, fmt.Errorf("test error on http.NewRequest")
				}
			},
			checks: model.CheckResult(
				model.CheckResultError("test error on http.NewRequest"),
			),
		},
		{
			name: "client-do-error",
			before: func(n *WebhookNotifier) {
				n.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("test http new request error")
					},
				}
			},
			checks: model.CheckResult(
				model.CheckResultError("test http new request error"),
			),
		},
		{
			name:   "http-status-code-not-ok",
			config: &Config{Endpoint: "http://localhost:8080/webhook"},
			message: &model.Notification{
				Event: model.EventType("test-event"),
				Data:  "test-data",
			},
			before: func(n *WebhookNotifier) {
				n.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusForbidden,
							Body:       io.NopCloser(bytes.NewBufferString(`Forbidden`)),
						}, nil
					},
				}
			},
			checks: model.CheckResult(
				model.CheckResultError("webhook returned non-OK status: 403"),
			),
		},
		{
			name: "http-status-code-ok",
			config: &Config{
				Endpoint: "http://localhost:8080/webhook",
				Headers: map[string]string{
					"Header-XYZ": "xyz",
				},
			},
			message: &model.Notification{
				Event: model.EventType("test-event"),
				Data:  "test-data",
			},
			before: func(n *WebhookNotifier) {
				n.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(`Ok`)),
						}, nil
					},
				}
			},
			checks: model.CheckResult(
				model.CheckResultError(""),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(tt.config)

			if tt.before != nil {
				tt.before(n)
			}

			r := n.Deliver(tt.message)

			for _, c := range tt.checks {
				c(t, n, r)
			}
		})
	}
}

func TestWebhookNotifier_getClient(t *testing.T) {
	var (
		checkClientType = func(clientType interface{}) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*WebhookNotifier)
				expected := reflect.TypeOf(clientType)
				actual := reflect.TypeOf(n.client)
				assert.Equalf(t, expected, actual, "getClient type = %v, expected %v", actual, expected)
			}
		}

		checkClientInsecure = func(want bool) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				var (
					insecureSkipVerify bool
					n, _               = np.(*WebhookNotifier)
					transportType      = reflect.TypeOf(n.client.(*http.Client).Transport)
				)

				if transportType == reflect.TypeOf(&http.Transport{}) {
					insecureSkipVerify = n.client.(*http.Client).Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify
				}

				assert.Equalf(t, want, insecureSkipVerify, "getClient insecure = %t, expected %t", insecureSkipVerify, want)
			}
		}

		tests = []struct {
			name   string
			config *Config
			before func(*WebhookNotifier)
			checks []model.TestCheckNotifierFn
		}{
			{
				name: "existing-client",
				before: func(n *WebhookNotifier) {
					n.client = &mockHTTPClient{}
				},
				checks: model.CheckNotifier(
					checkClientType(&mockHTTPClient{}),
				),
			},
			{
				name: "empty-client",
				before: func(n *WebhookNotifier) {
					n.client = nil
				},
				checks: model.CheckNotifier(
					checkClientType(&http.Client{}),
					checkClientInsecure(false),
				),
			},
			{
				name:   "insecure-client",
				config: &Config{Insecure: true},
				before: func(n *WebhookNotifier) {
					n.client = nil
				},
				checks: model.CheckNotifier(
					checkClientType(&http.Client{}),
					checkClientInsecure(true),
				),
			},
		}
	)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(tt.config)

			if tt.before != nil {
				tt.before(n)
			}

			_ = n.getClient()

			for _, c := range tt.checks {
				c(t, n)
			}
		})
	}
}

func TestWebhookNotifier_Run(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "test:", log.LstdFlags)

		tests = []struct {
			name    string
			config  *Config
			message *model.Notification
			before  func(*WebhookNotifier)
			checks  []model.TestCheckNotifierFn
		}{
			{
				name:   "fail-forbidden",
				config: &Config{Logger: logger},
				before: func(n *WebhookNotifier) {
					n.client = &mockHTTPClient{
						DoFunc: func(req *http.Request) (*http.Response, error) {
							return &http.Response{
								StatusCode: http.StatusForbidden,
								Body:       io.NopCloser(bytes.NewBufferString(`Forbidden`)),
							}, nil
						},
					}
				},
				checks: []model.TestCheckNotifierFn{
					model.CheckLoggerError(&buf, "webhook returned non-OK status: 403"),
				},
			},
			{
				name:   "success",
				config: &Config{Logger: logger},
				before: func(n *WebhookNotifier) {
					n.client = &mockHTTPClient{
						DoFunc: func(req *http.Request) (*http.Response, error) {
							return &http.Response{
								StatusCode: http.StatusOK,
								Body:       io.NopCloser(bytes.NewBufferString(`Ok`)),
							}, nil
						},
					}
				},
				checks: []model.TestCheckNotifierFn{
					model.CheckLoggerError(&buf, ""),
				},
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(tt.config)

			// set mock client
			if tt.before != nil {
				tt.before(n)
			}

			// start the runner
			go n.Run()
			time.Sleep(10 * time.Millisecond)

			// send a single message and close the channel
			go func() {
				n.Channel <- &model.Notification{Data: "Test message"}
				close(n.Channel)
			}()
			time.Sleep(10 * time.Millisecond)

			for _, c := range tt.checks {
				c(t, n)
			}

			buf.Reset()
		})
	}
}
