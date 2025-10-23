package httpbox

import (
	"encoding/json"
	"net/http"
)

const (
	ContentTypeTextPlain = "text/plain; charset=utf-8"
	ContentTypeJSON      = "application/json; charset=utf-8"
)

func WriteJSON(w http.ResponseWriter, code int, data any) error {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(code)

	return nil
}
