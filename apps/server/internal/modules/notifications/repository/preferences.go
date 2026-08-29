package notificationsrepository

import (
	"context"
	"errors"
	"fmt"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) GetPreferences(ctx context.Context, access notificationsdomain.WorkspaceAccess) (notificationsdomain.Preferences, error) {
	if err := access.Validate(); err != nil {
		return notificationsdomain.Preferences{}, err
	}
	defaults, err := defaultPreferencesJSON()
	if err != nil {
		return notificationsdomain.Preferences{}, err
	}
	row, err := repository.queries.GetNotificationPreferences(ctx, notificationssql.GetNotificationPreferencesParams{
		WorkspaceID:        access.WorkspaceID,
		ActorID:            access.ActorID,
		DefaultPreferences: defaults,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notificationsdomain.Preferences{}, fmt.Errorf("get notification preferences: %w", notificationsdomain.ErrForbidden)
		}
		return notificationsdomain.Preferences{}, fmt.Errorf("get notification preferences: %w", err)
	}
	return toPreferences(row.PreferenceID, row.UserID, row.WorkspaceID, row.Preferences, row.CreatedAt, row.UpdatedAt)
}

func (repository *Repository) UpdatePreference(ctx context.Context, command notificationsdomain.UpdatePreference) (notificationsdomain.Preferences, error) {
	command.Patch = command.Patch.Normalized(command.Type)
	if err := command.Validate(); err != nil {
		return notificationsdomain.Preferences{}, err
	}
	defaults, err := defaultPreferencesJSON()
	if err != nil {
		return notificationsdomain.Preferences{}, err
	}
	email, emailPresent := patchValue(command.Patch.Email.Value())
	inApp, inAppPresent := patchValue(command.Patch.InApp.Value())
	row, err := repository.queries.UpdateNotificationPreference(ctx, notificationssql.UpdateNotificationPreferenceParams{
		WorkspaceID:        command.Access.WorkspaceID,
		ActorID:            command.Access.ActorID,
		DefaultPreferences: defaults,
		PreferenceType:     string(command.Type),
		EmailPresent:       emailPresent,
		EmailEnabled:       email,
		SupportsInApp:      command.Type.SupportsInAppDelivery(),
		InAppPresent:       inAppPresent,
		InAppEnabled:       inApp,
		UpdatedAt:          command.At,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notificationsdomain.Preferences{}, fmt.Errorf("update notification preference: %w", notificationsdomain.ErrForbidden)
		}
		return notificationsdomain.Preferences{}, mapWriteError("update notification preference", err)
	}
	return toPreferences(row.PreferenceID, row.UserID, row.WorkspaceID, row.Preferences, row.CreatedAt, row.UpdatedAt)
}

func patchValue(value *bool, specified bool) (bool, bool) {
	if !specified || value == nil {
		return false, specified
	}
	return *value, true
}
