package sprints_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	"github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type repositoryStub struct {
	query sprintdomain.ListQuery
	now   time.Time
}

func (stub *repositoryStub) List(_ context.Context, query sprintdomain.ListQuery) ([]sprintdomain.Sprint, error) {
	stub.query = query
	return []sprintdomain.Sprint{}, nil
}
func (*repositoryStub) Running(context.Context, uuid.UUID, uuid.UUID, time.Time) ([]sprintdomain.Sprint, error) {
	return nil, nil
}
func (*repositoryStub) GetByID(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (sprintdomain.Sprint, error) {
	return sprintdomain.Sprint{}, nil
}
func (*repositoryStub) Create(context.Context, sprintdomain.CreateCommand) (sprintdomain.Sprint, error) {
	return sprintdomain.Sprint{}, nil
}
func (*repositoryStub) Update(context.Context, sprintdomain.UpdateCommand) (sprintdomain.Sprint, error) {
	return sprintdomain.Sprint{}, nil
}
func (*repositoryStub) Delete(context.Context, sprintdomain.DeleteCommand) error { return nil }
func (stub *repositoryStub) GetAnalytics(_ context.Context, _, _, _ uuid.UUID, now time.Time) (sprintdomain.Analytics, error) {
	stub.now = now
	return sprintdomain.Analytics{}, nil
}

func TestListRejectsUnknownDynamicFilter(t *testing.T) {
	t.Parallel()

	service := sprints.New(testLogger(), &repositoryStub{})
	_, err := service.List(context.Background(), uuid.New(), uuid.New(), map[string]any{"workspace_id; DROP TABLE sprints": uuid.New()})
	if !errors.Is(err, sprintdomain.ErrInvalid) {
		t.Fatalf("error = %v, want invalid", err)
	}
}

func TestAnalyticsUsesInjectedClock(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC)
	repository := &repositoryStub{}
	service := sprints.NewWithClock(testLogger(), repository, testkit.NewFixedClock(now))
	if _, err := service.GetAnalytics(context.Background(), uuid.New(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("get analytics: %v", err)
	}
	if !repository.now.Equal(now) {
		t.Fatalf("repository now = %v, want %v", repository.now, now)
	}
}

func testLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelDebug, "sprints-test")
}
