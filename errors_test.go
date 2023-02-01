package errors_test

import (
	"fmt"
	"testing"

	"github.com/mailgun/errors"
	"github.com/mailgun/errors/callstack"
	"github.com/stretchr/testify/assert"
)

type ErrTest struct {
	Msg string
}

func (e *ErrTest) Error() string {
	return e.Msg
}

func (e *ErrTest) Is(target error) bool {
	_, ok := target.(*ErrTest)
	return ok
}

type ErrHasFields struct {
	M string
	F map[string]interface{}
}

func (e *ErrHasFields) Error() string {
	return e.M
}

func (e *ErrHasFields) Is(target error) bool {
	_, ok := target.(*ErrHasFields)
	return ok
}

func (e *ErrHasFields) Fields() map[string]interface{} {
	return e.F
}

func TestLast(t *testing.T) {
	err := errors.New("bottom")
	err = errors.Wrap(err, "last")
	err = errors.Wrap(err, "second")
	err = errors.Wrap(err, "first")
	err = fmt.Errorf("wrapped: %w", err)

	// errors.As() returns the "first" error in the chain with a stack trace
	var first callstack.HasStackTrace
	assert.True(t, errors.As(err, &first))
	assert.Equal(t, "first: second: last: bottom", first.(error).Error())

	// errors.Last() returns the last error in the chain with a stack trace
	var last callstack.HasStackTrace
	assert.True(t, errors.Last(err, &last))
	assert.Equal(t, "last: bottom", last.(error).Error())

	// If no stack trace is found, then should not set target and should return false
	assert.False(t, errors.Last(errors.New("no stack"), &last))
	assert.Equal(t, "last: bottom", last.(error).Error())
}
