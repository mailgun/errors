package errors_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mailgun/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToMapToLogrusFindsLastStackTrace(t *testing.T) {
	err := errors.New("this is an error")

	// --- Should report this line number for the stack in the chain ---
	err = errors.Wrap(err, "last")
	// ----------------------------------

	err = errors.Wrap(err, "second")
	err = errors.Wrap(err, "first")

	t.Run("ToMap() finds the last stack in the chain", func(t *testing.T) {
		m := errors.ToMap(err)
		assert.NotNil(t, m)
		assert.Equal(t, 21, m["excLineNum"])
	})

	t.Run("ToLogrus() finds the last stack in the chain", func(t *testing.T) {
		f := errors.ToLogrus(err)
		require.NotNil(t, f)
		b := bytes.Buffer{}
		logrus.SetOutput(&b)
		logrus.WithFields(f).Info("test logrus fields")
		logrus.SetOutput(os.Stdout)
		assert.Contains(t, b.String(), "excLineNum=21")
	})
}

func TestWithFields(t *testing.T) {
	err := &ErrTest{Msg: "query error"}
	wrap := errors.WithFields{"key1": "value1"}.Wrap(err, "message")
	assert.NotNil(t, wrap)

	t.Run("Unwrap should return ErrTest", func(t *testing.T) {
		u := errors.Unwrap(wrap)
		require.NotNil(t, u)
		assert.Equal(t, "query error", u.Error())
	})

	t.Run("Extract fields as a normal map", func(t *testing.T) {
		m := errors.ToMap(wrap)
		require.NotNil(t, m)

		assert.Equal(t, "value1", m["key1"])
		assert.Regexp(t, ".*/fields_test.go", m["excFileName"])
		assert.Regexp(t, "\\d*", m["excLineNum"])
		assert.Equal(t, "message: query error", m["excValue"])
		assert.Equal(t, "errors_test.TestWithFields", m["excFuncName"])
		assert.Equal(t, "*errors_test.ErrTest", m["excType"])
		assert.Len(t, m, 6)
	})

	t.Run("Can use errors.Is() from std `errors` package", func(t *testing.T) {
		assert.True(t, errors.Is(err, &ErrTest{}))
		assert.True(t, errors.Is(wrap, &ErrTest{}))
	})

	t.Run("Can use errors.As() from std `errors` package", func(t *testing.T) {
		myErr := &ErrTest{}
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
		assert.Contains(t, b.String(), `excType="*errors_test.ErrTest"`)
		assert.Contains(t, b.String(), "key1=value1")
		assert.Contains(t, b.String(), "excFuncName=errors_test.TestWithFields")
		assert.Regexp(t, "excFileName=.*/fields_test.go", b.String())
		assert.Regexp(t, "excLineNum=\\d*", b.String())

		// OUTPUT: time="2023-01-26T10:37:48-05:00" level=info msg="test logrus fields"
		//   excFileName=errors/fields_test.go excFuncName=errors_test.TestWithFields
		//   excLineNum=18 excType="*errors_test.ErrTest" excValue="message: query error" key1=value1
		// t.Log(b.String())

		assert.Equal(t, "message: query error", wrap.Error())
		out := fmt.Sprintf("%+v", wrap)
		assert.True(t, strings.Contains(out, `message: query error (key1=value1)`))
	})

	t.Run("ToLogrus() should extract the error with StackTrace() from the chain", func(t *testing.T) {
		// This error has no  StackTrace() method
		err := fmt.Errorf("I have no stack trace: %w", wrap)
		// ToLogrus() should find the wrapped error in the chain that has the StackTrace() method.
		f := errors.ToLogrus(err)
		// t.Log(f)

		assert.Regexp(t, ".*/fields_test.go", f["excFileName"])
		assert.Regexp(t, "\\d*", f["excLineNum"])
		assert.Equal(t, "I have no stack trace: message: query error", f["excValue"])
		assert.Equal(t, "errors_test.TestWithFields", f["excFuncName"])
		assert.Equal(t, "*errors.withFields", f["excType"])
		assert.Equal(t, "value1", f["key1"])
		assert.Len(t, f, 6)

		require.NotNil(t, f)
	})

	t.Run("Wrap() should return nil, if error is nil", func(t *testing.T) {
		got := errors.WithFields{"some": "context"}.Wrap(nil, "no error")
		assert.Nil(t, got)
	})

	t.Run("Wrapf() should return nil, if error is nil", func(t *testing.T) {
		got := errors.WithFields{"some": "context"}.Wrapf(nil, "no '%d' error", 1)
		assert.Nil(t, got)
	})
}

func TestErrorf(t *testing.T) {
	err := errors.New("this is an error")
	wrap := errors.WithFields{"key1": "value1", "key2": "value2"}.Wrap(err, "message")
	err = fmt.Errorf("wrapped: %w", wrap)
	assert.Equal(t, fmt.Sprintf("%s", err), "wrapped: message: this is an error")
}

func TestNestedWithFields(t *testing.T) {
	err := errors.New("this is an error")
	err = errors.WithFields{"key1": "value1"}.Wrap(err, "message")
	err = errors.Wrap(err, "second")
	err = errors.WithFields{"key2": "value2"}.Wrap(err, "message")
	err = errors.Wrap(err, "first")

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

func TestWithFieldsFmtDirectives(t *testing.T) {
	t.Run("Wrap() with a message", func(t *testing.T) {
		err := errors.WithFields{"key1": "value1"}.Wrap(errors.New("error"), "shit happened")
		assert.Equal(t, "shit happened: error", fmt.Sprintf("%s", err))
		assert.Equal(t, "shit happened: error", fmt.Sprintf("%v", err))
		assert.Equal(t, "shit happened: error (key1=value1)", fmt.Sprintf("%+v", err))
		assert.Equal(t, "*errors.withFields", fmt.Sprintf("%T", err))
	})

	t.Run("Wrap() without a message", func(t *testing.T) {
		err := errors.WithFields{"key1": "value1"}.Wrap(errors.New("error"), "")
		assert.Equal(t, "error", fmt.Sprintf("%s", err))
		assert.Equal(t, "error", fmt.Sprintf("%v", err))
		assert.Equal(t, "error (key1=value1)", fmt.Sprintf("%+v", err))
		assert.Equal(t, "*errors.withFields", fmt.Sprintf("%T", err))
	})
}

func TestWithFieldsErrorValue(t *testing.T) {
	err := io.EOF
	wrap := errors.WithFields{"key1": "value1"}.Wrap(err, "message")
	assert.True(t, errors.Is(wrap, io.EOF))
}

func TestHasFields(t *testing.T) {
	hf := &ErrHasFields{M: "error", F: map[string]interface{}{"file": "errors.go"}}
	err := errors.WithFields{"key1": "value1"}.Wrap(hf, "")
	m := errors.ToMap(err)
	require.NotNil(t, m)
	assert.Equal(t, "value1", m["key1"])
	assert.Equal(t, "errors.go", m["file"])
}

func TestWithFieldsError(t *testing.T) {
	t.Run("WithFields.Error() should create a new error", func(t *testing.T) {
		err := errors.WithFields{"key1": "value1"}.Error("error")
		m := errors.ToMap(err)
		require.NotNil(t, m)
		assert.Equal(t, "value1", m["key1"])
		assert.Equal(t, "error", err.Error())
	})

	t.Run("WithFields.Errorf() should create a new error", func(t *testing.T) {
		err := errors.WithFields{"key1": "value1"}.Errorf("error '%d'", 1)
		m := errors.ToMap(err)
		require.NotNil(t, m)
		assert.Equal(t, "value1", m["key1"])
		assert.Equal(t, "error '1'", err.Error())
	})
}

func TestWithFieldsWithStack(t *testing.T) {
	err := errors.WithFields{"key1": "value1"}.WithStack(io.EOF)
	m := errors.ToMap(err)
	require.NotNil(t, m)
	assert.Equal(t, "value1", m["key1"])
	assert.Equal(t, io.EOF.Error(), err.Error())
}
