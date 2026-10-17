package chatservice

import (
	"context"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type markAllThreadsRepositoryStub struct {
	Repository
	snapshot func(context.Context, domain.ActorScope) ([]domain.ThreadReadTarget, error)
	within   func(context.Context, func(context.Context) error) error
	markRead func(context.Context, domain.ReadCommand) (domain.ReadState, error)
}

func (r markAllThreadsRepositoryStub) UnreadThreadReadTargets(ctx context.Context, scope domain.ActorScope) ([]domain.ThreadReadTarget, error) {
	return r.snapshot(ctx, scope)
}

func (r markAllThreadsRepositoryStub) WithinUnit(ctx context.Context, operation func(context.Context) error) error {
	return r.within(ctx, operation)
}

func (r markAllThreadsRepositoryStub) MarkRead(ctx context.Context, command domain.ReadCommand) (domain.ReadState, error) {
	return r.markRead(ctx, command)
}

func TestMarkAllThreadsReadRequiresWriteScopeBeforeSnapshot(t *testing.T) {
	readOnly := testScope(t)
	readOnly.Actor.Scopes = auth.MustScopeSet(auth.ScopeChatRead)
	foreign := testScope(t)
	foreign.WorkspaceID = uuid.New()
	for _, scope := range []domain.ActorScope{{}, readOnly, foreign} {
		require.ErrorIs(t, New(nil).MarkAllThreadsRead(t.Context(), scope), domain.ErrForbidden)
	}
}

func TestMarkAllThreadsReadRetainsSnapshotAcrossMutationRetries(t *testing.T) {
	scope := testScope(t)
	targets := []domain.ThreadReadTarget{
		{ConversationID: uuid.New(), ThreadRootID: uuid.New(), ThroughMessageID: uuid.New()},
		{ConversationID: uuid.New(), ThreadRootID: uuid.New(), ThroughMessageID: uuid.New()},
	}
	snapshots, attempts := 0, 0
	var calls []domain.ReadCommand
	type unitKey struct{}
	repo := markAllThreadsRepositoryStub{
		snapshot: func(_ context.Context, requested domain.ActorScope) ([]domain.ThreadReadTarget, error) {
			snapshots++
			require.Equal(t, scope, requested)
			return append([]domain.ThreadReadTarget(nil), targets...), nil
		},
		within: func(ctx context.Context, operation func(context.Context) error) error {
			for range 2 {
				attempts++
				require.NoError(t, operation(context.WithValue(ctx, unitKey{}, attempts)))
				// A newer reply and a newly unread thread become visible before retry.
				targets[0].ThroughMessageID = uuid.New()
				targets = append(targets, domain.ThreadReadTarget{ConversationID: uuid.New(), ThreadRootID: uuid.New(), ThroughMessageID: uuid.New()})
			}
			return nil
		},
		markRead: func(ctx context.Context, command domain.ReadCommand) (domain.ReadState, error) {
			require.Equal(t, attempts, ctx.Value(unitKey{}))
			calls = append(calls, command)
			return domain.ReadState{}, nil
		},
	}
	require.NoError(t, New(repo).MarkAllThreadsRead(t.Context(), scope))
	require.Equal(t, 1, snapshots)
	require.Len(t, calls, 4)
	require.Equal(t, calls[:2], calls[2:], "retries must use the original explicit reply IDs")
	for _, command := range calls {
		require.Equal(t, scope, command.Scope)
		require.NotNil(t, command.ThreadRootID, "the timeline read cursor must not advance")
	}
}

func TestMarkAllThreadsReadDoesNotMutateAfterEmptyOrFailedSnapshot(t *testing.T) {
	failure := errors.New("snapshot unavailable")
	for _, snapshotError := range []error{nil, failure} {
		repo := markAllThreadsRepositoryStub{
			snapshot: func(context.Context, domain.ActorScope) ([]domain.ThreadReadTarget, error) {
				return nil, snapshotError
			},
		}
		require.ErrorIs(t, New(repo).MarkAllThreadsRead(t.Context(), testScope(t)), snapshotError)
	}
}

func TestMarkAllThreadsReadPropagatesMutationFailure(t *testing.T) {
	failure := domain.ErrNotFound
	calls := 0
	repo := markAllThreadsRepositoryStub{
		snapshot: func(context.Context, domain.ActorScope) ([]domain.ThreadReadTarget, error) {
			return []domain.ThreadReadTarget{{}, {}}, nil
		},
		within: func(ctx context.Context, operation func(context.Context) error) error {
			err := operation(ctx)
			require.ErrorIs(t, err, failure, "the unit must receive the failure to roll back")
			return err
		},
		markRead: func(context.Context, domain.ReadCommand) (domain.ReadState, error) {
			calls++
			return domain.ReadState{}, failure
		},
	}
	require.ErrorIs(t, New(repo).MarkAllThreadsRead(t.Context(), testScope(t)), failure)
	require.Equal(t, 1, calls)
}
