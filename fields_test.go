package errors_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mailgun/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithFields(t *testing.T) {
	err := &TestError{Msg: "query error"}
	wrap := errors.WithFields{"key1": "value1"}.Wrap(err, "message")
	assert.NotNil(t, wrap)

	t.Run("Unwrap should return TestError", func(t *testing.T) {
		u := errors.Unwrap(wrap)
		require.NotNil(t, u)
		assert.Equal(t, "query error", u.Error())
	})

	t.Run("Extract fields as a normal map", func(t *testing.T) {
		m := errors.ToMap(wrap)
		require.NotNil(t, m)
		assert.Equal(t, "value1", m["key1"])
	})

	t.Run("Can use errors.Is() from std `errors` package", func(t *testing.T) {
		assert.True(t, errors.Is(err, &TestError{}))
		assert.True(t, errors.Is(wrap, &TestError{}))
	})

	t.Run("Can use errors.As() from std `errors` package", func(t *testing.T) {
		myErr := &TestError{}
		assert.True(t, errors.As(wrap, &myErr))
		assert.Equal(t, myErr.Msg, "query error")
	})

	t.Run("Extract as Logrus fields", func(t *testing.T) {
		f := errors.ToLogrus(wrap)
		require.NotNil(t, f)
		b := bytes.Buffer{}
		logrus.SetOutput(&b)
		logrus.WithFields(f).Info("test logrus fields")
		logrus.SetOutput(os.Stdout)
		assert.Contains(t, b.String(), "test logrus fields")
		assert.Contains(t, b.String(), `excValue="message: query error"`)
		assert.Contains(t, b.String(), `excType="*errors_test.TestError"`)
		assert.Contains(t, b.String(), "key1=value1")
		// TODO: Should include stack trace information file.go:123 and function name
		assert.Equal(t, "message: query error", wrap.Error())
		out := fmt.Sprintf("%+v", wrap)
		assert.True(t, strings.Contains(out, `message: query error (key1=value1)`))
	})
}

func TestErrorf(t *testing.T) {
	err := errors.New("this is an error")
	wrap := errors.WithFields{"key1": "value1", "key2": "value2"}.Wrap(err, "message")
	err = fmt.Errorf("wrapped: %w", wrap)

	// Output: 'final: wrapped: message: this is an error (key1=value1, key2=value2)'
	out := fmt.Sprintf("final: %s", err)
	t.Log(out)
	assert.Contains(t, out, "final: wrapped: message: this is an error")
	assert.Contains(t, out, "key1=value1")
	assert.Contains(t, out, "key2=value2")
}

func TestNestedWithFields(t *testing.T) {
	err := errors.New("this is an error")
	err = errors.WithFields{"key1": "value1"}.Wrap(err, "message")
	err = errors.WithFields{"key2": "value2"}.Wrap(err, "message")

	t.Run("ToMap() collects all values from nested fields", func(t *testing.T) {
		m := errors.ToMap(err)
		assert.NotNil(t, m)
		assert.Equal(t, "value1", m["key1"])
		assert.Equal(t, "value2", m["key2"])
	})

	t.Run("ToLogrus() collects all values from nested fields", func(t *testing.T) {
		f := errors.ToLogrus(err)
		require.NotNil(t, f)
		b := bytes.Buffer{}
		logrus.SetOutput(&b)
		logrus.WithFields(f).Info("test logrus fields")
		logrus.SetOutput(os.Stdout)
		assert.Contains(t, b.String(), "test logrus fields")
		assert.Contains(t, b.String(), "key1=value1")
		assert.Contains(t, b.String(), "key2=value2")
	})
}

func TestFmtDirectives(t *testing.T) {
	t.Run("Wrap() with a message", func(t *testing.T) {
		err := errors.WithFields{"key1": "value1"}.Wrap(errors.New("error"), "shit happened")
		assert.Equal(t, "value: shit happened: error (key1=value1)", fmt.Sprintf("value: %v", err))
		assert.Equal(t, "value+: shit happened: error (key1=value1)", fmt.Sprintf("value+: %+v", err))
		assert.Equal(t, "type: *errors.withFields", fmt.Sprintf("type: %T", err))
	})

	t.Run("Wrap() without a message", func(t *testing.T) {
		err := errors.WithFields{"key1": "value1"}.Wrap(errors.New("error"), "")
		assert.Equal(t, "value: error (key1=value1)", fmt.Sprintf("value: %v", err))
		assert.Equal(t, "value+: error (key1=value1)", fmt.Sprintf("value+: %+v", err))
		assert.Equal(t, "type: *errors.withFields", fmt.Sprintf("type: %T", err))
	})
}
