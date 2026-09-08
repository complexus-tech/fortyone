package githubadapter

import (
	"context"
	"testing"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
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
	adapter := NewStoryService(backend, uuid.New())
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
	actor        auth.Actor
	systemReadID uuid.UUID
	userReads    int
	activity     stories.CoreActivity
	newComment   stories.CoreNewComment
	comment      stories.CoreComment
}

func (stub *storyBackendStub) Get(
	_ context.Context,
	storyID, workspaceID uuid.UUID,
) (stories.CoreSingleStory, error) {
	stub.userReads++
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

func (stub *storyBackendStub) RecordActivity(ctx context.Context, activity stories.CoreActivity) error {
	stub.actor, _ = auth.GetActor(ctx)
	stub.activity = activity
	return nil
}

func (stub *storyBackendStub) CreateCommentExternal(
	ctx context.Context,
	_, _ uuid.UUID,
	comment stories.CoreNewComment,
) (stories.CoreComment, error) {
	stub.actor, _ = auth.GetActor(ctx)
	stub.newComment = comment
	return stub.comment, nil
}

func (backend *storyBackendStub) GetForSystem(ctx context.Context, actorID, storyID, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	backend.systemReadID = actorID
	return stories.CoreSingleStory{ID: storyID, Workspace: workspaceID}, nil
}

func TestBackgroundGitHubReadsUseSystemAndInteractiveReadsKeepUser(t *testing.T) {
	t.Parallel()
	backend := &storyBackendStub{}
	systemID, storyID, workspaceID := uuid.New(), uuid.New(), uuid.New()
	adapter := NewStoryService(backend, systemID)
	_, err := adapter.Get(context.Background(), storyID, workspaceID)
	require.NoError(t, err)
	require.Equal(t, systemID, backend.systemReadID)
	require.Zero(t, backend.userReads)
	_, err = adapter.Get(auth.SetUserID(context.Background(), uuid.New()), storyID, workspaceID)
	require.NoError(t, err)
	require.Equal(t, 1, backend.userReads)
}

func TestGitHubMappedAuthorPreservesAuthorityAndRejectsImpersonation(t *testing.T) {
	t.Parallel()
	systemID, authorID, workspaceID, teamID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	teams, err := auth.RestrictedTeamAccess(teamID)
	require.NoError(t, err)
	system, err := auth.NewActor(systemID, auth.PrincipalSystem, uuid.Nil,
		auth.MustScopeSet(auth.ScopeStoriesWrite, auth.ScopeCommentsWrite), teams)
	require.NoError(t, err)
	system, err = system.WithWorkspace(workspaceID)
	require.NoError(t, err)
	ctx, err := auth.SetActor(context.Background(), system)
	require.NoError(t, err)
	backend := &storyBackendStub{}
	adapter := NewStoryService(backend, systemID)
	activity := github.StoryActivity{UserID: authorID, WorkspaceID: workspaceID, StoryID: uuid.New()}
	require.NoError(t, adapter.RecordActivity(ctx, activity))
	require.Equal(t, authorID, backend.actor.PrincipalID)
	require.Equal(t, auth.PrincipalHumanUser, backend.actor.Kind)
	require.Equal(t, system.Scopes, backend.actor.Scopes)
	require.Equal(t, teams, backend.actor.TeamAccess)
	_, err = adapter.CreateCommentExternal(ctx, authorID, workspaceID, github.NewStoryComment{UserID: authorID, StoryID: activity.StoryID, Comment: "Reviewed"})
	require.NoError(t, err)
	require.Equal(t, authorID, backend.actor.PrincipalID)
	original, err := auth.GetActor(ctx)
	require.NoError(t, err)
	require.Equal(t, system, original)
	require.ErrorIs(t, adapter.RecordActivity(auth.SetUserID(context.Background(), uuid.New()), activity), stories.ErrStoryMutationForbidden)
	activity.WorkspaceID = uuid.New()
	require.ErrorIs(t, adapter.RecordActivity(ctx, activity), stories.ErrStoryMutationForbidden)
}
