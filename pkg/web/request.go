package web

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type validator interface {
	Validate() error
}

// Params returns the parameters from the request.
func Params(r *http.Request, key string) string {
	params := mux.Vars(r)
	return params[key]
}

// Decode decodes the body of a request into a given interface.
func Decode(r *http.Request, v any) error {
	decoder := json.NewDecoder(r.Body)
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
