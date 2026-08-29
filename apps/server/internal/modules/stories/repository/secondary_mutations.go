package storiesrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type authorizedSecondaryTarget struct {
	ID                 uuid.UUID
	TeamID             uuid.UUID
	Title              string
	ReporterID         *uuid.UUID
	AssigneeID         *uuid.UUID
	ActorWorkspaceRole string
	DeletedAt          *time.Time
	ArchivedAt         *time.Time
	UpdatedAt          time.Time
}

func (r *repo) ApplySecondaryStoryLifecycle(
	ctx context.Context,
	command storydomain.SecondaryLifecycleCommand,
) (storydomain.SecondaryLifecycleResult, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.SecondaryLifecycleResult{}, err
	}
	if err := command.Validate(); err != nil {
		return storydomain.SecondaryLifecycleResult{}, err
	}
	storyIDs, _ := storydomain.NormalizeSecondaryMutationIDs(command.StoryIDs)
	events := secondaryEventsByStory(command.Events)
	result := storydomain.SecondaryLifecycleResult{StoryIDs: append([]uuid.UUID(nil), storyIDs...)}
	var hardDeleteRoute attachmentObjectStorageRoute
	if command.Action == storydomain.SecondaryMutationHardDelete {
		var err error
		hardDeleteRoute, err = r.interactiveHardDeleteStorageRoute()
		if err != nil {
			return storydomain.SecondaryLifecycleResult{}, err
		}
	}

	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		targets, err := authorizeSecondaryTargets(ctx, queries, command.Scope, storyIDs, command.ChangedAt)
		if err != nil {
			return err
		}
		if command.Action == storydomain.SecondaryMutationSoftDelete || command.Action == storydomain.SecondaryMutationHardDelete {
			if err := authorizeSecondaryDeletion(command.Scope, targets); err != nil {
				return err
			}
		}

		var changed, orphaned []uuid.UUID
		if command.Action == storydomain.SecondaryMutationHardDelete {
			hardDeleteResult, hardDeleteErr := applyInteractiveHardDelete(
				ctx,
				queries,
				hardDeleteRoute,
				command.Scope.WorkspaceID,
				storyIDs,
				command.ChangedAt,
			)
			if hardDeleteErr != nil {
				return hardDeleteErr
			}
			changed = hardDeleteResult.storyIDs
			orphaned = hardDeleteResult.retiredAttachmentIDs
			result.AttachmentObjectDeletionDeferred = true
		} else {
			var lifecycleErr error
			changed, orphaned, lifecycleErr = applySecondaryLifecycleState(
				ctx, queries, command.Action, command.Scope.WorkspaceID, storyIDs, command.ChangedAt,
			)
			if lifecycleErr != nil {
				return lifecycleErr
			}
		}
		changed = orderSecondarySubset(storyIDs, changed)
		for _, storyID := range changed {
			event, exists := events[storyID]
			if !exists {
				return fmt.Errorf("%w: missing durable event for changed story", storydomain.ErrInvalidMutation)
			}
			if err := insertMutationEvent(ctx, queries, event); err != nil {
				return err
			}
		}
		result.ChangedStoryIDs = changed
		result.OrphanedAttachmentIDs = orphaned
		return nil
	})
	if err != nil {
		return storydomain.SecondaryLifecycleResult{}, mapMutationDatabaseError(err)
	}
	return result, nil
}

func (r *repo) ReplaceStoryLabels(
	ctx context.Context,
	command storydomain.ReplaceStoryLabelsCommand,
) (storydomain.ReplacementResult, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.ReplacementResult{}, err
	}
	if err := command.Validate(); err != nil {
		return storydomain.ReplacementResult{}, err
	}
	next, _ := storydomain.NormalizeSecondaryReplacementIDs(command.LabelIDs)
	sortUUIDs(next)
	result := storydomain.ReplacementResult{CurrentIDs: append([]uuid.UUID(nil), next...)}
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		targets, err := authorizeSecondaryTargets(
			ctx, queries, command.Scope, []uuid.UUID{command.StoryID}, command.Event.OccurredAt,
		)
		if err != nil {
			return err
		}
		if targets[0].DeletedAt != nil {
			return storydomain.ErrNotFound
		}
		previous, err := queries.ListStoryLabelsForUpdate(ctx, storyreadsql.ListStoryLabelsForUpdateParams{StoryID: command.StoryID})
		if err != nil {
			return fmt.Errorf("list story labels for update: %w", err)
		}
		sortUUIDs(previous)
		result.PreviousIDs = append([]uuid.UUID(nil), previous...)
		if slices.Equal(previous, next) {
			return nil
		}
		if err := queries.DeleteStoryLabelsForUpdate(ctx, storyreadsql.DeleteStoryLabelsForUpdateParams{StoryID: command.StoryID}); err != nil {
			return fmt.Errorf("clear story labels: %w", err)
		}
		if len(next) > 0 {
			workspaceID, teamID := command.Scope.WorkspaceID, targets[0].TeamID
			inserted, err := queries.InsertAuthorizedStoryLabels(ctx, storyreadsql.InsertAuthorizedStoryLabelsParams{
				StoryID: command.StoryID, LabelIds: next, WorkspaceID: &workspaceID, TeamID: &teamID,
			})
			if err != nil {
				return fmt.Errorf("insert authorized story labels: %w", err)
			}
			if len(inserted) != len(next) {
				return fmt.Errorf("%w: one or more labels are outside the story scope", storydomain.ErrInvalidMutation)
			}
		}
		if err := persistSecondaryReplacementActivity(
			ctx, queries, command.Scope, command.StoryID, command.Activity, previous, next,
		); err != nil {
			return err
		}
		if err := insertMutationEvent(ctx, queries, command.Event); err != nil {
			return err
		}
		result.Changed = true
		return nil
	})
	if err != nil {
		return storydomain.ReplacementResult{}, mapMutationDatabaseError(err)
	}
	return result, nil
}

func (r *repo) ReplaceStoryCollaborators(
	ctx context.Context,
	command storydomain.ReplaceStoryCollaboratorsCommand,
) (storydomain.ReplacementResult, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.ReplacementResult{}, err
	}
	if err := command.Validate(); err != nil {
		return storydomain.ReplacementResult{}, err
	}
	next, _ := storydomain.NormalizeSecondaryReplacementIDs(command.CollaboratorIDs)
	sortUUIDs(next)
	result := storydomain.ReplacementResult{CurrentIDs: append([]uuid.UUID(nil), next...)}
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		targets, err := authorizeSecondaryTargets(
			ctx, queries, command.Scope, []uuid.UUID{command.StoryID}, command.Event.OccurredAt,
		)
		if err != nil {
			return err
		}
		target := targets[0]
		result.AssigneeID = target.AssigneeID
		if target.DeletedAt != nil {
			return storydomain.ErrNotFound
		}
		previous, err := queries.ListStoryCollaboratorsForUpdate(ctx, storyreadsql.ListStoryCollaboratorsForUpdateParams{StoryID: command.StoryID})
		if err != nil {
			return fmt.Errorf("list story collaborators for update: %w", err)
		}
		sortUUIDs(previous)
		result.PreviousIDs = append([]uuid.UUID(nil), previous...)
		if slices.Equal(previous, next) {
			return nil
		}
		if len(next) > 0 {
			valid, err := queries.CountValidStoryCollaborators(ctx, storyreadsql.CountValidStoryCollaboratorsParams{
				CollaboratorIds: next, TeamID: target.TeamID, AssigneeID: target.AssigneeID,
			})
			if err != nil {
				return fmt.Errorf("validate story collaborators: %w", err)
			}
			if valid != int64(len(next)) {
				return fmt.Errorf("%w: collaborators must be active non-system team members and cannot be the assignee", storydomain.ErrInvalidMutation)
			}
		}
		if err := queries.DeleteStoryCollaboratorsForUpdate(ctx, storyreadsql.DeleteStoryCollaboratorsForUpdateParams{StoryID: command.StoryID}); err != nil {
			return fmt.Errorf("clear story collaborators: %w", err)
		}
		if len(next) > 0 {
			inserted, err := queries.InsertStoryCollaboratorsForUpdate(ctx, storyreadsql.InsertStoryCollaboratorsForUpdateParams{
				StoryID: command.StoryID, TeamID: target.TeamID, CollaboratorIds: next,
			})
			if err != nil {
				return fmt.Errorf("insert story collaborators: %w", err)
			}
			if inserted != int64(len(next)) {
				return fmt.Errorf("%w: collaborator replacement was incomplete", storydomain.ErrMutationConflict)
			}
		}
		if err := persistSecondaryReplacementActivity(
			ctx, queries, command.Scope, command.StoryID, command.Activity, previous, next,
		); err != nil {
			return err
		}
		if err := insertMutationEvent(ctx, queries, command.Event); err != nil {
			return err
		}
		result.Changed = true
		return nil
	})
	if err != nil {
		return storydomain.ReplacementResult{}, mapMutationDatabaseError(err)
	}
	return result, nil
}

// SetStoryWatching changes only the acting user's delivery preference. It is
// intentionally not a story.updated event because no shared story state changes.
func (r *repo) SetStoryWatching(
	ctx context.Context,
	scope storydomain.MutationScope,
	storyID, actorID uuid.UUID,
	watching bool,
) error {
	if err := r.mutationConfigured(); err != nil {
		return err
	}
	if err := scope.Validate(); err != nil {
		return err
	}
	if storyID == uuid.Nil || actorID != scope.Actor.PrincipalID || !scope.Actor.IsUserActor() {
		return storydomain.ErrMutationForbidden
	}
	return r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		if _, err := authorizeSecondaryTargets(ctx, queries, scope, []uuid.UUID{storyID}, time.Now().UTC()); err != nil {
			return err
		}
		state, err := queries.GetStoryWatchStateForUpdate(ctx, storyreadsql.GetStoryWatchStateForUpdateParams{
			ActorID: &actorID, StoryID: storyID, WorkspaceID: scope.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return storydomain.ErrMutationForbidden
		}
		if err != nil {
			return fmt.Errorf("authorize story watch preference: %w", err)
		}
		if !scope.Actor.TeamAccess.Allows(state.TeamID) {
			return storydomain.ErrMutationForbidden
		}
		if watching {
			if err := queries.DeleteStoryNotificationMute(ctx, storyreadsql.DeleteStoryNotificationMuteParams{StoryID: storyID, ActorID: actorID}); err != nil {
				return fmt.Errorf("unmute story notifications: %w", err)
			}
			if !state.HasAutomaticAudienceRole {
				if err := queries.InsertStoryWatcher(ctx, storyreadsql.InsertStoryWatcherParams{StoryID: storyID, TeamID: state.TeamID, ActorID: actorID}); err != nil {
					return fmt.Errorf("watch story: %w", err)
				}
			}
			return nil
		}
		if err := queries.DeleteStoryWatcher(ctx, storyreadsql.DeleteStoryWatcherParams{StoryID: storyID, ActorID: actorID}); err != nil {
			return fmt.Errorf("stop watching story: %w", err)
		}
		if state.HasAutomaticAudienceRole {
			if err := queries.InsertStoryNotificationMute(ctx, storyreadsql.InsertStoryNotificationMuteParams{StoryID: storyID, TeamID: state.TeamID, ActorID: actorID}); err != nil {
				return fmt.Errorf("mute story notifications: %w", err)
			}
			return nil
		}
		if err := queries.DeleteStoryNotificationMute(ctx, storyreadsql.DeleteStoryNotificationMuteParams{StoryID: storyID, ActorID: actorID}); err != nil {
			return fmt.Errorf("clear story notification mute: %w", err)
		}
		return nil
	})
}

func (r *repo) ListStoryNotificationAudience(ctx context.Context, storyID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	if err := r.mutationConfigured(); err != nil {
		return nil, err
	}
	rows, err := r.reads.ListStoryNotificationAudience(ctx, storyreadsql.ListStoryNotificationAudienceParams{
		StoryID: storyID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list story notification audience: %w", err)
	}
	result := make([]uuid.UUID, 0, len(rows))
	for _, userID := range rows {
		if userID != nil {
			result = append(result, *userID)
		}
	}
	return result, nil
}

func authorizeSecondaryTargets(
	ctx context.Context,
	queries *storyreadsql.Queries,
	scope storydomain.MutationScope,
	storyIDs []uuid.UUID,
	now time.Time,
) ([]authorizedSecondaryTarget, error) {
	rows, err := queries.AuthorizeSecondaryStoryTargets(ctx, storyreadsql.AuthorizeSecondaryStoryTargetsParams{
		ActorID: scope.Actor.PrincipalID, ActorKind: string(scope.Actor.Kind),
		ActorCredentialID: scope.Actor.CredentialID, Now: now.UTC(), StoryIds: storyIDs,
		WorkspaceID: scope.WorkspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("authorize secondary story targets: %w", err)
	}
	if len(rows) != len(storyIDs) {
		return nil, storydomain.ErrNotFound
	}
	result := make([]authorizedSecondaryTarget, 0, len(rows))
	for _, row := range rows {
		if !row.ActorAuthorized || !scope.Actor.TeamAccess.Allows(row.TeamID) {
			return nil, storydomain.ErrMutationForbidden
		}
		result = append(result, authorizedSecondaryTarget{
			ID: row.ID, TeamID: row.TeamID, Title: row.Title, ReporterID: row.ReporterID, AssigneeID: row.AssigneeID,
			ActorWorkspaceRole: row.ActorWorkspaceRole, DeletedAt: row.DeletedAt, ArchivedAt: row.ArchivedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return result, nil
}

func authorizeSecondaryDeletion(scope storydomain.MutationScope, targets []authorizedSecondaryTarget) error {
	if scope.Actor.Kind == platformauth.PrincipalSystem {
		return nil
	}
	if !scope.Actor.IsUserActor() {
		return storydomain.ErrMutationForbidden
	}
	for _, target := range targets {
		isReporter := target.ReporterID != nil && *target.ReporterID == scope.Actor.PrincipalID
		if !isReporter && target.ActorWorkspaceRole != "admin" {
			return storydomain.ErrMutationForbidden
		}
	}
	return nil
}

func applySecondaryLifecycleState(
	ctx context.Context,
	queries *storyreadsql.Queries,
	action storydomain.SecondaryMutationAction,
	workspaceID uuid.UUID,
	storyIDs []uuid.UUID,
	changedAt time.Time,
) ([]uuid.UUID, []uuid.UUID, error) {
	switch action {
	case storydomain.SecondaryMutationSoftDelete:
		changedAt = changedAt.UTC()
		ids, err := queries.SoftDeleteSecondaryStories(ctx, storyreadsql.SoftDeleteSecondaryStoriesParams{
			ChangedAt: &changedAt, StoryIds: storyIDs, WorkspaceID: workspaceID,
		})
		return ids, nil, err
	case storydomain.SecondaryMutationRestore:
		ids, err := queries.RestoreSecondaryStories(ctx, storyreadsql.RestoreSecondaryStoriesParams{
			ChangedAt: changedAt.UTC(), StoryIds: storyIDs, WorkspaceID: workspaceID,
		})
		return ids, nil, err
	case storydomain.SecondaryMutationArchive:
		changedAt = changedAt.UTC()
		ids, err := queries.ArchiveSecondaryStories(ctx, storyreadsql.ArchiveSecondaryStoriesParams{
			ChangedAt: &changedAt, StoryIds: storyIDs, WorkspaceID: workspaceID,
		})
		return ids, nil, err
	case storydomain.SecondaryMutationUnarchive:
		ids, err := queries.UnarchiveSecondaryStories(ctx, storyreadsql.UnarchiveSecondaryStoriesParams{
			ChangedAt: changedAt.UTC(), StoryIds: storyIDs, WorkspaceID: workspaceID,
		})
		return ids, nil, err
	default:
		return nil, nil, storydomain.ErrInvalidMutation
	}
}

func secondaryEventsByStory(events []storydomain.MutationEvent) map[uuid.UUID]storydomain.MutationEvent {
	result := make(map[uuid.UUID]storydomain.MutationEvent, len(events))
	for _, event := range events {
		result[event.StoryID] = event
	}
	return result
}

func orderSecondarySubset(order, values []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range order {
		if _, exists := set[value]; exists {
			result = append(result, value)
		}
	}
	return result
}

func sortUUIDs(values []uuid.UUID) {
	slices.SortFunc(values, func(left, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
}

func persistSecondaryReplacementActivity(
	ctx context.Context,
	queries *storyreadsql.Queries,
	scope storydomain.MutationScope,
	storyID uuid.UUID,
	activity *storydomain.MutationActivity,
	previous, current []uuid.UUID,
) error {
	if activity == nil {
		return nil
	}
	oldValue, err := json.Marshal(previous)
	if err != nil {
		return fmt.Errorf("encode previous story relationship activity value: %w", err)
	}
	newValue, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("encode current story relationship activity value: %w", err)
	}
	resolved := *activity
	resolved.OldValue = oldValue
	resolved.NewValue = newValue
	if err := resolved.Validate(scope, storyID); err != nil {
		return err
	}
	return upsertMutationActivity(ctx, queries, resolved, true)
}
