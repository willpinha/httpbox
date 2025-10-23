package httpbox

import (
	"errors"
	"log/slog"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)

	if err != nil {
		handleError(w, err)
		return
	}
}

func handleError(w http.ResponseWriter, err error) {
	var httpErr *Error

	// This avoids leaking internal error details to the client. The library user
	// should wrap errors in httpbox.Error to provide proper status codes and messages
	if !errors.As(err, &httpErr) {
		httpErr = NewError(http.StatusInternalServerError, "Unexpected error occurred",
			WithInternalError(err),
			WithLog(),
		)
	}

	// The only possible error is if the Details field contains non-serializable data
	if err := WriteJSON(w, httpErr.Code, httpErr); err != nil {
		failedMsg := "Failed to serialize error details"

		httpErr.Details = failedMsg

		slog.Error(failedMsg, "error", err, "original_error", httpErr.Err)

		// Since we overwrite Details, we ignore the error here as it will not occur
		WriteJSON(w, httpErr.Code, httpErr)
	}

	if httpErr.ShouldLog() {
		slog.Error(httpErr.Message, "code", httpErr.Code, "details", httpErr.Details, "error", httpErr.Err)
	}
}

func (h Handler) WithMiddlewares(middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func AdaptHandler(h http.Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}
}
