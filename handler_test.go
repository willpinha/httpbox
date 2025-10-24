package httpbox

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name         string
		handler      Handler
		expectedCode int
		expectedBody string
	}{
		{
			name: "successful request",
			handler: Handler(func(w http.ResponseWriter, r *http.Request) error {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`"ok"`))
				return nil
			}),
			expectedCode: http.StatusOK,
			expectedBody: `"ok"`,
		},
		{
			name: "error without details",
			handler: Handler(func(w http.ResponseWriter, r *http.Request) error {
				return NewError(http.StatusNotFound, "not found")
			}),
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"not found"}`,
		},
		{
			name: "error with details",
			handler: Handler(func(w http.ResponseWriter, r *http.Request) error {
				return NewError(http.StatusBadRequest, "bad request", WithDetails("invalid input"))
			}),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"bad request","details":"invalid input"}`,
		},
		{
			name: "unknown error",
			handler: Handler(func(w http.ResponseWriter, r *http.Request) error {
				return errors.New("something went wrong")
			}),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":500,"message":"Unexpected error occurred"}`,
		},
		{
			name: "error with non-serializable details",
			handler: Handler(func(w http.ResponseWriter, r *http.Request) error {
				// Create a non-serializable type (channel)
				nonSerializable := make(chan int)
				return NewError(http.StatusBadRequest, "test error", WithDetails(nonSerializable))
			}),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"test error","details":"failed to serialize error details"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			tt.handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
		})
	}
}
