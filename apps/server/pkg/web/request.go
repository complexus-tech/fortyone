package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

const DefaultMaxJSONBodyBytes int64 = 1 << 20

var (
	ErrInvalidJSONContentType = errors.New("request content type must be application/json")
	ErrMultipleJSONValues     = errors.New("request body must contain exactly one JSON value")
)

// validator is an interface for validating the request.
type validator interface {
	Validate() error
}

// Params returns the parameters from the request.
func Params(r *http.Request, key string) string {
	return r.PathValue(key)
}

// Decode reads one bounded JSON value and validates it. JSON endpoints reject
// missing or non-JSON content types, unknown fields, and trailing values.
func Decode(r *http.Request, v any) error {
	return DecodeWithLimit(r, v, DefaultMaxJSONBodyBytes)
}

// DecodeJSON reads one bounded JSON value without applying struct or explicit
// request validation. It is reserved for dynamic JSON objects whose handler
// immediately enforces an explicit field allowlist and domain validation.
func DecodeJSON(r *http.Request, v any) error {
	return decodeJSONWithLimit(r, v, DefaultMaxJSONBodyBytes)
}

// DecodeWithLimit applies a route-specific maximum body size. Use Decode for
// the platform default and this function only when a smaller contract is
// appropriate. Raw signed webhooks must use their provider verifier instead.
func DecodeWithLimit(r *http.Request, v any, maxBodyBytes int64) error {
	if err := decodeJSONWithLimit(r, v, maxBodyBytes); err != nil {
		return err
	}

	if err := ValidateStruct(v); err != nil {
		return err
	}
	if val, ok := v.(validator); ok {
		if err := val.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func decodeJSONWithLimit(r *http.Request, v any, maxBodyBytes int64) error {
	if maxBodyBytes <= 0 {
		return errors.New("maximum JSON body size must be positive")
	}
	if err := requireJSONContentType(r.Header.Get("Content-Type")); err != nil {
		return err
	}

	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return HumanizeJSONDecodeError(err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ErrMultipleJSONValues
		}
		return fmt.Errorf("%w: %v", ErrMultipleJSONValues, HumanizeJSONDecodeError(err))
	}

	return nil
}

func requireJSONContentType(header string) error {
	if strings.TrimSpace(header) == "" {
		return ErrInvalidJSONContentType
	}

	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSONContentType, err)
	}
	mediaType = strings.ToLower(mediaType)
	if mediaType != "application/json" && !strings.HasSuffix(mediaType, "+json") {
		return ErrInvalidJSONContentType
	}
	return nil
}
