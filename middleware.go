package httpbox

import (
	"log/slog"
	"net/http"
)

type Middleware func(Handler) Handler

func applyMiddlewares(h Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type accessResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bodySize   int
}

func newAccessResponseWriter(w http.ResponseWriter) *accessResponseWriter {
	return &accessResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (arw *accessResponseWriter) WriteHeader(statusCode int) {
	arw.statusCode = statusCode
	arw.ResponseWriter.WriteHeader(statusCode)
}

func (arw *accessResponseWriter) Write(b []byte) (int, error) {
	size, err := arw.ResponseWriter.Write(b)
	arw.bodySize += size
	return size, err
}

func AccessLogMiddleware() Middleware {
	return func(h Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			arw := newAccessResponseWriter(w)

			err := h(arw, r)

			reqGroup := slog.Group("req",
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
				slog.String("remote_addr", r.RemoteAddr),
			)

			resGroup := slog.Group("res",
				slog.Int("status", arw.statusCode),
				slog.Int("body_size", arw.bodySize),
			)

			slog.Info("Access", reqGroup, resGroup)

			return err
		}
	}
}
