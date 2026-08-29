package safeio

import (
	"bytes"
	"errors"
	"math"
	"testing"
)

func TestReadAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload []byte
		limit   int64
		wantErr error
	}{
		{name: "below limit", payload: []byte("abc"), limit: 4},
		{name: "exact limit", payload: []byte("abcd"), limit: 4},
		{name: "over limit", payload: []byte("abcde"), limit: 4, wantErr: ErrLimitExceeded},
		{name: "zero limit", payload: []byte("a"), limit: 0, wantErr: ErrInvalidLimit},
		{name: "negative limit", payload: []byte("a"), limit: -1, wantErr: ErrInvalidLimit},
		{name: "overflowing limit", payload: []byte("a"), limit: math.MaxInt64, wantErr: ErrInvalidLimit},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := ReadAll(bytes.NewReader(test.payload), test.limit)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ReadAll() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr == nil && !bytes.Equal(got, test.payload) {
				t.Fatalf("ReadAll() = %q, want %q", got, test.payload)
			}
		})
	}
}

func TestReadAllRejectsNilReader(t *testing.T) {
	t.Parallel()

	if _, err := ReadAll(nil, 1); err == nil {
		t.Fatal("ReadAll() accepted a nil reader")
	}
}
