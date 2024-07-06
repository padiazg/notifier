package notification

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCheckNotifierFn func(*testing.T, Notifier)
type TestCheckResultFn func(*testing.T, Notifier, *Result)

func CheckNotifier(fns ...TestCheckNotifierFn) []TestCheckNotifierFn {
	return fns
}

func CheckResult(fns ...TestCheckResultFn) []TestCheckResultFn {
	return fns
}

func CheckLoggerError(buf *bytes.Buffer, want string) TestCheckNotifierFn {
	return func(t *testing.T, n Notifier) {
		var (
			logMsg  = buf.String()
			wantLog = want != ""
			hasLog  = logMsg != ""
		)

		t.Helper()

		assert.Equalf(t, wantLog, hasLog, "checkLoggerError = %v, expected %t", hasLog, wantLog)

		if wantLog {
			assert.Containsf(t, logMsg, want, "checkLoggerError text=%s, want %s", logMsg, want)
		}
	}
}

func CheckResultError(want string) TestCheckResultFn {
	return func(t *testing.T, np Notifier, r *Result) {
		var (
			wantError = want != ""
			hasError  = r.Error != nil
		)

		t.Helper()

		assert.Equalf(t, wantError, hasError, "checkResultError = %v, expected %t", hasError, wantError)
		if wantError {
			if assert.NotNil(t, r.Error, "Error is nil, expected not to") {
				errorMsg := r.Error.Error()
				assert.Containsf(t, errorMsg, want, "checkResultError text=%s, want %s", errorMsg, want)
			}
		}
	}
}
