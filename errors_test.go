package errors_test

import (
	"fmt"
	"testing"

	stderr "errors"

	"github.com/mailgun/errors"
	"github.com/mailgun/errors/callstack"
	"github.com/stretchr/testify/assert"
)

// Ensure errors.Is remains stdlib compliant
func TestIs(t *testing.T) {
	err := errors.New("bottom")
	top := fmt.Errorf("top: %w", err)

	assert.True(t, stderr.Is(top, err))
	assert.True(t, errors.Is(top, err))
}

// Ensure errors.As remains stdlib compliant
func TestAs(t *testing.T) {
	err := &ErrTest{Msg: "bottom"}
	top := fmt.Errorf("top: %w", err)

	var exp *ErrTest
	assert.True(t, stderr.As(top, &exp))
	assert.True(t, errors.As(top, &exp))
}

func TestLast(t *testing.T) {
	err := errors.New("bottom")
	err = errors.Wrap(err, "last")
	err = errors.Wrap(err, "second")
	err = errors.Wrap(err, "first")
	err = errors.Errorf("wrapped: %w", err)

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
	F map[string]any
}

func (e *ErrHasFields) Error() string {
	return e.M
}

func (e *ErrHasFields) Is(target error) bool {
	_, ok := target.(*ErrHasFields)
	return ok
}

func (e *ErrHasFields) HasFields() map[string]any {
	return e.F
}
