package dummy

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/padiazg/notifier/notification"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "test-logger", log.LstdFlags)

		checkConfigSet = func() notification.TestCheckNotifierFn {
			return func(t *testing.T, n notification.Notifier) {
				t.Helper()
				dn, _ := n.(*DummyNotifier)
				assert.NotNilf(t, dn.Config, "configSet Config expexted to be not nil")
			}
		}

		checkConfigName = func(name string) notification.TestCheckNotifierFn {
			return func(t *testing.T, n notification.Notifier) {
				t.Helper()
				dn, _ := n.(*DummyNotifier)
				if name != "" {
					assert.Equalf(t, name, dn.Config.Name, "checkConfigName = %s, expected %s", dn.Config.Name, name)
					return
				}

				var (
					test = dn.Type() + `[abcdef0-9]{8}`
					re   = regexp.MustCompile(test)
				)

				assert.Regexpf(t, re, dn.Config.Name, "checkConfigName = %v doesn't appply for %v", dn.Config.Name, test)
			}
		}

		checkLogger = func(mark string) notification.TestCheckNotifierFn {
			return func(t *testing.T, n notification.Notifier) {
				dn, _ := n.(*DummyNotifier)

				dn.Logger.Printf(mark)
				if got := buf.String(); !strings.Contains(got, mark) {
					t.Errorf("Notify log = %s, expected %s", got, mark)
				}
			}
		}
	)

	tests := []struct {
		name   string
		config *Config
		checks []notification.TestCheckNotifierFn
	}{
		{
			name:   "success-empty-config",
			config: nil,
			checks: notification.CheckNotifier(
				checkConfigSet(),
				checkConfigName(""),
				checkLogger(""),
			),
		},
		{
			name: "success-full-config",
			config: &Config{
				Name:   "dummy-test",
				Logger: logger,
			},
			checks: notification.CheckNotifier(
				checkConfigSet(),
				checkConfigName("dummy-test"),
				checkLogger("test-logger-a"),
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

			buf.Reset()
		})
	}
}

func TestDummyNotifier_Type(t *testing.T) {
	d := New(&Config{})
	got := d.Type()
	assert.Equalf(t, got, "dummy", "DummyNotifier.Type() = %v, want dummy", got)
}

func TestDummyNotifier_Connect(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "fail",
			config: &Config{
				Name:         "dummy-01",
				ConnectError: fmt.Errorf("test"),
			},
			wantErr: true,
		},
		{
			name: "success",
			config: &Config{
				Name:         "dummy-01",
				ConnectError: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			d := New(tt.config)
			if err := d.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("DummyNotifier.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDummyNotifier_Close(t *testing.T) {
	d := New(&Config{Name: "dummy-01"})
	err := d.Close()
	assert.NoErrorf(t, err, "Close error = %+v, no error expected", err)
}

func TestDummyNotifier_Name(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name:   "notifier-a",
			config: &Config{Name: "notifier-a"},
			want:   "notifier-a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DummyNotifier{Config: tt.config}
			if got := d.Name(); got != tt.want {
				t.Errorf("DummyNotifier.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyNotifier_GetChannel(t *testing.T) {
	d := New(&Config{})
	go d.StartWorker()
	time.Sleep(100 * time.Millisecond)
	got := d.GetChannel()
	assert.NotNilf(t, got, "DummyNotifier.GetChannel() = nil, want not nil")
	time.Sleep(100 * time.Millisecond)
	defer func() { close(d.Channel) }()
}

func TestDummyNotifier_Notify(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "", log.LstdFlags)
	)

	tests := []struct {
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
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantPanic {
					t.Errorf("Notify has panicked, expected not to")
				}
			}()

			dn := &DummyNotifier{
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

func TestDummyNotifier_Exists(t *testing.T) {
	tests := []struct {
		name   string
		before func(d *DummyNotifier) *notification.Notification
		want   bool
	}{
		{
			name: "found",
			before: func(d *DummyNotifier) *notification.Notification {
				n := &notification.Notification{
					ID:    "payload-01",
					Event: notification.EventType("test"),
					Data:  nil,
				}
				d.lock.Lock()
				defer d.lock.Unlock()
				d.in = append(d.in, n)

				return n
			},
			want: true,
		},
		{
			name: "not-found",
			before: func(d *DummyNotifier) *notification.Notification {
				return &notification.Notification{
					ID:    "payload-01",
					Event: notification.EventType("test"),
					Data:  nil,
				}
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				d = New(&Config{})
				n *notification.Notification
			)

			if tt.before != nil {
				n = tt.before(d)
			}

			if got := d.Exists(n); got != tt.want {
				t.Errorf("DummyNotifier.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyNotifier_First(t *testing.T) {
	tests := []struct {
		name   string
		before func(d *DummyNotifier) *notification.Notification
		want   bool
	}{
		{
			name: "found",
			before: func(d *DummyNotifier) *notification.Notification {
				n := &notification.Notification{
					ID:    "payload-01",
					Event: notification.EventType("test"),
					Data:  nil,
				}
				d.lock.Lock()
				defer d.lock.Unlock()
				d.in = append(d.in, n)

				return n
			},
			want: true,
		},
		{
			name: "not-found",
			before: func(d *DummyNotifier) *notification.Notification {
				return &notification.Notification{
					ID:    "payload-01",
					Event: notification.EventType("test"),
					Data:  nil,
				}
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				d = New(&Config{})
				n *notification.Notification
			)

			if tt.before != nil {
				n = tt.before(d)
			}

			got := d.First()
			if (got == n) != tt.want {
				t.Errorf("DummyNotifier.First() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyNotifier_In(t *testing.T) {
	tests := []struct {
		name   string
		before func(d *DummyNotifier) []*notification.Notification
	}{
		{
			name: "success-nil-notifications",
			before: func(d *DummyNotifier) []*notification.Notification {
				var n []*notification.Notification
				d.in = n
				return n
			},
		},
		{
			name: "success-empty-notifications",
			before: func(d *DummyNotifier) []*notification.Notification {
				n := []*notification.Notification{}
				d.in = n
				return n
			},
		},
		{
			name: "success-one-notification",
			before: func(d *DummyNotifier) []*notification.Notification {
				n := []*notification.Notification{
					{ID: "001", Event: notification.EventType("test"), Data: "test"},
				}
				d.in = n
				return n
			},
		},
		{
			name: "success-two-notification",
			before: func(d *DummyNotifier) []*notification.Notification {
				n := []*notification.Notification{
					{ID: "001", Event: notification.EventType("test"), Data: "test"},
					{ID: "002", Event: notification.EventType("test"), Data: "test"},
				}
				d.in = n
				return n
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				d = New(&Config{})
				n []*notification.Notification
			)

			if tt.before != nil {
				n = tt.before(d)
			}

			if got := d.In(); !reflect.DeepEqual(got, n) {
				t.Errorf("DummyNotifier.In() = %v, want %v", got, n)
			}
		})
	}
}
