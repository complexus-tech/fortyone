package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadBoundedBodyPreservesExactBytesAtLimit(t *testing.T) {
	t.Parallel()

	const raw = "{\n  \"event\": \"created\"\n}"
	request := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(raw))
	body, err := ReadBoundedBody(httptest.NewRecorder(), request, int64(len(raw)))
	if err != nil {
		t.Fatalf("ReadBoundedBody() error = %v", err)
	}
	if string(body) != raw {
		t.Fatalf("ReadBoundedBody() = %q, want exact %q", body, raw)
	}
}

func TestReadBoundedBodyRejectsOversizedPayload(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("12345"))
	body, err := ReadBoundedBody(httptest.NewRecorder(), request, 4)
	if !errors.Is(err, ErrRequestBodyTooLarge) {
		t.Fatalf("ReadBoundedBody() error = %v, want ErrRequestBodyTooLarge", err)
	}
	if body != nil {
		t.Fatalf("ReadBoundedBody() body = %q, want nil", body)
	}
}

func TestReadBoundedBodyRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
	if _, err := ReadBoundedBody(httptest.NewRecorder(), request, 0); !errors.Is(err, ErrInvalidBodyLimit) {
		t.Fatalf("zero limit error = %v, want ErrInvalidBodyLimit", err)
	}
	if _, err := ReadBoundedBody(httptest.NewRecorder(), nil, 1); err == nil {
		t.Fatal("nil request error = nil")
	}
}

func FuzzReadBoundedBodyNeverPanics(f *testing.F) {
	f.Add([]byte("{}"))
	f.Add([]byte{0x00, 0xff, '\n'})

	f.Fuzz(func(t *testing.T, raw []byte) {
		request := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(raw)))
		_, _ = ReadBoundedBody(httptest.NewRecorder(), request, 256)
	})
}
