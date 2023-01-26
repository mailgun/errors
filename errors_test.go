package errors_test

type TestError struct {
	Msg string
}

func (e *TestError) Error() string {
	return e.Msg
}

func (e *TestError) Is(target error) bool {
	_, ok := target.(*TestError)
	return ok
}
