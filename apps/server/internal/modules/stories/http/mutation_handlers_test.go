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
		name string
		body string
	}{
		{name: "unknown field", body: `{"reporterId":"` + uuid.NewString() + `"}`},
		{name: "system owned field", body: `{"autoSchedulingStatus":"planning"}`},
		{name: "null required field", body: `{"title":null}`},
		{name: "fractional duration", body: `{"estimatedDurationMinutes":12.5}`},
		{name: "empty patch", body: `{}`},
		{name: "trailing JSON", body: `{"title":"valid"}{"title":"second"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := invokeStoryMutationHandler(t, (&Handlers{}).Update, test.body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
		})
	}
}

func TestBulkUpdateHandlerRejectsUnknownOuterAndMalformedTypedPatch(t *testing.T) {
	t.Parallel()

	storyID := uuid.NewString()
	tests := []struct {
		name string
		body string
	}{
		{
			name: "unknown outer field",
			body: `{"storyIds":["` + storyID + `"],"updates":{"title":"valid"},"atomic":true}`,
		},
		{
			name: "missing updates",
			body: `{"storyIds":["` + storyID + `"]}`,
		},
		{
			name: "null updates",
			body: `{"storyIds":["` + storyID + `"],"updates":null}`,
		},
		{
			name: "unknown patch field",
			body: `{"storyIds":["` + storyID + `"],"updates":{"workspaceId":"` + uuid.NewString() + `"}}`,
		},
		{
			name: "malformed story ids",
			body: `{"storyIds":["not-a-uuid"],"updates":{"title":"valid"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := invokeStoryMutationHandler(t, (&Handlers{}).BulkUpdate, test.body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
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
