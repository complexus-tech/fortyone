package feedback

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func (s *Service) CreateComment(ctx context.Context, input CoreCommentInput) (CoreComment, error) {
	if input.WorkspaceID == uuid.Nil || input.ItemID == uuid.Nil || input.AuthorID == uuid.Nil {
		return CoreComment{}, invalidInput("workspace, feedback, and author are required")
	}
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" {
		return CoreComment{}, invalidInput("comment body is required")
	}
	item, err := s.getInternalItem(ctx, input.WorkspaceID, input.ItemID, input.AuthorID)
	if err != nil {
		return CoreComment{}, err
	}
	if item.MergedIntoItemID != nil {
		return CoreComment{}, ErrMergeConflict
	}
	return s.createComment(ctx, input, item)
}

func (s *Service) CreatePublicComment(ctx context.Context, input CorePublicCommentInput) (CoreComment, error) {
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" {
		return CoreComment{}, invalidInput("comment body is required")
	}
	if utf8.RuneCountInString(input.Body) > maxPublicFeedbackCommentCharacters {
		return CoreComment{}, invalidInputf("comment body must be %d characters or fewer", maxPublicFeedbackCommentCharacters)
	}

	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(input.PortalSlug))
	if err != nil {
		return CoreComment{}, err
	}
	item, err := s.repo.GetItemByPortal(ctx, portal.ID, input.ItemID)
	if err != nil {
		return CoreComment{}, err
	}
	if item.WorkspaceID != portal.WorkspaceID {
		return CoreComment{}, ErrNotFound
	}
	if item.MergedIntoItemID != nil {
		return CoreComment{}, ErrMergeConflict
	}
	if input.Participant != nil {
		participant := *input.Participant
		if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.BlockedAt != nil || participant.Kind == ContributorKindAnonymous {
			return CoreComment{}, ErrAuthenticationRequired
		}
		if portal.ParticipationMode == ParticipationModeAccountRequired {
			return CoreComment{}, ErrParticipationNotAllowed
		}
		var parent *CoreComment
		if input.ParentID != nil {
			parentComment, err := s.getInternalComment(ctx, portal.WorkspaceID, item.ID, *input.ParentID, input.AuthorID)
			if err != nil {
				return CoreComment{}, err
			}
			if parentComment.ParentID != nil {
				return CoreComment{}, invalidInput("replies can only be one level deep")
			}
			parent = &parentComment
		}
		comment, err := s.nextRepo.CreateContributorComment(ctx, CoreContributorCommentInput{
			WorkspaceID: portal.WorkspaceID,
			PortalID:    portal.ID,
			ItemID:      item.ID,
			Participant: participant,
			ParentID:    input.ParentID,
			Body:        input.Body,
		})
		if err != nil {
			return CoreComment{}, err
		}
		s.publishGuestCommentAccountNotifications(ctx, item, parent, comment, participant)
		s.enqueueDeliveries(ctx, portal.ID, &item.ID, nil, participant.ID, "feedback.comment.created", "comment:"+comment.ID.String(), "New reply on "+item.Title, comment.Body, s.publicItemURL(portal.Slug, item.Slug))
		return comment, nil
	}
	if input.AuthorID == uuid.Nil {
		return CoreComment{}, ErrAuthenticationRequired
	}

	return s.createComment(ctx, CoreCommentInput{
		WorkspaceID: portal.WorkspaceID,
		ItemID:      item.ID,
		AuthorID:    input.AuthorID,
		ParentID:    input.ParentID,
		Body:        input.Body,
	}, item)
}

func (s *Service) createComment(ctx context.Context, input CoreCommentInput, item CoreItem) (CoreComment, error) {
	var parent *CoreComment
	if input.ParentID != nil {
		if *input.ParentID == uuid.Nil {
			return CoreComment{}, invalidInput("parent comment is required")
		}
		parentComment, err := s.getInternalComment(ctx, input.WorkspaceID, input.ItemID, *input.ParentID, input.AuthorID)
		if err != nil {
			return CoreComment{}, err
		}
		if parentComment.ParentID != nil {
			return CoreComment{}, invalidInput("replies can only be one level deep")
		}
		parent = &parentComment
	}

	var comment CoreComment
	scope, scoped, err := s.coreScope(ctx, input.WorkspaceID, input.AuthorID)
	if err != nil {
		return CoreComment{}, err
	}
	if scoped {
		comment, err = s.scopedCoreRepo.CreateCommentScoped(ctx, scope, input)
	} else {
		comment, err = s.repo.CreateComment(ctx, input)
	}
	if err != nil {
		return CoreComment{}, err
	}

	recipients := map[uuid.UUID]struct{}{}
	if shouldNotify(item.AuthorID, input.AuthorID) {
		recipients[item.AuthorID] = struct{}{}
	}
	if parent != nil && shouldNotify(parent.AuthorID, input.AuthorID) {
		recipients[parent.AuthorID] = struct{}{}
	}
	s.addAccountItemFollowers(ctx, item, input.AuthorID, recipients)
	for recipientID := range recipients {
		s.publish(ctx, events.Event{
			Type: events.FeedbackCommentCreated,
			Payload: events.FeedbackCommentCreatedPayload{
				CommentID:     comment.ID,
				FeedbackID:    item.ID,
				FeedbackTitle: item.Title,
				FeedbackSlug:  item.Slug,
				WorkspaceID:   item.WorkspaceID,
				RecipientID:   recipientID,
				Content:       comment.Body,
				IsReply:       parent != nil && recipientID == parent.AuthorID,
			},
			Timestamp: time.Now(),
			ActorID:   input.AuthorID,
		})
	}
	if s.nextRepo != nil {
		participant, participantErr := s.nextRepo.GetParticipantByUser(ctx, item.PortalID, input.AuthorID)
		if participantErr == nil {
			portal, portalErr := s.repo.GetPortal(ctx, item.WorkspaceID, item.PortalID)
			if portalErr == nil {
				s.enqueueDeliveries(ctx, item.PortalID, &item.ID, nil, participant.ID, "feedback.comment.created", "comment:"+comment.ID.String(), "New reply on "+item.Title, comment.Body, s.publicItemURL(portal.Slug, item.Slug))
			}
		}
	}
	return comment, nil
}

func (s *Service) publishGuestCommentAccountNotifications(ctx context.Context, item CoreItem, parent *CoreComment, comment CoreComment, participant CoreParticipant) {
	recipients := make(map[uuid.UUID]bool)
	if item.AuthorID != uuid.Nil {
		recipients[item.AuthorID] = false
	}
	if parent != nil && parent.AuthorID != uuid.Nil {
		recipients[parent.AuthorID] = true
	}
	followers := make(map[uuid.UUID]struct{})
	s.addAccountItemFollowers(ctx, item, uuid.Nil, followers)
	for recipientID := range followers {
		if _, exists := recipients[recipientID]; !exists {
			recipients[recipientID] = false
		}
	}
	actorName := strings.TrimSpace(participant.DisplayName)
	if participant.PublicMasked || actorName == "" {
		actorName = "Anonymous"
	}
	for recipientID, isReply := range recipients {
		s.publish(ctx, events.Event{
			Type: events.FeedbackCommentCreated,
			Payload: events.FeedbackCommentCreatedPayload{
				CommentID: comment.ID, FeedbackID: item.ID, FeedbackTitle: item.Title, FeedbackSlug: item.Slug,
				WorkspaceID: item.WorkspaceID, RecipientID: recipientID, ActorContributorID: participant.ID,
				ActorName: actorName, Content: comment.Body, IsReply: isReply,
			},
			Timestamp: s.now().UTC(),
			ActorID:   s.guestNotificationActorID,
		})
	}
}

func (s *Service) addAccountItemFollowers(ctx context.Context, item CoreItem, actorID uuid.UUID, recipients map[uuid.UUID]struct{}) {
	if s.nextRepo == nil {
		return
	}
	followers, err := s.nextRepo.ListAccountItemFollowers(ctx, item.PortalID, item.ID)
	if err != nil {
		s.logNextPhaseError(ctx, "list account feedback item followers", err)
		return
	}
	for _, recipientID := range followers {
		if shouldNotify(recipientID, actorID) || (actorID == uuid.Nil && recipientID != uuid.Nil) {
			recipients[recipientID] = struct{}{}
		}
	}
}

func (s *Service) publishAccountStatusNotifications(ctx context.Context, item CoreItem, actorID, eventID uuid.UUID, occurredAt time.Time) {
	recipients := make(map[uuid.UUID]struct{})
	if shouldNotify(item.AuthorID, actorID) {
		recipients[item.AuthorID] = struct{}{}
	}
	s.addAccountItemFollowers(ctx, item, actorID, recipients)
	for recipientID := range recipients {
		s.publish(ctx, events.Event{
			Type: events.FeedbackStatusUpdated,
			Payload: events.FeedbackStatusUpdatedPayload{
				EventID: eventID, FeedbackID: item.ID, FeedbackTitle: item.Title, FeedbackSlug: item.Slug,
				WorkspaceID: item.WorkspaceID, RecipientID: recipientID, Status: item.Status,
			},
			Timestamp: occurredAt,
			ActorID:   actorID,
		})
	}
}

func (s *Service) publish(ctx context.Context, event events.Event) {
	if s.publisher == nil {
		return
	}
	if err := s.publisher.Publish(context.WithoutCancel(ctx), event); err != nil && s.log != nil {
		s.log.Error(ctx, "failed to publish feedback notification event", "event_type", event.Type, "error", err)
	}
}

func shouldNotify(recipientID, actorID uuid.UUID) bool {
	return recipientID != uuid.Nil && actorID != uuid.Nil && recipientID != actorID
}
