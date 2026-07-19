package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodePreservesRequestBodyTooLargeError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/feedback",
		strings.NewReader(`{"body":"this request is intentionally too large"}`),
	)
	request.Body = http.MaxBytesReader(recorder, request.Body, 16)
	var input struct {
		Body string `json:"body"`
	}

	err := Decode(request, &input)

	if !errors.Is(err, ErrRequestBodyTooLarge) {
		t.Fatalf("Decode() error = %v, want ErrRequestBodyTooLarge", err)
	}
}
