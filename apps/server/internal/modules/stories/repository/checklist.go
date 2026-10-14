package storiesrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type checklistQueries interface {
	AuthorizeSecondaryStoryTargets(context.Context, storyreadsql.AuthorizeSecondaryStoryTargetsParams) ([]storyreadsql.AuthorizeSecondaryStoryTargetsRow, error)
	ListStoryChecklist(context.Context, storyreadsql.ListStoryChecklistParams) ([]storyreadsql.ListStoryChecklistRow, error)
	InsertStoryChecklistItem(context.Context, storyreadsql.InsertStoryChecklistItemParams) error
	UpdateStoryChecklistItem(context.Context, storyreadsql.UpdateStoryChecklistItemParams) (int64, error)
	DeleteStoryChecklistItem(context.Context, storyreadsql.DeleteStoryChecklistItemParams) (int64, error)
	ClearStoryChecklist(context.Context, storyreadsql.ClearStoryChecklistParams) error
	ReorderStoryChecklist(context.Context, storyreadsql.ReorderStoryChecklistParams) (int64, error)
	CompleteStoryChecklistConversion(context.Context, storyreadsql.CompleteStoryChecklistConversionParams) error
	FindChecklistConversionStatus(context.Context, storyreadsql.FindChecklistConversionStatusParams) (uuid.UUID, error)
	TouchChecklistStory(context.Context, storyreadsql.TouchChecklistStoryParams) error
	GetStoryMutationSnapshot(context.Context, storyreadsql.GetStoryMutationSnapshotParams) (storyreadsql.GetStoryMutationSnapshotRow, error)
	NextStorySequence(context.Context, storyreadsql.NextStorySequenceParams) (int32, error)
	CreateStoryMutation(context.Context, storyreadsql.CreateStoryMutationParams) (storyreadsql.CreateStoryMutationRow, error)
	InsertStoryAssociation(context.Context, storyreadsql.InsertStoryAssociationParams) (storyreadsql.InsertStoryAssociationRow, error)
	InsertStoryMutationEvent(context.Context, storyreadsql.InsertStoryMutationEventParams) error
	UpsertStoryMutationActivity(context.Context, storyreadsql.UpsertStoryMutationActivityParams) (storyreadsql.UpsertStoryMutationActivityRow, error)
}

func (r *repo) ListChecklist(ctx context.Context, scope storydomain.ReadScope, storyID uuid.UUID) ([]storydomain.ChecklistItem, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	if storyID == uuid.Nil {
		return nil, storydomain.ErrInvalidReadQuery
	}
	var items []storydomain.ChecklistItem
	// Read authorization and target hydration from one consistent snapshot, so a
	// concurrent team move or membership removal cannot expose another team.
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		_, err := queries.GetVisibleStory(ctx, storyreadsql.GetVisibleStoryParams{ActorID: scope.ActorID, StoryID: storyID, WorkspaceID: scope.WorkspaceID, BypassActorMembership: false, UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess, AllowedTeamIds: cloneUUIDs(scope.AllowedTeamIDs)})
		if err != nil {
			return visibleStoryReadError(err)
		}
		items, err = listChecklist(ctx, queries, scope.WorkspaceID, storyID)
		return err
	})
	return items, err
}

func listChecklist(ctx context.Context, queries interface {
	ListStoryChecklist(context.Context, storyreadsql.ListStoryChecklistParams) ([]storyreadsql.ListStoryChecklistRow, error)
}, workspaceID, storyID uuid.UUID) ([]storydomain.ChecklistItem, error) {
	rows, err := queries.ListStoryChecklist(ctx, storyreadsql.ListStoryChecklistParams{StoryID: storyID, WorkspaceID: workspaceID})
	if err != nil {
		return nil, fmt.Errorf("list story checklist: %w", err)
	}
	items := make([]storydomain.ChecklistItem, 0, len(rows))
	for _, row := range rows {
		item := storydomain.ChecklistItem{ID: row.ID, StoryID: row.StoryID, Title: row.Title, Completed: row.Completed, Position: int(row.Position), ConvertedAt: row.ConvertedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
		if row.ConvertedID != nil && row.ConvertedTitle != nil && row.ConvertedTeamID != nil {
			item.ConvertedStory = &storydomain.ChecklistStory{ID: *row.ConvertedID, SequenceID: intValue(row.ConvertedSequenceID), Title: *row.ConvertedTitle, TeamID: *row.ConvertedTeamID, StatusID: row.ConvertedStatusID}
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *repo) checklistTransaction(ctx context.Context, operation func(checklistQueries) error) error {
	if r.runChecklistTransaction != nil {
		return r.runChecklistTransaction(ctx, operation)
	}
	return r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error { return operation(storyreadsql.New(tx)) })
}

func (r *repo) ApplyChecklistMutation(ctx context.Context, command storydomain.ChecklistCommand) (storydomain.ChecklistMutationResult, error) {
	if err := command.Validate(); err != nil {
		return storydomain.ChecklistMutationResult{}, err
	}
	var result storydomain.ChecklistMutationResult
	var sourceTeamID uuid.UUID
	var err error
	for attempt := 0; attempt < storyMutationTransactionAttempts; attempt++ {
		result = storydomain.ChecklistMutationResult{}
		err = r.checklistTransaction(ctx, func(queries checklistQueries) error {
			targets, err := authorizeSecondaryTargets(ctx, queries, command.Scope, []uuid.UUID{command.StoryID}, command.Event.OccurredAt)
			if err != nil {
				return err
			}
			target := targets[0]
			if command.Scope.Actor.IsUserActor() && target.ActorWorkspaceRole != "member" && target.ActorWorkspaceRole != "admin" {
				return storydomain.ErrMutationForbidden
			}
			if target.DeletedAt != nil {
				return storydomain.ErrNotFound
			}
			if target.ArchivedAt != nil {
				return fmt.Errorf("%w: archived stories cannot be edited", storydomain.ErrInvalidMutation)
			}
			sourceTeamID = target.TeamID
			items, err := listChecklist(ctx, queries, command.Scope.WorkspaceID, command.StoryID)
			if err != nil {
				return err
			}
			result.Items = items
			index := slices.IndexFunc(items, func(item storydomain.ChecklistItem) bool { return item.ID == command.ItemID })
			switch command.Action {
			case storydomain.ChecklistAdd:
				if len(items) >= storydomain.MaxChecklistItems {
					return fmt.Errorf("%w: a checklist can contain at most %d items", storydomain.ErrInvalidMutation, storydomain.MaxChecklistItems)
				}
				position := 0
				if len(items) > 0 {
					position = items[len(items)-1].Position + 1
				}
				err = queries.InsertStoryChecklistItem(ctx, storyreadsql.InsertStoryChecklistItemParams{ID: command.ItemID, StoryID: command.StoryID, Title: *command.Title, Position: int32(position), Now: command.Event.OccurredAt})
			case storydomain.ChecklistUpdate:
				if index < 0 {
					return storydomain.ErrNotFound
				}
				item := items[index]
				if item.ConvertedAt != nil && command.Completed != nil && !*command.Completed {
					return fmt.Errorf("%w: converted checklist items remain completed", storydomain.ErrInvalidMutation)
				}
				if (command.Title == nil || *command.Title == item.Title) && (command.Completed == nil || *command.Completed == item.Completed) {
					return nil
				}
				_, err = queries.UpdateStoryChecklistItem(ctx, storyreadsql.UpdateStoryChecklistItemParams{ID: command.ItemID, StoryID: command.StoryID, Title: command.Title, Completed: command.Completed, Now: command.Event.OccurredAt})
			case storydomain.ChecklistDelete:
				if index < 0 {
					return storydomain.ErrNotFound
				}
				_, err = queries.DeleteStoryChecklistItem(ctx, storyreadsql.DeleteStoryChecklistItemParams{ID: command.ItemID, StoryID: command.StoryID})
			case storydomain.ChecklistClear:
				if len(items) == 0 {
					return nil
				}
				err = queries.ClearStoryChecklist(ctx, storyreadsql.ClearStoryChecklistParams{StoryID: command.StoryID})
			case storydomain.ChecklistReorder:
				if err := storydomain.ValidateChecklistOrder(items, command.ItemIDs); err != nil {
					return err
				}
				unchanged := true
				for i, item := range items {
					if item.ID != command.ItemIDs[i] {
						unchanged = false
						break
					}
				}
				if unchanged {
					return nil
				}
				_, err = queries.ReorderStoryChecklist(ctx, storyreadsql.ReorderStoryChecklistParams{StoryID: command.StoryID, ItemIds: command.ItemIDs, Now: command.Event.OccurredAt})
			case storydomain.ChecklistConvert:
				if index < 0 {
					return storydomain.ErrNotFound
				}
				// The owning story row lock is held through commit, making repeated and
				// concurrent conversion requests return the same retained item.
				if items[index].ConvertedAt != nil {
					return nil
				}
				result.CreatedStory, err = convertChecklistItem(ctx, queries, command, items[index])
			}
			if err != nil {
				return fmt.Errorf("apply checklist %s: %w", command.Action, err)
			}
			if err := queries.TouchChecklistStory(ctx, storyreadsql.TouchChecklistStoryParams{StoryID: command.StoryID, WorkspaceID: command.Scope.WorkspaceID, Now: command.Event.OccurredAt}); err != nil {
				return err
			}
			if err := insertMutationEvent(ctx, queries, command.Event); err != nil {
				return err
			}
			result.Items, err = listChecklist(ctx, queries, command.Scope.WorkspaceID, command.StoryID)
			if err != nil {
				return err
			}
			if err := insertChecklistActivities(ctx, queries, command, items, result.Items); err != nil {
				return err
			}
			result.Changed = true
			return nil
		})
		if err == nil {
			return result, nil
		}
		if !isStorySequenceUniqueConflict(err) && !platformdatabase.IsRetryableTransactionError(err) {
			break
		}
		if isStorySequenceUniqueConflict(err) && sourceTeamID != uuid.Nil {
			if syncErr := r.reads.SynchronizeStorySequence(ctx, storyreadsql.SynchronizeStorySequenceParams{WorkspaceID: command.Scope.WorkspaceID, TeamID: sourceTeamID}); syncErr != nil {
				return storydomain.ChecklistMutationResult{}, syncErr
			}
		}
	}
	return storydomain.ChecklistMutationResult{}, mapMutationDatabaseError(err)
}

func convertChecklistItem(ctx context.Context, queries checklistQueries, command storydomain.ChecklistCommand, item storydomain.ChecklistItem) (*storydomain.ChecklistStory, error) {
	params := mutationQueryScope(command.Scope, command.Event.OccurredAt)
	params.StoryID = command.StoryID
	snapshot, err := queries.GetStoryMutationSnapshot(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storydomain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	source := mutationSnapshotToStory(snapshot)
	statusID, err := queries.FindChecklistConversionStatus(ctx, storyreadsql.FindChecklistConversionStatusParams{TeamID: source.Team})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: the team needs an unstarted or backlog status", storydomain.ErrInvalidMutation)
	}
	if err != nil {
		return nil, err
	}
	sequence, err := queries.NextStorySequence(ctx, storyreadsql.NextStorySequenceParams{WorkspaceID: command.Scope.WorkspaceID, TeamID: source.Team})
	if err != nil {
		return nil, err
	}
	// New work inherits planning context, never completion, assignment or parentage.
	story := storydomain.Story{
		ID: uuid.New(), Title: item.Title,
		Team: source.Team, Workspace: source.Workspace,
		Objective: source.Objective, Sprint: source.Sprint, KeyResult: source.KeyResult,
		Status: &statusID, Priority: "No Priority",
		AutoSchedulingStatus: storydomain.AutoSchedulingStatusOff,
		Reporter:             command.Scope.ActivityUser,
		CreatedAt:            command.Event.OccurredAt, UpdatedAt: command.Event.OccurredAt,
	}
	createParams, err := storyCreateParams(story, sequence)
	if err != nil {
		return nil, err
	}
	created, err := queries.CreateStoryMutation(ctx, createParams)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storydomain.ErrInvalidMutation
	}
	if err != nil {
		return nil, err
	}
	if _, err := queries.InsertStoryAssociation(ctx, storyreadsql.InsertStoryAssociationParams{AssociationID: uuid.New(), FromStoryID: command.StoryID, ToStoryID: created.ID, AssociationType: "related", WorkspaceID: command.Scope.WorkspaceID}); err != nil {
		return nil, err
	}
	if err := queries.CompleteStoryChecklistConversion(ctx, storyreadsql.CompleteStoryChecklistConversionParams{ID: item.ID, StoryID: command.StoryID, ConvertedStoryID: &created.ID, Now: &command.Event.OccurredAt}); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(map[string]any{
		"storyId": created.ID, "workspaceId": story.Workspace, "teamId": story.Team,
		"title": story.Title, "assigneeId": nil, "reporterId": story.Reporter,
	})
	if err != nil {
		return nil, err
	}
	event := storydomain.MutationEvent{
		ID: uuid.New(), WorkspaceID: story.Workspace, StoryID: created.ID,
		Type: storydomain.MutationEventStoryCreated, Actor: command.Scope.Actor,
		Payload: payload, OccurredAt: command.Event.OccurredAt,
	}
	if err := insertMutationEvent(ctx, queries, event); err != nil {
		return nil, err
	}
	if command.Scope.ActivityUser != nil {
		titleJSON, err := json.Marshal(story.Title)
		if err != nil {
			return nil, err
		}
		activity := storydomain.MutationActivity{
			ID: uuid.New(), StoryID: created.ID, UserID: *command.Scope.ActivityUser,
			Type: "create", Field: "story", CurrentValue: story.Title,
			NewValue: titleJSON, OldValue: json.RawMessage("null"),
			WorkspaceID: story.Workspace, CreatedAt: command.Event.OccurredAt,
		}
		if err := upsertMutationActivity(ctx, queries, activity, false); err != nil {
			return nil, err
		}
	}
	return &storydomain.ChecklistStory{ID: created.ID, SequenceID: intValue(created.SequenceID), Title: created.Title, TeamID: created.TeamID, StatusID: created.StatusID}, nil
}
