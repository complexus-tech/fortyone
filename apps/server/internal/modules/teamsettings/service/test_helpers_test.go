package teamsettings

import (
	"context"
	"io"
	"log/slog"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type repositoryStub struct {
	isTeamMember bool
	err          error
	committed    bool
	sprint       CoreTeamSprintSettings
	story        CoreTeamStoryAutomationSettings
	estimation   CoreTeamEstimationSettings
}

func (r *repositoryStub) IsActiveTeamMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error) {
	return r.isTeamMember, r.err
}

func (r *repositoryStub) GetSprintSettings(context.Context, uuid.UUID, uuid.UUID) (CoreTeamSprintSettings, error) {
	return r.sprint, r.err
}

func (r *repositoryStub) UpdateSprintSettings(context.Context, uuid.UUID, uuid.UUID, CoreUpdateTeamSprintSettings, AuditActor) (CoreTeamSprintSettings, error) {
	if r.err == nil {
		r.committed = true
	}
	return r.sprint, r.err
}

func (r *repositoryStub) ReconcileSprintSchedule(context.Context, CoreTeamSprintSettings, AuditActor) (int, error) {
	return 0, r.err
}

func (r *repositoryStub) GetStoryAutomationSettings(context.Context, uuid.UUID, uuid.UUID) (CoreTeamStoryAutomationSettings, error) {
	return r.story, r.err
}

func (r *repositoryStub) UpdateStoryAutomationSettings(context.Context, uuid.UUID, uuid.UUID, CoreUpdateTeamStoryAutomationSettings, AuditActor) (CoreTeamStoryAutomationSettings, error) {
	if r.err == nil {
		r.committed = true
	}
	return r.story, r.err
}

func (r *repositoryStub) GetEstimationSettings(context.Context, uuid.UUID, uuid.UUID) (CoreTeamEstimationSettings, error) {
	return r.estimation, r.err
}

func (r *repositoryStub) UpdateEstimationSettings(context.Context, uuid.UUID, uuid.UUID, CoreUpdateTeamEstimationSettings, AuditActor) (CoreTeamEstimationSettings, error) {
	if r.err == nil {
		r.committed = true
	}
	return r.estimation, r.err
}

func testLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelError, "teamsettings-test")
}
