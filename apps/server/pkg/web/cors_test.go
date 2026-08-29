package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOriginPolicyUsesExactConfiguredOrigins(t *testing.T) {
	t.Parallel()

	policy, err := NewOriginPolicy("https://app.fortyone.app, http://localhost:3000")
	if err != nil {
		t.Fatalf("new origin policy: %v", err)
	}
	tests := []struct {
		origin string
		want   string
	}{
		{origin: "https://app.fortyone.app", want: "https://app.fortyone.app"},
		{origin: "http://localhost:3000", want: "http://localhost:3000"},
		{origin: "https://APP.FORTYONE.APP", want: "https://APP.FORTYONE.APP"},
		{origin: "https://preview.fortyone.app"},
		{origin: "https://app.fortyone.app.attacker.example"},
		{origin: "null"},
	}
	for _, test := range tests {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Origin", test.origin)
		if got := policy.AllowedOrigin(request); got != test.want {
			t.Errorf("AllowedOrigin(%q) = %q, want %q", test.origin, got, test.want)
		}
	}
}

func TestOriginPolicyRejectsUnsafeConfiguration(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"*",
		"null",
		"ftp://app.fortyone.app",
		"https://app.fortyone.app/path",
		"https://app.fortyone.app?query=yes",
		"https://user@app.fortyone.app",
		tooManyOrigins(),
	}
	for _, raw := range tests {
		if _, err := NewOriginPolicy(raw); !errors.Is(err, ErrInvalidAllowedOrigin) {
			t.Errorf("NewOriginPolicy(%q) error = %v, want ErrInvalidAllowedOrigin", raw, err)
		}
	}
}

func tooManyOrigins() string {
	origins := make([]string, 0, maxAllowedOrigins+1)
	for index := 0; index <= maxAllowedOrigins; index++ {
		origins = append(origins, fmt.Sprintf("https://app-%d.example", index))
	}
	return strings.Join(origins, ",")
}

func TestOriginPolicyRequiresHTTPSWhenRequested(t *testing.T) {
	t.Parallel()

	secure, err := NewOriginPolicy("https://app.fortyone.app")
	if err != nil {
		t.Fatalf("secure policy: %v", err)
	}
	if err := secure.ValidateHTTPS(); err != nil {
		t.Fatalf("ValidateHTTPS: %v", err)
	}
	local, err := NewOriginPolicy("http://localhost:3000")
	if err != nil {
		t.Fatalf("local policy: %v", err)
	}
	if err := local.ValidateHTTPS(); !errors.Is(err, ErrInvalidAllowedOrigin) {
		t.Fatalf("ValidateHTTPS error = %v, want ErrInvalidAllowedOrigin", err)
	}
}
