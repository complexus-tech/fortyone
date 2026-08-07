package storieshttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestRedirectStoryMediaUsesTemporaryNoStoreResponse(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/story-media", nil)
	recorder := httptest.NewRecorder()

	redirectStoryMedia(recorder, request, "https://storage.test/fresh-signature")

	response := recorder.Result()
	defer response.Body.Close()
	if response.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusTemporaryRedirect)
	}
	if location := response.Header.Get("Location"); location != "https://storage.test/fresh-signature" {
		t.Fatalf("location = %q", location)
	}
	if cacheControl := response.Header.Get("Cache-Control"); cacheControl != "private, no-store" {
		t.Fatalf("cache-control = %q", cacheControl)
	}
	if contentTypeOptions := response.Header.Get("X-Content-Type-Options"); contentTypeOptions != "nosniff" {
		t.Fatalf("x-content-type-options = %q", contentTypeOptions)
	}
}

func TestStoryMediaRoutesRequireAuthentication(t *testing.T) {
	storyID := uuid.New()
	attachmentID := uuid.New()
	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/workspaces/acme/stories/" + storyID.String() + "/media"},
		{method: http.MethodGet, path: "/workspaces/acme/stories/" + storyID.String() + "/media/" + attachmentID.String()},
		{method: http.MethodDelete, path: "/workspaces/acme/stories/" + storyID.String() + "/media/" + attachmentID.String()},
	}

	app := web.New(make(chan os.Signal, 1), nil)
	Routes(Config{SecretKey: "story-media-test-secret"}, app)
	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
			}
		})
	}
}
