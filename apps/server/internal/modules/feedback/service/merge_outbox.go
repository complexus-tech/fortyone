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

type mergeSnapshot struct {
	SchemaVersion     int         `json:"schemaVersion"`
	MergeEventID      uuid.UUID   `json:"mergeEventId"`
	WorkspaceID       uuid.UUID   `json:"workspaceId"`
	PortalID          uuid.UUID   `json:"portalId"`
	SourceItemID      uuid.UUID   `json:"sourceItemId"`
	TargetItemID      uuid.UUID   `json:"targetItemId"`
	MergedByUserID    uuid.UUID   `json:"mergedByUserId"`
	MergedAt          time.Time   `json:"mergedAt"`
	SourceTitle       string      `json:"sourceTitle"`
	SourceSlug        string      `json:"sourceSlug"`
	TargetTitle       string      `json:"targetTitle"`
	TargetSlug        string      `json:"targetSlug"`
	SourceFollowerIDs []uuid.UUID `json:"sourceFollowerIds"`
}

type permanentMergeError struct {
	err error
}

func (e permanentMergeError) Error() string { return e.err.Error() }
func (e permanentMergeError) Unwrap() error { return e.err }

const (
	mergeOutboxBatchSize    = 50
	mergeOutboxStaleAfter   = 10 * time.Minute
	mergeOutboxMaxAttempts  = 8
	mergeOutboxInitialDelay = 30 * time.Second
	mergeOutboxMaxDelay     = 15 * time.Minute
)

func (s *Service) dispatchReadyMergeEvents(ctx context.Context) error {
	if s.nextRepo == nil || s.publisher == nil || s.tasks == nil || s.security == nil || strings.TrimSpace(s.websiteURL) == "" {
		return ErrFeatureUnavailable
	}
	merges, err := s.nextRepo.ClaimMergeOutboxEvents(ctx, mergeOutboxBatchSize, mergeOutboxStaleAfter)
	if err != nil {
		return fmt.Errorf("claim feedback item merge outbox: %w", err)
	}

	var lifecycleErrors error
	for _, merge := range merges {
		dispatchErr := s.dispatchMergeEvent(ctx, merge)
		if dispatchErr == nil {
			if err := s.nextRepo.CompleteMergeOutboxEvent(ctx, merge.EventID, merge.ClaimToken); err != nil {
				// Completion is an uncertain write. Let stale-claim recovery replay
				// the recipient-deduped effects rather than overwriting its state.
				lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("complete feedback merge %s: %w", merge.EventID, err))
			}
			continue
		}

		terminal := merge.AttemptCount >= mergeOutboxMaxAttempts
		var permanent permanentMergeError
		if errors.As(dispatchErr, &permanent) {
			terminal = true
		}
		retryAt := s.now().UTC().Add(mergeRetryDelay(merge.AttemptCount))
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		releaseErr := s.nextRepo.RetryMergeOutboxEvent(
			releaseCtx,
			merge.EventID,
			merge.ClaimToken,
			dispatchErr.Error(),
			retryAt,
			terminal,
		)
		cancel()
		if releaseErr != nil {
			lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("release feedback merge %s after %v: %w", merge.EventID, dispatchErr, releaseErr))
		} else {
			s.logNextPhaseError(ctx, "dispatch feedback item merge", dispatchErr)
		}
	}
	return lifecycleErrors
}

func (s *Service) dispatchMergeEvent(ctx context.Context, merge CoreMergeOutboxEvent) error {
	snapshot, err := parseMergeSnapshot(merge)
	if err != nil {
		return err
	}
	if merge.EventID == uuid.Nil || merge.WorkspaceID == uuid.Nil || merge.PortalID == uuid.Nil ||
		merge.SourceItemID == uuid.Nil || merge.TargetItemID == uuid.Nil || merge.ActorID == uuid.Nil ||
		merge.ClaimToken == uuid.Nil || merge.MergedAt.IsZero() {
		return errors.New("merge outbox envelope is incomplete")
	}
	recipients, err := s.nextRepo.ListMergeRecipients(ctx, merge.PortalID, merge.TargetItemID, snapshot.SourceFollowerIDs)
	if err != nil {
		return fmt.Errorf("list merge recipients: %w", err)
	}
	portal, err := s.repo.GetPortal(ctx, merge.WorkspaceID, merge.PortalID)
	if err != nil {
		return fmt.Errorf("load merge portal: %w", err)
	}
	destinationURL := fmt.Sprintf(
		"%s/portal/%s/feedback/%s",
		strings.TrimRight(s.websiteURL, "/"),
		url.PathEscape(portal.Slug),
		url.PathEscape(snapshot.TargetSlug),
	)
	effectiveActorID := merge.ActorID
	for _, recipient := range recipients {
		if recipient.Kind != ContributorKindAccount || recipient.UserID == uuid.Nil || recipient.UserID == merge.ActorID {
			continue
		}
		resolver, ok := s.repo.(NotificationActorResolver)
		if !ok {
			return ErrFeatureUnavailable
		}
		effectiveActorID, err = resolver.ResolveNotificationActor(ctx, merge.ActorID, s.guestNotificationActorID)
		if err != nil {
			return fmt.Errorf("resolve merge notification actor: %w", err)
		}
		break
	}

	seenAccounts := make(map[uuid.UUID]struct{}, len(recipients))
	seenContributors := make(map[uuid.UUID]struct{}, len(recipients))
	for _, recipient := range recipients {
		switch recipient.Kind {
		case ContributorKindAccount:
			if recipient.UserID == uuid.Nil || recipient.UserID == merge.ActorID || recipient.UserID == effectiveActorID {
				continue
			}
			if _, duplicate := seenAccounts[recipient.UserID]; duplicate {
				continue
			}
			seenAccounts[recipient.UserID] = struct{}{}
			if err := s.publisher.Publish(ctx, events.Event{
				Type: events.FeedbackItemMerged,
				Payload: events.FeedbackItemMergedPayload{
					MergeEventID: merge.EventID, SourceItemID: merge.SourceItemID,
					TargetItemID: merge.TargetItemID, TargetItemTitle: snapshot.TargetTitle,
					TargetItemSlug: snapshot.TargetSlug, WorkspaceID: merge.WorkspaceID,
					RecipientID: recipient.UserID,
				},
				Timestamp: merge.MergedAt.UTC(),
				ActorID:   effectiveActorID,
			}); err != nil {
				return fmt.Errorf("publish merge notification for account %s: %w", recipient.UserID, err)
			}
		case ContributorKindVerifiedGuest, ContributorKindExternal:
			if recipient.ContributorID == uuid.Nil {
				continue
			}
			if _, duplicate := seenContributors[recipient.ContributorID]; duplicate {
				continue
			}
			seenContributors[recipient.ContributorID] = struct{}{}
			if err := s.persistAndEnqueueMergeDelivery(ctx, merge, snapshot, recipient.ContributorID, destinationURL); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) persistAndEnqueueMergeDelivery(ctx context.Context, merge CoreMergeOutboxEvent, snapshot mergeSnapshot, contributorID uuid.UUID, destinationURL string) error {
	deliveryID := stableMergeDeliveryID(merge.EventID, contributorID)
	_, unsubscribeHash, err := feedbacksecurity.DeriveUnsubscribeTokenWithKey(s.security.unsubscribeKey[:], deliveryID)
	if err != nil {
		return fmt.Errorf("derive merge unsubscribe token: %w", err)
	}
	targetID := merge.TargetItemID
	delivery, _, err := s.nextRepo.CreateContributorDelivery(ctx, CoreCreateDeliveryInput{
		DeliveryID: deliveryID, PortalID: merge.PortalID, ContributorID: contributorID,
		ItemID: &targetID, EventType: "feedback.item.merged",
		DedupeKey:      "item-merge:" + merge.EventID.String(),
		Subject:        "Feedback merged into " + snapshot.TargetTitle,
		Message:        "Your feedback is now tracked in " + snapshot.TargetTitle + ".",
		DestinationURL: destinationURL, TokenHash: unsubscribeHash,
	})
	if err != nil {
		return fmt.Errorf("persist merge delivery for contributor %s: %w", contributorID, err)
	}
	// A missing delivery is an intentional suppression because the contributor
	// was blocked, removed, or opted out after the immutable audience snapshot.
	if delivery.ID == uuid.Nil {
		return nil
	}
	if err := s.tasks.EnqueueFeedbackContributorDelivery(tasks.FeedbackContributorDeliveryPayload{
		DeliveryID: delivery.ID,
	}); err != nil {
		return fmt.Errorf("enqueue merge delivery %s: %w", delivery.ID, err)
	}
	return nil
}

func parseMergeSnapshot(merge CoreMergeOutboxEvent) (mergeSnapshot, error) {
	var snapshot mergeSnapshot
	if err := json.Unmarshal(merge.Payload, &snapshot); err != nil {
		return mergeSnapshot{}, permanentMergeError{err: fmt.Errorf("decode merge snapshot: %w", err)}
	}
	if snapshot.SchemaVersion != 1 || snapshot.MergeEventID != merge.EventID ||
		snapshot.WorkspaceID != merge.WorkspaceID || snapshot.PortalID != merge.PortalID ||
		snapshot.SourceItemID != merge.SourceItemID || snapshot.TargetItemID != merge.TargetItemID ||
		snapshot.MergedByUserID != merge.ActorID || !snapshot.MergedAt.Equal(merge.MergedAt) ||
		strings.TrimSpace(snapshot.TargetTitle) == "" || strings.TrimSpace(snapshot.TargetSlug) == "" {
		return mergeSnapshot{}, permanentMergeError{err: errors.New("merge snapshot does not match its outbox envelope")}
	}
	seen := make(map[uuid.UUID]struct{}, len(snapshot.SourceFollowerIDs))
	for _, contributorID := range snapshot.SourceFollowerIDs {
		if contributorID == uuid.Nil {
			return mergeSnapshot{}, permanentMergeError{err: errors.New("merge snapshot contains an empty follower id")}
		}
		if _, duplicate := seen[contributorID]; duplicate {
			return mergeSnapshot{}, permanentMergeError{err: errors.New("merge snapshot contains duplicate follower ids")}
		}
		seen[contributorID] = struct{}{}
	}
	return snapshot, nil
}

func stableMergeDeliveryID(eventID, contributorID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("feedback-item-merge:v1:"+eventID.String()+":"+contributorID.String()))
}

func mergeRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 20 {
		shift = 20
	}
	delay := mergeOutboxInitialDelay * time.Duration(1<<shift)
	if delay > mergeOutboxMaxDelay {
		return mergeOutboxMaxDelay
	}
	return delay
}
