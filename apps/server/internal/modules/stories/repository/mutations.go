package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const storyMutationTransactionAttempts = 3

func (r *repo) mutationConfigured() error {
	if r == nil || r.pool == nil || r.reads == nil {
		return errors.New("story mutation repository is not configured")
	}
	return nil
}

func mutationQueryScope(scope storydomain.MutationScope, now time.Time) storyreadsql.GetStoryMutationSnapshotParams {
	return storyreadsql.GetStoryMutationSnapshotParams{
		WorkspaceID: scope.WorkspaceID, ActorKind: string(scope.Actor.Kind),
		ActorID: scope.Actor.PrincipalID, ActorCredentialID: scope.Actor.CredentialID,
		Now: now.UTC(),
	}
}

func (r *repo) PrepareStoryMutation(
	ctx context.Context,
	scope storydomain.MutationScope,
	teamID uuid.UUID,
	keyResultID *uuid.UUID,
) (storydomain.MutationPreconditions, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.MutationPreconditions{}, err
	}
	if err := scope.Validate(); err != nil {
		return storydomain.MutationPreconditions{}, err
	}
	if teamID == uuid.Nil || !scope.Actor.TeamAccess.Allows(teamID) {
		return storydomain.MutationPreconditions{}, storydomain.ErrMutationForbidden
	}
	authorized, err := r.reads.AuthorizeStoryCreate(ctx, storyreadsql.AuthorizeStoryCreateParams{
		TeamID: teamID, WorkspaceID: scope.WorkspaceID,
		ActorKind: string(scope.Actor.Kind), ActorID: scope.Actor.PrincipalID,
		ActorCredentialID: scope.Actor.CredentialID, Now: time.Now().UTC(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return storydomain.MutationPreconditions{}, storydomain.ErrMutationForbidden
	}
	if err != nil {
		return storydomain.MutationPreconditions{}, fmt.Errorf("authorize story mutation preconditions: %w", err)
	}
	result := storydomain.MutationPreconditions{EstimateScheme: authorized.EstimateScheme}
	if keyResultID == nil {
		return result, nil
	}
	if *keyResultID == uuid.Nil {
		return storydomain.MutationPreconditions{}, storydomain.ErrInvalidMutation
	}
	keyResult, err := r.reads.GetStoryMutationKeyResult(ctx, storyreadsql.GetStoryMutationKeyResultParams{
		KeyResultID: *keyResultID, WorkspaceID: scope.WorkspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return storydomain.MutationPreconditions{}, storydomain.ErrNotFound
	}
	if err != nil {
		return storydomain.MutationPreconditions{}, fmt.Errorf("resolve story mutation key result: %w", err)
	}
	result.KeyResult = &storydomain.MutationKeyResultReference{
		ObjectiveID: keyResult.ObjectiveID,
		Name:        keyResult.Name,
	}
	return result, nil
}

func (r *repo) GetStoryForMutation(
	ctx context.Context,
	scope storydomain.MutationScope,
	storyID uuid.UUID,
) (storydomain.Story, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.Story{}, err
	}
	if err := scope.Validate(); err != nil {
		return storydomain.Story{}, err
	}
	params := mutationQueryScope(scope, time.Now())
	params.StoryID = storyID
	row, err := r.reads.GetStoryMutationSnapshot(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return storydomain.Story{}, storydomain.ErrNotFound
	}
	if err != nil {
		return storydomain.Story{}, fmt.Errorf("get story mutation snapshot: %w", err)
	}
	story := mutationSnapshotToStory(row)
	if !scope.Actor.TeamAccess.Allows(story.Team) {
		return storydomain.Story{}, storydomain.ErrNotFound
	}
	return story, nil
}

func (r *repo) CreateStoryMutation(
	ctx context.Context,
	command storydomain.CreateStoryCommand,
) (storydomain.CreateStoryResult, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.CreateStoryResult{}, err
	}
	if err := command.Validate(); err != nil {
		return storydomain.CreateStoryResult{}, err
	}
	if !command.Scope.Actor.TeamAccess.Allows(command.Story.Team) {
		return storydomain.CreateStoryResult{}, storydomain.ErrMutationForbidden
	}
	labels, err := normalizeMutationLabelIDs(command.LabelIDs)
	if err != nil {
		return storydomain.CreateStoryResult{}, err
	}

	var created storydomain.Story
	for attempt := 1; attempt <= storyMutationTransactionAttempts; attempt++ {
		err = r.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
			queries := storyreadsql.New(tx)
			now := command.Event.OccurredAt.UTC()
			if _, authorizeErr := queries.AuthorizeStoryCreate(ctx, storyreadsql.AuthorizeStoryCreateParams{
				TeamID: command.Story.Team, WorkspaceID: command.Scope.WorkspaceID,
				ActorKind: string(command.Scope.Actor.Kind), ActorID: command.Scope.Actor.PrincipalID,
				ActorCredentialID: command.Scope.Actor.CredentialID, Now: now,
			}); authorizeErr != nil {
				if errors.Is(authorizeErr, pgx.ErrNoRows) {
					return storydomain.ErrMutationForbidden
				}
				return fmt.Errorf("authorize story creation: %w", authorizeErr)
			}
			sequence, sequenceErr := queries.NextStorySequence(ctx, storyreadsql.NextStorySequenceParams{
				WorkspaceID: command.Scope.WorkspaceID, TeamID: command.Story.Team,
			})
			if sequenceErr != nil {
				return fmt.Errorf("advance story sequence: %w", sequenceErr)
			}
			params, mapErr := storyCreateParams(command.Story, sequence)
			if mapErr != nil {
				return mapErr
			}
			row, createErr := queries.CreateStoryMutation(ctx, params)
			if createErr != nil {
				if errors.Is(createErr, pgx.ErrNoRows) {
					return fmt.Errorf("%w: a referenced resource is outside the story scope", storydomain.ErrInvalidMutation)
				}
				return fmt.Errorf("insert story mutation: %w", createErr)
			}
			if len(labels) > 0 {
				workspaceID, teamID := command.Scope.WorkspaceID, command.Story.Team
				inserted, labelErr := queries.InsertAuthorizedStoryLabels(ctx, storyreadsql.InsertAuthorizedStoryLabelsParams{
					StoryID: row.ID, LabelIds: labels, WorkspaceID: &workspaceID, TeamID: &teamID,
				})
				if labelErr != nil {
					return fmt.Errorf("insert story labels: %w", labelErr)
				}
				if len(inserted) != len(labels) {
					return fmt.Errorf("%w: one or more labels are outside the story scope", storydomain.ErrInvalidMutation)
				}
			}
			if command.Activity != nil {
				if activityErr := upsertMutationActivity(ctx, queries, *command.Activity, false); activityErr != nil {
					return activityErr
				}
			}
			if eventErr := insertMutationEvent(ctx, queries, command.Event); eventErr != nil {
				return eventErr
			}
			created = createdMutationToStory(row, command.Story.EstimateScheme, labels)
			return nil
		})
		if err == nil {
			return storydomain.CreateStoryResult{Story: created, Created: true}, nil
		}
		if isCreationKeyConflict(err) && command.Story.CreationKey != nil {
			return r.loadIdempotentStory(ctx, command)
		}
		if !isStorySequenceUniqueConflict(err) && !platformdatabase.IsRetryableTransactionError(err) {
			return storydomain.CreateStoryResult{}, mapMutationDatabaseError(err)
		}
		if isStorySequenceUniqueConflict(err) {
			if syncErr := r.reads.SynchronizeStorySequence(ctx, storyreadsql.SynchronizeStorySequenceParams{
				WorkspaceID: command.Scope.WorkspaceID, TeamID: command.Story.Team,
			}); syncErr != nil {
				return storydomain.CreateStoryResult{}, fmt.Errorf("synchronize story sequence: %w", syncErr)
			}
		}
	}
	return storydomain.CreateStoryResult{}, fmt.Errorf("create story mutation after %d attempts: %w", storyMutationTransactionAttempts, err)
}

func (r *repo) loadIdempotentStory(
	ctx context.Context,
	command storydomain.CreateStoryCommand,
) (storydomain.CreateStoryResult, error) {
	key := strings.TrimSpace(*command.Story.CreationKey)
	if command.Scope.Actor.Kind == platformauth.PrincipalOAuthApplication {
		row, err := r.reads.GetOAuthApplicationStoryCreationReplay(ctx, storyreadsql.GetOAuthApplicationStoryCreationReplayParams{
			WorkspaceID: command.Scope.WorkspaceID, ExternalCreationKey: &key,
			ActorKind: string(command.Scope.Actor.Kind), ActorID: command.Scope.Actor.PrincipalID,
			ActorCredentialID: command.Scope.Actor.CredentialID, Now: time.Now().UTC(),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return storydomain.CreateStoryResult{}, storydomain.ErrMutationForbidden
		}
		if err != nil {
			return storydomain.CreateStoryResult{}, fmt.Errorf("get OAuth application story creation replay: %w", err)
		}
		story := oauthApplicationCreationReplayToStory(row)
		story.CreatedNow = false
		return storydomain.CreateStoryResult{Story: story, Created: false}, nil
	}
	id, err := r.reads.GetStoryIDByCreationKey(ctx, storyreadsql.GetStoryIDByCreationKeyParams{
		WorkspaceID: command.Scope.WorkspaceID, ExternalCreationKey: &key,
	})
	if err != nil {
		return storydomain.CreateStoryResult{}, fmt.Errorf("get story by creation key: %w", err)
	}
	story, err := r.GetStoryForMutation(ctx, command.Scope, id)
	if err != nil {
		return storydomain.CreateStoryResult{}, err
	}
	story.CreatedNow = false
	return storydomain.CreateStoryResult{Story: story, Created: false}, nil
}

func (r *repo) ApplyStoryMutation(
	ctx context.Context,
	command storydomain.UpdateStoryCommand,
) (storydomain.UpdateStoryResult, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.UpdateStoryResult{}, err
	}
	if err := command.Validate(); err != nil {
		return storydomain.UpdateStoryResult{}, err
	}
	var result storydomain.UpdateStoryResult
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		now := command.Event.OccurredAt.UTC()
		params, err := storyPatchParams(command, now)
		if err != nil {
			return err
		}
		row, err := queries.ApplyStoryPatch(ctx, params)
		if err != nil {
			return fmt.Errorf("apply story patch: %w", err)
		}
		switch {
		case !row.StoryExists:
			return storydomain.ErrNotFound
		case !row.ActorAuthorized:
			return storydomain.ErrMutationForbidden
		case !row.ReferencesValid:
			return fmt.Errorf("%w: a referenced resource is outside the story scope", storydomain.ErrInvalidMutation)
		case !row.StoryUpdated:
			return storydomain.ErrMutationConflict
		}
		if command.ReconcileMedia {
			orphaned, reconcileErr := reconcileMutationMedia(ctx, queries, command)
			if reconcileErr != nil {
				return reconcileErr
			}
			result.OrphanedAttachmentIDs = orphaned
		}
		for _, activity := range command.Activities {
			if err := upsertMutationActivity(ctx, queries, activity, true); err != nil {
				return err
			}
		}
		if err := insertMutationEvent(ctx, queries, command.Event); err != nil {
			return err
		}
		result.UpdatedAt = now
		return nil
	})
	if err != nil {
		return storydomain.UpdateStoryResult{}, mapMutationDatabaseError(err)
	}
	return result, nil
}

func (r *repo) DeleteStoryMutation(
	ctx context.Context,
	command storydomain.DeleteStoryCommand,
) (storydomain.DeleteStoryResult, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.DeleteStoryResult{}, err
	}
	if err := command.Validate(); err != nil {
		return storydomain.DeleteStoryResult{}, err
	}
	var result storydomain.DeleteStoryResult
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		now := command.Event.OccurredAt.UTC()
		row, err := queries.DeleteStoryMutation(ctx, storyreadsql.DeleteStoryMutationParams{
			StoryID: command.StoryID, WorkspaceID: command.Scope.WorkspaceID,
			ActorKind: string(command.Scope.Actor.Kind), ActorID: command.Scope.Actor.PrincipalID,
			ActorCredentialID: command.Scope.Actor.CredentialID, Now: now,
			DeletedAt: &now, ExpectedUpdatedAt: command.ExpectedUpdatedAt.UTC(),
		})
		if err != nil {
			return fmt.Errorf("delete story mutation: %w", err)
		}
		switch {
		case !row.StoryExists:
			return storydomain.ErrNotFound
		case !row.ActorAuthorized:
			return storydomain.ErrMutationForbidden
		case !row.DeletionPermitted:
			return storydomain.ErrMutationForbidden
		case !row.StoryDeleted:
			return storydomain.ErrMutationConflict
		}
		if command.Activity != nil {
			if err := upsertMutationActivity(ctx, queries, *command.Activity, false); err != nil {
				return err
			}
		}
		if err := insertMutationEvent(ctx, queries, command.Event); err != nil {
			return err
		}
		result = storydomain.DeleteStoryResult{Deleted: true, DeletedAt: row.DeletedAt}
		return nil
	})
	if err != nil {
		return storydomain.DeleteStoryResult{}, mapMutationDatabaseError(err)
	}
	return result, nil
}
