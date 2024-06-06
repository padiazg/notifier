package dummy

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/padiazg/notifier/notification"
	"github.com/stretchr/testify/assert"
)

func TestDummyNotifier_Connect(t *testing.T) {
	type fields struct {
		Config *Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "fail",
			fields: fields{
				Config: &Config{
					Name:         "dummy-01",
					ConnectError: fmt.Errorf("test"),
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				Config: &Config{
					Name:         "dummy-01",
					ConnectError: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(tt.fields.Config)
			if err := d.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("DummyNotifier.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDummyNotifier_Close(t *testing.T) {
	type fields struct {
		Config *Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Config: &Config{Name: "dummy-01"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(tt.fields.Config)
			if err := d.Close(); (err != nil) != tt.wantErr {
				t.Errorf("DummyNotifier.Close() error = %v, wantErr %v", err, tt.wantErr)
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
	type fields struct {
		Config *Config
	}
	type args struct {
		payload *notification.Notification
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "fail-empty-payload",
			fields: fields{
				Config: &Config{Name: "dummy-01"},
			},
			args: args{
				payload: nil,
			},
		},
		{
			name: "fail-unknown-payload",
			fields: fields{
				Config: &Config{Name: "dummy-01"},
			},
			args: args{
				payload: &notification.Notification{
					ID:    "payload-01",
					Event: notification.EventType("test"),
					Data:  "wrong-type",
				},
			},
		},
		{
			name: "success",
			fields: fields{
				Config: &Config{Name: "dummy-01"},
			},
			args: args{
				payload: &notification.Notification{
					ID:    "payload-01",
					Event: notification.EventType("test"),
					Data:  &notification.Result{Success: true},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			d := New(tt.fields.Config)
			go d.StartWorker()
			time.Sleep(100 * time.Millisecond)
			d.Notify(tt.args.payload)
			time.Sleep(250 * time.Millisecond)
			defer close(d.Channel)
			// TODO: test here
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

func TestDummyNotifier_Type(t *testing.T) {
	d := New(&Config{})
	got := d.Type()
	assert.Equalf(t, got, "dummy", "DummyNotifier.Type() = %v, want dummy", got)
}
