package storieshttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestChecklistRoutesRequireAuthentication(t *testing.T) {
	root := "/workspaces/acme/stories/" + uuid.NewString() + "/checklist"
	item := root + "/" + uuid.NewString()
	app := web.New(make(chan os.Signal, 1), nil)
	Routes(Config{SecretKey: "checklist-test-secret"}, app)
	for _, test := range []struct{ method, path string }{{"GET", root}, {"POST", root}, {"PATCH", item}, {"DELETE", item}, {"DELETE", root + "/items"}, {"PUT", root + "/order"}, {"POST", item + "/convert"}} {
		t.Run(test.method+test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			app.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, nil))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d", recorder.Code)
			}
		})
	}
}
func TestChecklistPatchRejectsExplicitNullBeforeMutation(t *testing.T) {
	for _, body := range []string{`{"title":null}`, `{"completed":null}`, `{"title":3}`, `{"completed":"yes"}`} {
		t.Run(body, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h := &Handlers{}
			if err := h.UpdateChecklistItem(context.Background(), w, r); err != nil {
				t.Fatal(err)
			}
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}
