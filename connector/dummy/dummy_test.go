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

	"github.com/padiazg/notifier/model"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "test-logger", log.LstdFlags)

		checkConfigSet = func() model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*DummyNotifier)
				assert.NotNilf(t, n.Config, "configSet Config expexted to be not nil")
			}
		}

		checkConfigName = func(name string) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*DummyNotifier)
				if name != "" {
					assert.Equalf(t, name, n.Config.Name, "checkConfigName = %s, expected %s", n.Config.Name, name)
					return
				}

				var (
					test = n.Type() + `[abcdef0-9]{8}`
					re   = regexp.MustCompile(test)
				)

				assert.Regexpf(t, re, n.Config.Name, "checkConfigName = %v doesn't appply for %v", n.Config.Name, test)
			}
		}

		checkLogger = func(mark string) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*DummyNotifier)

				if assert.NotEmptyf(t, n.Logger, "checkLogger Logger is empty, expected to be set") {
					n.Logger.Printf(mark)
					if got := buf.String(); !strings.Contains(got, mark) {
						t.Errorf("checkLogger log = %s, expected %s", got, mark)
					}
				}
			}
		}
	)

	tests := []struct {
		name   string
		config *Config
		checks []model.TestCheckNotifierFn
	}{
		{
			name:   "success-empty-config",
			config: nil,
			checks: model.CheckNotifier(
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
			checks: model.CheckNotifier(
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
	var (
		n    = New(&Config{})
		got  = n.Type()
		want = "dummy"
	)

	assert.Equalf(t, got, want, "Type() = %v, want %s", got, want)
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
			n := New(tt.config)
			if err := n.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("DummyNotifier.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDummyNotifier_Close(t *testing.T) {
	n := New(&Config{Name: "dummy-01"})
	err := n.Close()
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
	n := New(&Config{})
	got := n.GetChannel()
	assert.NotNilf(t, got, "DummyNotifier.GetChannel() = nil, want not nil")
	time.Sleep(10 * time.Millisecond)
	defer func() { close(n.Channel) }()
}

func TestDummyNotifier_Notify(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "", log.LstdFlags)
	)

	tests := []struct {
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
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantPanic {
					t.Errorf("Notify has panicked, expected not to")
				}
			}()

			n := &DummyNotifier{
				Config:  &Config{Logger: logger},
				Channel: tt.channel,
			}

			n.Notify(tt.payload)

			model.CheckLoggerError(&buf, tt.wantLog)

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

func TestWebhookNotifier_Run(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "test:", log.LstdFlags)

		tests = []struct {
			name    string
			config  *Config
			message *model.Notification
			checks  []model.TestCheckNotifierFn
		}{
			{
				name:   "fail",
				config: &Config{Logger: logger},
				message: &model.Notification{
					Data: "must-fail",
				},
				checks: []model.TestCheckNotifierFn{
					model.CheckLoggerError(&buf, "unexpected type"),
				},
			},
			{
				name:   "success",
				config: &Config{Logger: logger},
				message: &model.Notification{
					Data: &model.Result{Success: true},
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

			// start the runner
			go n.Run()
			time.Sleep(10 * time.Millisecond)

			// send a single message and close the channel
			go func() {
				n.Channel <- tt.message
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

func TestDummyNotifier_Exists(t *testing.T) {
	tests := []struct {
		name   string
		before func(d *DummyNotifier) *model.Notification
		want   bool
	}{
		{
			name: "found",
			before: func(d *DummyNotifier) *model.Notification {
				n := &model.Notification{
					ID:    "payload-01",
					Event: model.EventType("test"),
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
			before: func(d *DummyNotifier) *model.Notification {
				return &model.Notification{
					ID:    "payload-01",
					Event: model.EventType("test"),
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
				n *model.Notification
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
		before func(d *DummyNotifier) *model.Notification
		want   bool
	}{
		{
			name: "found",
			before: func(d *DummyNotifier) *model.Notification {
				n := &model.Notification{
					ID:    "payload-01",
					Event: model.EventType("test"),
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
			before: func(d *DummyNotifier) *model.Notification {
				return &model.Notification{
					ID:    "payload-01",
					Event: model.EventType("test"),
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
				n *model.Notification
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
		before func(d *DummyNotifier) []*model.Notification
	}{
		{
			name: "success-nil-notifications",
			before: func(d *DummyNotifier) []*model.Notification {
				var n []*model.Notification
				d.in = n
				return n
			},
		},
		{
			name: "success-empty-notifications",
			before: func(d *DummyNotifier) []*model.Notification {
				n := []*model.Notification{}
				d.in = n
				return n
			},
		},
		{
			name: "success-one-notification",
			before: func(d *DummyNotifier) []*model.Notification {
				n := []*model.Notification{
					{ID: "001", Event: model.EventType("test"), Data: "test"},
				}
				d.in = n
				return n
			},
		},
		{
			name: "success-two-notification",
			before: func(d *DummyNotifier) []*model.Notification {
				n := []*model.Notification{
					{ID: "001", Event: model.EventType("test"), Data: "test"},
					{ID: "002", Event: model.EventType("test"), Data: "test"},
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
				n []*model.Notification
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
