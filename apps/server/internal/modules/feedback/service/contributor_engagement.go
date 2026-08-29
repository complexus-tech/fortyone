package feedback

import (
	"context"
	"fmt"
	"github.com/complexus-tech/projects-api/pkg/feedbacksecurity"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"net/url"
	"strings"
	"time"
)

func (s *Service) FollowItem(ctx context.Context, portalSlug string, itemID uuid.UUID, participant CoreParticipant, following bool) (CoreFollowState, error) {
	portal, item, err := s.publicPortalItem(ctx, portalSlug, itemID)
	if err != nil {
		return CoreFollowState{}, err
	}
	if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.Kind == ContributorKindAnonymous {
		return CoreFollowState{}, ErrAuthenticationRequired
	}
	return s.nextRepo.SetItemFollow(ctx, item.ID, participant.ID, following)
}
func (s *Service) GetItemFollow(ctx context.Context, portalSlug string, itemID uuid.UUID, participant CoreParticipant) (CoreFollowState, error) {
	portal, item, err := s.publicPortalItem(ctx, portalSlug, itemID)
	if err != nil {
		return CoreFollowState{}, err
	}
	if participant.PortalID != portal.ID || participant.ID == uuid.Nil {
		return CoreFollowState{}, ErrAuthenticationRequired
	}
	return s.nextRepo.GetItemFollow(ctx, item.ID, participant.ID)
}
func (s *Service) GetContributorPreferences(ctx context.Context, portalSlug string, participant CoreParticipant) (CoreContributorPreferences, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorPreferences{}, err
	}
	if participant.PortalID != portal.ID {
		return CoreContributorPreferences{}, ErrAuthenticationRequired
	}
	return s.nextRepo.GetContributorPreferences(ctx, portal.ID, participant.ID)
}
func (s *Service) SetPortalEmailPreference(ctx context.Context, portalSlug string, participant CoreParticipant, enabled bool) (CoreContributorPreferences, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorPreferences{}, err
	}
	if participant.PortalID != portal.ID {
		return CoreContributorPreferences{}, ErrAuthenticationRequired
	}
	return s.nextRepo.SetPortalEmailPreference(ctx, portal.ID, participant.ID, enabled)
}
func (s *Service) GetUnreadUpdateCount(ctx context.Context, portalSlug string, participant CoreParticipant) (int, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return 0, err
	}
	if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.Kind == ContributorKindAnonymous {
		return 0, ErrAuthenticationRequired
	}
	return s.nextRepo.GetUnreadUpdateCount(ctx, portal.ID, participant.ID)
}
func (s *Service) MarkUpdatesSeen(ctx context.Context, portalSlug string, participant CoreParticipant) (time.Time, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return time.Time{}, err
	}
	if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.Kind == ContributorKindAnonymous {
		return time.Time{}, ErrAuthenticationRequired
	}
	return s.nextRepo.MarkUpdatesSeen(ctx, portal.ID, participant.ID)
}
func (s *Service) enqueueDeliveries(ctx context.Context, portalID uuid.UUID, itemID, updateID *uuid.UUID, actorContributorID uuid.UUID, eventType, eventKey, subject, message, destinationURL string) {
	if s.nextRepo == nil || s.tasks == nil {
		return
	}
	recipients, err := s.nextRepo.ListDeliveryRecipients(ctx, portalID, itemID, updateID, actorContributorID)
	if err != nil {
		s.logNextPhaseError(ctx, "list feedback delivery recipients", err)
		return
	}
	for _, recipient := range recipients {
		if recipient.Kind != ContributorKindVerifiedGuest && recipient.Kind != ContributorKindExternal {
			continue
		}
		deliveryID := uuid.New()
		_, unsubscribeHash, err := feedbacksecurity.DeriveUnsubscribeTokenWithKey(s.security.unsubscribeKey[:], deliveryID)
		if err != nil {
			s.logNextPhaseError(ctx, "generate feedback unsubscribe token", err)
			continue
		}
		dedupeKey := eventKey + ":" + recipient.ContributorID.String()
		delivery, created, err := s.nextRepo.CreateContributorDelivery(ctx, CoreCreateDeliveryInput{DeliveryID: deliveryID, PortalID: portalID, ContributorID: recipient.ContributorID, ItemID: itemID, UpdateID: updateID, EventType: eventType, DedupeKey: dedupeKey, Subject: subject, Message: message, DestinationURL: destinationURL, TokenHash: unsubscribeHash})
		if err != nil {
			s.logNextPhaseError(ctx, "create feedback contributor delivery", err)
			continue
		}
		if !created {
			continue
		}
		if err := s.tasks.EnqueueFeedbackContributorDelivery(tasks.FeedbackContributorDeliveryPayload{DeliveryID: delivery.ID}); err != nil {
			s.logNextPhaseError(ctx, "enqueue feedback contributor delivery", err)
		}
	}
}
func (s *Service) enqueueStatusDeliveries(ctx context.Context, item CoreItem, actorID uuid.UUID, transitionKey string) {
	if s.nextRepo == nil {
		return
	}
	actorContributorID := uuid.Nil
	if actorID != uuid.Nil {
		if participant, err := s.nextRepo.GetParticipantByUser(ctx, item.PortalID, actorID); err == nil {
			actorContributorID = participant.ID
		}
	}
	portal, err := s.repo.GetPortal(ctx, item.WorkspaceID, item.PortalID)
	if err != nil {
		s.logNextPhaseError(ctx, "load feedback portal for status delivery", err)
		return
	}
	if transitionKey == "" {
		transitionKey = item.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	s.enqueueDeliveries(ctx, item.PortalID, &item.ID, nil, actorContributorID, "feedback.status.updated", "status:"+item.ID.String()+":"+item.Status+":"+transitionKey, item.Title+" is now "+strings.ReplaceAll(item.Status, "_", " "), item.RoadmapSummaryText(), s.publicItemURL(portal.Slug, item.Slug))
}

func (s *Service) NotifyLinkedStoryStatusTransition(ctx context.Context, workspaceID, storyID, actorID uuid.UUID, occurredAt time.Time) error {
	if s.nextRepo == nil {
		return ErrFeatureUnavailable
	}
	if workspaceID == uuid.Nil || storyID == uuid.Nil || actorID == uuid.Nil || occurredAt.IsZero() {
		return invalidInput("workspace, story, actor, and transition time are required")
	}
	items, err := s.nextRepo.ListPrimaryStoryItems(ctx, workspaceID, storyID)
	if err != nil {
		return err
	}
	transitionKey := storyID.String() + ":" + occurredAt.UTC().Format(time.RFC3339Nano)
	for _, item := range // NotifyLinkedStoryStatusTransition bridges projected story workflow changes
	// into the same account and contributor delivery paths as direct status edits.
	// The event id and delivery key are stable across stream redelivery.
	items {
		if !shouldNotifyFeedbackStatusTransition(item.Status, item.RoadmapSummary) {
			continue
		}
		eventID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(transitionKey+":"+item.ID.String()+":"+item.Status))
		s.publishAccountStatusNotifications(ctx, item, actorID, eventID, occurredAt.UTC())
		s.enqueueStatusDeliveries(ctx, item, actorID, eventID.String())
	}
	return nil
}

func shouldNotifyFeedbackStatusTransition(status string, publicExplanation *string) bool {
	switch status {
	case StatusPlanned, StatusInProgress, StatusCompleted:
		return true
	case StatusClosed:
		return publicExplanation != nil && strings.TrimSpace(*publicExplanation) != ""
	default:
		return false
	}
}
func (s *Service) publicPortalItem(ctx context.Context, portalSlug string, itemID uuid.UUID) (CorePortal, CoreItem, error) {
	if itemID == uuid.Nil {
		return CorePortal{}, CoreItem{}, invalidInput("feedback id is required")
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CorePortal{}, CoreItem{}, err
	}
	item, err := s.repo.GetItemByPortal(ctx, portal.ID, itemID)
	if err != nil {
		return CorePortal{}, CoreItem{}, err
	}
	if item.WorkspaceID != portal.WorkspaceID {
		return CorePortal{}, CoreItem{}, ErrNotFound
	}
	if item.MergedIntoItemID != nil {
		return CorePortal{}, CoreItem{}, ErrMergeConflict
	}
	return portal, item, nil
}
func (s *Service) publicItemURL(portalSlug, itemSlug string) string {
	return fmt.Sprintf("%s/portal/%s/feedback/%s", s.websiteURL, url.PathEscape(portalSlug), url.PathEscape(itemSlug))
}
func (s *Service) logNextPhaseError(ctx context.Context, operation string, err error) {
	if s.log != nil {
		s.log.Error(ctx, operation, "error", err)
	}
}

// shouldNotifyFeedbackStatusTransition is the single public-notification
// policy for feedback workflow changes. Internal triage is silent. Closing an
// item is public only when the transition supplies an explanation; callers
// without transition input (for example linked-story projection) pass the
// item's current public roadmap summary.
