package workerbootstrap

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type accountDeletionFinalizerStub struct {
	calls int
	limit int
	ctx   context.Context
	err   error
}

type accountSubscriberCleanupStub struct {
	calls  int
	err    error
	update usersdomain.SubscriberUpdate
}

func (s *accountSubscriberCleanupStub) Dispatch(context.Context) (int, error) {
	s.calls++
	return 0, s.err
}
func (s *accountSubscriberCleanupStub) UpdateActiveAccount(_ context.Context, update usersdomain.SubscriberUpdate) error {
	s.update = update
	return nil
}

func TestSubscriberUpdateWorkerUsesLifecycleGuard(t *testing.T) {
	t.Parallel()
	subscribers := &accountSubscriberCleanupStub{}
	mux := buildTaskMux(taskMuxDependencies{SubscriberCleanup: subscribers})
	require.NoError(t, mux.ProcessTask(t.Context(), asynq.NewTask(tasks.TypeSubscriberUpdate, []byte(`{"email":"member@example.com","fullName":"Stale queued name"}`))))
	require.Equal(t, "member@example.com", subscribers.update.Email)
	require.ErrorIs(t, mux.ProcessTask(t.Context(), asynq.NewTask(tasks.TypeSubscriberUpdate, []byte(`{"email":123}`))), asynq.SkipRetry)
}

func TestSubscriberOnboardingWorkersPreserveListsAndOriginalIdentity(t *testing.T) {
	t.Parallel()
	subscribers := &accountSubscriberCleanupStub{}
	mux := buildTaskMux(taskMuxDependencies{SubscriberCleanup: subscribers})
	userID := uuid.New()
	payload := []byte(`{"userId":"` + userID.String() + `","email":"member@example.com","workspaceSlug":"acme","workspaceName":"Acme"}`)
	for _, taskType := range []string{tasks.TypeUserOnboardingStart, tasks.TypeWorkspaceTrialStart} {
		require.NoError(t, mux.ProcessTask(t.Context(), asynq.NewTask(taskType, payload)))
		require.Equal(t, userID, subscribers.update.UserID)
		if taskType == tasks.TypeUserOnboardingStart {
			require.Equal(t, []int64{6}, subscribers.update.ListIDs)
		} else {
			require.Equal(t, []int64{12}, subscribers.update.ListIDs)
			require.Equal(t, "acme", subscribers.update.Attributes["WORKSPACE_SLUG"])
		}
	}
}

func (s *accountDeletionFinalizerStub) FinalizePending(ctx context.Context, limit int) (int, error) {
	s.calls++
	s.limit = limit
	s.ctx = ctx
	return 1, s.err
}

type accountDeletionScheduleStub struct {
	spec string
	task *asynq.Task
	opts map[asynq.OptionType]interface{}
	err  error
}

func (s *accountDeletionScheduleStub) Register(spec string, task *asynq.Task, opts ...asynq.Option) (string, error) {
	s.spec, s.task = spec, task
	s.opts = make(map[asynq.OptionType]interface{}, len(opts))
	for _, opt := range opts {
		s.opts[opt.Type()] = opt.Value()
	}
	return "account-deletion-test", s.err
}

func TestAccountDeletionWorkerSchedulesBoundedFinalization(t *testing.T) {
	t.Parallel()
	mux := asynq.NewServeMux()
	scheduler := &accountDeletionScheduleStub{}
	manager := &accountDeletionFinalizerStub{}
	log := logger.NewWithJSON(io.Discard, slog.LevelDebug, "test")
	require.NoError(t, registerAccountDeletionFinalization(mux, scheduler, manager, &accountSubscriberCleanupStub{}, log))
	require.Equal(t, "*/1 * * * *", scheduler.spec)
	require.Equal(t, accountDeletionFinalizationTask, scheduler.task.Type())
	require.Empty(t, scheduler.task.Payload())
	require.Equal(t, "cleanup", scheduler.opts[asynq.QueueOpt])
	require.Equal(t, 2, scheduler.opts[asynq.MaxRetryOpt])
	require.Equal(t, 4*time.Minute, scheduler.opts[asynq.TimeoutOpt])
	require.Equal(t, 55*time.Second, scheduler.opts[asynq.UniqueOpt])
	ctx := context.Background()
	require.NoError(t, mux.ProcessTask(ctx, scheduler.task))
	require.Equal(t, 1, manager.calls)
	require.Equal(t, 50, manager.limit)
	require.Equal(t, ctx, manager.ctx)
}

func TestAccountDeletionWorkerPreservesFailureForRetry(t *testing.T) {
	t.Parallel()
	mux := asynq.NewServeMux()
	scheduler := &accountDeletionScheduleStub{}
	manager := &accountDeletionFinalizerStub{err: errors.New("database unavailable")}
	log := logger.NewWithJSON(io.Discard, slog.LevelDebug, "test")
	subscribers := &accountSubscriberCleanupStub{}
	require.NoError(t, registerAccountDeletionFinalization(mux, scheduler, manager, subscribers, log))
	require.ErrorIs(t, mux.ProcessTask(context.Background(), scheduler.task), manager.err)
	manager.err = nil
	require.NoError(t, mux.ProcessTask(context.Background(), scheduler.task))
	require.Equal(t, 2, manager.calls)
	require.Equal(t, 2, subscribers.calls)
}

func TestAccountDeletionWorkerHonorsCancellationBeforeDatabaseWork(t *testing.T) {
	t.Parallel()
	mux := asynq.NewServeMux()
	scheduler := &accountDeletionScheduleStub{}
	manager := &accountDeletionFinalizerStub{}
	log := logger.NewWithJSON(io.Discard, slog.LevelDebug, "test")
	require.NoError(t, registerAccountDeletionFinalization(mux, scheduler, manager, &accountSubscriberCleanupStub{}, log))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, mux.ProcessTask(ctx, scheduler.task), context.Canceled)
	require.Zero(t, manager.calls)
}

func TestAccountDeletionWorkerScheduleFailurePreventsStartup(t *testing.T) {
	t.Parallel()
	scheduler := &accountDeletionScheduleStub{err: errors.New("scheduler unavailable")}
	log := logger.NewWithJSON(io.Discard, slog.LevelDebug, "test")
	err := registerAccountDeletionFinalization(asynq.NewServeMux(), scheduler, &accountDeletionFinalizerStub{}, &accountSubscriberCleanupStub{}, log)
	require.ErrorIs(t, err, scheduler.err)
}
