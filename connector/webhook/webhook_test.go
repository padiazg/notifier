package webhook

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCheckFn func(*testing.T, *WebhookNotifier)

var (
	check = func(fns ...testCheckFn) []testCheckFn { return fns }
)

func checkConfigSet() testCheckFn {
	return func(t *testing.T, wn *WebhookNotifier) {
		t.Helper()
		assert.NotNilf(t, wn.Config, "configSet Config expexted to be not nil")
	}
}

func checkName(name string) testCheckFn {
	return func(t *testing.T, wn *WebhookNotifier) {
		t.Helper()
		if name != "" {
			assert.Equalf(t, name, wn.Config.Name, "checkName = %s, expected %s", wn.Config.Name, name)
			return
		}

		var (
			test = wn.Type() + `[abcdef0-9]{8}`
			re   = regexp.MustCompile(test)
		)

		assert.Regexpf(t, re, wn.Config.Name, "checkName() = %v doesn't appply for %v", wn.Config.Name, test)
	}
}

func checkEndpoint(endpoint string) testCheckFn {
	return func(t *testing.T, wn *WebhookNotifier) {
		t.Helper()
		assert.Equalf(t, endpoint, wn.Config.Endpoint, "checkEndpoint = %s, expected %s", wn.Config.Endpoint, endpoint)
	}
}

func checkInsecure(insecure bool) testCheckFn {
	return func(t *testing.T, wn *WebhookNotifier) {
		t.Helper()
		assert.Equalf(t, insecure, wn.Config.Insecure, "checkInsecure = %t, expected %t", wn.Config.Endpoint, insecure)
	}
}

func checkHeaderExist(headerKey, headerValue string) testCheckFn {
	return func(t *testing.T, wn *WebhookNotifier) {
		t.Helper()
		value, ok := wn.Config.Headers[headerKey]
		if !ok {
			t.Errorf("checkHeaderExist %s header key not registered", headerKey)
		}

		assert.Equalf(t, headerValue, value, "checkInsecure = %t, expected %t", value, headerValue)
	}
}

func TestNewWebhookNotifier(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		checks []testCheckFn
	}{
		{
			name:   "success-empty-config",
			config: nil,
			checks: check(
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
			checks: check(
				checkConfigSet(),
				checkName("webhook-01"),
				checkEndpoint("http://localhost:8081/webhook"),
				checkInsecure(true),
				checkHeaderExist("x-token-id", "abcdef"),
				checkHeaderExist("y-feature", "123456"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := NewWebhookNotifier(tt.config)
			for _, c := range tt.checks {
				c(t, n)
			}
		})
	}
}
