package engine

import (
	"fmt"
	"testing"
	"time"

	amqp "github.com/padiazg/notifier/connector/amqp10"
	"github.com/padiazg/notifier/connector/dummy"
	"github.com/padiazg/notifier/connector/webhook"
	"github.com/padiazg/notifier/model"
	"github.com/stretchr/testify/assert"
)

type engineTestCheckFn func(*testing.T, *Engine)
type notificationCheckFn func(*testing.T, *Engine, *model.Notification)

var (
	checkEngine        = func(fns ...engineTestCheckFn) []engineTestCheckFn { return fns }
	checkNotifications = func(fns ...notificationCheckFn) []notificationCheckFn { return fns }
	errors             []error
)

func registerError(err error) {
	errors = append(errors, err)
}

func clearErrors() {
	errors = []error{}
}

func hasOnError(has bool) engineTestCheckFn {
	return func(t *testing.T, e *Engine) {
		t.Helper()
		if has {
			assert.NotNilf(t, e.OnError, "hasOnError errors expected, none produced")
		} else {
			assert.Nil(t, e.OnError, "hasOnError = [%+v], no errors expected", errors)
		}
	}
}

func hasNotifiers(count int) engineTestCheckFn {
	return func(t *testing.T, e *Engine) {
		t.Helper()
		q := len(e.notifiers)
		assert.Equalf(t, count, q, "hasNotifiers count=%d, expected %d", q, count)
	}
}

func hasErrors(has bool) engineTestCheckFn {
	return func(t *testing.T, e *Engine) {
		t.Helper()
		if has {
			assert.NotEmptyf(t, errors, "hasErrors errors expected, none produced")
		} else {
			assert.Emptyf(t, errors, "hasErrors = [%+v], no errors expected", errors)
		}
	}
}

func hasErrorsNotification(has bool) notificationCheckFn {
	return func(t *testing.T, e *Engine, n *model.Notification) {
		t.Helper()
		if has {
			assert.NotEmptyf(t, errors, "hasErrorsNotification errors expected, none produced")
		} else {
			assert.Emptyf(t, errors, "hasErrorsNotification = [%+v], no errors expected", errors)
		}
	}
}

func notificationReceived() notificationCheckFn {
	return func(t *testing.T, e *Engine, n *model.Notification) {
		t.Helper()
		if len(n.Channels) == 0 {
			for _, nt := range e.notifiers {
				assert.Truef(t, nt.(*dummy.DummyNotifier).Exists(n), "notificationReceived not found, expected %+v", n)
			}
		} else {
			for _, name := range n.Channels {
				nt, ok := e.notifiers[name]
				if !ok {
					t.Errorf("%s channel not found", name)
					continue
				}
				assert.Truef(t, nt.(*dummy.DummyNotifier).Exists(n), "notificationReceived not found, expected %+v", n)
			}
		}
	}
}

// call this checker with a single notifier and a single notification
func notificationHasId() notificationCheckFn {
	return func(t *testing.T, e *Engine, n *model.Notification) {
		t.Helper()
		for _, nt := range e.notifiers {
			data := nt.(*dummy.DummyNotifier).First()
			assert.NotNilf(t, data, "notificationHasId not found, expected at least one")
			if data != nil {
				assert.NotEmptyf(t, data.ID, "notificationHasId ID is empty, expected not empty")
			}
		}
	}
}

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
		notifiers []model.Notifier
		checks    []engineTestCheckFn
	}{
		{
			name: "one-notifier",
			notifiers: []model.Notifier{
				webhook.New(&webhook.Config{}),
			},
			checks: checkEngine(
				hasNotifiers(1),
			),
		},
		{
			name: "two-notifiers",
			notifiers: []model.Notifier{
				webhook.New(&webhook.Config{}),
				amqp.New(&amqp.Config{}),
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
		notifiers []model.Notifier
		checks    []engineTestCheckFn
	}{
		{
			name: "connect-error",
			notifiers: []model.Notifier{
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
			notifiers: []model.Notifier{
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
		message   *model.Notification
		checks    []notificationCheckFn
		notifiers []model.Notifier
		// before  func(e *Engine)
	}{
		{
			name: "success-empty-message",
			notifiers: []model.Notifier{
				dummy.New(&dummy.Config{Name: "dummy-01"}),
			},
			message: nil,
		},
		{
			name: "success-empty-message-id",
			notifiers: []model.Notifier{
				dummy.New(&dummy.Config{Name: "dummy-01"}),
			},
			message: &model.Notification{
				Event: model.EventType("test"),
				Data: &model.Result{
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
			notifiers: []model.Notifier{
				dummy.New(&dummy.Config{Name: "dummy-01"}),
			},
			message: &model.Notification{
				ID:    "msg-01",
				Event: model.EventType("test"),
				Data: &model.Result{
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
			notifiers: []model.Notifier{
				dummy.New(&dummy.Config{Name: "dummy-01"}),
				dummy.New(&dummy.Config{Name: "dummy-02"}),
			},
			message: &model.Notification{
				ID:    "msg-01",
				Event: model.EventType("test"),
				Data: &model.Result{
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
			notifiers: []model.Notifier{
				dummy.New(&dummy.Config{Name: "dummy-01"}),
				dummy.New(&dummy.Config{Name: "dummy-02"}),
			},
			message: &model.Notification{
				ID:       "msg-01",
				Event:    model.EventType("test"),
				Channels: []string{"dummy-02"},
				Data: &model.Result{
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
			notifiers: []model.Notifier{
				dummy.New(&dummy.Config{Name: "dummy-01"}),
				dummy.New(&dummy.Config{Name: "dummy-02"}),
			},
			message: &model.Notification{
				ID:       "msg-01",
				Event:    model.EventType("test"),
				Channels: []string{"dummy-03"},
				Data: &model.Result{
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
			time.Sleep(100 * time.Millisecond)
			e.Dispatch(tt.message)
			time.Sleep(100 * time.Millisecond)
			e.Stop()

			for _, c := range tt.checks {
				c(t, e, tt.message)
			}
		})
	}
}
