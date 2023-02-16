package errors

import (
	"fmt"
	"io"

	"github.com/mailgun/errors/callstack"
)

// Wrap wraps the error and attaches stack information to the error
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		stack:   callstack.New(1),
		wrapped: err,
		msg:     msg,
	}
}

// Wrapf is identical to Wrap but formats the error before wrapping.
func Wrapf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		stack:   callstack.New(1),
		wrapped: err,
		msg:     fmt.Sprintf(format, a...),
	}
}

type wrappedError struct {
	msg     string
	wrapped error
	stack   *callstack.CallStack
}

func (e *wrappedError) Unwrap() error {
	return e.wrapped
}

func (e *wrappedError) Is(target error) bool {
	_, ok := target.(*wrappedError)
	return ok
}

func (e *wrappedError) Error() string {
	if e.msg == NoMsg {
		return e.wrapped.Error()
	}
	return e.msg + ": " + e.wrapped.Error()
}

func (e *wrappedError) StackTrace() callstack.StackTrace {
	if child, ok := e.wrapped.(callstack.HasStackTrace); ok {
		return child.StackTrace()
	}
	return e.stack.StackTrace()
}

func (e *wrappedError) Format(s fmt.State, verb rune) {
	_, _ = io.WriteString(s, e.Error())
}
