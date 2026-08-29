// Package safeio provides small, transport-neutral resource limits for data
// read from providers and other untrusted streams.
package safeio

import (
	"errors"
	"io"
	"math"
)

var (
	ErrInvalidLimit  = errors.New("read limit must be positive")
	ErrLimitExceeded = errors.New("reader exceeds the configured limit")
)

// ReadAll reads one stream without trusting external length metadata. The
// extra byte distinguishes an exact-boundary stream from a truncated one.
func ReadAll(reader io.Reader, maxBytes int64) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("reader is required")
	}
	if maxBytes <= 0 || maxBytes == math.MaxInt64 {
		return nil, ErrInvalidLimit
	}

	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrLimitExceeded
	}
	return data, nil
}
