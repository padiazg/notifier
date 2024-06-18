package notification

import "testing"

type TestCheckNotifierFn func(*testing.T, Notifier)
type TestCheckResultFn func(*testing.T, Notifier, *Result)

var (
	CheckNotifier = func(fns ...TestCheckNotifierFn) []TestCheckNotifierFn { return fns }
	CheckResult   = func(fns ...TestCheckResultFn) []TestCheckResultFn { return fns }
)
