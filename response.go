package httpbox

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, code int, data any) error {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	return nil
}

func WriteXML(w http.ResponseWriter, code int, data any) error {
	if err := xml.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(code)

	return nil
}

func WriteBytes(w http.ResponseWriter, code int, contentType string, data []byte) error {
	return WriteFromReader(w, bytes.NewReader(data), code, contentType)
}

func WriteFromReader(w http.ResponseWriter, r io.Reader, code int, contentType string) error {
	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)

	return nil
}
