package calendarrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) UpsertConnection(ctx context.Context, input calendar.CoreConnectionUpsert) (calendar.CoreConnection, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreConnection{}, err
	}

	var connection calendar.CoreConnection
	err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, input.UserID); err != nil {
			return err
		}

		existing, err := queries.LockExistingCalendarConnection(ctx, calendarsql.LockExistingCalendarConnectionParams{
			UserID:   input.UserID,
			Provider: string(input.Provider),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("lock existing calendar connection: %w", err)
		}
		if err == nil {
			if existing.CleanupPending {
				return calendar.ErrCalendarCleanupPending
			}
			if strings.TrimSpace(existing.ProviderAccountID) != "" && existing.ProviderAccountID != input.ProviderAccountID {
				return calendar.ErrCalendarAccountChangePending
			}
		}

		row, err := queries.UpsertCalendarConnection(ctx, calendarsql.UpsertCalendarConnectionParams{
			WorkspaceID:          input.WorkspaceID,
			UserID:               input.UserID,
			CredentialGeneration: uuid.New(),
			ProviderAccountID:    input.ProviderAccountID,
			Provider:             string(input.Provider),
			CanWrite:             (calendar.CoreConnection{Provider: input.Provider, Scopes: input.Scopes}).CanWriteEvents(),
			ConnectedEmail:       input.ConnectedEmail,
			Timezone:             input.Timezone,
			TokenPayload:         input.TokenPayload,
			Scopes:               input.Scopes,
		})
		if err != nil {
			return fmt.Errorf("upsert calendar connection: %w", err)
		}
		if err := queries.ReactivateCalendarOutboxAfterAuthorizationRefresh(ctx, calendarsql.ReactivateCalendarOutboxAfterAuthorizationRefreshParams{
			UserID:   input.UserID,
			Provider: string(input.Provider),
		}); err != nil {
			return fmt.Errorf("reactivate calendar outbox after authorization refresh: %w", err)
		}
		connection = toCoreConnection(row)
		return nil
	})
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("upsert calendar connection transaction: %w", err)
	}
	return connection, nil
}

func (r *Repo) SetPrimaryConnection(ctx context.Context, _ uuid.UUID, userID, connectionID uuid.UUID) (calendar.CoreConnection, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreConnection{}, err
	}

	var connection calendar.CoreConnection
	err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, userID); err != nil {
			return err
		}
		row, err := queries.LockCalendarConnectionForPrimary(ctx, calendarsql.LockCalendarConnectionForPrimaryParams{
			UserID:       userID,
			ConnectionID: connectionID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return calendar.ErrCalendarNotFound
		}
		if err != nil {
			return fmt.Errorf("load primary calendar connection: %w", err)
		}
		selected := toCoreConnection(row)
		if !selected.CanWriteEvents() {
			return calendar.ErrCalendarReauthorizationRequired
		}
		if selected.IsPrimary {
			connection = selected
			return nil
		}
		if err := queries.ClearPrimaryCalendarConnection(ctx, calendarsql.ClearPrimaryCalendarConnectionParams{UserID: userID}); err != nil {
			return fmt.Errorf("clear primary calendar connection: %w", err)
		}
		row, err = queries.SetPrimaryCalendarConnection(ctx, calendarsql.SetPrimaryCalendarConnectionParams{ConnectionID: connectionID})
		if err != nil {
			return fmt.Errorf("set primary calendar connection: %w", err)
		}
		connection = toCoreConnection(row)
		return nil
	})
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("set primary calendar connection transaction: %w", err)
	}
	return connection, nil
}

func (r *Repo) UpdateConnectionToken(ctx context.Context, connection calendar.CoreConnection, tokenPayload string) error {
	if err := r.configured(); err != nil {
		return err
	}
	rows, err := r.queries.UpdateCalendarConnectionToken(ctx, calendarsql.UpdateCalendarConnectionTokenParams{
		TokenPayload:         tokenPayload,
		ConnectionID:         connection.ID,
		CredentialGeneration: connection.CredentialGeneration,
	})
	if err != nil {
		return fmt.Errorf("update calendar connection token: %w", err)
	}
	if rows == 0 {
		return calendar.ErrCalendarSyncSuperseded
	}
	return nil
}

// BeginConnectionSync rotates the generation before provider I/O. A later sync
// attempt supersedes every earlier in-flight attempt.
func (r *Repo) BeginConnectionSync(ctx context.Context, connection calendar.CoreConnection) (calendar.CoreConnection, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreConnection{}, err
	}
	var started calendar.CoreConnection
	err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, connection.UserID); err != nil {
			return err
		}
		row, err := queries.BeginCalendarConnectionSync(ctx, calendarsql.BeginCalendarConnectionSyncParams{
			NextCredentialGeneration:    uuid.New(),
			ConnectionID:                connection.ID,
			WorkspaceID:                 connection.WorkspaceID,
			UserID:                      connection.UserID,
			Provider:                    string(connection.Provider),
			CurrentCredentialGeneration: connection.CredentialGeneration,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return calendar.ErrCalendarSyncSuperseded
		}
		if err != nil {
			return fmt.Errorf("start calendar sync: %w", err)
		}
		started = toCoreConnection(row)
		return nil
	})
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("begin calendar sync transaction: %w", err)
	}
	return started, nil
}

func (r *Repo) RevokeConnection(ctx context.Context, _ uuid.UUID, userID, connectionID uuid.UUID) error {
	if err := r.configured(); err != nil {
		return err
	}
	if err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, userID); err != nil {
			return err
		}
		revoked, err := queries.LockCalendarConnectionForRevocation(ctx, calendarsql.LockCalendarConnectionForRevocationParams{
			UserID:       userID,
			ConnectionID: connectionID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return calendar.ErrCalendarNotFound
		}
		if err != nil {
			return fmt.Errorf("lock revoked calendar connection: %w", err)
		}
		rows, err := queries.MarkCalendarConnectionCleanupPending(ctx, calendarsql.MarkCalendarConnectionCleanupPendingParams{
			UserID:       userID,
			ConnectionID: connectionID,
		})
		if err != nil {
			return fmt.Errorf("revoke calendar connection: %w", err)
		}
		if rows == 0 {
			return calendar.ErrCalendarNotFound
		}
		if err := queries.ReactivateCalendarOutboxForCleanup(ctx, calendarsql.ReactivateCalendarOutboxForCleanupParams{
			UserID:   userID,
			Provider: revoked.Provider,
		}); err != nil {
			return fmt.Errorf("reactivate calendar cleanup outbox: %w", err)
		}
		mappings, err := queries.ListMayaScheduleMirrorsForCleanup(ctx, calendarsql.ListMayaScheduleMirrorsForCleanupParams{
			UserID:   userID,
			Provider: revoked.Provider,
		})
		if err != nil {
			return fmt.Errorf("list Maya schedule mirrors for calendar cleanup: %w", err)
		}
		for _, mapping := range mappings {
			event := cleanupScheduleEvent(mapping, revoked.Provider)
			if err := enqueueScheduleEventOutbox(ctx, queries, mapping.WorkspaceID, userID, &mapping.BlockID, calendar.Provider(revoked.Provider), calendar.ScheduleEventOperationDelete, event, "", true); err != nil {
				return err
			}
		}
		if err := queries.DetachMayaScheduleMirrors(ctx, calendarsql.DetachMayaScheduleMirrorsParams{UserID: userID, Provider: revoked.Provider}); err != nil {
			return fmt.Errorf("detach Maya schedule mirrors during calendar cleanup: %w", err)
		}
		if revoked.IsPrimary {
			if err := queries.PromoteReplacementPrimaryCalendarConnection(ctx, calendarsql.PromoteReplacementPrimaryCalendarConnectionParams{
				UserID:              userID,
				GoogleReadScope:     calendar.GoogleCalendarEventsReadonlyScope,
				GoogleOwnedScope:    calendar.GoogleCalendarEventsOwnedScope,
				MicrosoftWriteScope: calendar.MicrosoftCalendarReadWriteScope,
			}); err != nil {
				return fmt.Errorf("promote replacement primary calendar connection: %w", err)
			}
		}
		if err := queries.DeleteCalendarConnectionEvents(ctx, calendarsql.DeleteCalendarConnectionEventsParams{ConnectionID: connectionID}); err != nil {
			return fmt.Errorf("delete calendar events for revoked connection: %w", err)
		}
		if err := queries.DeleteCalendarConnectionBusyWindows(ctx, calendarsql.DeleteCalendarConnectionBusyWindowsParams{ConnectionID: connectionID}); err != nil {
			return fmt.Errorf("delete calendar busy windows for revoked connection: %w", err)
		}
		if err := queries.DeleteDrainedCalendarConnection(ctx, calendarsql.DeleteDrainedCalendarConnectionParams{ConnectionID: connectionID}); err != nil {
			return fmt.Errorf("delete calendar connection without pending provider cleanup: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("revoke calendar connection transaction: %w", err)
	}
	return nil
}

func cleanupScheduleEvent(mapping calendarsql.ListMayaScheduleMirrorsForCleanupRow, provider string) calendar.ExternalScheduleEventInput {
	calendarID := "primary"
	if mapping.ExternalCalendarID != nil && strings.TrimSpace(*mapping.ExternalCalendarID) != "" {
		calendarID = strings.TrimSpace(*mapping.ExternalCalendarID)
	}
	eventID := calendar.StableGoogleScheduleEventID(mapping.BlockID)
	if provider == string(calendar.ProviderMicrosoft) {
		eventID = "pending:" + mapping.BlockID.String()
	}
	if mapping.ExternalEventID != nil && strings.TrimSpace(*mapping.ExternalEventID) != "" {
		eventID = strings.TrimSpace(*mapping.ExternalEventID)
	}
	storyID := uuid.Nil
	if mapping.StoryID != nil {
		storyID = *mapping.StoryID
	}
	return calendar.ExternalScheduleEventInput{
		CalendarID:  calendarID,
		EventID:     eventID,
		BlockID:     mapping.BlockID,
		StoryID:     storyID,
		WorkspaceID: mapping.WorkspaceID,
	}
}

func (r *Repo) ReplaceCalendarSnapshot(ctx context.Context, connection calendar.CoreConnection, snapshot calendar.CalendarSyncSnapshot) error {
	if err := r.configured(); err != nil {
		return err
	}
	if err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, connection.UserID); err != nil {
			return err
		}
		_, err := queries.LockCalendarConnectionForSnapshot(ctx, calendarsql.LockCalendarConnectionForSnapshotParams{
			ConnectionID:         connection.ID,
			WorkspaceID:          connection.WorkspaceID,
			UserID:               connection.UserID,
			Provider:             string(connection.Provider),
			CredentialGeneration: connection.CredentialGeneration,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return calendar.ErrCalendarSyncSuperseded
		}
		if err != nil {
			return fmt.Errorf("lock active calendar connection: %w", err)
		}
		if connection.CanReadEventDetails() && !snapshot.CanReadEventDetails {
			if err := queries.RemoveCalendarConnectionScope(ctx, calendarsql.RemoveCalendarConnectionScopeParams{
				Scope:        calendar.GoogleCalendarEventsReadonlyScope,
				ConnectionID: connection.ID,
			}); err != nil {
				return fmt.Errorf("downgrade calendar event detail access: %w", err)
			}
		}
		if err := queries.InvalidateMayaScheduleMirrorHashes(ctx, calendarsql.InvalidateMayaScheduleMirrorHashesParams{
			UserID:   connection.UserID,
			Provider: string(connection.Provider),
		}); err != nil {
			return fmt.Errorf("invalidate Maya schedule mirrors before full calendar sync: %w", err)
		}

		syncGeneration := uuid.New()
		for _, event := range snapshot.Events {
			params, err := calendarEventUpsertParams(connection, event, syncGeneration)
			if err != nil {
				return err
			}
			if err := queries.UpsertCalendarEvent(ctx, params); err != nil {
				return fmt.Errorf("upsert calendar event: %w", err)
			}
		}
		if err := queries.DeleteStaleCalendarEvents(ctx, calendarsql.DeleteStaleCalendarEventsParams{
			ConnectionID:   connection.ID,
			SyncGeneration: syncGeneration,
		}); err != nil {
			return fmt.Errorf("delete stale calendar events: %w", err)
		}
		if err := queries.DeleteCalendarBusyWindows(ctx, calendarsql.DeleteCalendarBusyWindowsParams{ConnectionID: connection.ID}); err != nil {
			return fmt.Errorf("delete existing calendar busy windows: %w", err)
		}
		for _, window := range snapshot.BusyWindows {
			if err := queries.InsertCalendarBusyWindow(ctx, calendarBusyWindowParams(connection, window)); err != nil {
				return fmt.Errorf("insert calendar busy window: %w", err)
			}
		}
		if err := queries.UpdateCalendarSnapshotMetadata(ctx, calendarsql.UpdateCalendarSnapshotMetadataParams{
			SyncToken:    strings.TrimSpace(snapshot.NextSyncToken),
			Timezone:     strings.TrimSpace(snapshot.Timezone),
			ConnectionID: connection.ID,
		}); err != nil {
			return fmt.Errorf("store calendar sync metadata: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("replace calendar snapshot transaction: %w", err)
	}
	return nil
}

func calendarEventUpsertParams(connection calendar.CoreConnection, event calendar.CoreCalendarEvent, syncGeneration uuid.UUID) (calendarsql.UpsertCalendarEventParams, error) {
	organizer, err := marshalOptionalCalendarParticipant(event.Organizer)
	if err != nil {
		return calendarsql.UpsertCalendarEventParams{}, err
	}
	attendees, err := json.Marshal(event.Attendees)
	if err != nil {
		return calendarsql.UpsertCalendarEventParams{}, fmt.Errorf("encode calendar event attendees: %w", err)
	}
	startDate, err := parseOptionalCalendarDate(event.StartDate)
	if err != nil {
		return calendarsql.UpsertCalendarEventParams{}, fmt.Errorf("parse calendar event start date: %w", err)
	}
	endDate, err := parseOptionalCalendarDate(event.EndDate)
	if err != nil {
		return calendarsql.UpsertCalendarEventParams{}, fmt.Errorf("parse calendar event end date: %w", err)
	}
	return calendarsql.UpsertCalendarEventParams{
		ConnectionID:     connection.ID,
		WorkspaceID:      connection.WorkspaceID,
		UserID:           connection.UserID,
		Provider:         string(connection.Provider),
		CalendarID:       event.CalendarID,
		ProviderEventID:  event.ProviderEventID,
		Title:            event.Title,
		Description:      event.Description,
		Location:         event.Location,
		MeetingURL:       event.MeetingURL,
		HtmlLink:         event.HTMLLink,
		Organizer:        organizer,
		Attendees:        attendees,
		AttendeesOmitted: event.AttendeesOmitted,
		IsAllDay:         event.IsAllDay,
		StartDate:        startDate,
		EndDate:          endDate,
		StartAt:          event.StartAt,
		EndAt:            event.EndAt,
		Visibility:       event.Visibility,
		IsPrivate:        event.IsPrivate,
		SourceHash:       event.SourceHash,
		SyncGeneration:   syncGeneration,
	}, nil
}

func calendarBusyWindowParams(connection calendar.CoreConnection, window calendar.CoreBusyWindow) calendarsql.InsertCalendarBusyWindowParams {
	return calendarsql.InsertCalendarBusyWindowParams{
		ConnectionID:    connection.ID,
		WorkspaceID:     connection.WorkspaceID,
		UserID:          connection.UserID,
		Provider:        string(connection.Provider),
		ProviderEventID: window.ProviderEventID,
		CalendarID:      window.CalendarID,
		Title:           nil,
		StartAt:         window.StartAt,
		EndAt:           window.EndAt,
		Status:          string(window.Status),
		Transparency:    string(window.Transparency),
		IsPrivate:       window.IsPrivate,
		SourceHash:      window.SourceHash,
	}
}

func marshalOptionalCalendarParticipant(participant *calendar.CoreCalendarParticipant) ([]byte, error) {
	if participant == nil {
		return nil, nil
	}
	value, err := json.Marshal(participant)
	if err != nil {
		return nil, fmt.Errorf("encode calendar event organizer: %w", err)
	}
	return value, nil
}

func parseOptionalCalendarDate(value *string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.DateOnly, strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (r *Repo) MarkConnectionSynced(ctx context.Context, workspaceID, connectionID, credentialGeneration uuid.UUID, syncedAt time.Time) error {
	if err := r.configured(); err != nil {
		return err
	}
	rows, err := r.queries.MarkCalendarConnectionSynced(ctx, calendarsql.MarkCalendarConnectionSyncedParams{
		SyncedAt:             &syncedAt,
		WorkspaceID:          workspaceID,
		ConnectionID:         connectionID,
		CredentialGeneration: credentialGeneration,
	})
	if err != nil {
		return fmt.Errorf("mark calendar connection synced: %w", err)
	}
	if rows == 0 {
		return calendar.ErrCalendarSyncSuperseded
	}
	return nil
}

func (r *Repo) MarkConnectionSyncFailed(ctx context.Context, workspaceID, connectionID, credentialGeneration uuid.UUID, message string) error {
	if err := r.configured(); err != nil {
		return err
	}
	rows, err := r.queries.MarkCalendarConnectionSyncFailed(ctx, calendarsql.MarkCalendarConnectionSyncFailedParams{
		SyncError:            &message,
		WorkspaceID:          workspaceID,
		ConnectionID:         connectionID,
		CredentialGeneration: credentialGeneration,
	})
	if err != nil {
		return fmt.Errorf("mark calendar connection sync failed: %w", err)
	}
	if rows == 0 {
		return calendar.ErrCalendarSyncSuperseded
	}
	return nil
}
