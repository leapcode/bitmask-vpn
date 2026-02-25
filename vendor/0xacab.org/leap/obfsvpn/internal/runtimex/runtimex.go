// Package runtimex contains runtime checks and utilities.
package runtimex

import "log"

// logFatal is a variable that can be overridden in tests to avoid calling os.Exit.
var logFatal = log.Fatal

// AssertNotErrorOrPanic calls panic if the given error is not nil.
func AssertNotErrorOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// AssertNotErrorWithReturnOrPanic is like [AssertNotErrorOrPanic] but
// also allows to return a specific value.
func AssertNotErrorWithReturnOrPanic[T any](v T, err error) T {
	AssertNotErrorOrPanic(err)
	return v
}

// AssertTrueOrExit calls [log.Fatal] if the given condition is false.
func AssertTrueOrExit(condition bool, v ...any) {
	if !condition {
		logFatal(v...)
	}
}

// AssertNotErrorOrExit calls [log.Fatal] if the given error is not nil.
func AssertNotErrorOrExit(err error) {
	if err != nil {
		logFatal(err)
	}
}
