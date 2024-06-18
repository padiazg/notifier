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
	"strings"
	"testing"
	"time"

	"github.com/padiazg/notifier/notification"
	"github.com/stretchr/testify/assert"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func checkName(name string) notification.TestCheckNotifierFn {
	return func(t *testing.T, n notification.Notifier) {
		t.Helper()
		wn, _ := n.(*WebhookNotifier)
		if name != "" {
			assert.Equalf(t, name, wn.Name(), "checkName = %s, expected %s", wn.Name(), name)
			return
		}

		var (
			test = wn.Type() + `[abcdef0-9]{8}`
			re   = regexp.MustCompile(test)
		)

		assert.Regexpf(t, re, wn.Name(), "checkName = %v doesn't appply for %v", wn.Name(), test)
	}
}

func TestNew(t *testing.T) {
	var (
		checkConfigSet = func() notification.TestCheckNotifierFn {
			return func(t *testing.T, n notification.Notifier) {
				t.Helper()
				wn, _ := n.(*WebhookNotifier)
				assert.NotNilf(t, wn.Config, "configSet Config expexted to be not nil")
			}
		}

		checkEndpoint = func(endpoint string) notification.TestCheckNotifierFn {
			return func(t *testing.T, n notification.Notifier) {
				t.Helper()
				wn, _ := n.(*WebhookNotifier)
				assert.Equalf(t, endpoint, wn.Config.Endpoint, "checkEndpoint = %s, expected %s", wn.Config.Endpoint, endpoint)
			}
		}

		checkInsecure = func(insecure bool) notification.TestCheckNotifierFn {
			return func(t *testing.T, n notification.Notifier) {
				t.Helper()
				wn, _ := n.(*WebhookNotifier)
				assert.Equalf(t, insecure, wn.Config.Insecure, "checkInsecure = %t, expected %t", wn.Config.Endpoint, insecure)
			}
		}

		checkHeaderExist = func(headerKey, headerValue string) notification.TestCheckNotifierFn {
			return func(t *testing.T, n notification.Notifier) {
				t.Helper()
				wn, _ := n.(*WebhookNotifier)
				value, ok := wn.Config.Headers[headerKey]
				if !ok {
					t.Errorf("checkHeaderExist %s header key not registered", headerKey)
				}

				assert.Equalf(t, headerValue, value, "checkInsecure = %t, expected %t", value, headerValue)
			}
		}

		tests = []struct {
			name   string
			config *Config
			checks []notification.TestCheckNotifierFn
		}{
			{
				name:   "success-empty-config",
				config: nil,
				checks: notification.CheckNotifier(
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
				checks: notification.CheckNotifier(
					checkConfigSet(),
					checkName("webhook-01"),
					checkEndpoint("http://localhost:8081/webhook"),
					checkInsecure(true),
					checkHeaderExist("x-token-id", "abcdef"),
					checkHeaderExist("y-feature", "123456"),
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
	assert.Equalf(t, got, "webhook", "Type = %v, want webhook", got)
}

func TestWebhookNotifier_Connect(t *testing.T) {
	d := New(&Config{})
	got := d.Connect()
	assert.Nilf(t, got, "Connect = %v, want nil", got)
}

func TestWebhookNotifier_Close(t *testing.T) {
	d := New(&Config{})
	got := d.Close()
	assert.Nilf(t, got, "Close = %v, want nil", got)
}

func TestWebhookNotifier_Name(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		checks []notification.TestCheckNotifierFn
	}{
		{
			name:   "success-no-name-set",
			config: nil,
			checks: notification.CheckNotifier(
				checkName(""),
			),
		},
		{
			name:   "success-name-set",
			config: &Config{Name: "webhook-01"},
			checks: notification.CheckNotifier(
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

func TestWebhookNotifier_GetChannel(t *testing.T) {
	d := New(&Config{})
	go d.StartWorker()
	time.Sleep(100 * time.Millisecond)
	got := d.GetChannel()
	assert.NotNilf(t, got, "WebhookNotifier.GetChannel() = nil, want not nil")
	time.Sleep(100 * time.Millisecond)
	defer func() { close(d.Channel) }()
}

// TODO: replicate TestDummyNotifier_Notify approach
func TestWebhookNotifier_Notify(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "", log.LstdFlags)

		tests = []struct {
			name      string
			channel   chan *notification.Notification
			payload   *notification.Notification
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
				channel:   make(chan *notification.Notification, 1),
				payload:   nil,
				wantLog:   "payload is nil",
				wantPanic: false,
			},
			{
				name:      "valid-payload",
				channel:   make(chan *notification.Notification, 1),
				payload:   &notification.Notification{},
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

			if gotLog := buf.String(); !strings.Contains(gotLog, tt.wantLog) {
				t.Errorf("Notify log = %s, expected %s", gotLog, tt.wantLog)
			}

			if tt.wantValue {
				select {
				case value := <-tt.channel:
					if !reflect.DeepEqual(value, tt.payload) {
						t.Errorf("Notify payload = %+v, expected %+v", value, tt.payload)
					}

				default:
					t.Errorf("Notify payload is empty, expected %+v", tt.payload)
				}
			}

			buf.Reset()
		})
	}
}

func TestWebhookNotifier_SendNotification(t *testing.T) {
	var (
		orig_jsonMarshal    = jsonMarshal
		orig_httpNewRequest = httpNewRequest

		checkNotifyError = func(want bool, msg string) notification.TestCheckResultFn {
			return func(t *testing.T, n notification.Notifier, r *notification.Result) {
				t.Helper()
				hasError := want && (r.Error != nil)
				assert.Equalf(t, want, hasError, "checkNotifyError = %v, expected %t", hasError, want)
				if msg != "" {
					errorMsg := r.Error.Error()
					assert.Equalf(t, msg, errorMsg, "checkNotifyError text=%s, want %s", errorMsg, msg)
				}
			}
		}

		tests = []struct {
			name    string
			config  *Config
			message *notification.Notification
			before  func(*WebhookNotifier)
			checks  []notification.TestCheckResultFn
		}{
			{
				name:   "json-marshal-error",
				config: nil,
				before: func(_ *WebhookNotifier) {
					jsonMarshal = func(_ any) ([]byte, error) { return nil, fmt.Errorf("test error on json.Marshal") }
				},
				checks: notification.CheckResult(
					checkNotifyError(true, "test error on json.Marshal"),
				),
			},
			{
				name:   "http-newrequest-error",
				config: nil,
				before: func(_ *WebhookNotifier) {
					httpNewRequest = func(_, _ string, _ io.Reader) (*http.Request, error) {
						return nil, fmt.Errorf("test error on http.NewRequest")
					}
				},
				checks: notification.CheckResult(
					checkNotifyError(true, "test error on http.NewRequest"),
				),
			},
			{
				name: "client-do-error",
				before: func(wn *WebhookNotifier) {
					wn.client = &mockHTTPClient{
						DoFunc: func(req *http.Request) (*http.Response, error) {
							return nil, errors.New("test http new request error")
						},
					}
				},
				checks: notification.CheckResult(
					checkNotifyError(true, "test http new request error"),
				),
			},
			{
				name:   "http-status-code-not-ok",
				config: &Config{Endpoint: "http://localhost:8080/webhook"},
				message: &notification.Notification{
					Event: notification.EventType("test-event"),
					Data:  "test-data",
				},
				before: func(wn *WebhookNotifier) {
					wn.client = &mockHTTPClient{
						DoFunc: func(req *http.Request) (*http.Response, error) {
							return &http.Response{
								StatusCode: http.StatusForbidden,
								Body:       io.NopCloser(bytes.NewBufferString(`Forbidden`)),
							}, nil
						},
					}
				},
				checks: notification.CheckResult(
					checkNotifyError(true, "webhook returned non-OK status: 403"),
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
				message: &notification.Notification{
					Event: notification.EventType("test-event"),
					Data:  "test-data",
				},
				before: func(wn *WebhookNotifier) {
					wn.client = &mockHTTPClient{
						DoFunc: func(req *http.Request) (*http.Response, error) {

							return &http.Response{
								StatusCode: http.StatusOK,
								Body:       io.NopCloser(bytes.NewBufferString(`Ok`)),
							}, nil
						},
					}
				},
				checks: notification.CheckResult(
					checkNotifyError(false, ""),
				),
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				jsonMarshal = orig_jsonMarshal
				httpNewRequest = orig_httpNewRequest
			}()

			wn := New(tt.config)

			if tt.before != nil {
				tt.before(wn)
			}

			r := wn.SendNotification(tt.message)

			for _, c := range tt.checks {
				c(t, wn, r)
			}
		})
	}
}

func TestWebhookNotifier_getClient(t *testing.T) {
	var checkClientType = func(clientType interface{}) notification.TestCheckNotifierFn {
		return func(t *testing.T, n notification.Notifier) {
			wn, _ := n.(*WebhookNotifier)
			expected := reflect.TypeOf(clientType)
			actual := reflect.TypeOf(wn.client)
			assert.Equalf(t, expected, actual, "getClient type = %v, expected %v", actual, expected)
		}
	}

	var checkClientInsecure = func(want bool) notification.TestCheckNotifierFn {
		return func(t *testing.T, n notification.Notifier) {
			var (
				insecureSkipVerify bool
				wn, _              = n.(*WebhookNotifier)
				transportType      = reflect.TypeOf(wn.client.(*http.Client).Transport)
			)

			if transportType == reflect.TypeOf(&http.Transport{}) {
				insecureSkipVerify = wn.client.(*http.Client).Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify
			}

			assert.Equalf(t, want, insecureSkipVerify, "getClient insecure = %t, expected %t", insecureSkipVerify, want)
		}
	}

	tests := []struct {
		name   string
		config *Config
		before func(*WebhookNotifier)
		checks []notification.TestCheckNotifierFn
	}{
		{
			name: "existing-client",
			before: func(wn *WebhookNotifier) {
				wn.client = &mockHTTPClient{}
			},
			checks: notification.CheckNotifier(
				checkClientType(&mockHTTPClient{}),
			),
		},
		{
			name: "empty-client",
			before: func(wn *WebhookNotifier) {
				wn.client = nil
			},
			checks: notification.CheckNotifier(
				checkClientType(&http.Client{}),
				checkClientInsecure(false),
			),
		},
		{
			name:   "insecure-client",
			config: &Config{Insecure: true},
			before: func(wn *WebhookNotifier) {
				wn.client = nil
			},
			checks: notification.CheckNotifier(
				checkClientType(&http.Client{}),
				checkClientInsecure(true),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			wn := New(tt.config)

			if tt.before != nil {
				tt.before(wn)
			}

			_ = wn.getClient()

			for _, c := range tt.checks {
				c(t, wn)
			}
		})
	}
}

// func TestWebhookNotifier_StartWorker(t *testing.T) {
// 	var (
// 		d       = New(&Config{})
// 		c       = d.GetChannel()
// 		payload = &notification.Notification{ID: "12345"}
// 	)

// 	go d.StartWorker()
// 	time.Sleep(100 * time.Millisecond)

// 	// case 1: valid payload
// 	go d.Notify(payload)

// 	receivedNotification := <-c

// 	time.Sleep(100 * time.Millisecond)

// 	defer func() { close(d.Channel) }()
// }
