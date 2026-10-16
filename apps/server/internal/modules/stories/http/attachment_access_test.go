package storieshttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestStoryAttachmentAccessNegotiatesPreviewAndDownloadWithoutCaching(t *testing.T) {
	accessURL := "https://storage.test/file?signature=fresh"
	for _, accept := range []string{"image/png", "application/json"} {
		t.Run(accept, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/attachment", nil)
			request.Header.Set("Accept", accept)
			recorder := httptest.NewRecorder()
			recorder.Header().Set("Vary", "Origin")
			if err := respondStoryAttachmentAccess(context.Background(), recorder, request, accessURL); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(recorder.Header().Get("Cache-Control"), "no-store") {
				t.Fatal("fresh signatures must not be cached")
			}
			if values := strings.Join(recorder.Header().Values("Vary"), ","); !strings.Contains(values, "Origin") || !strings.Contains(values, "Accept") {
				t.Fatalf("Vary = %s", values)
			}
			if accept == "application/json" {
				var result struct {
					Data struct {
						URL string `json:"url"`
					} `json:"data"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if recorder.Code != http.StatusOK || result.Data.URL != accessURL || recorder.Header().Get("Location") != "" {
					t.Fatalf("download resolution: status=%d body=%s", recorder.Code, recorder.Body.String())
				}
			} else if recorder.Code != http.StatusTemporaryRedirect || recorder.Header().Get("Location") != accessURL {
				t.Fatalf("preview: status=%d location=%s", recorder.Code, recorder.Header().Get("Location"))
			}
		})
	}
}

func TestStoryAttachmentResolverRequiresAuthentication(t *testing.T) {
	app := web.New(make(chan os.Signal, 1), nil)
	Routes(Config{SecretKey: "attachment-access-test"}, app)
	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/workspaces/acme/stories/"+uuid.NewString()+"/attachments/"+uuid.NewString(), nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", recorder.Code)
	}
	if storyAttachmentStatus(attachments.ErrCommentReference) != http.StatusConflict {
		t.Fatal("referenced comment files must return a conflict")
	}
}

type attachmentReadRepositoryStub struct {
	calls int
	scope stories.StoryReadScope
}

func (repository *attachmentReadRepositoryStub) AuthorizeVisibleStoryAttachmentRead(_ context.Context, scope stories.StoryReadScope, _ uuid.UUID) error {
	repository.calls++
	repository.scope = scope
	return stories.ErrNotFound
}

func TestStoryAttachmentResolverNeverSignsWithoutStoryReadAccess(t *testing.T) {
	for _, test := range []struct {
		name      string
		scopes    platformauth.ScopeSet
		status    int
		wantCalls int
	}{
		{"hidden or deleted story", platformauth.MustScopeSet(platformauth.ScopeStoriesRead), http.StatusNotFound, 1},
		{"credential without story read scope", platformauth.MustScopeSet(platformauth.ScopeTeamsRead), http.StatusForbidden, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspaceID, actorID, storyID, attachmentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
			repository := &attachmentReadRepositoryStub{}
			// A nil attachment service also proves denial happens before signing.
			handler := &Handlers{stories: stories.New(nil, repository, nil, nil)}
			actor, err := platformauth.NewActor(actorID, platformauth.PrincipalPersonalToken, uuid.New(), test.scopes, platformauth.UnrestrictedTeamAccess())
			if err != nil {
				t.Fatal(err)
			}
			ctx, err := platformauth.SetActor(t.Context(), actor)
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodGet, "/attachment", nil)
			request.SetPathValue("workspaceSlug", "acme")
			request.SetPathValue("id", storyID.String())
			request.SetPathValue("attachmentId", attachmentID.String())
			recorder := httptest.NewRecorder()
			resolver := mutationWorkspaceResolverStub{workspace: mid.WorkspaceInfo{ID: workspaceID, Slug: "acme", UserRole: string(mid.RoleMember)}}
			if err := mid.Workspace(nil, resolver)(handler.ResolveStoryAttachment)(ctx, recorder, request); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != test.status || repository.calls != test.wantCalls {
				t.Fatalf("status=%d calls=%d, want status=%d calls=%d", recorder.Code, repository.calls, test.status, test.wantCalls)
			}
			if repository.calls > 0 && (repository.scope.ActorID != actorID || repository.scope.WorkspaceID != workspaceID) {
				t.Fatalf("story visibility received wrong actor/tenant: %#v", repository.scope)
			}
		})
	}
}
