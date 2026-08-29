package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const DefaultMaxPathParameterBytes = 512

var (
	ErrMissingPathParameter = errors.New("path parameter is required")
	ErrInvalidPathParameter = errors.New("path parameter is invalid")
	ErrPathParameterTooLong = errors.New("path parameter is too long")
)

// PathParameterError deliberately omits the raw value. Invitation tokens and
// other bearer values can appear in paths and must not be copied into logs or
// error responses by a generic parser.
type PathParameterError struct {
	Name  string
	Cause error
}

func (e *PathParameterError) Error() string {
	return fmt.Sprintf("%s: %v", e.Name, e.Cause)
}

func (e *PathParameterError) Unwrap() error {
	return e.Cause
}

func RequiredPathParameter(r *http.Request, name string, maxBytes int) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", &PathParameterError{Name: "path parameter", Cause: ErrInvalidPathParameter}
	}
	if maxBytes <= 0 {
		return "", &PathParameterError{Name: name, Cause: ErrInvalidPathParameter}
	}
	value := strings.TrimSpace(r.PathValue(name))
	if value == "" {
		return "", &PathParameterError{Name: name, Cause: ErrMissingPathParameter}
	}
	if len(value) > maxBytes {
		return "", &PathParameterError{Name: name, Cause: ErrPathParameterTooLong}
	}
	return value, nil
}

// UUIDPathParameter parses a required, non-zero UUID without ever including
// the attacker-controlled raw path value in its error.
func UUIDPathParameter(r *http.Request, name string) (uuid.UUID, error) {
	value, err := RequiredPathParameter(r, name, DefaultMaxPathParameterBytes)
	if err != nil {
		return uuid.Nil, err
	}
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil {
		return uuid.Nil, &PathParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidPathParameter}
	}
	return parsed, nil
}
