package integrationrequestsrepository

import (
	"context"
	"fmt"
	"strings"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	integrationrequestssql "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository/sqlc"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

func (r *Repo) UpdatePending(ctx context.Context, workspaceID, requestID, userID uuid.UUID, input integrationrequestdomain.UpdateRequestInput) (integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.IntegrationRequest{}, err
	}
	var result integrationrequestdomain.IntegrationRequest
	err := r.withinTransaction(ctx, func(queries integrationrequestssql.Querier) error {
		currentRow, err := queries.LockPendingIntegrationRequestForUpdate(ctx, integrationrequestssql.LockPendingIntegrationRequestForUpdateParams{
			WorkspaceID: workspaceID, RequestID: requestID, ActorID: userID,
		})
		if err != nil {
			return mapNotFound("lock pending integration request for update", err)
		}
		current, err := integrationRequestFromSQL(currentRow)
		if err != nil {
			return err
		}
		params, err := updateParams(current, workspaceID, requestID, userID, input)
		if err != nil {
			return err
		}
		if err := validatePendingRequestProperties(
			ctx, queries, current.WorkspaceID, current.TeamID, params.StatusID, params.AssigneeID,
			params.ObjectiveID, params.KeyResultID, params.SprintID, params.LabelIds, params.StartDate, params.EndDate,
		); err != nil {
			return err
		}
		updated, err := queries.UpdatePendingIntegrationRequest(ctx, params)
		if err != nil {
			return mapNotFound("update pending integration request", err)
		}
		result, err = integrationRequestFromSQL(updated)
		return err
	})
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, fmt.Errorf("update integration request transaction: %w", err)
	}
	return result, nil
}

// ReserveAcceptance durably fences a pending request before story creation.
// Repeated calls return the existing reservation so a crashed conversion can
// resume with the original actor and the same story idempotency key.
func (r *Repo) ReserveAcceptance(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.IntegrationRequest{}, err
	}
	var result integrationrequestdomain.IntegrationRequest
	err := r.withinTransaction(ctx, func(queries integrationrequestssql.Querier) error {
		currentRow, err := queries.LockPendingIntegrationRequestForAcceptance(ctx, integrationrequestssql.LockPendingIntegrationRequestForAcceptanceParams{
			WorkspaceID: workspaceID, RequestID: requestID, ActorID: userID,
		})
		if err != nil {
			return mapNotFound("lock pending integration request for acceptance", err)
		}
		current, err := integrationRequestFromSQL(currentRow)
		if err != nil {
			return err
		}
		if current.AcceptanceState == integrationrequestdomain.AcceptanceStateReserved {
			result = current
			return nil
		}
		if current.AcceptanceState != integrationrequestdomain.AcceptanceStateIdle {
			return integrationrequestdomain.ErrRequestNotPending
		}
		labelIDs, err := requestLabelIDs(current.LabelIDs, nil)
		if err != nil {
			return err
		}
		if err := validatePendingRequestProperties(
			ctx, queries, current.WorkspaceID, current.TeamID, current.StatusID, current.AssigneeID,
			current.ObjectiveID, current.KeyResultID, current.SprintID, labelIDs, current.StartDate, current.EndDate,
		); err != nil {
			return err
		}
		reserved, err := queries.ReserveIntegrationRequestAcceptance(ctx, integrationrequestssql.ReserveIntegrationRequestAcceptanceParams{
			ActorID: uuidPointer(userID), WorkspaceID: workspaceID, RequestID: requestID,
		})
		if err != nil {
			return mapNotFound("reserve integration request acceptance", err)
		}
		result, err = integrationRequestFromSQL(reserved)
		return err
	})
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, fmt.Errorf("reserve integration request acceptance transaction: %w", err)
	}
	return result, nil
}

func (r *Repo) MarkAccepted(ctx context.Context, workspaceID, requestID, storyID, acceptedByUserID uuid.UUID) (integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.IntegrationRequest{}, err
	}
	row, err := r.queries.MarkIntegrationRequestAccepted(ctx, integrationrequestssql.MarkIntegrationRequestAcceptedParams{
		StoryID: uuidPointer(storyID), ActorID: uuidPointer(acceptedByUserID), WorkspaceID: workspaceID, RequestID: requestID,
	})
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, mapNotFound("mark integration request accepted", err)
	}
	return integrationRequestFromSQL(row)
}

func (r *Repo) MarkDeclined(ctx context.Context, workspaceID, requestID, declinedByUserID uuid.UUID) (integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.IntegrationRequest{}, err
	}
	row, err := r.queries.MarkIntegrationRequestDeclined(ctx, integrationrequestssql.MarkIntegrationRequestDeclinedParams{
		ActorID: uuidPointer(declinedByUserID), WorkspaceID: workspaceID, RequestID: requestID,
	})
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, mapNotFound("mark integration request declined", err)
	}
	return integrationRequestFromSQL(row)
}

func updateParams(current integrationrequestdomain.IntegrationRequest, workspaceID, requestID, actorID uuid.UUID, input integrationrequestdomain.UpdateRequestInput) (integrationrequestssql.UpdatePendingIntegrationRequestParams, error) {
	title := current.Title
	if input.Title != nil {
		title = strings.TrimSpace(*input.Title)
	}
	priority := current.Priority
	if input.Priority != nil {
		priority = strings.TrimSpace(*input.Priority)
	}
	estimatedDuration := optionalValue(current.EstimatedDurationMinutes, input.EstimatedDurationMinutes)
	minimumFocusBlock := optionalValue(current.MinimumFocusBlockMinutes, input.MinimumFocusBlockMinutes)
	if err := storydomain.ValidateScheduling(estimatedDuration, minimumFocusBlock); err != nil {
		return integrationrequestssql.UpdatePendingIntegrationRequestParams{}, fmt.Errorf("%w: %w", integrationrequestdomain.ErrInvalidRequestProperty, err)
	}
	estimatedDurationSQL, err := optionalInt32(estimatedDuration)
	if err != nil {
		return integrationrequestssql.UpdatePendingIntegrationRequestParams{}, fmt.Errorf("convert estimated duration: %w", err)
	}
	minimumFocusBlockSQL, err := optionalInt32(minimumFocusBlock)
	if err != nil {
		return integrationrequestssql.UpdatePendingIntegrationRequestParams{}, fmt.Errorf("convert minimum focus block: %w", err)
	}
	labelIDs, err := requestLabelIDs(current.LabelIDs, input.LabelIDs)
	if err != nil {
		return integrationrequestssql.UpdatePendingIntegrationRequestParams{}, err
	}
	return integrationrequestssql.UpdatePendingIntegrationRequestParams{
		Title: title, Description: optionalValue(current.Description, input.Description),
		StatusID: optionalValue(current.StatusID, input.StatusID), Priority: priority,
		AssigneeID:               optionalValue(current.AssigneeID, input.AssigneeID),
		EstimateUnit:             optionalValue(current.EstimateValue, input.EstimateValue),
		EstimatedDurationMinutes: estimatedDurationSQL, MinimumFocusBlockMinutes: minimumFocusBlockSQL,
		ObjectiveID: optionalValue(current.ObjectiveID, input.ObjectiveID),
		KeyResultID: optionalValue(current.KeyResultID, input.KeyResultID),
		SprintID:    optionalValue(current.SprintID, input.SprintID),
		StartDate:   optionalValue(current.StartDate, input.StartDate), EndDate: optionalValue(current.EndDate, input.EndDate),
		LabelIds: labelIDs, WorkspaceID: workspaceID, RequestID: requestID, ActorID: actorID,
	}, nil
}
