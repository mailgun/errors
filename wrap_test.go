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

func TestWrap(t *testing.T) {
	err := &ErrTest{Msg: "query error"}
	wrap := errors.Wrap(err, "message")
	assert.NotNil(t, wrap)
	wrapf := errors.Wrapf(err, "message: %d", 1)
	assert.NotNil(t, wrapf)

	t.Run("Wrap/Wrapf should return wrap the error", func(t *testing.T) {
		assert.Equal(t, "message: query error", wrap.Error())
		assert.Equal(t, "message: 1: query error", wrapf.Error())
	})

	t.Run("Unwrap should return ErrTest", func(t *testing.T) {
		u := errors.Unwrap(wrap)
		require.NotNil(t, u)
		assert.Equal(t, "query error", u.Error())

		uf := errors.Unwrap(wrapf)
		require.NotNil(t, uf)
		assert.Equal(t, "query error", uf.Error())
	})

	t.Run("Extract stack fields as a normal map", func(t *testing.T) {
		mf := errors.ToMap(wrapf)
		assert.Len(t, mf, 5)

		m := errors.ToMap(wrap)
		require.NotNil(t, m)
		assert.Regexp(t, ".*/wrap_test.go", m["excFileName"])
		assert.Regexp(t, "\\d*", m["excLineNum"])
		assert.Equal(t, "message: query error", m["excValue"])
		assert.Equal(t, "errors_test.TestWrap", m["excFuncName"])
		assert.Equal(t, "*errors_test.ErrTest", m["excType"])
		assert.Len(t, m, 5)
	})

	t.Run("Can use errors.Is() from std `errors` package", func(t *testing.T) {
		assert.True(t, errors.Is(err, &ErrTest{}))
		assert.True(t, errors.Is(wrap, &ErrTest{}))
		assert.True(t, errors.Is(wrapf, &ErrTest{}))
	})

	t.Run("Can use errors.As() from std `errors` package", func(t *testing.T) {
		myErr := &ErrTest{}
		assert.True(t, errors.As(wrap, &myErr))
		assert.Equal(t, myErr.Msg, "query error")
		assert.True(t, errors.As(wrapf, &myErr))
		assert.Equal(t, myErr.Msg, "query error")
	})

	t.Run("Extract as Logrus fields", func(t *testing.T) {
		ff := errors.ToLogrus(wrapf)
		assert.Len(t, ff, 5)

		f := errors.ToLogrus(wrap)

		require.NotNil(t, f)
		b := bytes.Buffer{}
		logrus.SetOutput(&b)
		logrus.WithFields(f).Info("test logrus fields")
		logrus.SetOutput(os.Stdout)
		assert.Contains(t, b.String(), "test logrus fields")
		assert.Contains(t, b.String(), `excValue="message: query error"`)
		assert.Contains(t, b.String(), `excType="*errors_test.ErrTest"`)
		assert.Contains(t, b.String(), "excFuncName=errors_test.TestWrap")
		assert.Regexp(t, "excFileName=.*/wrap_test.go", b.String())
		assert.Regexp(t, "excLineNum=\\d*", b.String())

		// OUTPUT: time="2023-01-26T10:37:48-05:00" level=info msg="test logrus fields"
		//   excFileName=errors/fields_test.go excFuncName=errors_test.TestWithFields
		//   excLineNum=18 excType="*errors_test.ErrTest" excValue="message: query error" key1=value1
		t.Log(b.String())

		assert.Equal(t, "message: query error", wrap.Error())
		out := fmt.Sprintf("%+v", wrap)
		assert.True(t, strings.Contains(out, `message: query error`))
	})

	t.Run("Wrap() should return nil, if error is nil", func(t *testing.T) {
		assert.Nil(t, errors.Wrap(nil, "no error"))
	})

	t.Run("Wrapf() should return nil, if error is nil", func(t *testing.T) {
		assert.Nil(t, errors.Wrapf(nil, "no '%d' error", 1))
	})
}

func TestWrapFmtDirectives(t *testing.T) {
	t.Run("Wrapf()", func(t *testing.T) {
		err := errors.Wrapf(errors.New("error"), "shit happened '%d'", 1)
		assert.Equal(t, "shit happened '1': error", fmt.Sprintf("%s", err))
		assert.Equal(t, "shit happened '1': error", fmt.Sprintf("%v", err))
		assert.Equal(t, "shit happened '1': error", fmt.Sprintf("%+v", err))
		assert.Equal(t, "*errors.wrappedError", fmt.Sprintf("%T", err))
	})

	t.Run("Wrap()", func(t *testing.T) {
		err := errors.Wrapf(errors.New("error"), "")
		assert.Equal(t, "error", fmt.Sprintf("%s", err))
		assert.Equal(t, "error", fmt.Sprintf("%v", err))
		assert.Equal(t, "error", fmt.Sprintf("%+v", err))
		assert.Equal(t, "*errors.wrappedError", fmt.Sprintf("%T", err))
	})
}

func TestWrapErrorValue(t *testing.T) {
	err := io.EOF
	wrap := errors.Wrap(err, "message")
	assert.True(t, errors.Is(wrap, io.EOF))
}
