package feedback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/feedbacksecurity"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

const (
	publicationOutboxBatchSize    = 50
	publicationOutboxStaleAfter   = 10 * time.Minute
	publicationOutboxMaxAttempts  = 8
	publicationOutboxInitialDelay = 30 * time.Second
	publicationOutboxMaxDelay     = 15 * time.Minute
)

type publicationSnapshot struct {
	SchemaVersion       int                      `json:"schemaVersion"`
	PublicationEventID  uuid.UUID                `json:"publicationEventId"`
	UpdateID            uuid.UUID                `json:"updateId"`
	WorkspaceID         uuid.UUID                `json:"workspaceId"`
	PortalID            uuid.UUID                `json:"portalId"`
	PortalSlug          string                   `json:"portalSlug"`
	PublishedByUserID   uuid.UUID                `json:"publishedByUserId"`
	PublicationSequence int64                    `json:"publicationSequence"`
	PublishedAt         time.Time                `json:"publishedAt"`
	Slug                string                   `json:"slug"`
	Title               string                   `json:"title"`
	Summary             *string                  `json:"summary,omitempty"`
	Body                string                   `json:"body"`
	CoverImageURL       *string                  `json:"coverImageUrl,omitempty"`
	LinkedItems         []CoreUpdateItem         `json:"linkedItems"`
	ContributorAudience []uuid.UUID              `json:"contributorAudience"`
	AccountAudience     []publicationAccountPair `json:"accountAudience"`
}

type publicationAccountPair struct {
	UserID uuid.UUID `json:"userId"`
	ItemID uuid.UUID `json:"itemId"`
}

type permanentPublicationError struct {
	err error
}

func (e permanentPublicationError) Error() string { return e.err.Error() }
func (e permanentPublicationError) Unwrap() error { return e.err }

// DispatchReadyOutboxEvents is the shared worker entry point for durable
// feedback outboxes. Additional feedback aggregate outboxes can be composed
// here without adding another scheduler or task handler.
func (s *Service) DispatchReadyOutboxEvents(ctx context.Context) error {
	// Attempt both aggregate outboxes on every tick. A transient failure in one
	// must not starve ready work in the other.
	return errors.Join(
		s.dispatchReadyPublicationEvents(ctx),
		s.dispatchReadyMergeEvents(ctx),
	)
}

func (s *Service) dispatchReadyPublicationEvents(ctx context.Context) error {
	repository, ok := s.repo.(PublicationOutboxRepository)
	if !ok || s.publisher == nil || s.tasks == nil || s.security == nil || strings.TrimSpace(s.websiteURL) == "" {
		return ErrFeatureUnavailable
	}
	eventsToDispatch, err := repository.ClaimPublicationOutboxEvents(ctx, publicationOutboxBatchSize, publicationOutboxStaleAfter)
	if err != nil {
		return fmt.Errorf("claim feedback Update publication outbox: %w", err)
	}

	var lifecycleErrors error
	for _, publication := range eventsToDispatch {
		dispatchErr := s.dispatchPublicationEvent(ctx, repository, publication)
		if dispatchErr == nil {
			if err := repository.CompletePublicationOutboxEvent(ctx, publication.EventID, publication.ClaimToken); err != nil {
				// Completion is an uncertain write: do not overwrite it with a
				// retry transition. A stale-claim recovery will safely replay the
				// deduped side effects if the completion did not commit.
				lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("complete feedback publication %s: %w", publication.EventID, err))
			}
			continue
		}

		terminal := publication.AttemptCount >= publicationOutboxMaxAttempts
		var permanent permanentPublicationError
		if errors.As(dispatchErr, &permanent) {
			terminal = true
		}
		retryAt := s.now().UTC().Add(publicationRetryDelay(publication.AttemptCount))
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		releaseErr := repository.RetryPublicationOutboxEvent(
			releaseCtx,
			publication.EventID,
			publication.ClaimToken,
			dispatchErr.Error(),
			retryAt,
			terminal,
		)
		cancel()
		if releaseErr != nil {
			lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("release feedback publication %s after %v: %w", publication.EventID, dispatchErr, releaseErr))
		} else {
			s.logNextPhaseError(ctx, "dispatch feedback Update publication", dispatchErr)
		}
	}
	return lifecycleErrors
}

func (s *Service) dispatchPublicationEvent(ctx context.Context, repository PublicationOutboxRepository, publication CorePublicationOutboxEvent) error {
	snapshot, err := parsePublicationSnapshot(publication)
	if err != nil {
		return err
	}
	itemIDs := make([]uuid.UUID, 0, len(snapshot.LinkedItems))
	for _, item := range snapshot.LinkedItems {
		itemIDs = append(itemIDs, item.ID)
	}
	deliveryRecipients, err := repository.ListPublicationDeliveryRecipients(
		ctx,
		publication.PortalID,
		snapshot.ContributorAudience,
		itemIDs,
	)
	if err != nil {
		return fmt.Errorf("list contributor recipients: %w", err)
	}
	accountAudience := make([]CoreAccountUpdateRecipient, 0, len(snapshot.AccountAudience))
	for _, recipient := range snapshot.AccountAudience {
		accountAudience = append(accountAudience, CoreAccountUpdateRecipient{
			UserID: recipient.UserID,
			ItemID: recipient.ItemID,
		})
	}
	accountRecipients, err := repository.ListAccountPublicationRecipients(ctx, publication.PortalID, accountAudience)
	if err != nil {
		return fmt.Errorf("list account recipients: %w", err)
	}
	effectiveActorID := publication.ActorID
	if len(accountRecipients) > 0 {
		effectiveActorID, err = repository.ResolveNotificationActor(ctx, publication.ActorID, s.guestNotificationActorID)
		if err != nil {
			return fmt.Errorf("resolve publication notification actor: %w", err)
		}
	}

	destinationURL := fmt.Sprintf(
		"%s/portal/%s/updates/%s",
		strings.TrimRight(s.websiteURL, "/"),
		url.PathEscape(snapshot.PortalSlug),
		url.PathEscape(snapshot.Slug),
	)
	message := snapshot.Title
	if snapshot.Summary != nil && strings.TrimSpace(*snapshot.Summary) != "" {
		message = strings.TrimSpace(*snapshot.Summary)
	}
	eventKeyPrefix := "update-publication:" + publication.EventID.String()
	var linkedItemID *uuid.UUID
	if len(itemIDs) == 1 {
		itemID := itemIDs[0]
		linkedItemID = &itemID
	}
	updateID := publication.UpdateID
	for _, recipient := range deliveryRecipients {
		if recipient.ContributorID == uuid.Nil || (recipient.Kind != ContributorKindVerifiedGuest && recipient.Kind != ContributorKindExternal) {
			continue
		}
		deliveryID := stablePublicationDeliveryID(publication.EventID, recipient.ContributorID)
		_, unsubscribeHash, err := feedbacksecurity.DeriveUnsubscribeTokenWithKey(s.security.unsubscribeKey[:], deliveryID)
		if err != nil {
			return fmt.Errorf("derive contributor unsubscribe token: %w", err)
		}
		delivery, _, err := repository.CreateContributorDelivery(ctx, CoreCreateDeliveryInput{
			DeliveryID: deliveryID, PortalID: publication.PortalID,
			ContributorID: recipient.ContributorID, ItemID: linkedItemID, UpdateID: &updateID,
			EventType: "feedback.update.published", DedupeKey: eventKeyPrefix + ":" + recipient.ContributorID.String(),
			Subject: snapshot.Title, Message: message, DestinationURL: destinationURL,
			TokenHash: unsubscribeHash,
		})
		if err != nil {
			return fmt.Errorf("persist contributor delivery %s: %w", recipient.ContributorID, err)
		}
		// A nil delivery means the contributor became ineligible (blocked,
		// missing email, or unsubscribed) between audience selection and the
		// guarded insert. That is a completed suppression, not a retry.
		if delivery.ID == uuid.Nil {
			continue
		}
		if err := s.tasks.EnqueueFeedbackContributorDelivery(tasks.FeedbackContributorDeliveryPayload{
			DeliveryID: delivery.ID,
		}); err != nil {
			return fmt.Errorf("enqueue contributor delivery %s: %w", delivery.ID, err)
		}
	}

	seenAccounts := make(map[uuid.UUID]struct{}, len(accountRecipients))
	for _, recipient := range accountRecipients {
		if recipient.UserID == uuid.Nil || recipient.ItemID == uuid.Nil ||
			recipient.UserID == publication.ActorID || recipient.UserID == effectiveActorID {
			continue
		}
		if _, duplicate := seenAccounts[recipient.UserID]; duplicate {
			continue
		}
		seenAccounts[recipient.UserID] = struct{}{}
		if err := s.publisher.Publish(ctx, events.Event{
			Type: events.FeedbackUpdatePublished,
			Payload: events.FeedbackUpdatePublishedPayload{
				PublicationEventID: publication.EventID, PublicationSequence: publication.PublicationSequence,
				UpdateID: publication.UpdateID, LinkedItemID: recipient.ItemID,
				WorkspaceID: publication.WorkspaceID, RecipientID: recipient.UserID,
				UpdateTitle: snapshot.Title, UpdateSlug: snapshot.Slug,
			},
			Timestamp: publication.PublishedAt.UTC(),
			ActorID:   effectiveActorID,
		}); err != nil {
			return fmt.Errorf("publish account notification for %s: %w", recipient.UserID, err)
		}
	}
	return nil
}

func parsePublicationSnapshot(publication CorePublicationOutboxEvent) (publicationSnapshot, error) {
	var snapshot publicationSnapshot
	if err := json.Unmarshal(publication.Payload, &snapshot); err != nil {
		return publicationSnapshot{}, permanentPublicationError{err: fmt.Errorf("decode publication snapshot: %w", err)}
	}
	if snapshot.SchemaVersion != 1 || publication.EventID == uuid.Nil || publication.UpdateID == uuid.Nil ||
		publication.WorkspaceID == uuid.Nil || publication.PortalID == uuid.Nil || publication.ActorID == uuid.Nil ||
		publication.ClaimToken == uuid.Nil || publication.PublicationSequence <= 0 || publication.PublishedAt.IsZero() {
		return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication outbox envelope is incomplete")}
	}
	if snapshot.PublicationEventID != publication.EventID || snapshot.UpdateID != publication.UpdateID ||
		snapshot.WorkspaceID != publication.WorkspaceID || snapshot.PortalID != publication.PortalID ||
		snapshot.PublishedByUserID != publication.ActorID || snapshot.PublicationSequence != publication.PublicationSequence ||
		!snapshot.PublishedAt.Equal(publication.PublishedAt) {
		return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot does not match its outbox envelope")}
	}
	if strings.TrimSpace(snapshot.PortalSlug) == "" || strings.TrimSpace(snapshot.Slug) == "" || strings.TrimSpace(snapshot.Title) == "" {
		return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot is missing delivery content")}
	}
	if snapshot.LinkedItems == nil || snapshot.ContributorAudience == nil || snapshot.AccountAudience == nil {
		return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot is missing its immutable audience")}
	}
	seenItems := make(map[uuid.UUID]struct{}, len(snapshot.LinkedItems))
	for _, item := range snapshot.LinkedItems {
		if item.ID == uuid.Nil {
			return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot contains an empty linked item id")}
		}
		if _, duplicate := seenItems[item.ID]; duplicate {
			return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot contains duplicate linked items")}
		}
		seenItems[item.ID] = struct{}{}
	}
	seenContributors := make(map[uuid.UUID]struct{}, len(snapshot.ContributorAudience))
	for _, contributorID := range snapshot.ContributorAudience {
		if contributorID == uuid.Nil {
			return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot contains an empty contributor audience id")}
		}
		if _, duplicate := seenContributors[contributorID]; duplicate {
			return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot contains duplicate contributor audience ids")}
		}
		seenContributors[contributorID] = struct{}{}
	}
	seenAccounts := make(map[publicationAccountPair]struct{}, len(snapshot.AccountAudience))
	for _, recipient := range snapshot.AccountAudience {
		if recipient.UserID == uuid.Nil || recipient.ItemID == uuid.Nil {
			return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot contains an incomplete account audience entry")}
		}
		if _, linked := seenItems[recipient.ItemID]; !linked {
			return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot account audience references an unlinked item")}
		}
		if _, duplicate := seenAccounts[recipient]; duplicate {
			return publicationSnapshot{}, permanentPublicationError{err: errors.New("publication snapshot contains duplicate account audience entries")}
		}
		seenAccounts[recipient] = struct{}{}
	}
	return snapshot, nil
}

func stablePublicationDeliveryID(eventID, contributorID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("feedback-update-publication:v1:"+eventID.String()+":"+contributorID.String()))
}

func publicationRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 20 {
		shift = 20
	}
	delay := publicationOutboxInitialDelay * time.Duration(1<<shift)
	if delay > publicationOutboxMaxDelay {
		return publicationOutboxMaxDelay
	}
	return delay
}
