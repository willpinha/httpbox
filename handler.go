package httpbox

import "net/http"

type Handler func(w http.ResponseWriter, r *http.Request) error

func HelloWorld() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var err error

		if err != nil {
			return NewError(
				http.StatusInternalServerError,
				WithMessage("Ops"),
				WithDetails("Something went wrong"),
				WithInternalError(err),
			)
		}

		return nil
	}
}
