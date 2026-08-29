package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseURLFormKeepsBodyValuesSeparateFromQuery(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/oauth/token?token=query-secret", strings.NewReader("token=body-secret&scope=read"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	if err := ParseURLForm(httptest.NewRecorder(), request, 1024); err != nil {
		t.Fatalf("ParseURLForm() error = %v", err)
	}
	if got := request.PostForm.Get("token"); got != "body-secret" {
		t.Fatalf("PostForm token = %q, want body-secret", got)
	}
	if got := request.Form["token"]; len(got) != 2 {
		t.Fatalf("combined Form token values = %#v, want body and query values", got)
	}
}

func TestParseURLFormRejectsOversizedAndWrongContentType(t *testing.T) {
	t.Parallel()

	oversized := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader("token=12345"))
	oversized.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := ParseURLForm(httptest.NewRecorder(), oversized, 4); !errors.Is(err, ErrRequestBodyTooLarge) {
		t.Fatalf("oversized form error = %v, want ErrRequestBodyTooLarge", err)
	}

	jsonRequest := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(`{"token":"secret"}`))
	jsonRequest.Header.Set("Content-Type", "application/json")
	if err := ParseURLForm(httptest.NewRecorder(), jsonRequest, 1024); !errors.Is(err, ErrInvalidFormContentType) {
		t.Fatalf("JSON form error = %v, want ErrInvalidFormContentType", err)
	}
}

func FuzzParseURLFormNeverPanics(f *testing.F) {
	f.Add("grant_type=authorization_code&code=abc")
	f.Add("token=%ZZ")
	f.Add("a=1&a=2")

	f.Fuzz(func(t *testing.T, body string) {
		request := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		_ = ParseURLForm(httptest.NewRecorder(), request, 256)
	})
}
