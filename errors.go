package errors

import (
	"errors"
	"reflect"
)

// NoMsg is a small indicator in the code that "" is intentional and there
// is no message include with the Wrap()
const NoMsg = ""

// Import all the standard errors functions as a convenience.

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(text string) error {
	return errors.New(text)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Last finds the last error in err's chain that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method `As(any) bool` such that
// As(target) returns true.
//
// An error type might provide an As() method so it can be treated as if it were a
// different error type.
//
// Last panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
//
// NOTE: Last() is much slower than As(). Therefore As() should always be used
// unless you absolutely need Last() to retrieve the last error in the error chain
// that matches the target.
func Last(err error, target any) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflect.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	targetType := typ.Elem()
	if targetType.Kind() != reflect.Interface && !targetType.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	var found error
	for err != nil {
		if reflect.TypeOf(err).AssignableTo(targetType) {
			found = err
		}
		if x, ok := err.(interface{ As(any) bool }); ok && x.As(target) {
			found = err
		}
		err = Unwrap(err)
	}
	if found != nil {
		val.Elem().Set(reflect.ValueOf(found))
		return true
	}
	return false
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()
