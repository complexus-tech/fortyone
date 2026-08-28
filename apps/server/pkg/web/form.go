package web

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"
)

var ErrInvalidFormContentType = errors.New("request content type must be application/x-www-form-urlencoded")

// ParseURLForm parses an application/x-www-form-urlencoded body under an
// explicit whole-request limit. Callers handling credentials must read values
// from Request.PostForm so query parameters cannot substitute for body fields.
func ParseURLForm(w http.ResponseWriter, r *http.Request, maxBodyBytes int64) error {
	if w == nil {
		return errors.New("form response writer is required")
	}
	if r == nil || r.Body == nil {
		return errors.New("form request body is required")
	}
	if maxBodyBytes <= 0 {
		return ErrInvalidBodyLimit
	}

	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(r.Header.Get("Content-Type")))
	if err != nil || !strings.EqualFold(mediaType, "application/x-www-form-urlencoded") {
		return ErrInvalidFormContentType
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := r.ParseForm(); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return fmt.Errorf("%w: maximum is %d bytes", ErrRequestBodyTooLarge, maxBodyBytes)
		}
		return fmt.Errorf("parse URL-encoded form: %w", err)
	}
	return nil
}
