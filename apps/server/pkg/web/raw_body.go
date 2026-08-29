package web

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrInvalidBodyLimit = errors.New("request body limit must be positive")

// ReadBoundedBody reads an exact, unmodified request body under a hard byte
// limit. It is intended for signature-verified webhooks and other protocols
// where decoding and re-encoding before verification would change the signed
// bytes. The request body is consumed exactly once.
func ReadBoundedBody(w http.ResponseWriter, r *http.Request, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, ErrInvalidBodyLimit
	}
	if r == nil || r.Body == nil {
		return nil, errors.New("request body is required")
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	body, err := io.ReadAll(r.Body)
	if err == nil {
		return body, nil
	}

	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		return nil, fmt.Errorf("%w: maximum is %d bytes", ErrRequestBodyTooLarge, maxBytes)
	}
	return nil, fmt.Errorf("read request body: %w", err)
}
