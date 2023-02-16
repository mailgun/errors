package errors_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/ahmetb/go-linq"
	"github.com/mailgun/errors"
	"github.com/mailgun/errors/callstack"
	"github.com/stretchr/testify/assert"
)

// NOTE: Line numbers matter to this test
func TestWrapWithFieldsAndWithStack(t *testing.T) {
	// NOTE: The stack from StackTrace() should report this line
	// not the Fields line below
	s := errors.WithStack(&ErrTest{Msg: "error"})

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
	assert.Contains(t, fmt.Sprintf("%+v", stack), "errors/stack_test.go:18")
	assert.Equal(t, "errors_test.TestWrapWithFieldsAndWithStack", caller.Func)
	assert.Equal(t, 18, caller.LineNo)
}

func TestWithStack(t *testing.T) {
	err := errors.WithStack(io.EOF)

	var files []string
	var funcs []string
	if cast, ok := err.(callstack.HasStackTrace); ok {
		for _, frame := range cast.StackTrace() {
			files = append(files, fmt.Sprintf("%s", frame))
			funcs = append(funcs, fmt.Sprintf("%n", frame))
		}
	}
	assert.True(t, linq.From(files).Contains("stack_test.go"))
	assert.True(t, linq.From(funcs).Contains("TestWithStack"), funcs)
}

func TestWithStackWrapped(t *testing.T) {
	err := errors.WithStack(&ErrTest{Msg: "query error"})
	err = fmt.Errorf("wrapped: %w", err)

	// Can use errors.Is() from std `errors` package
	assert.True(t, errors.Is(err, &ErrTest{}))

	// Can use errors.As() from std `errors` package
	myErr := &ErrTest{}
	assert.True(t, errors.As(err, &myErr))
	assert.Equal(t, myErr.Msg, "query error")
}

func TestFormatWithStack(t *testing.T) {
	tests := []struct {
		err    error
		Name   string
		format string
		want   []string
	}{{
		Name:   "withStack() string",
		err:    errors.WithStack(io.EOF),
		format: "%s",
		want:   []string{"EOF"},
	}, {
		Name:   "withStack() value",
		err:    errors.WithStack(io.EOF),
		format: "%v",
		want:   []string{"EOF"},
	}, {
		Name:   "withStack() value plus",
		err:    errors.WithStack(io.EOF),
		format: "%+v",
		want: []string{
			"EOF",
			"github.com/mailgun/errors_test.TestFormatWithStack",
		},
	}, {
		Name:   "withStack(errors.New()) string",
		err:    errors.WithStack(errors.New("error")),
		format: "%s",
		want:   []string{"error"},
	}, {
		Name:   "withStack(errors.New()) value",
		err:    errors.WithStack(errors.New("error")),
		format: "%v",
		want:   []string{"error"},
	}, {
		Name:   "withStack(errors.New()) value plus",
		err:    errors.WithStack(errors.New("error")),
		format: "%+v",
		want: []string{
			"error",
			"github.com/mailgun/errors_test.TestFormatWithStack",
			"errors/stack_test.go",
		},
	}, {
		Name:   "errors.WithStack(errors.WithStack(io.EOF)) value plus",
		err:    errors.WithStack(errors.WithStack(io.EOF)),
		format: "%+v",
		want: []string{"EOF",
			"github.com/mailgun/errors_test.TestFormatWithStack",
			"github.com/mailgun/errors_test.TestFormatWithStack",
		},
	}, {
		Name:   "deeply nested stack",
		err:    errors.WithStack(errors.WithStack(fmt.Errorf("message: %w", io.EOF))),
		format: "%+v",
		want: []string{"EOF",
			"message",
			"github.com/mailgun/errors_test.TestFormatWithStack",
		},
	}, {
		Name:   "WithStack with fmt.Errorf()",
		err:    errors.WithStack(fmt.Errorf("error%d", 1)),
		format: "%+v",
		want: []string{"error1",
			"github.com/mailgun/errors_test.TestFormatWithStack",
		},
	}}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			out := fmt.Sprintf(tt.format, tt.err)
			//t.Log(out)

			for _, line := range tt.want {
				assert.Contains(t, out, line)
			}
		})
	}
}
