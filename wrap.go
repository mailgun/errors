package errors

import "fmt"

// Wrap is deprecated use fmt.Errorf() instead
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Wrapf is deprecated use fmt.Errorf() instead
func Wrapf(err error, msg string, a ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(msg, a...), err)
}
