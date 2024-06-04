package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/padiazg/notifier/connector/amqp"
	"github.com/padiazg/notifier/connector/dummy"
	"github.com/padiazg/notifier/connector/webhook"
	"github.com/padiazg/notifier/notification"
)

func TestNewEngine(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		checks []engineTestCheckFn
	}{
		{
			name:   "default",
			config: nil,
			checks: checkEngine(
				hasOnError(false),
				hasNotifiers(0),
			),
		},
		{
			name: "init-on-error",
			config: &Config{
				OnError: func(err error) {},
			},
			checks: checkEngine(
				hasOnError(true),
				hasNotifiers(0),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(tt.config)

			for _, c := range tt.checks {
				c(t, e)
			}
		})
	}
}

func TestEngine_RegisterNotifier(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		notifiers []notification.Notifier
		checks    []engineTestCheckFn
	}{
		{
			name: "one-notifier",
			notifiers: []notification.Notifier{
				webhook.NewWebhookNotifier(&webhook.Config{}),
			},
			checks: checkEngine(
				hasNotifiers(1),
			),
		},
		{
			name: "two-notifiers",
			notifiers: []notification.Notifier{
				webhook.NewWebhookNotifier(&webhook.Config{}),
				amqp.NewAMQPNotifier(&amqp.Config{Protocol: amqp.ProtocolAMQP10}),
			},
			checks: checkEngine(
				hasNotifiers(2),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(tt.config)

			for _, n := range tt.notifiers {
				e.RegisterNotifier(n)
			}

			for _, c := range tt.checks {
				c(t, e)
			}
		})
	}
}

func TestEngine_Start(t *testing.T) {
	tests := []struct {
		name      string
		notifiers []notification.Notifier
		checks    []engineTestCheckFn
	}{
		{
			name: "connect-error",
			notifiers: []notification.Notifier{
				&dummy.DummyNotifier{
					Config: &dummy.Config{
						Name:         "dummy-01",
						ConnectError: fmt.Errorf("connecting"),
					},
				},
			},
			checks: checkEngine(
				hasErrors(true),
			),
		},
		{
			name: "success",
			notifiers: []notification.Notifier{
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-01"}},
			},
			checks: checkEngine(
				hasErrors(false),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				c = &Config{OnError: registerError}
				e = NewEngine(c)
			)

			clearErrors()

			for _, n := range tt.notifiers {
				e.RegisterNotifier(n)
			}

			e.Start()
			time.Sleep(250 * time.Millisecond)
			e.Stop()
			time.Sleep(250 * time.Millisecond)

			for _, c := range tt.checks {
				c(t, e)
			}

		})
	}
}

func TestEngine_Dispatch(t *testing.T) {
	tests := []struct {
		name      string
		message   *notification.Notification
		checks    []notificationCheckFn
		notifiers []notification.Notifier
		// before  func(e *Engine)
	}{
		{
			name: "success-empty-message",
			notifiers: []notification.Notifier{
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-01"}},
			},
			message: nil,
		},
		{
			name: "success-empty-message-id",
			notifiers: []notification.Notifier{
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-01"}},
			},
			message: &notification.Notification{
				Event: notification.EventType("test"),
				Data: &notification.Result{
					Success: true,
					Error:   nil,
				},
			},
			checks: checkNotifications(
				notificationHasId(),
			),
		},
		{
			name: "success-one-notifier",
			notifiers: []notification.Notifier{
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-01"}},
			},
			message: &notification.Notification{
				ID:    "msg-01",
				Event: notification.EventType("test"),
				Data: &notification.Result{
					Success: true,
					Error:   nil,
				},
			},
			checks: checkNotifications(
				notificationReceived(),
			),
		},
		{
			name: "success-two-notifiers",
			notifiers: []notification.Notifier{
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-01"}},
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-02"}},
			},
			message: &notification.Notification{
				ID:    "msg-01",
				Event: notification.EventType("test"),
				Data: &notification.Result{
					Success: true,
					Error:   nil,
				},
			},
			checks: checkNotifications(
				notificationReceived(),
			),
		},
		{
			name: "success-two-notifiers-one-reveived",
			notifiers: []notification.Notifier{
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-01"}},
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-02"}},
			},
			message: &notification.Notification{
				ID:       "msg-01",
				Event:    notification.EventType("test"),
				Channels: []string{"dummy-02"},
				Data: &notification.Result{
					Success: true,
					Error:   nil,
				},
			},
			checks: checkNotifications(
				notificationReceived(),
			),
		},
		{
			name: "fail-missing-channel",
			notifiers: []notification.Notifier{
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-01"}},
				&dummy.DummyNotifier{Config: &dummy.Config{Name: "dummy-02"}},
			},
			message: &notification.Notification{
				ID:       "msg-01",
				Event:    notification.EventType("test"),
				Channels: []string{"dummy-03"},
				Data: &notification.Result{
					Success: true,
					Error:   nil,
				},
			},
			checks: checkNotifications(
				hasErrorsNotification(true),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				c = &Config{OnError: registerError}
				e = NewEngine(c)
			)

			clearErrors()

			for _, n := range tt.notifiers {
				e.RegisterNotifier(n)
			}

			e.Start()
			e.Dispatch(tt.message)
			time.Sleep(250 * time.Millisecond)
			e.Stop()

			for _, c := range tt.checks {
				c(t, e, tt.message)
			}
		})
	}
}
