package apiv1http

import (
	"context"
	"net/http"
	"testing"
	"time"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type guardedPatchStub struct {
	patch   stories.StoryPatch
	version time.Time
	calls   int
	err     error
}

func (stub *guardedPatchStub) UpdatePatchIfUnchanged(_ context.Context, _, _ uuid.UUID, version time.Time, patch stories.StoryPatch) error {
	stub.calls++
	stub.version, stub.patch = version, patch
	return stub.err
}

func TestPublicPatchPreservesOmittedAndClearedFields(t *testing.T) {
	t.Parallel()
	version := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	clear := []openapiv1.ComponentsResourcesUpdateStoryRequestClearFields{"assigneeId", "endDate"}
	patch, err := publicStoryPatch(openapiv1.ComponentsResourcesUpdateStoryRequest{ExpectedUpdatedAt: version, ClearFields: &clear})
	require.NoError(t, err)
	require.False(t, patch.Title.Specified())
	value, specified := patch.AssigneeID.Value()
	require.True(t, specified)
	require.Nil(t, value)
	assignee := uuid.New()
	_, err = publicStoryPatch(openapiv1.ComponentsResourcesUpdateStoryRequest{ExpectedUpdatedAt: version, AssigneeId: &assignee, ClearFields: &clear})
	require.ErrorIs(t, err, stories.ErrInvalidStoryMutation)
}

func TestPublicPatchRejectsConflictAndMissingScopeBeforeWrite(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	version := time.Now().UTC()
	title := "Updated campaign"
	body := openapiv1.ComponentsResourcesUpdateStoryRequest{ExpectedUpdatedAt: version, Title: &title}
	writer := &guardedPatchStub{err: stories.ErrStoryChanged}
	server := &server{storyPatches: writer}
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeStoriesWrite)
	ctx, err := platformauth.SetActor(t.Context(), actor)
	require.NoError(t, err)
	response, err := server.UpdateStory(ctx, openapiv1.UpdateStoryRequestObject{WorkspaceId: workspaceID, StoryId: uuid.New(), Body: &body})
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, response.(mutationProblem).status)
	require.Equal(t, version, writer.version)
	require.Equal(t, 1, writer.calls)
	reader := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeStoriesRead)
	ctx, err = platformauth.SetActor(t.Context(), reader)
	require.NoError(t, err)
	response, err = server.UpdateStory(ctx, openapiv1.UpdateStoryRequestObject{WorkspaceId: workspaceID, Body: &body})
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, response.(mutationProblem).status)
	require.Equal(t, 1, writer.calls)
}
