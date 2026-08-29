package notificationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) ListKeyResultAudience(ctx context.Context, query notificationsdomain.KeyResultAudienceQuery) ([]notificationsdomain.KeyResultAudienceMember, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListKeyResultNotificationAudience(ctx, notificationssql.ListKeyResultNotificationAudienceParams{
		ActorID:     query.ActorID,
		WorkspaceID: query.WorkspaceID,
		KeyResultID: query.KeyResultID,
	})
	if err != nil {
		return nil, fmt.Errorf("list key result notification audience: %w", err)
	}
	result := make([]notificationsdomain.KeyResultAudienceMember, 0, len(rows))
	for _, row := range rows {
		result = append(result, notificationsdomain.KeyResultAudienceMember{
			RecipientID:   row.RecipientID,
			KeyResultID:   row.KeyResultID,
			ObjectiveID:   row.ObjectiveID,
			KeyResultName: row.KeyResultName,
			ObjectiveName: row.ObjectiveName,
		})
	}
	return result, nil
}

func (repository *Repository) GetEmailDelivery(ctx context.Context, query notificationsdomain.EmailNotificationQuery) (*notificationsdomain.EmailNotification, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	row, err := repository.queries.GetNotificationEmailDelivery(ctx, notificationssql.GetNotificationEmailDeliveryParams{
		NotificationID: query.NotificationID,
		RecipientID:    query.Scope.RecipientID,
		WorkspaceID:    query.Scope.WorkspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get notification email delivery: %w", err)
	}
	notificationType, entityType, err := deliveryTypes(row.Type, row.EntityType)
	if err != nil {
		return nil, fmt.Errorf("map notification email delivery: %w", err)
	}
	return &notificationsdomain.EmailNotification{
		NotificationID:   row.NotificationID,
		RecipientID:      row.RecipientID,
		WorkspaceID:      row.WorkspaceID,
		NotificationType: notificationType,
		EntityType:       entityType,
		EntityID:         row.EntityID,
		Title:            row.Title,
		Message:          json.RawMessage(row.Message),
		UserEmail:        row.UserEmail,
		UserName:         row.UserName,
		ActorName:        row.ActorName,
		WorkspaceName:    row.WorkspaceName,
		WorkspaceSlug:    row.WorkspaceSlug,
		WorkspaceRole:    row.WorkspaceRole,
		EmailEnabled:     row.EmailEnabled,
		FeedbackSlug:     row.FeedbackSlug,
	}, nil
}

func (repository *Repository) ListEmailDigest(ctx context.Context, scope notificationsdomain.DeliveryScope) (*notificationsdomain.EmailDigest, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListNotificationEmailDigestDeliveries(ctx, notificationssql.ListNotificationEmailDigestDeliveriesParams{
		RecipientID: scope.RecipientID,
		WorkspaceID: scope.WorkspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list notification email digest deliveries: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	items := make([]notificationsdomain.EmailDigestItem, 0, len(rows))
	for _, row := range rows {
		if row.CreatedAt == nil {
			return nil, fmt.Errorf("notification %s has null created_at", row.NotificationID)
		}
		notificationType, entityType, err := deliveryTypes(row.Type, row.EntityType)
		if err != nil {
			return nil, fmt.Errorf("map digest notification %s: %w", row.NotificationID, err)
		}
		items = append(items, notificationsdomain.EmailDigestItem{
			NotificationID:   row.NotificationID,
			NotificationType: notificationType,
			EntityType:       entityType,
			EntityID:         row.EntityID,
			Title:            row.Title,
			Message:          json.RawMessage(row.Message),
			CreatedAt:        *row.CreatedAt,
			ActorName:        row.ActorName,
			FeedbackSlug:     row.FeedbackSlug,
		})
	}
	return &notificationsdomain.EmailDigest{
		RecipientID:   rows[0].RecipientID,
		WorkspaceID:   rows[0].WorkspaceID,
		UserEmail:     rows[0].UserEmail,
		UserName:      rows[0].UserName,
		WorkspaceName: rows[0].WorkspaceName,
		WorkspaceSlug: rows[0].WorkspaceSlug,
		WorkspaceRole: rows[0].WorkspaceRole,
		Items:         items,
	}, nil
}

func (repository *Repository) ListDeliveryTeamIDs(ctx context.Context, scope notificationsdomain.DeliveryScope) ([]uuid.UUID, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	teamIDs, err := repository.queries.ListNotificationDeliveryTeamIDs(ctx, notificationssql.ListNotificationDeliveryTeamIDsParams{
		WorkspaceID: scope.WorkspaceID,
		RecipientID: scope.RecipientID,
	})
	if err != nil {
		return nil, fmt.Errorf("list notification delivery teams: %w", err)
	}
	return teamIDs, nil
}

func (repository *Repository) MarkEmailSent(ctx context.Context, command notificationsdomain.MarkEmailSent) error {
	if err := command.Validate(); err != nil {
		return err
	}
	_, err := repository.queries.MarkNotificationEmailsSent(ctx, notificationssql.MarkNotificationEmailsSentParams{
		SentAt:          command.At,
		RecipientID:     command.Scope.RecipientID,
		WorkspaceID:     command.Scope.WorkspaceID,
		NotificationIds: command.NotificationIDs,
	})
	return mapWriteError("mark notification emails sent", err)
}

func deliveryTypes(
	notificationType notificationssql.NotificationType,
	entityType notificationssql.EntityType,
) (notificationsdomain.NotificationType, notificationsdomain.EntityType, error) {
	mappedNotificationType, err := notificationsdomain.ParseNotificationType(string(notificationType))
	if err != nil {
		return "", "", err
	}
	mappedEntityType, err := notificationsdomain.ParseEntityType(string(entityType))
	if err != nil {
		return "", "", err
	}
	if !mappedNotificationType.SupportsEntity(mappedEntityType) {
		return "", "", fmt.Errorf(
			"%w: notification type %q is incompatible with entity type %q",
			notificationsdomain.ErrInvalid,
			mappedNotificationType,
			mappedEntityType,
		)
	}
	return mappedNotificationType, mappedEntityType, nil
}
