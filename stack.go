package errors

import (
	"fmt"
	"io"

	"github.com/mailgun/errors/callstack"
)

// WithStack annotates err with a stack trace at the point WithStack was called.
// If err is nil, WithStack returns nil.
func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return &withStack{
		err,
		callstack.New(1),
	}
}

type withStack struct {
	error
	*callstack.CallStack
}

func (w *withStack) Unwrap() error { return w.error }

func (w *withStack) Is(target error) bool {
	_, ok := target.(*withStack)
	return ok
}

func (w *withStack) Fields() map[string]interface{} {
	if child, ok := w.error.(HasFields); ok {
		return child.Fields()
	}
	return nil
}

func (w *withStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%+v", w.Unwrap())
			w.CallStack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, w.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", w.Error())
	}
}
