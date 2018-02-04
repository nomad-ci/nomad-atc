package bitbucket

type Error struct {
	status    string
	context   *string
	message   *string
	exception *string
}

func NewError(status string) *Error {
	return &Error{
		status: status,
	}
}

func (e Error) Error() string {
	switch {
	case e.message != nil:
		return *e.message
	case e.context != nil:
		return *e.context
	case e.exception != nil:
		return *e.exception
	}

	return e.status
}

func (e *Error) WithContext(context string) *Error {
	e.context = &context
	return e
}

func (e *Error) Context() string {
	if e.context == nil {
		return ""
	}

	return *e.context
}

func (e *Error) WithMessage(message string) *Error {
	e.message = &message
	return e
}

func (e *Error) Message() string {
	if e.message == nil {
		return ""
	}

	return *e.message
}

func (e *Error) WithExceptionName(exceptionName string) *Error {
	e.exception = &exceptionName
	return e
}

func (e *Error) ExceptionName() string {
	if e.exception == nil {
		return ""
	}

	return *e.exception
}
