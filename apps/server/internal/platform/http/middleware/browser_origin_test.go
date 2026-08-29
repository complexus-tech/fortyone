package mid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/web"
)

func TestRequireTrustedBrowserOrigin(t *testing.T) {
	t.Parallel()

	policy, err := web.NewOriginPolicy("https://app.fortyone.app")
	if err != nil {
		t.Fatalf("new origin policy: %v", err)
	}
	tests := []struct {
		name       string
		method     string
		cookie     bool
		origin     string
		fetchSite  string
		wantStatus int
		wantNext   bool
	}{
		{name: "safe method", method: http.MethodGet, cookie: true, wantStatus: http.StatusNoContent, wantNext: true},
		{name: "non-browser credential", method: http.MethodPost, wantStatus: http.StatusNoContent, wantNext: true},
		{name: "configured origin", method: http.MethodPost, cookie: true, origin: "https://app.fortyone.app", wantStatus: http.StatusNoContent, wantNext: true},
		{name: "same-origin metadata fallback", method: http.MethodDelete, cookie: true, fetchSite: "same-origin", wantStatus: http.StatusNoContent, wantNext: true},
		{name: "sibling origin denied", method: http.MethodPost, cookie: true, origin: "https://preview.fortyone.app", wantStatus: http.StatusForbidden},
		{name: "same-site metadata denied", method: http.MethodPatch, cookie: true, fetchSite: "same-site", wantStatus: http.StatusForbidden},
		{name: "missing browser metadata denied", method: http.MethodPut, cookie: true, wantStatus: http.StatusForbidden},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			called := false
			handler := RequireTrustedBrowserOrigin(policy)(func(_ context.Context, writer http.ResponseWriter, _ *http.Request) error {
				called = true
				writer.WriteHeader(http.StatusNoContent)
				return nil
			})
			request := httptest.NewRequest(test.method, "/resource", nil)
			if test.cookie {
				request.AddCookie(&http.Cookie{Name: authCookieName, Value: "opaque-session"})
			}
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			if test.fetchSite != "" {
				request.Header.Set("Sec-Fetch-Site", test.fetchSite)
			}
			recorder := httptest.NewRecorder()
			if err := handler(context.Background(), recorder, request); err != nil {
				t.Fatalf("middleware returned error: %v", err)
			}
			if recorder.Code != test.wantStatus || called != test.wantNext {
				t.Fatalf("status/called = %d/%t, want %d/%t", recorder.Code, called, test.wantStatus, test.wantNext)
			}
		})
	}
}
