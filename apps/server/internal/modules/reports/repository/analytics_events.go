package reportsrepository

import (
	"context"
	"encoding/json"
	"fmt"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
)

const maxAnalyticsEventPropertiesBytes = 64 << 10

func (r *repo) CreateWorkspaceAnalyticsEvent(ctx context.Context, input reportdomain.CoreWorkspaceAnalyticsEventInput) error {
	if err := r.authorize(ctx, input.UserID, input.WorkspaceID); err != nil {
		return err
	}

	properties, err := json.Marshal(input.Properties)
	if err != nil {
		return fmt.Errorf("marshaling analytics event properties: %w", reportdomain.ErrInvalidWorkspaceAnalyticsEvent)
	}
	if len(properties) > maxAnalyticsEventPropertiesBytes {
		return reportdomain.ErrInvalidWorkspaceAnalyticsEvent
	}

	rows, err := r.queries.CreateWorkspaceAnalyticsEvent(ctx, reportssql.CreateWorkspaceAnalyticsEventParams{
		WorkspaceID: input.WorkspaceID,
		ActorID:     input.UserID,
		TeamID:      input.TeamID,
		StoryID:     input.StoryID,
		ObjectiveID: input.ObjectiveID,
		SprintID:    input.SprintID,
		KeyResultID: input.KeyResultID,
		EventName:   input.EventName,
		Surface:     input.Surface,
		Properties:  properties,
		OccurredAt:  input.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("inserting workspace analytics event: %w", err)
	}
	if rows != 1 {
		return reportdomain.ErrReportsAccessDenied
	}

	return nil
}
