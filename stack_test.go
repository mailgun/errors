package errors_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/mailgun/errors"
	"github.com/mailgun/errors/callstack"
	"github.com/stretchr/testify/assert"
)

// NOTE: Line numbers matter to this test
func TestWrapWithFieldsAndStack(t *testing.T) {
	// NOTE: The stack from StackTrace() should report this line
	// not the Fields line below
	s := errors.Stack(&ErrTest{Msg: "error"})

	err := errors.Fields{"key1": "value1"}.Wrap(s, "context")

	myErr := &ErrTest{}
	assert.True(t, errors.Is(err, &ErrTest{}))
	assert.True(t, errors.As(err, &myErr))
	assert.Equal(t, myErr.Msg, "error")
	assert.Equal(t, "context: error", err.Error())

	// Extract the stack from the error chain
	var stack callstack.HasStackTrace
	// Extract the stack info if provided
	assert.True(t, errors.As(err, &stack))

	trace := stack.StackTrace()
	caller := callstack.GetLastFrame(trace)
	assert.Contains(t, fmt.Sprintf("%+v", stack), "errors/stack_test.go:17")
	assert.Equal(t, "errors_test.TestWrapWithFieldsAndStack", caller.Func)
	assert.Equal(t, 17, caller.LineNo)
}

func TestStack(t *testing.T) {
	err := errors.Stack(io.EOF)

	var files []string
	var funcs []string
	if cast, ok := err.(callstack.HasStackTrace); ok {
		for _, frame := range cast.StackTrace() {
			files = append(files, fmt.Sprintf("%s", frame))
			funcs = append(funcs, fmt.Sprintf("%n", frame))
		}
	}
	assert.Contains(t, files, "stack_test.go")
	assert.Contains(t, funcs, "TestStack")
}

func TestStackWrapped(t *testing.T) {
	err := errors.Stack(&ErrTest{Msg: "query error"})
	err = fmt.Errorf("wrapped: %w", err)

	// Can use errors.Is() from std `errors` package
	assert.True(t, errors.Is(err, &ErrTest{}))

	// Can use errors.As() from std `errors` package
	myErr := &ErrTest{}
	assert.True(t, errors.As(err, &myErr))
	assert.Equal(t, myErr.Msg, "query error")
}

func TestFormatStack(t *testing.T) {
	tests := []struct {
		err    error
		Name   string
		format string
		want   []string
	}{{
		Name:   "stack() string",
		err:    errors.Stack(io.EOF),
		format: "%s",
		want:   []string{"EOF"},
	}, {
		Name:   "stack() value",
		err:    errors.Stack(io.EOF),
		format: "%v",
		want:   []string{"EOF"},
	}, {
		Name:   "stack() value plus",
		err:    errors.Stack(io.EOF),
		format: "%+v",
		want: []string{
			"EOF",
			"github.com/mailgun/errors_test.TestFormatStack",
		},
	}, {
		Name:   "stack(errors.New()) string",
		err:    errors.Stack(errors.New("error")),
		format: "%s",
		want:   []string{"error"},
	}, {
		Name:   "stack(errors.New()) value",
		err:    errors.Stack(errors.New("error")),
		format: "%v",
		want:   []string{"error"},
	}, {
		Name:   "stack(errors.New()) value plus",
		err:    errors.Stack(errors.New("error")),
		format: "%+v",
		want: []string{
			"error",
			"github.com/mailgun/errors_test.TestFormatStack",
			"errors/stack_test.go",
		},
	}, {
		Name:   "errors.Stack(errors.Stack(io.EOF)) value plus",
		err:    errors.Stack(errors.Stack(io.EOF)),
		format: "%+v",
		want: []string{"EOF",
			"github.com/mailgun/errors_test.TestFormatStack",
			"github.com/mailgun/errors_test.TestFormatStack",
		},
	}, {
		Name:   "deeply nested stack",
		err:    errors.Stack(errors.Stack(fmt.Errorf("message: %w", io.EOF))),
		format: "%+v",
		want: []string{"EOF",
			"message",
			"github.com/mailgun/errors_test.TestFormatStack",
		},
	}, {
		Name:   "Stack with fmt.Errorf()",
		err:    errors.Stack(fmt.Errorf("error%d", 1)),
		format: "%+v",
		want: []string{"error1",
			"github.com/mailgun/errors_test.TestFormatStack",
		},
	}}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			out := fmt.Sprintf(tt.format, tt.err)
			// t.Log(out)

			for _, line := range tt.want {
				assert.Contains(t, out, line)
			}
		})
	}
}

// Ensure errors.Stack() returns an that works with `github.com/pkf/errors.Cause()`
func TestStackCause(t *testing.T) {
	err := errors.Stack(io.EOF)
	assert.Equal(t, io.EOF, pkgErrorCause(err))
}

// pkgErrorCause is identical to github.com/pkg/errors.Cause()
// Much of our existing code uses that function and depends
// on our errors conforming to the old style Causer interface.
func pkgErrorCause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
