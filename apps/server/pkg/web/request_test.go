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
	request.Header.Set("Content-Type", "application/json")
	request.Body = http.MaxBytesReader(recorder, request.Body, 16)
	var input struct {
		Body string `json:"body"`
	}

	err := Decode(request, &input)

	if !errors.Is(err, ErrRequestBodyTooLarge) {
		t.Fatalf("Decode() error = %v, want ErrRequestBodyTooLarge", err)
	}
}

func TestDecodeEnforcesJSONContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		body        string
		wantErr     error
	}{
		{name: "valid", contentType: "application/json; charset=utf-8", body: `{"name":"Ada"}`},
		{name: "structured JSON type", contentType: "application/merge-patch+json", body: `{"name":"Ada"}`},
		{name: "missing content type", body: `{"name":"Ada"}`, wantErr: ErrInvalidJSONContentType},
		{name: "wrong content type", contentType: "text/plain", body: `{"name":"Ada"}`, wantErr: ErrInvalidJSONContentType},
		{name: "unknown field", contentType: "application/json", body: `{"name":"Ada","admin":true}`},
		{name: "trailing value", contentType: "application/json", body: `{"name":"Ada"} {"name":"Grace"}`, wantErr: ErrMultipleJSONValues},
		{name: "empty", contentType: "application/json", body: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(test.body))
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}
			var input struct {
				Name string `json:"name"`
			}

			err := Decode(request, &input)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("Decode() error = %v, want %v", err, test.wantErr)
				}
				return
			}
			if test.name == "unknown field" {
				if err == nil || err.Error() != "admin is not a valid field" {
					t.Fatalf("Decode() error = %v, want unknown-field error", err)
				}
				return
			}
			if test.name == "empty" {
				if err == nil || err.Error() != "request body is required" {
					t.Fatalf("Decode() error = %v, want required-body error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
		})
	}
}

func TestDecodeRunsExplicitValidation(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("name is required")
	input := &validatingRequest{err: wantErr}
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Ada"}`))
	request.Header.Set("Content-Type", "application/json")

	err := Decode(request, input)

	if !errors.Is(err, wantErr) {
		t.Fatalf("Decode() error = %v, want validation error", err)
	}
}

func TestDecodeRunsSharedTagValidationBeforeExplicitValidation(t *testing.T) {
	t.Parallel()

	input := &validatingRequest{err: errors.New("explicit validation should not run")}
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":""}`))
	request.Header.Set("Content-Type", "application/json")

	err := Decode(request, input)
	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("Decode() error = %v, want ValidationError", err)
	}
	if len(validationError.Violations) != 1 || validationError.Violations[0].Field != "name" {
		t.Fatalf("violations = %#v, want name violation", validationError.Violations)
	}
}

func TestDecodeWithLimitRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"too long"}`))
	request.Header.Set("Content-Type", "application/json")
	var input struct {
		Name string `json:"name"`
	}

	err := DecodeWithLimit(request, &input, 8)

	if !errors.Is(err, ErrRequestBodyTooLarge) {
		t.Fatalf("DecodeWithLimit() error = %v, want ErrRequestBodyTooLarge", err)
	}
}

func FuzzDecodeWithLimitNeverPanics(f *testing.F) {
	f.Add([]byte(`{"name":"Ada"}`), "application/json")
	f.Add([]byte(`{"name":"Ada"} {"name":"Grace"}`), "application/json")
	f.Add([]byte{0xff, 0x00, '{'}, "application/json")
	f.Add([]byte(`{"name":1}`), "application/merge-patch+json")

	f.Fuzz(func(t *testing.T, body []byte, contentType string) {
		request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(string(body)))
		request.Header.Set("Content-Type", contentType)
		var input struct {
			Name string `json:"name"`
		}
		_ = DecodeWithLimit(request, &input, 256)
	})
}

type validatingRequest struct {
	Name string `json:"name" validate:"required"`
	err  error
}

func (r validatingRequest) Validate() error {
	return r.err
}
