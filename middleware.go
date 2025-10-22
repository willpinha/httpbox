package httpbox

import "net/http"

type Middleware func(Handler) Handler

func HelloWorld() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
}

func AccessLogMiddleware() Middleware {
	return func(h Handler) Handler {
		return nil
	}
}

func X() {
	mux := http.NewServeMux()

	mux.Handle("GET /hello-world", HelloWorld())
	mux.Handle("POST /hello-world", HelloWorld())
	mux.Handle("PUT /hello-world", HelloWorld())
	mux.Handle("PATCH /hello-world", HelloWorld())
	mux.Handle("DELETE /hello-world", HelloWorld())

	handler := AdaptHandler(mux).WithMiddlewares(
		AccessLogMiddleware(),
	)

	http.ListenAndServe(":8080", handler)
}
