package githubadapter

import (
	"context"
	"testing"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRequestStoreMapsOnlyGitHubOwnedContractAndClonesMetadata(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	backend := &requestBackendStub{
		upsertResult: integrationrequests.CoreIntegrationRequest{
			WorkspaceID: workspaceID,
			Provider:    integrationrequests.ProviderGitHub,
			SourceType:  integrationrequests.SourceTypeIssue,
			Metadata:    map[string]any{"repository_id": "repository"},
		},
	}
	metadata := map[string]any{"repository_id": "repository"}
	store := NewRequestStore(backend)

	result, err := store.UpsertPending(context.Background(), github.UpsertIntegrationRequestInput{
		WorkspaceID:      workspaceID,
		TeamID:           uuid.New(),
		Provider:         integrationrequests.ProviderGitHub,
		SourceType:       integrationrequests.SourceTypeIssue,
		SourceExternalID: "41",
		Title:            "Typed boundaries",
		Priority:         "High",
		Metadata:         metadata,
	})
	require.NoError(t, err)
	require.Equal(t, "41", backend.upsertInput.SourceExternalID)
	require.Equal(t, "High", backend.upsertInput.Priority)
	require.Equal(t, "repository", result.Metadata["repository_id"])

	metadata["repository_id"] = "mutated"
	backend.upsertResult.Metadata["repository_id"] = "also-mutated"
	require.Equal(t, "repository", backend.upsertInput.Metadata["repository_id"])
	require.Equal(t, "repository", result.Metadata["repository_id"])
}

func TestStoryServiceMapsActivityAndCommentCommands(t *testing.T) {
	t.Parallel()

	backend := &storyBackendStub{comment: stories.CoreComment{ID: uuid.New()}}
	adapter := NewStoryService(backend)
	workspaceID, storyID, actorID := uuid.New(), uuid.New(), uuid.New()
	reason := "provider automation"

	err := adapter.RecordActivity(context.Background(), github.StoryActivity{
		StoryID: storyID, UserID: actorID, WorkspaceID: workspaceID,
		Type: "link", Field: "github_review", CurrentValue: "approved",
		NewValue: "https://example.invalid/review", Reason: &reason,
	})
	require.NoError(t, err)
	require.Equal(t, storyID, backend.activity.StoryID)
	require.Equal(t, &reason, backend.activity.Reason)

	comment, err := adapter.CreateCommentExternal(context.Background(), actorID, workspaceID, github.NewStoryComment{
		StoryID: storyID, UserID: actorID, Comment: "Reviewed", Mentions: []uuid.UUID{uuid.New()},
	})
	require.NoError(t, err)
	require.Equal(t, backend.comment.ID, comment.ID)
	require.Equal(t, "Reviewed", backend.newComment.Comment)
	require.Len(t, backend.newComment.Mentions, 1)
}

type requestBackendStub struct {
	upsertInput  integrationrequests.CoreUpsertRequestInput
	upsertResult integrationrequests.CoreIntegrationRequest
}

func (stub *requestBackendStub) UpsertPending(
	_ context.Context,
	input integrationrequests.CoreUpsertRequestInput,
) (integrationrequests.CoreIntegrationRequest, error) {
	stub.upsertInput = input
	return stub.upsertResult, nil
}

func (stub *requestBackendStub) Get(
	_ context.Context,
	_, _ uuid.UUID,
) (integrationrequests.CoreIntegrationRequest, error) {
	return stub.upsertResult, nil
}

type storyBackendStub struct {
	activity   stories.CoreActivity
	newComment stories.CoreNewComment
	comment    stories.CoreComment
}

func (stub *storyBackendStub) Get(
	_ context.Context,
	storyID, workspaceID uuid.UUID,
) (stories.CoreSingleStory, error) {
	return stories.CoreSingleStory{ID: storyID, Workspace: workspaceID}, nil
}

func (stub *storyBackendStub) UpdateExternalWithReason(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	uuid.UUID,
	map[string]any,
	string,
) error {
	return nil
}

func (stub *storyBackendStub) RecordActivity(_ context.Context, activity stories.CoreActivity) error {
	stub.activity = activity
	return nil
}

func (stub *storyBackendStub) CreateCommentExternal(
	_ context.Context,
	_, _ uuid.UUID,
	comment stories.CoreNewComment,
) (stories.CoreComment, error) {
	stub.newComment = comment
	return stub.comment, nil
}
