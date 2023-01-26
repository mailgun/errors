package errors_test

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
