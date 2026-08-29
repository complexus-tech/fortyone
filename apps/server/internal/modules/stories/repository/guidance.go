package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

const maximumOverdueStoryGuidanceBatchSize = 100

var (
	errInvalidStoryGuidanceRecipient = errors.New("story guidance query returned an invalid recipient")
	errInvalidStoryGuidanceItem      = errors.New("story guidance query returned an invalid item")
)

// ListOverdueStoryGuidanceRecipients returns a stable, bounded page of
// currently eligible story-guidance recipients after cursor.
func (r *repo) ListOverdueStoryGuidanceRecipients(
	ctx context.Context,
	asOf time.Time,
	cursor *storydomain.OverdueGuidanceCursor,
	limit int,
) ([]storydomain.OverdueGuidanceRecipient, error) {
	if r == nil || r.reads == nil {
		return nil, errReadRepositoryNotConfigured
	}
	if limit < 1 || limit > maximumOverdueStoryGuidanceBatchSize {
		return nil, fmt.Errorf("%w: story guidance limit must be between 1 and %d", storydomain.ErrInvalidReadQuery, maximumOverdueStoryGuidanceBatchSize)
	}
	asOf, err := storyGuidanceUTCDate(asOf)
	if err != nil {
		return nil, err
	}
	resultLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, fmt.Errorf("%w: story guidance limit: %v", storydomain.ErrInvalidReadQuery, err)
	}

	params := storyreadsql.ListOverdueStoryGuidanceRecipientsParams{AsOf: asOf, ResultLimit: resultLimit}
	if cursor != nil {
		if cursor.AssigneeID == uuid.Nil || cursor.WorkspaceID == uuid.Nil {
			return nil, fmt.Errorf("%w: story guidance cursor requires assignee and workspace", storydomain.ErrInvalidReadQuery)
		}
		params.HasCursor = true
		params.AfterAssigneeID = &cursor.AssigneeID
		params.AfterWorkspaceID = cursor.WorkspaceID
	}

	rows, err := r.reads.ListOverdueStoryGuidanceRecipients(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list overdue story guidance recipients: %w", err)
	}
	recipients := make([]storydomain.OverdueGuidanceRecipient, 0, len(rows))
	for _, row := range rows {
		if row.AssigneeID == nil || *row.AssigneeID == uuid.Nil || row.WorkspaceID == uuid.Nil {
			return nil, errInvalidStoryGuidanceRecipient
		}
		recipients = append(recipients, storydomain.OverdueGuidanceRecipient{
			AssigneeID:    *row.AssigneeID,
			AssigneeEmail: row.AssigneeEmail,
			AssigneeName:  row.AssigneeName,
			WorkspaceID:   row.WorkspaceID,
			WorkspaceName: row.WorkspaceName,
			WorkspaceSlug: row.WorkspaceSlug,
			EmailEnabled:  row.EmailEnabled,
		})
	}
	return recipients, nil
}

// ListOverdueStoryGuidanceItems loads story deadline signals for one eligible
// assignee in one workspace.
func (r *repo) ListOverdueStoryGuidanceItems(
	ctx context.Context,
	asOf time.Time,
	assigneeID uuid.UUID,
	workspaceID uuid.UUID,
) ([]storydomain.OverdueGuidanceStory, error) {
	if r == nil || r.reads == nil {
		return nil, errReadRepositoryNotConfigured
	}
	if assigneeID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: story guidance assignee and workspace are required", storydomain.ErrInvalidReadQuery)
	}
	asOf, err := storyGuidanceUTCDate(asOf)
	if err != nil {
		return nil, err
	}

	rows, err := r.reads.ListOverdueStoryGuidanceItems(ctx, storyreadsql.ListOverdueStoryGuidanceItemsParams{
		AsOf:        asOf,
		AssigneeID:  &assigneeID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list overdue story guidance items: %w", err)
	}
	stories := make([]storydomain.OverdueGuidanceStory, 0, len(rows))
	for _, row := range rows {
		if row.SequenceID == nil || *row.SequenceID < 1 || row.EndDate == nil || row.AssigneeID == nil ||
			*row.AssigneeID == uuid.Nil || row.WorkspaceID == uuid.Nil || row.TeamID == uuid.Nil || row.StatusCategory == nil {
			return nil, errInvalidStoryGuidanceItem
		}
		stories = append(stories, storydomain.OverdueGuidanceStory{
			ID:             row.ID,
			Title:          row.Title,
			EndDate:        row.EndDate.UTC(),
			AssigneeID:     *row.AssigneeID,
			AssigneeEmail:  row.AssigneeEmail,
			AssigneeName:   row.AssigneeName,
			WorkspaceID:    row.WorkspaceID,
			WorkspaceName:  row.WorkspaceName,
			WorkspaceSlug:  row.WorkspaceSlug,
			TeamID:         row.TeamID,
			TeamName:       row.TeamName,
			TeamCode:       row.TeamCode,
			SequenceID:     int(*row.SequenceID),
			StatusName:     row.StatusName,
			StatusCategory: *row.StatusCategory,
			DeadlineStatus: row.DeadlineStatus,
			DaysDifference: int(row.DaysDifference),
			EmailEnabled:   true,
		})
	}
	return stories, nil
}

func storyGuidanceUTCDate(value time.Time) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, fmt.Errorf("%w: story guidance as-of time is required", storydomain.ErrInvalidReadQuery)
	}
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC), nil
}
