package objectivesrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

const maximumOverdueObjectiveGuidanceBatchSize = 100

var (
	errObjectiveRepositoryNotConfigured  = errors.New("objectives repository is not configured")
	errInvalidObjectiveGuidanceRecipient = errors.New("objective guidance query returned an invalid recipient")
	errInvalidObjectiveGuidanceItem      = errors.New("objective guidance query returned an invalid item")
)

// ListOverdueObjectiveGuidanceRecipients returns a stable, bounded page of
// currently eligible objective-guidance recipients after cursor.
func (repository *Repository) ListOverdueObjectiveGuidanceRecipients(
	ctx context.Context,
	asOf time.Time,
	cursor *objectivesdomain.OverdueGuidanceCursor,
	limit int,
) ([]objectivesdomain.OverdueGuidanceRecipient, error) {
	if repository == nil || repository.queries == nil {
		return nil, errObjectiveRepositoryNotConfigured
	}
	if limit < 1 || limit > maximumOverdueObjectiveGuidanceBatchSize {
		return nil, fmt.Errorf("%w: objective guidance limit must be between 1 and %d", objectivesdomain.ErrInvalid, maximumOverdueObjectiveGuidanceBatchSize)
	}
	asOf, err := objectiveGuidanceUTCDate(asOf)
	if err != nil {
		return nil, err
	}
	resultLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, fmt.Errorf("%w: objective guidance limit: %v", objectivesdomain.ErrInvalid, err)
	}

	params := objectivessql.ListOverdueObjectiveGuidanceRecipientsParams{AsOf: asOf, ResultLimit: resultLimit}
	if cursor != nil {
		if cursor.LeadUserID == uuid.Nil || cursor.WorkspaceID == uuid.Nil {
			return nil, fmt.Errorf("%w: objective guidance cursor requires lead and workspace", objectivesdomain.ErrInvalid)
		}
		params.HasCursor = true
		params.AfterLeadUserID = uuidPointer(cursor.LeadUserID)
		params.AfterWorkspaceID = uuidPointer(cursor.WorkspaceID)
	}

	rows, err := repository.queries.ListOverdueObjectiveGuidanceRecipients(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list overdue objective guidance recipients: %w", mapDatabaseError(err))
	}
	recipients := make([]objectivesdomain.OverdueGuidanceRecipient, 0, len(rows))
	for _, row := range rows {
		if row.LeadUserID == nil || *row.LeadUserID == uuid.Nil || row.WorkspaceID == uuid.Nil {
			return nil, errInvalidObjectiveGuidanceRecipient
		}
		recipients = append(recipients, objectivesdomain.OverdueGuidanceRecipient{
			LeadUserID:    *row.LeadUserID,
			LeadEmail:     row.LeadEmail,
			LeadName:      row.LeadName,
			WorkspaceID:   row.WorkspaceID,
			WorkspaceName: row.WorkspaceName,
			WorkspaceSlug: row.WorkspaceSlug,
			EmailEnabled:  row.EmailEnabled,
		})
	}
	return recipients, nil
}

// ListOverdueObjectiveGuidanceItems loads the objective and key-result signals
// for one eligible lead in one workspace.
func (repository *Repository) ListOverdueObjectiveGuidanceItems(
	ctx context.Context,
	asOf time.Time,
	leadUserID uuid.UUID,
	workspaceID uuid.UUID,
) ([]objectivesdomain.OverdueGuidanceObjective, error) {
	if repository == nil || repository.queries == nil {
		return nil, errObjectiveRepositoryNotConfigured
	}
	if leadUserID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: objective guidance lead and workspace are required", objectivesdomain.ErrInvalid)
	}
	asOf, err := objectiveGuidanceUTCDate(asOf)
	if err != nil {
		return nil, err
	}

	rows, err := repository.queries.ListOverdueObjectiveGuidanceItems(ctx, objectivessql.ListOverdueObjectiveGuidanceItemsParams{
		AsOf:        asOf,
		LeadUserID:  uuidPointer(leadUserID),
		WorkspaceID: uuidPointer(workspaceID),
	})
	if err != nil {
		return nil, fmt.Errorf("list overdue objective guidance items: %w", mapDatabaseError(err))
	}
	objectives := make([]objectivesdomain.OverdueGuidanceObjective, 0, len(rows))
	for _, row := range rows {
		if row.EndDate == nil || row.LeadUserID == nil || row.WorkspaceID == nil || row.TeamID == nil ||
			*row.LeadUserID == uuid.Nil || *row.WorkspaceID == uuid.Nil || *row.TeamID == uuid.Nil {
			return nil, errInvalidObjectiveGuidanceItem
		}
		objectives = append(objectives, objectivesdomain.OverdueGuidanceObjective{
			ID:             row.ObjectiveID,
			Name:           row.Name,
			EndDate:        row.EndDate.UTC(),
			LeadUserID:     *row.LeadUserID,
			LeadEmail:      row.LeadEmail,
			LeadName:       row.LeadName,
			WorkspaceID:    *row.WorkspaceID,
			WorkspaceName:  row.WorkspaceName,
			WorkspaceSlug:  row.WorkspaceSlug,
			TeamID:         *row.TeamID,
			DeadlineStatus: row.DeadlineStatus,
			DaysDifference: int(row.DaysDifference),
			KeyResults:     string(row.KeyResults),
		})
	}
	return objectives, nil
}

func objectiveGuidanceUTCDate(value time.Time) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, fmt.Errorf("%w: objective guidance as-of time is required", objectivesdomain.ErrInvalid)
	}
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC), nil
}
