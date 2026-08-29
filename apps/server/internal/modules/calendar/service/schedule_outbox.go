package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/api/googleapi"
)

func (s *Service) DispatchScheduleEventOutbox(ctx context.Context, userID uuid.UUID) error {
	if s.repo == nil || userID == uuid.Nil {
		return ErrCalendarNotConfigured
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return err
	}
	var dispatchErr error
	lockErr := scheduleRepo.WithScheduleEventDispatchLock(ctx, userID, func(outbox ScheduleEventOutboxStore) error {
		connection, cleanupPending, err := s.repo.GetScheduleEventDispatchConnection(ctx, userID)
		if err != nil {
			if errors.Is(err, ErrCalendarNotFound) {
				return nil
			}
			return err
		}
		if cleanupPending && !connection.CanDeleteOwnedEvents() {
			return terminallyFinalizeScheduleCleanup(
				ctx,
				outbox,
				userID,
				connection.Provider,
				"Calendar cleanup could not call the provider because the connection never granted event deletion access.",
			)
		}
		if !cleanupPending && !connection.CanWriteEvents() {
			return nil
		}
		provider, err := s.provider(connection.Provider)
		if err != nil {
			if cleanupPending {
				return terminallyFinalizeScheduleCleanup(ctx, outbox, userID, connection.Provider, "Calendar cleanup could not initialize the provider writer.")
			}
			return err
		}
		eventWriter, ok := provider.(CalendarEventWriter)
		if !ok {
			if cleanupPending {
				return terminallyFinalizeScheduleCleanup(ctx, outbox, userID, connection.Provider, "Calendar cleanup provider does not support event deletion.")
			}
			return ErrCalendarNotConfigured
		}
		token, err := s.tokenForConnection(ctx, connection, provider)
		if err != nil {
			if cleanupPending {
				return terminallyFinalizeScheduleCleanup(ctx, outbox, userID, connection.Provider, "Calendar cleanup credentials could not be decrypted.")
			}
			return err
		}
		for {
			items, err := outbox.ListPendingScheduleEventOutbox(ctx, userID, connection.Provider, 100)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return outbox.DeleteCleanupPendingConnectionIfDrained(ctx, userID, connection.Provider)
			}
			for index, item := range items {
				var event ExternalScheduleEventInput
				isMember := false
				if !cleanupPending {
					var membershipErr error
					isMember, membershipErr = s.repo.WorkspaceMemberExists(ctx, item.WorkspaceID, userID)
					if membershipErr != nil {
						return membershipErr
					}
				}
				needsEventPayload := item.Operation == ScheduleEventOperationUpsert && !cleanupPending && isMember
				if needsEventPayload {
					if err := json.Unmarshal(item.Payload, &event); err != nil {
						if markErr := outbox.MarkScheduleEventOutboxFailed(ctx, item, "Calendar event payload could not be decoded.", true); markErr != nil {
							return errors.Join(err, markErr)
						}
						if releaseErr := outbox.ReleaseScheduleEventOutbox(ctx, scheduleOutboxIDs(items[index+1:])); releaseErr != nil {
							return errors.Join(err, releaseErr)
						}
						if cleanupPending {
							if finalizeErr := outbox.DeleteCleanupPendingConnectionIfDrained(ctx, userID, connection.Provider); finalizeErr != nil {
								return errors.Join(err, finalizeErr)
							}
						}
						dispatchErr = fmt.Errorf("decode calendar schedule event outbox: %w", err)
						return nil
					}
				}
				forceDelete := cleanupPending || item.Operation == ScheduleEventOperationDelete || item.Operation == ScheduleEventOperationUpsert && !isMember
				if !forceDelete && item.Operation == ScheduleEventOperationUpsert {
					current, currentErr := outbox.ScheduleEventUpsertIsCurrent(ctx, item, event)
					if currentErr != nil {
						return currentErr
					}
					forceDelete = !current
				}
				var writeErr error
				writeResult := ExternalScheduleEventResult{EventID: item.ProviderEventID}
				failureMessage := "Calendar event update failed."
				if forceDelete {
					writeErr = eventWriter.DeleteScheduleEvent(ctx, token, item.CalendarID, item.ProviderEventID)
					failureMessage = "Calendar event cleanup failed."
				} else {
					writeResult, writeErr = eventWriter.UpsertScheduleEvent(ctx, token, event)
				}
				if writeErr != nil {
					failureMessage = fmt.Sprintf("%s %v", failureMessage, writeErr)
					terminal := isPermanentCalendarWriteError(writeErr) || item.AttemptCount >= maximumScheduleEventOutboxAttempts
					if markErr := outbox.MarkScheduleEventOutboxFailed(ctx, item, failureMessage, terminal); markErr != nil {
						return errors.Join(writeErr, markErr)
					}
					if releaseErr := outbox.ReleaseScheduleEventOutbox(ctx, scheduleOutboxIDs(items[index+1:])); releaseErr != nil {
						return errors.Join(writeErr, releaseErr)
					}
					if terminal && cleanupPending {
						if finalizeErr := outbox.DeleteCleanupPendingConnectionIfDrained(ctx, userID, connection.Provider); finalizeErr != nil {
							return errors.Join(writeErr, finalizeErr)
						}
					}
					dispatchErr = writeErr
					return nil
				}
				processedItem := item
				if forceDelete {
					processedItem.Operation = ScheduleEventOperationDelete
				} else if strings.TrimSpace(writeResult.EventID) != "" {
					processedItem.ProviderEventID = strings.TrimSpace(writeResult.EventID)
					event.EventID = processedItem.ProviderEventID
				}
				if err := outbox.MarkScheduleEventOutboxProcessed(ctx, processedItem, ScheduleEventSyncHash(event)); err != nil {
					return err
				}
			}
		}
	})
	return errors.Join(lockErr, dispatchErr)
}

func terminallyFinalizeScheduleCleanup(ctx context.Context, outbox ScheduleEventOutboxStore, userID uuid.UUID, provider Provider, message string) error {
	for {
		items, err := outbox.ListPendingScheduleEventOutbox(ctx, userID, provider, 100)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return outbox.DeleteCleanupPendingConnectionIfDrained(ctx, userID, provider)
		}
		for _, item := range items {
			if err := outbox.MarkScheduleEventOutboxFailed(ctx, item, message, true); err != nil {
				return err
			}
		}
	}
}

func isPermanentCalendarWriteError(err error) bool {
	var graphErr *MicrosoftGraphError
	if errors.As(err, &graphErr) {
		if graphErr.StatusCode < 400 || graphErr.StatusCode >= 500 {
			return false
		}
		switch graphErr.StatusCode {
		case http.StatusRequestTimeout, http.StatusConflict, http.StatusTooManyRequests:
			return false
		default:
			return true
		}
	}
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.Code < 400 || apiErr.Code >= 500 {
		return false
	}
	switch apiErr.Code {
	case 408, 409, 429:
		return false
	}
	for _, item := range apiErr.Errors {
		switch item.Reason {
		case "rateLimitExceeded", "userRateLimitExceeded", "backendError":
			return false
		}
	}
	return true
}

func (s *Service) DispatchReadyScheduleEventOutbox(ctx context.Context) (int, error) {
	if s.repo == nil {
		return 0, ErrCalendarNotConfigured
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return 0, err
	}
	userIDs, err := scheduleRepo.ListReadyScheduleEventOutboxUsers(ctx, 100)
	if err != nil {
		return 0, err
	}
	var dispatchErr error
	for _, userID := range userIDs {
		if err := s.DispatchScheduleEventOutbox(ctx, userID); err != nil {
			dispatchErr = errors.Join(dispatchErr, err)
		}
	}
	return len(userIDs), dispatchErr
}

func scheduleOutboxIDs(items []CoreScheduleEventOutbox) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.ID != uuid.Nil {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

func (s *Service) scheduleReconciliationRepository() (ScheduleReconciliationRepository, error) {
	repo, ok := s.repo.(ScheduleReconciliationRepository)
	if !ok {
		return nil, ErrCalendarNotConfigured
	}
	return repo, nil
}
