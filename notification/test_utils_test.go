package notification

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckNotifier(t *testing.T) {
	var (
		testFn = func(*testing.T, Notifier) {}

		tests = []struct {
			name string
			fns  []TestCheckNotifierFn
			want []TestCheckNotifierFn
		}{
			{
				name: "nil-list",
				fns:  nil,
				want: nil,
			},
			{
				name: "empty-list",
				fns:  []TestCheckNotifierFn{},
				want: []TestCheckNotifierFn{},
			},
			{
				name: "single-element",
				fns:  []TestCheckNotifierFn{testFn},
				want: []TestCheckNotifierFn{testFn},
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckNotifier(tt.fns...)
			assert.Equal(t, tt.fns, got)
		})
	}
}

func TestCheckResult(t *testing.T) {
	var (
		testFn = func(*testing.T, Notifier, *Result) {}

		tests = []struct {
			name string
			fns  []TestCheckResultFn
			want []TestCheckResultFn
		}{
			{
				name: "nil-list",
				fns:  nil,
				want: nil,
			},
			{
				name: "empty-list",
				fns:  []TestCheckResultFn{},
				want: []TestCheckResultFn{},
			},
			{
				name: "single-element",
				fns:  []TestCheckResultFn{testFn},
				want: []TestCheckResultFn{testFn},
			},
		}
	)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckResult(tt.fns...)
			assert.Equal(t, tt.fns, got)
		})
	}
}

func TestCheckLoggerError(t *testing.T) {
	tests := []struct {
		name string
		buf  *bytes.Buffer
		want string
	}{
		{
			name: "success",
			buf:  &bytes.Buffer{},
			want: "",
		},
		{
			name: "error",
			buf:  bytes.NewBufferString(`test-error`),
			want: "test-error",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			CheckLoggerError(tt.buf, tt.want)(t, nil)
		})
	}
}

func TestCheckResultError(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		result *Result
	}{
		{
			name: "success",
			want: "",
			result: &Result{
				Error:   nil,
				Success: true,
			},
		},
		{
			name: "error",
			want: "test-error",
			result: &Result{
				Error:   fmt.Errorf("test-error"),
				Success: false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			CheckResultError(tt.want)(t, nil, tt.result)
		})
	}
}
