package web

import (
	"encoding/json"
	"net/http"
)

// validator is an interface for validating the request.
type validator interface {
	Validate() error
}

// Params returns the parameters from the request.
func Params(r *http.Request, key string) string {
	return r.PathValue(key)
}

// Decode reads and decodes the JSON body of a request into the provided value.
// It returns an error if the body contains unknown fields or if validation fails
// (when the value implements the validator interface).
func Decode(r *http.Request, v any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	if val, ok := v.(validator); ok {
		if err := val.Validate(); err != nil {
			return err
		}
	}
	return nil
}
