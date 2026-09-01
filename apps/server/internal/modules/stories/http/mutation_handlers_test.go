package storieshttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type mutationWorkspaceResolverStub struct {
	workspace mid.WorkspaceInfo
}

func (resolver mutationWorkspaceResolverStub) ResolveCurrentWorkspace(
	context.Context,
	string,
	uuid.UUID,
) (mid.WorkspaceInfo, error) {
	return resolver.workspace, nil
}

func (mutationWorkspaceResolverStub) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func TestUpdateHandlerRejectsUnknownSystemOwnedAndAmbiguousFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		wantMessage string
	}{
		{
			name:        "unknown field",
			body:        `{"reporterId":"` + uuid.NewString() + `"}`,
			wantMessage: `unknown story update field \"reporterId\"`,
		},
		{
			name:        "system owned field",
			body:        `{"autoSchedulingStatus":"planning"}`,
			wantMessage: `unknown story update field \"autoSchedulingStatus\"`,
		},
		{name: "null required field", body: `{"title":null}`, wantMessage: "title cannot be null"},
		{
			name:        "fractional duration",
			body:        `{"estimatedDurationMinutes":12.5}`,
			wantMessage: "invalid estimatedDurationMinutes",
		},
		{name: "empty patch", body: `{}`, wantMessage: "at least one field is required"},
		{
			name:        "trailing JSON",
			body:        `{"title":"valid"}{"title":"second"}`,
			wantMessage: "request body must contain exactly one JSON value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := invokeStoryMutationHandler(t, (&Handlers{}).Update, test.body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), test.wantMessage) {
				t.Fatalf("body = %s, want message containing %q", recorder.Body.String(), test.wantMessage)
			}
		})
	}
}

func TestBulkUpdateHandlerRejectsUnknownOuterAndMalformedTypedPatch(t *testing.T) {
	t.Parallel()

	storyID := uuid.NewString()
	tests := []struct {
		name        string
		body        string
		wantMessage string
	}{
		{
			name:        "unknown outer field",
			body:        `{"storyIds":["` + storyID + `"],"updates":{"title":"valid"},"atomic":true}`,
			wantMessage: `unknown bulk story update field \"atomic\"`,
		},
		{
			name:        "missing updates",
			body:        `{"storyIds":["` + storyID + `"]}`,
			wantMessage: "updates is required",
		},
		{
			name:        "null updates",
			body:        `{"storyIds":["` + storyID + `"],"updates":null}`,
			wantMessage: "updates is required",
		},
		{
			name:        "unknown patch field",
			body:        `{"storyIds":["` + storyID + `"],"updates":{"workspaceId":"` + uuid.NewString() + `"}}`,
			wantMessage: `unknown story update field \"workspaceId\"`,
		},
		{
			name:        "malformed story ids",
			body:        `{"storyIds":["not-a-uuid"],"updates":{"title":"valid"}}`,
			wantMessage: "invalid storyIds",
		},
		{
			name:        "empty patch",
			body:        `{"storyIds":["` + storyID + `"],"updates":{}}`,
			wantMessage: "at least one field is required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := invokeStoryMutationHandler(t, (&Handlers{}).BulkUpdate, test.body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), test.wantMessage) {
				t.Fatalf("body = %s, want message containing %q", recorder.Body.String(), test.wantMessage)
			}
		})
	}
}

func invokeStoryMutationHandler(t *testing.T, handler web.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()

	workspaceID := uuid.New()
	request := httptest.NewRequest(http.MethodPatch, "/workspaces/mutation/stories/"+uuid.NewString(), strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.SetPathValue("workspaceSlug", "mutation")
	request.SetPathValue("id", uuid.NewString())
	recorder := httptest.NewRecorder()
	resolver := mutationWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
		ID: workspaceID, Slug: "mutation", UserRole: string(mid.RoleMember),
	}}
	wrapped := mid.Workspace(nil, resolver)(handler)
	ctx := platformauth.SetUserID(t.Context(), uuid.New())
	if err := wrapped(ctx, recorder, request); err != nil {
		t.Fatalf("invoke story mutation handler: %v", err)
	}
	return recorder
}
