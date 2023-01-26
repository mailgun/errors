package errors_test

import (
	"testing"

	"github.com/mailgun/errors"
	"github.com/stretchr/testify/assert"
)

func TestWrapfNil(t *testing.T) {
	got := errors.WithFields{"some": "context"}.Wrapf(nil, "no '%d' error", 1)
	assert.Nil(t, got)
}

func TestWrapNil(t *testing.T) {
	got := errors.WithFields{"some": "context"}.Wrap(nil, "no error")
	assert.Nil(t, got)
}
