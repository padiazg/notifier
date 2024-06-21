package notification

import "testing"

type TestCheckNotifierFn func(*testing.T, Notifier)
type TestCheckResultFn func(*testing.T, Notifier, *Result)

func CheckNotifier(fns ...TestCheckNotifierFn) []TestCheckNotifierFn {
	return fns
}

func CheckResult(fns ...TestCheckResultFn) []TestCheckResultFn {
	return fns
}
