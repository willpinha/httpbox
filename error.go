package httpbox

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
	err     error  `json:"-"`
}

type ErrorOption func(*Error)

func (e *Error) Error() string {
	return e.Message
}

func NewError(code int, opts ...ErrorOption) *Error {
	err := &Error{
		Code: code,
	}

	for _, opt := range opts {
		opt(err)
	}

	return err
}

func WithMessage(message string) ErrorOption {
	return func(err *Error) {
		err.Message = message
	}
}

func WithDetails(details any) ErrorOption {
	return func(err *Error) {
		err.Details = details
	}
}

func WithInternalError(internalErr error) ErrorOption {
	return func(err *Error) {
		err.err = internalErr
	}
}
