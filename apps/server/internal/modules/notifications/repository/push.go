package notificationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) RegisterPushDevice(ctx context.Context, command notificationsdomain.RegisterPushDevice) (notificationsdomain.PushDevice, error) {
	command.Token = strings.TrimSpace(command.Token)
	if err := command.Validate(); err != nil {
		return notificationsdomain.PushDevice{}, err
	}
	row, err := repository.queries.RegisterNotificationPushDevice(ctx, notificationssql.RegisterNotificationPushDeviceParams{
		UserID:        command.UserID,
		ExpoPushToken: command.Token,
		Platform:      string(command.Platform),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notificationsdomain.PushDevice{}, fmt.Errorf("register push device: %w", notificationsdomain.ErrForbidden)
		}
		return notificationsdomain.PushDevice{}, mapWriteError("register push device", err)
	}
	return notificationsdomain.PushDevice{
		ID: row.DeviceID, UserID: row.UserID, Token: row.ExpoPushToken,
		Platform:  notificationsdomain.PushPlatform(row.Platform),
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (repository *Repository) UnregisterPushDevice(ctx context.Context, userID uuid.UUID, token string) error {
	if userID == uuid.Nil || strings.TrimSpace(token) == "" {
		return fmt.Errorf("%w: user and token are required", notificationsdomain.ErrInvalid)
	}
	_, err := repository.queries.UnregisterNotificationPushDevice(ctx, notificationssql.UnregisterNotificationPushDeviceParams{
		UserID: userID, ExpoPushToken: strings.TrimSpace(token),
	})
	return mapWriteError("unregister push device", err)
}

func (repository *Repository) ListPushTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user ID is required", notificationsdomain.ErrInvalid)
	}
	tokens, err := repository.queries.ListNotificationPushTokensForUser(ctx, notificationssql.ListNotificationPushTokensForUserParams{
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list notification push tokens: %w", err)
	}
	return tokens, nil
}

func (repository *Repository) GetPushDelivery(ctx context.Context, notificationID uuid.UUID) (*notificationsdomain.PushDelivery, error) {
	if notificationID == uuid.Nil {
		return nil, fmt.Errorf("%w: notification ID is required", notificationsdomain.ErrInvalid)
	}
	rows, err := repository.queries.GetNotificationPushDelivery(ctx, notificationssql.GetNotificationPushDeliveryParams{NotificationID: notificationID})
	if err != nil {
		return nil, fmt.Errorf("get push delivery: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	entityType, err := notificationsdomain.ParseEntityType(string(rows[0].EntityType))
	if err != nil {
		return nil, fmt.Errorf("map push delivery: %w", err)
	}
	tokens := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.ExpoPushToken != "" {
			tokens = append(tokens, row.ExpoPushToken)
		}
	}
	var message notificationsdomain.NotificationMessage
	if err := json.Unmarshal(rows[0].Message, &message); err != nil {
		return nil, fmt.Errorf("map push delivery message: %w", err)
	}
	publicMessage, err := json.Marshal(message.Public())
	if err != nil {
		return nil, fmt.Errorf("marshal public push delivery message: %w", err)
	}
	return &notificationsdomain.PushDelivery{
		NotificationID: rows[0].NotificationID, RecipientID: rows[0].RecipientID,
		WorkspaceID: rows[0].WorkspaceID, EntityType: entityType, EntityID: rows[0].EntityID,
		WorkspaceSlug: rows[0].WorkspaceSlug, Title: rows[0].Title,
		Message: publicMessage, Tokens: tokens,
	}, nil
}

func (repository *Repository) MarkPushSent(ctx context.Context, notificationID uuid.UUID, at time.Time) error {
	if notificationID == uuid.Nil || at.IsZero() {
		return fmt.Errorf("%w: notification ID and delivery time are required", notificationsdomain.ErrInvalid)
	}
	_, err := repository.queries.MarkNotificationPushSent(ctx, notificationssql.MarkNotificationPushSentParams{
		NotificationID: notificationID, SentAt: at,
	})
	return mapWriteError("mark notification push sent", err)
}

func (repository *Repository) DisablePushDevices(ctx context.Context, tokens []string, at time.Time) error {
	if len(tokens) == 0 {
		return nil
	}
	_, err := repository.queries.DisableNotificationPushDevices(ctx, notificationssql.DisableNotificationPushDevicesParams{
		ExpoPushTokens: tokens, DisabledAt: at,
	})
	return mapWriteError("disable notification push devices", err)
}
