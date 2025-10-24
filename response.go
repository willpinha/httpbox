package httpbox

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, code int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}

func WriteXML(w http.ResponseWriter, code int, data any) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(code)

	if err := xml.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}

func WriteBytes(w http.ResponseWriter, code int, contentType string, data []byte) error {
	return WriteFromReader(w, bytes.NewReader(data), code, contentType)
}

func WriteFromReader(w http.ResponseWriter, r io.Reader, code int, contentType string) error {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)

	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	return nil
}
