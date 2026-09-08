package workerbootstrap

import (
	"context"
	"errors"
	"testing"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type workerMayaStoryReaderStub struct {
	scopes []storydomain.MutationScope
	ids    []uuid.UUID
	err    error
}

func (r *workerMayaStoryReaderStub) GetStoryForMutation(_ context.Context, scope storydomain.MutationScope, storyID uuid.UUID) (storydomain.Story, error) {
	r.scopes = append(r.scopes, scope)
	r.ids = append(r.ids, storyID)
	estimate := int16(3)
	return storydomain.Story{ID: storyID, Workspace: scope.WorkspaceID, EstimateScheme: "fibonacci", EstimateValue: &estimate}, r.err
}

func TestWorkerMayaStoryReadsBindSystemIdentityForEachWorkspace(t *testing.T) {
	t.Parallel()
	reader := &workerMayaStoryReaderStub{}
	systemID := uuid.New()
	service := workerMayaStories{reader: reader, actorID: systemID}
	ctx := context.Background() // Queue tasks have no authenticated HTTP actor.
	for _, workspaceID := range []uuid.UUID{uuid.New(), uuid.New()} {
		storyID := uuid.New()
		story, err := service.Get(ctx, storyID, workspaceID)
		require.NoError(t, err)
		require.Equal(t, storyID, story.ID)
		require.Equal(t, workspaceID, story.Workspace)
		require.NotNil(t, story.EstimateLabel)
		scope := reader.scopes[len(reader.scopes)-1]
		require.NoError(t, scope.Validate())
		require.Equal(t, auth.PrincipalSystem, scope.Actor.Kind)
		require.Equal(t, systemID, scope.Actor.PrincipalID)
		require.Equal(t, workspaceID, scope.Actor.WorkspaceID)
		require.Equal(t, workspaceID, scope.WorkspaceID)
		require.Equal(t, []auth.Scope{auth.ScopeStoriesWrite}, scope.Actor.Scopes.Values())
		require.Equal(t, systemID, *scope.ActivityUser)
		require.Equal(t, storyID, reader.ids[len(reader.ids)-1])
	}
	_, err := auth.GetActor(ctx)
	require.ErrorIs(t, err, auth.ErrActorNotFound)
}

func TestWorkerMayaStoryReadsRejectMissingIdentity(t *testing.T) {
	t.Parallel()
	for _, missing := range []string{"reader", "actor", "workspace", "story"} {
		t.Run(missing, func(t *testing.T) {
			reader := &workerMayaStoryReaderStub{}
			service := workerMayaStories{reader: reader, actorID: uuid.New()}
			storyID, workspaceID := uuid.New(), uuid.New()
			switch missing {
			case "reader":
				service.reader = nil
			case "actor":
				service.actorID = uuid.Nil
			case "workspace":
				workspaceID = uuid.Nil
			case "story":
				storyID = uuid.Nil
			}
			_, err := service.Get(context.Background(), storyID, workspaceID)
			require.Error(t, err)
			require.Empty(t, reader.scopes)
		})
	}
}

func TestWorkerMayaStoryReadsPreserveRepositoryErrors(t *testing.T) {
	t.Parallel()
	for _, expected := range []error{storydomain.ErrNotFound, storydomain.ErrMutationForbidden, errors.New("database unavailable")} {
		reader := &workerMayaStoryReaderStub{err: expected}
		service := workerMayaStories{reader: reader, actorID: uuid.New()}
		_, err := service.Get(context.Background(), uuid.New(), uuid.New())
		require.ErrorIs(t, err, expected)
	}
}
