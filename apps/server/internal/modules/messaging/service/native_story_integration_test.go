//go:build integration

package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// Uses the real migration chain, proposal repository and confirmation runtime.
// Story writes use the actual service/repository, including actor attribution,
// activity snapshots and the durable mutation event in the same transaction.
func TestNativeDescriptionPostgresProposalReceiptAndOuterRollback(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	executor, scope, memory, _, _ := nativeFixture(t)
	ctx := t.Context()
	story := memory.persisted[scope.SourceScope.StoryID]
	originalMarkup := *story.DescriptionHTML + `<p><a href="https://example.com/spec">External spec</a> and <a href="/workspace/work/WEB-43">Related story</a></p>`
	story.DescriptionHTML = &originalMarkup
	for _, seed := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO users (user_id,username,email) VALUES ($1,$2,$3)`, []any{scope.UserID, "native-" + scope.UserID.String(), scope.UserID.String() + "@example.test"}},
		{`INSERT INTO workspaces (workspace_id,name,slug,created_by) VALUES ($1,'Native test',$2,$3)`, []any{scope.WorkspaceID, "native-" + scope.WorkspaceID.String(), scope.UserID}},
		{`INSERT INTO teams (team_id,workspace_id,name,code,color) VALUES ($1,$2,'Web','WEB','#111111')`, []any{story.Team, scope.WorkspaceID}},
		{`INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'member')`, []any{scope.WorkspaceID, scope.UserID}},
		{`INSERT INTO team_members (team_id,user_id) VALUES ($1,$2)`, []any{story.Team, scope.UserID}},
		{`INSERT INTO stories (id,workspace_id,team_id,sequence_id,title,priority,description,description_html,updated_at) VALUES ($1,$2,$3,42,$4,'Low',$5,$6,$7)`, []any{story.ID, story.Workspace, story.Team, story.Title, story.Description, story.DescriptionHTML, story.UpdatedAt}},
	} {
		_, err := postgres.Pool.Exec(ctx, seed.query, seed.args...)
		require.NoError(t, err)
	}
	log := logger.NewWithText(io.Discard, slog.LevelError, "test")
	backend := &nativePostgresStories{mutationStoriesStub: memory, pool: postgres.Pool, service: stories.New(log, storiesrepository.NewMutationRepository(log, postgres.Pool), nil, nil)}
	repo := messagingrepository.New(postgres.Pool)
	executor.mutations.stories = backend
	executor.mutations.store = repo
	proposal := nativeProposal(t, executor, scope, map[string]any{"description": "Persisted rewritten requirements."})
	inspection, err := executor.InspectStoryMutation(ctx, scope, proposal.Token)
	require.NoError(t, err)
	require.True(t, inspection.DirectEligible)
	id, _, err := batchConfirmationID(proposal.Token)
	require.NoError(t, err)
	binding := StoryMutationConfirmationBinding{ConfirmationID: id, WorkspaceID: scope.WorkspaceID, UserID: scope.UserID, TokenHash: storyMutationTokenHash(proposal.Token)}
	loaded, err := repo.LoadStoryMutationConfirmation(ctx, binding)
	require.NoError(t, err)
	require.Equal(t, StoryMutationUpdate, loaded.Operation)
	require.Contains(t, string(loaded.Proposal), "Persisted rewritten requirements.")

	rollback := errors.New("abort outer invocation receipt")
	err = database.WithinUnitOfWork(ctx, postgres.Pool, func(txCtx context.Context, _ pgx.Tx) error {
		_, err := executor.ConfirmStoryMutation(txCtx, scope, proposal.Token)
		if err != nil {
			return err
		}
		current, err := backend.Get(txCtx, story.ID, scope.WorkspaceID)
		if err != nil {
			return err
		}
		require.Equal(t, "Persisted rewritten requirements.\n\nExternal spec\n\nRelated story", *current.Description)
		pendingReceipt, err := repo.LoadStoryMutationConfirmation(txCtx, binding)
		if err != nil {
			return err
		}
		require.Equal(t, StoryMutationConfirmationApplied, pendingReceipt.Status)
		require.NotNil(t, pendingReceipt.Result)
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	current, err := backend.Get(ctx, story.ID, scope.WorkspaceID)
	require.NoError(t, err)
	require.Equal(t, "Old prose", *current.Description)
	loaded, err = repo.LoadStoryMutationConfirmation(ctx, binding)
	require.NoError(t, err)
	require.Equal(t, StoryMutationConfirmationPending, loaded.Status)
	require.Nil(t, loaded.Result)
	require.NotEmpty(t, loaded.Proposal)
	var activityCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM story_activities WHERE story_id=$1`, story.ID).Scan(&activityCount))
	require.Zero(t, activityCount, "audit records must roll back with the invocation")

	err = database.WithinUnitOfWork(ctx, postgres.Pool, func(txCtx context.Context, _ pgx.Tx) error {
		_, err := executor.ConfirmStoryMutation(txCtx, scope, proposal.Token)
		return err
	})
	require.NoError(t, err)
	current, err = backend.Get(ctx, story.ID, scope.WorkspaceID)
	require.NoError(t, err)
	require.Equal(t, "Persisted rewritten requirements.\n\nExternal spec\n\nRelated story", *current.Description)
	require.Contains(t, *current.DescriptionHTML, `data-attachment-id="image-id"`)
	require.Contains(t, *current.DescriptionHTML, `data-attachment-id="video-id"`)
	require.Contains(t, *current.DescriptionHTML, `<a href="https://example.com/spec">External spec</a>`)
	require.Contains(t, *current.DescriptionHTML, `<a href="/workspace/work/WEB-43">Related story</a>`)
	loaded, err = repo.LoadStoryMutationConfirmation(ctx, binding)
	require.NoError(t, err)
	require.Equal(t, StoryMutationConfirmationApplied, loaded.Status)
	require.Empty(t, loaded.Proposal, "completed prose is redacted from the proposal store")
	require.Equal(t, story.ID, loaded.Result.StoryID)
	var activityActor uuid.UUID
	var oldValue, newValue json.RawMessage
	var reason string
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT user_id,old_value,new_value,reason FROM story_activities WHERE story_id=$1 AND field_changed='description'`, story.ID).Scan(&activityActor, &oldValue, &newValue, &reason))
	require.Equal(t, scope.UserID, activityActor)
	require.Equal(t, nativeStoryMutationReason, reason)
	require.JSONEq(t, `"Old prose"`, string(oldValue))
	var newText string
	require.NoError(t, json.Unmarshal(newValue, &newText))
	require.Equal(t, *current.Description, newText)
	result, err := executor.ConfirmStoryMutation(ctx, scope, proposal.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, result.Status)
	replayed, err := backend.Get(ctx, story.ID, scope.WorkspaceID)
	require.NoError(t, err)
	require.Equal(t, current.UpdatedAt, replayed.UpdatedAt, "receipt replay cannot repeat the story update")

	stale := nativeProposal(t, executor, scope, map[string]any{"description": "Stale rewrite"})
	_, err = postgres.Pool.Exec(ctx, `UPDATE stories SET description='Concurrent human edit',updated_at=updated_at+INTERVAL '1 second' WHERE id=$1`, story.ID)
	require.NoError(t, err)
	_, err = executor.ConfirmStoryMutation(ctx, scope, stale.Token)
	require.ErrorIs(t, err, ErrStaleMutation)
	current, err = backend.Get(ctx, story.ID, scope.WorkspaceID)
	require.NoError(t, err)
	require.Equal(t, "Concurrent human edit", *current.Description)

	// A different participant can accept precisely this proposal, but the real
	// story service must authorize and audit that participant's own identity.
	parent := nativeProposal(t, executor, scope, map[string]any{"description": "Accepted by another participant.", "command_intent": "proposal"})
	responding := scope
	responding.UserID = uuid.New()
	for _, seed := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO users (user_id,username,email) VALUES ($1,$2,$3)`, []any{responding.UserID, "responding-" + responding.UserID.String(), responding.UserID.String() + "@example.test"}},
		{`INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'member')`, []any{scope.WorkspaceID, responding.UserID}},
		{`INSERT INTO team_members (team_id,user_id) VALUES ($1,$2)`, []any{story.Team, responding.UserID}},
		{`DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, []any{story.Team, scope.UserID}},
		{`DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, []any{scope.WorkspaceID, scope.UserID}},
	} {
		_, err := postgres.Pool.Exec(ctx, seed.query, seed.args...)
		require.NoError(t, err)
	}
	accepted, err := executor.RebindStoryMutation(ctx, scope, responding, parent.Token)
	require.NoError(t, err)
	err = database.WithinUnitOfWork(ctx, postgres.Pool, func(txCtx context.Context, _ pgx.Tx) error {
		_, err := executor.ConfirmStoryMutation(txCtx, responding, accepted.Token)
		return err
	})
	require.NoError(t, err)
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT user_id,reason,new_value FROM story_activities WHERE story_id=$1 AND field_changed='description' AND user_id=$2`, story.ID, responding.UserID).Scan(&activityActor, &reason, &newValue))
	require.Equal(t, responding.UserID, activityActor)
	require.Equal(t, nativeStoryMutationReason, reason)
	require.Contains(t, string(newValue), "Accepted by another participant.")
}

type nativePostgresStories struct {
	*mutationStoriesStub
	pool    *pgxpool.Pool
	service *stories.Service
}

type nativeStoryQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func (backend *nativePostgresStories) query(ctx context.Context) nativeStoryQuery {
	if tx, ok := database.UnitTransaction(ctx); ok {
		return tx
	}
	return backend.pool
}

func (backend *nativePostgresStories) Get(ctx context.Context, id, workspaceID uuid.UUID) (messagingStory, error) {
	var story messagingStory
	err := backend.query(ctx).QueryRow(ctx, `SELECT id,workspace_id,team_id,sequence_id,title,priority,description,description_html,updated_at FROM stories WHERE id=$1 AND workspace_id=$2`, id, workspaceID).Scan(&story.ID, &story.Workspace, &story.Team, &story.SequenceID, &story.Title, &story.Priority, &story.Description, &story.DescriptionHTML, &story.UpdatedAt)
	return story, err
}

func (backend *nativePostgresStories) UpdateExternalUserActionIfUnchanged(ctx context.Context, actorID, id, workspaceID uuid.UUID, expected time.Time, updates map[string]any) error {
	return backend.service.UpdateExternalUserActionIfUnchanged(ctx, actorID, id, workspaceID, expected, updates)
}

func (backend *nativePostgresStories) UpdateExternalUserActionWithReasonIfUnchanged(ctx context.Context, actorID, id, workspaceID uuid.UUID, expected time.Time, updates map[string]any, reason string) error {
	return backend.service.UpdateExternalUserActionWithReasonIfUnchanged(ctx, actorID, id, workspaceID, expected, updates, reason)
}
