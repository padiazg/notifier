package engine

import (
	"testing"

	"github.com/padiazg/notifier/connector/dummy"
	"github.com/padiazg/notifier/notification"
	"github.com/stretchr/testify/assert"
)

type engineTestCheckFn func(*testing.T, *Engine)
type notificationCheckFn func(*testing.T, *Engine, *notification.Notification)

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
			assert.NotNil(t, e.OnError)
		} else {
			assert.Nil(t, e.OnError)
		}
	}
}

func hasNotifiers(count int) engineTestCheckFn {
	return func(t *testing.T, e *Engine) {
		t.Helper()
		assert.Equal(t, count, len(e.notifiers))
	}
}

func hasErrors(has bool) engineTestCheckFn {
	return func(t *testing.T, e *Engine) {
		t.Helper()
		if has {
			assert.NotEmpty(t, errors)
		} else {
			assert.Empty(t, errors)
		}
	}
}

func hasErrorsNotification(has bool) notificationCheckFn {
	return func(t *testing.T, e *Engine, n *notification.Notification) {
		t.Helper()
		if has {
			assert.NotEmpty(t, errors)
		} else {
			assert.Empty(t, errors)
		}
	}
}

func notificationReceived() notificationCheckFn {
	return func(t *testing.T, e *Engine, n *notification.Notification) {
		t.Helper()
		if len(n.Channels) == 0 {
			for _, nt := range e.notifiers {
				assert.True(t, nt.(*dummy.DummyNotifier).Exists(n))
			}
		} else {
			for _, name := range n.Channels {
				nt, ok := e.notifiers[name]
				if !ok {
					t.Errorf("%s channel not found", name)
					continue
				}
				assert.True(t, nt.(*dummy.DummyNotifier).Exists(n))
			}
		}
	}
}

// call this checker with a single notifier and a single notification
func notificationHasId() notificationCheckFn {
	return func(t *testing.T, e *Engine, n *notification.Notification) {
		t.Helper()
		for _, nt := range e.notifiers {
			data := nt.(*dummy.DummyNotifier).First()
			assert.NotNil(t, data)
			if data != nil {
				assert.NotEmpty(t, data.ID)
			}
		}
	}
}
