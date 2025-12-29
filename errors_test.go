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

// Benchmarks for comparison purposes

/*
go test -bench=. -benchmem -count=5 -run="^$" ./...
goos: darwin
goarch: arm64
pkg: github.com/mailgun/errors
cpu: Apple M3 Pro
BenchmarkErrors-12    	 4825278	       221.8 ns/op	     328 B/op	       3 allocs/op
BenchmarkErrors-12    	 5548051	       216.4 ns/op	     328 B/op	       3 allocs/op
BenchmarkErrors-12    	 5548090	       215.5 ns/op	     328 B/op	       3 allocs/op
BenchmarkErrors-12    	 5548306	       215.7 ns/op	     328 B/op	       3 allocs/op
BenchmarkErrors-12    	 5557860	       215.8 ns/op	     328 B/op	       3 allocs/op
*/
func BenchmarkErrorsWrap(b *testing.B) {
	base := errors.New("init")
	for b.Loop() {
		errors.Wrap(base, "loop")
	}
}

/*
go test -bench=. -benchmem -count=5 -run="^$" ./...
goos: darwin
goarch: arm64
pkg: github.com/mailgun/errors
cpu: Apple M3 Pro
BenchmarkErrorsWrapf-12    	 4733302	       252.8 ns/op	     336 B/op	       4 allocs/op
BenchmarkErrorsWrapf-12    	 4750388	       252.1 ns/op	     336 B/op	       4 allocs/op
BenchmarkErrorsWrapf-12    	 4720851	       251.7 ns/op	     336 B/op	       4 allocs/op
BenchmarkErrorsWrapf-12    	 4740823	       252.3 ns/op	     336 B/op	       4 allocs/op
BenchmarkErrorsWrapf-12    	 4753670	       254.1 ns/op	     336 B/op	       4 allocs/op
*/
func BenchmarkErrorsWrapf(b *testing.B) {
	base := errors.New("init")
	for b.Loop() {
		errors.Wrapf(base, "loop %s", "two")
	}
}

/*
go test -bench=. -benchmem -count=5 -run="^$" ./...
goos: darwin
goarch: arm64
pkg: github.com/mailgun/errors
cpu: Apple M3 Pro
BenchmarkErrorsStack-12    	 5713897	       210.4 ns/op	     304 B/op	       3 allocs/op
BenchmarkErrorsStack-12    	 5677599	       210.1 ns/op	     304 B/op	       3 allocs/op
BenchmarkErrorsStack-12    	 5701461	       210.1 ns/op	     304 B/op	       3 allocs/op
BenchmarkErrorsStack-12    	 5655940	       210.1 ns/op	     304 B/op	       3 allocs/op
BenchmarkErrorsStack-12    	 5574022	       210.8 ns/op	     304 B/op	       3 allocs/op
*/
func BenchmarkErrorsStack(b *testing.B) {
	base := errors.New("init")
	for b.Loop() {
		_ = errors.Stack(base)
	}
}
