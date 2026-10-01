package workautomationsrepository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/workautomations/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ queries sql.Querier }

func New(pool *pgxpool.Pool) *Repository { return &Repository{queries: sql.New(pool)} }

func (r *Repository) List(ctx context.Context, actor, workspace, team uuid.UUID) ([]domain.Automation, error) {
	rows, err := r.queries.ListAutomationsForActor(ctx, sql.ListAutomationsForActorParams{ActorID: actor, WorkspaceID: workspace, TeamID: team})
	if err != nil {
		return nil, err
	}
	items := make([]domain.Automation, len(rows))
	for i, row := range rows {
		items[i] = mapRow(row)
	}
	return items, nil
}
func (r *Repository) Get(ctx context.Context, actor, workspace, id uuid.UUID) (domain.Automation, error) {
	row, err := r.queries.GetAutomationForActor(ctx, sql.GetAutomationForActorParams{ActorID: actor, WorkspaceID: workspace, ID: id})
	if err != nil {
		return domain.Automation{}, mapError(err)
	}
	return mapRow(sql.ListAutomationsForActorRow(row)), nil
}
func (r *Repository) Create(ctx context.Context, actor, workspace uuid.UUID, input domain.Input, next *time.Time) (domain.Automation, error) {
	row, err := r.queries.CreateAutomationForActor(ctx, sql.CreateAutomationForActorParams{ActorID: actor, WorkspaceID: workspace, TeamID: input.TeamID, Kind: input.Kind, Name: input.Name, Configuration: input.Configuration, NextRunAt: next})
	if err != nil {
		return domain.Automation{}, mapError(err)
	}
	return mapRow(sql.ListAutomationsForActorRow(row)), nil
}
func (r *Repository) Pause(ctx context.Context, actor, workspace, id uuid.UUID, paused bool, next *time.Time) (domain.Automation, error) {
	row, err := r.queries.PauseAutomationForActor(ctx, sql.PauseAutomationForActorParams{ActorID: actor, WorkspaceID: workspace, ID: id, Paused: paused, NextRunAt: next})
	if err != nil {
		return domain.Automation{}, mapError(err)
	}
	return mapRow(sql.ListAutomationsForActorRow(row)), nil
}
func (r *Repository) Archive(ctx context.Context, actor, workspace, id uuid.UUID) error {
	rows, err := r.queries.ArchiveAutomationForActor(ctx, sql.ArchiveAutomationForActorParams{ActorID: actor, WorkspaceID: workspace, ID: id})
	return changed(rows, err)
}
func (r *Repository) Runs(ctx context.Context, actor, workspace, id uuid.UUID) ([]domain.Run, error) {
	rows, err := r.queries.ListAutomationRunsForActor(ctx, sql.ListAutomationRunsForActorParams{ActorID: actor, WorkspaceID: workspace, ID: id})
	if err != nil {
		return nil, err
	}
	items := make([]domain.Run, len(rows))
	for i, row := range rows {
		items[i] = domain.Run{ID: row.ID, AutomationID: row.AutomationID, Occurrence: row.Occurrence, Status: row.Status, StoryID: row.StoryID, Error: row.Error, StartedAt: row.StartedAt, FinishedAt: row.FinishedAt}
	}
	return items, nil
}
func (r *Repository) Claim(ctx context.Context) (domain.Claimed, bool, error) {
	token := uuid.New()
	row, err := r.queries.ClaimAutomation(ctx, sql.ClaimAutomationParams{LeaseToken: &token})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Claimed{}, false, nil
	}
	if err != nil {
		return domain.Claimed{}, false, err
	}
	_, err = r.queries.FinishExhaustedAutomationRuns(ctx, sql.FinishExhaustedAutomationRunsParams{ID: row.ID, LeaseToken: &token})
	if err != nil {
		return domain.Claimed{}, false, err
	}
	return domain.Claimed{Automation: domain.Automation{ID: row.ID, TeamID: row.TeamID, OwnerID: row.OwnerID, Kind: row.Kind, Name: row.Name, Configuration: bytes.Clone(row.Configuration), Paused: row.Paused, NextRunAt: row.NextRunAt, LastRunAt: row.LastRunAt, LastError: row.LastError, CreatedAt: row.CreatedAt}, WorkspaceID: row.WorkspaceID, LeaseToken: token, DueAt: row.NextRunAt, EventAt: row.EventAt, EventID: row.EventID}, true, nil
}
func (r *Repository) Authorized(ctx context.Context, claim domain.Claimed) (bool, error) {
	allowed, err := r.queries.CurrentAutomationOwnerCanWrite(ctx, sql.CurrentAutomationOwnerCanWriteParams{ID: claim.ID, LeaseToken: &claim.LeaseToken})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return allowed, err
}
func (r *Repository) Events(ctx context.Context, claim domain.Claimed) ([]domain.Event, error) {
	rows, err := r.queries.ListAutomationEvents(ctx, sql.ListAutomationEventsParams{ID: claim.ID, LeaseToken: &claim.LeaseToken})
	if err != nil {
		return nil, err
	}
	items := make([]domain.Event, len(rows))
	for i, row := range rows {
		items[i] = domain.Event{ID: row.EventID, StoryID: row.SubjectID, Kind: row.EventType, CreatedAt: row.CreatedAt}
	}
	return items, nil
}
func (r *Repository) ClaimRun(ctx context.Context, claim domain.Claimed, occurrence string) (uuid.UUID, bool, error) {
	id, err := r.queries.ClaimAutomationRun(ctx, sql.ClaimAutomationRunParams{ID: claim.ID, LeaseToken: &claim.LeaseToken, Occurrence: occurrence})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return id, err == nil, err
}
func (r *Repository) CompleteRun(ctx context.Context, claim domain.Claimed, id uuid.UUID, status string, story *uuid.UUID, errorText string) error {
	rows, err := r.queries.CompleteAutomationRun(ctx, sql.CompleteAutomationRunParams{ID: id, LeaseToken: claim.LeaseToken, Status: status, StoryID: story, Error: errorText})
	return changed(rows, err)
}
func (r *Repository) Advance(ctx context.Context, claim domain.Claimed, _ domain.Event, errorText string) error {
	rows, err := r.queries.AdvanceAutomationEvent(ctx, sql.AdvanceAutomationEventParams{ID: claim.ID, LeaseToken: &claim.LeaseToken, LastError: errorText})
	return changed(rows, err)
}
func (r *Repository) Release(ctx context.Context, claim domain.Claimed, next *time.Time, errorText string) error {
	rows, err := r.queries.ReleaseAutomation(ctx, sql.ReleaseAutomationParams{ID: claim.ID, LeaseToken: &claim.LeaseToken, NextRunAt: next, LastError: errorText})
	return changed(rows, err)
}
func mapRow(row sql.ListAutomationsForActorRow) domain.Automation {
	return domain.Automation{ID: row.ID, TeamID: row.TeamID, OwnerID: row.OwnerID, Kind: row.Kind, Name: row.Name, Configuration: bytes.Clone(row.Configuration), Paused: row.Paused, NextRunAt: row.NextRunAt, LastRunAt: row.LastRunAt, LastError: row.LastError, CreatedAt: row.CreatedAt, CanEdit: row.CanEdit}
}
func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("automation storage: %w", err)
}
func changed(rows int64, err error) error {
	if err != nil {
		return mapError(err)
	}
	if rows != 1 {
		return domain.ErrNotFound
	}
	return nil
}
