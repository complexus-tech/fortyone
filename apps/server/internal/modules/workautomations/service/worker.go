package workautomations

import (
	"context"
	"fmt"
	"slices"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	"github.com/google/uuid"
)

const actionFailure = "Task action could not be completed. Review the current task fields and automation owner's access."

type WorkerRepository interface {
	Claim(context.Context) (domain.Claimed, bool, error)
	Authorized(context.Context, domain.Claimed) (bool, error)
	Events(context.Context, domain.Claimed) ([]domain.Event, error)
	ClaimRun(context.Context, domain.Claimed, string) (uuid.UUID, bool, error)
	CompleteRun(context.Context, domain.Claimed, uuid.UUID, string, *uuid.UUID, string) error
	Advance(context.Context, domain.Claimed, domain.Event, string) error
	Release(context.Context, domain.Claimed, *time.Time, string) error
}

// StoryActions is supplied at bootstrap. The adapter binds the current human
// owner to canonical story mutations, preserving permissions and outbox writes.
type StoryActions interface {
	Get(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domain.Story, error)
	Update(context.Context, uuid.UUID, uuid.UUID, domain.Story, domain.Actions, string) error
	Create(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, domain.Draft, string) (uuid.UUID, error)
}

type Worker struct {
	repository WorkerRepository
	stories    StoryActions
}

func NewWorker(repository WorkerRepository, stories StoryActions) *Worker {
	return &Worker{repository: repository, stories: stories}
}

func (w *Worker) DispatchBatch(ctx context.Context) (int, error) {
	seen := make(map[uuid.UUID]bool)
	processed := 0
	for processed < 16 {
		claim, found, err := w.repository.Claim(ctx)
		if err != nil {
			return processed, err
		}
		if !found {
			break
		}
		if seen[claim.ID] {
			if err := w.repository.Release(ctx, claim, claim.NextRunAt, claim.LastError); err != nil {
				return processed, err
			}
			break
		}
		seen[claim.ID] = true
		allowed, err := w.repository.Authorized(ctx, claim)
		if err != nil {
			return processed, err
		}
		if !allowed {
			if err := w.repository.Release(ctx, claim, claim.NextRunAt, "Automation owner no longer has access to this team."); err != nil {
				return processed, err
			}
			processed++
			continue
		}
		if claim.Kind == "rule" {
			err = w.rule(ctx, claim)
		} else {
			err = w.recurrence(ctx, claim)
		}
		if err != nil {
			return processed, fmt.Errorf("process automation: %w", err)
		}
		processed++
	}
	return processed, nil
}

func (w *Worker) rule(ctx context.Context, claim domain.Claimed) error {
	config, err := validateRule(claim.Configuration)
	if err != nil {
		return w.repository.Release(ctx, claim, nil, "Automation configuration is unavailable.")
	}
	events, err := w.repository.Events(ctx, claim)
	if err != nil {
		return err
	}
	lastError := claim.LastError
	for _, event := range events {
		if event.Kind == config.Trigger {
			run, found, err := w.repository.ClaimRun(ctx, claim, "event:"+event.ID.String())
			if err != nil {
				return err
			}
			if found {
				status, errorText := "skipped", ""
				story, err := w.stories.Get(ctx, claim.OwnerID, claim.WorkspaceID, event.StoryID)
				if err != nil {
					status, errorText = "failed", actionFailure
				} else if story.TeamID == claim.TeamID && matches(config.Conditions, story) && changes(config.Actions, story) {
					if err := w.stories.Update(ctx, claim.OwnerID, claim.WorkspaceID, story, config.Actions, "Team rule: "+claim.Name); err != nil {
						status, errorText = "failed", actionFailure
					} else {
						status = "succeeded"
					}
				}
				if err := w.repository.CompleteRun(ctx, claim, run, status, &event.StoryID, errorText); err != nil {
					return err
				}
				lastError = errorText
			}
		}
		if err := w.repository.Advance(ctx, claim, event, lastError); err != nil {
			return err
		}
	}
	return w.repository.Release(ctx, claim, nil, lastError)
}

func (w *Worker) recurrence(ctx context.Context, claim domain.Claimed) error {
	config, err := validateRecurrence(claim.Configuration)
	if err != nil || claim.DueAt == nil {
		return w.repository.Release(ctx, claim, nil, "Recurring task configuration is unavailable.")
	}
	occurrence := "schedule:" + claim.DueAt.UTC().Format(time.RFC3339)
	run, found, err := w.repository.ClaimRun(ctx, claim, occurrence)
	if err != nil {
		return err
	}
	lastError := ""
	if found {
		storyID, createErr := w.stories.Create(ctx, claim.OwnerID, claim.WorkspaceID, claim.TeamID, config.Draft, "team-automation:"+claim.ID.String()+":"+claim.DueAt.UTC().Format(time.RFC3339))
		status := "succeeded"
		var createdID *uuid.UUID
		if createErr != nil {
			status, lastError = "failed", actionFailure
		} else {
			createdID = &storyID
		}
		if err := w.repository.CompleteRun(ctx, claim, run, status, createdID, lastError); err != nil {
			return err
		}
	}
	next, err := NextOccurrence(config.Schedule, time.Now())
	if err != nil {
		return err
	}
	return w.repository.Release(ctx, claim, &next, lastError)
}

func matches(conditions domain.Conditions, story domain.Story) bool {
	if len(conditions.StatusIDs) > 0 && (story.StatusID == nil || !slices.Contains(conditions.StatusIDs, *story.StatusID)) {
		return false
	}
	if len(conditions.Priorities) > 0 && !slices.Contains(conditions.Priorities, story.Priority) {
		return false
	}
	if conditions.Unassigned && story.AssigneeID != nil {
		return false
	}
	if len(conditions.AssigneeIDs) > 0 && (story.AssigneeID == nil || !slices.Contains(conditions.AssigneeIDs, *story.AssigneeID)) {
		return false
	}
	return true
}

func changes(actions domain.Actions, story domain.Story) bool {
	return (actions.StatusID != nil && (story.StatusID == nil || *actions.StatusID != *story.StatusID)) ||
		(actions.Priority != nil && *actions.Priority != story.Priority) ||
		(actions.AssigneeID != nil && (story.AssigneeID == nil || *actions.AssigneeID != *story.AssigneeID)) ||
		(actions.ClearAssignee && story.AssigneeID != nil)
}
