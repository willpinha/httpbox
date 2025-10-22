package httpbox

import (
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

func (h Handler) WithMiddlewares(middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func handleError(w http.ResponseWriter, err error) {

}

func AdaptHandler(h http.Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}
}
