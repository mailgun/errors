package errors

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/mailgun/errors/callstack"
	"github.com/sirupsen/logrus"
)

// HasFields Implement this interface to pass along unstructured context to the logger
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

	// downstream context values have precedence as they are closer to the cause
	if child, ok := c.wrapped.(HasFields); ok {
		downstream := child.Fields()
		if downstream == nil {
			return result
		}

		for key, value := range downstream {
			result[key] = value
		}
	}
	return result
}

func (c *withFields) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if c.msg == "" {
			_, _ = fmt.Fprintf(s, "%+v (%s)", c.Unwrap(), c.FormatFields())
		} else {
			_, _ = fmt.Fprintf(s, "%s: %+v (%s)", c.msg, c.Unwrap(), c.FormatFields())
		}
	case 's', 'q':
		_, _ = io.WriteString(s, c.Error())
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

// ToMap Returns the context for the underlying error as map[string]interface{}
// If no context is available returns nil
func ToMap(err error) map[string]interface{} {
	var result map[string]interface{}

	if child, ok := err.(HasFields); ok {
		// Append the context map to our results
		result = make(map[string]interface{})
		for key, value := range child.Fields() {
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

	// Add the stack info if provided
	if cast, ok := err.(callstack.HasStackTrace); ok {
		trace := cast.StackTrace()
		caller := callstack.GetLastFrame(trace)
		result["excFuncName"] = caller.Func
		result["excLineNum"] = caller.LineNo
		result["excFileName"] = caller.File
	}

	// Add context if provided
	child, ok := err.(HasFields)
	if !ok {
		return result
	}

	// Append the context map to our results
	for key, value := range child.Fields() {
		result[key] = value
	}
	return result
}
