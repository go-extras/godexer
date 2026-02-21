// Package shared holds shared types for the godexer CLI commands.
package shared

import "fmt"

// Context carries global CLI state (flags set on the root command).
type Context struct{}

// ExitError is an error that carries a process exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

// NewExitError creates an ExitError with the given code and error.
func NewExitError(code int, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

// NewExitErrorf creates an ExitError with the given code and formatted message.
func NewExitErrorf(code int, format string, args ...any) *ExitError {
	return &ExitError{Code: code, Err: fmt.Errorf(format, args...)}
}
