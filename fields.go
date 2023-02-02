package errors

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/mailgun/errors/callstack"
	"github.com/sirupsen/logrus"
)

// HasFields Implement this interface to pass along unstructured context to the logger.
// It is the responsibility of Fields() implementation to unwrap the error chain and
// collect all errors that have `Fields()` defined.
type HasFields interface {
	Fields() map[string]interface{}
}

// HasFormat True if the interface has the format method (from fmt package)
type HasFormat interface {
	Format(st fmt.State, verb rune)
}

// WithFields Creates errors that conform to the `HasFields` interface
type WithFields map[string]interface{}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is call, and the format specifier.
// If err is nil, Wrapf returns nil.
func (f WithFields) Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &withFields{
		stack:   callstack.New(1),
		fields:  f,
		wrapped: err,
		msg:     fmt.Sprintf(format, args...),
	}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func (f WithFields) Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &withFields{
		stack:   callstack.New(1),
		fields:  f,
		wrapped: err,
		msg:     msg,
	}
}

// WithStack returns an error annotating err with a stack trace
// at the point WithStack is called
// If err is nil, WithStack returns nil.
func (f WithFields) WithStack(err error) error {
	if err == nil {
		return nil
	}
	return &withFields{
		stack:   callstack.New(1),
		fields:  f,
		wrapped: err,
	}
}

func (f WithFields) Error(msg string) error {
	return &withFields{
		stack:   callstack.New(1),
		fields:  f,
		wrapped: errors.New(msg),
		msg:     "",
	}
}

func (f WithFields) Errorf(format string, args ...interface{}) error {
	return &withFields{
		stack:   callstack.New(1),
		fields:  f,
		wrapped: fmt.Errorf(format, args...),
		msg:     "",
	}
}

type withFields struct {
	fields  WithFields
	msg     string
	wrapped error
	stack   *callstack.CallStack
}

func (c *withFields) Unwrap() error {
	return c.wrapped
}

func (c *withFields) Is(target error) bool {
	_, ok := target.(*withFields)
	return ok
}

func (c *withFields) Error() string {
	if c.msg == "" {
		return c.wrapped.Error()
	}
	return c.msg + ": " + c.wrapped.Error()
}

func (c *withFields) StackTrace() callstack.StackTrace {
	if child, ok := c.wrapped.(callstack.HasStackTrace); ok {
		return child.StackTrace()
	}
	return c.stack.StackTrace()
}

func (c *withFields) Fields() map[string]interface{} {
	result := make(map[string]interface{}, len(c.fields))
	for key, value := range c.fields {
		result[key] = value
	}

	// child fields have precedence as they are closer to the cause
	var f HasFields
	if errors.As(c.wrapped, &f) {
		child := f.Fields()
		if child == nil {
			return result
		}
		for key, value := range child {
			result[key] = value
		}
	}
	return result
}

func (c *withFields) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if c.msg == "" {
				_, _ = fmt.Fprintf(s, "%+v (%s)", c.Unwrap(), c.FormatFields())
				return
			}
			_, _ = fmt.Fprintf(s, "%s: %+v (%s)", c.msg, c.Unwrap(), c.FormatFields())
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, c.Error())
		return
	}
}

func (c *withFields) FormatFields() string {
	var buf bytes.Buffer
	var count int

	for key, value := range c.fields {
		if count > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%+v=%+v", key, value))
		count++
	}
	return buf.String()
}

// ToMap Returns the fields for the underlying error as map[string]interface{}
// If no fields are available returns nil
func ToMap(err error) map[string]interface{} {
	result := map[string]interface{}{
		"excValue": err.Error(),
		"excType":  fmt.Sprintf("%T", Unwrap(err)),
	}

	// Find any errors with StackTrace information if available
	var stack callstack.HasStackTrace
	if Last(err, &stack) {
		trace := stack.StackTrace()
		caller := callstack.GetLastFrame(trace)
		result["excFuncName"] = caller.Func
		result["excLineNum"] = caller.LineNo
		result["excFileName"] = caller.File
	}

	// Search the error chain for fields
	var f HasFields
	if errors.As(err, &f) {
		for key, value := range f.Fields() {
			result[key] = value
		}
	}
	return result
}

// ToLogrus Returns the context and stacktrace information for the underlying error as logrus.Fields{}
// returns empty logrus.Fields{} if err has no context or no stacktrace
//
//	logrus.WithFields(errors.ToLogrus(err)).WithField("tid", 1).Error(err)
func ToLogrus(err error) logrus.Fields {
	result := logrus.Fields{
		"excValue": err.Error(),
		"excType":  fmt.Sprintf("%T", Unwrap(err)),
	}

	// Find any errors with StackTrace information if available
	var stack callstack.HasStackTrace
	if Last(err, &stack) {
		trace := stack.StackTrace()
		caller := callstack.GetLastFrame(trace)
		result["excFuncName"] = caller.Func
		result["excLineNum"] = caller.LineNo
		result["excFileName"] = caller.File
	}

	// Search the error chain for fields
	var f HasFields
	if errors.As(err, &f) {
		for key, value := range f.Fields() {
			result[key] = value
		}
	}
	return result
}
