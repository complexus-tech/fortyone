package web

import (
	"errors"
	"fmt"
	"net/http"
)

const DefaultMultipartMemoryBytes int64 = 1 << 20

// ParseMultipartForm reads a multipart request through a hard body-size limit.
// maxBodyBytes includes MIME boundaries and field metadata, so callers should
// add a small, explicit overhead allowance to the largest accepted file.
// Temporary files are removed by RemoveMultipartForm.
func ParseMultipartForm(w http.ResponseWriter, r *http.Request, maxBodyBytes int64) error {
	if w == nil {
		return errors.New("multipart response writer is required")
	}
	if r == nil || r.Body == nil {
		return errors.New("multipart request body is required")
	}
	if maxBodyBytes <= 0 {
		return errors.New("maximum multipart body size must be positive")
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	// #nosec G120 -- MaxBytesReader above enforces the caller-provided hard body
	// limit; ParseMultipartForm's separate argument caps in-memory buffering.
	if err := r.ParseMultipartForm(DefaultMultipartMemoryBytes); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return fmt.Errorf("%w: maximum size is %d bytes", ErrRequestBodyTooLarge, maxBytesError.Limit)
		}
		return fmt.Errorf("parse multipart form: %w", err)
	}
	return nil
}

// RemoveMultipartForm releases any temporary files created while parsing.
// It is safe to call after a successful ParseMultipartForm.
func RemoveMultipartForm(r *http.Request) error {
	if r == nil || r.MultipartForm == nil {
		return nil
	}
	if err := r.MultipartForm.RemoveAll(); err != nil {
		return fmt.Errorf("remove multipart temporary files: %w", err)
	}
	return nil
}
