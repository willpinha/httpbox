package httpbox

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Err     error  `json:"-"`
	Log     bool   `json:"-"`
}

type ErrorOption func(*Error)

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) ShouldLog() bool {
	return e.Log
}

func NewError(code int, message string, opts ...ErrorOption) *Error {
	err := &Error{
		Code:    code,
		Message: message,
		Log:     false,
	}

	for _, opt := range opts {
		opt(err)
	}

	return err
}

func WithDetails(details any) ErrorOption {
	return func(err *Error) {
		err.Details = details
	}
}

func WithInternalError(internalErr error) ErrorOption {
	return func(err *Error) {
		err.Err = internalErr
	}
}

func WithLog() ErrorOption {
	return func(err *Error) {
		err.Log = true
	}
}
