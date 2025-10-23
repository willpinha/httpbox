package httpbox

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

type Validator interface {
	Validate() error
}

func verifyValidator(v any) error {
	validator, ok := v.(Validator)
	if !ok {
		return nil
	}

	return validator.Validate()
}

func ReadJSON[T any](r io.Reader) (T, error) {
	var v T

	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return v, NewError(http.StatusBadRequest, "invalid JSON body", WithDetails(err))
	}

	if err := verifyValidator(v); err != nil {
		return v, err
	}

	return v, nil
}

func ReadXML[T any](r io.Reader) (T, error) {
	var v T

	if err := xml.NewDecoder(r).Decode(&v); err != nil {
		return v, NewError(http.StatusBadRequest, "invalid XML body", WithDetails(err))
	}

	if err := verifyValidator(v); err != nil {
		return v, err
	}

	return v, nil
}

func ReadBytes(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, NewError(http.StatusBadRequest, "unable to read body", WithDetails(err))
	}

	return data, nil
}
