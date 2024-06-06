package amqp

import (
	"reflect"
	"testing"

	n "github.com/padiazg/notifier/notification"
)

func TestNewAMQPNotifier(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name string
		args args
		want n.Notifier
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAMQPNotifier(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAMQPNotifier() = %v, want %v", got, tt.want)
			}
		})
	}
}
