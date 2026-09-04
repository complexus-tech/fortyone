package feedback

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

func (s *Service) CreateItem(ctx context.Context, input CoreItemInput) (CoreItem, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.BoardID == uuid.Nil || (input.ContributorID == uuid.Nil && input.AuthorID == uuid.Nil) {
		return CoreItem{}, invalidInput("workspace, portal, board, and contributor are required")
	}
	if input.ContributorID == uuid.Nil {
		contributor, err := s.repo.GetOrCreateAccountContributor(ctx, input.PortalID, input.AuthorID)
		if err != nil {
			return CoreItem{}, err
		}
		input.ContributorID = contributor.ID
	}
	input.Source = strings.ToLower(strings.TrimSpace(input.Source))
	if input.Source == "" {
		input.Source = SubmissionSourceInternal
	}
	if !isValidSubmissionSource(input.Source) {
		return CoreItem{}, invalidInput("unsupported feedback submission source")
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.DescriptionHTML = strings.TrimSpace(input.DescriptionHTML)
	if input.Title == "" {
		return CoreItem{}, invalidInput("feedback title is required")
	}
	input.Slug = normalizeSlug(input.Slug)
	if input.Slug == publicRoadmapSlug {
		return CoreItem{}, invalidInput("feedback slug roadmap is reserved")
	}
	if input.Slug == "" {
		input.Slug = normalizeSlug(input.Title) + "-" + uuid.NewString()[:8]
	}
	return s.repo.CreateItem(ctx, input)
}

func (s *Service) CreatePublicItem(ctx context.Context, input CorePublicItemInput) (CorePublicItemResult, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.DescriptionHTML = strings.TrimSpace(input.DescriptionHTML)
	input.ParticipationIntent = strings.ToLower(strings.TrimSpace(input.ParticipationIntent))
	if input.Title == "" {
		return CorePublicItemResult{}, invalidInput("feedback title is required")
	}
	if utf8.RuneCountInString(input.Title) > maxPublicFeedbackTitleCharacters {
		return CorePublicItemResult{}, invalidInputf("feedback title must be %d characters or fewer", maxPublicFeedbackTitleCharacters)
	}
	if utf8.RuneCountInString(input.Description) > maxPublicFeedbackDescriptionCharacters {
		return CorePublicItemResult{}, invalidInputf("feedback description must be %d characters or fewer", maxPublicFeedbackDescriptionCharacters)
	}
	if utf8.RuneCountInString(input.DescriptionHTML) > maxPublicFeedbackDescriptionHTMLCharacters {
		return CorePublicItemResult{}, invalidInput("formatted feedback description is too large")
	}
	switch input.ParticipationIntent {
	case ParticipationIntentAccount:
		if input.AuthorID == uuid.Nil {
			return CorePublicItemResult{}, ErrAuthenticationRequired
		}
	case ParticipationIntentVerifiedGuest, ParticipationIntentExternal:
		if input.Participant == nil || input.Participant.ID == uuid.Nil || input.Participant.Kind != input.ParticipationIntent {
			return CorePublicItemResult{}, ErrAuthenticationRequired
		}
	case ParticipationIntentAnonymous:
	default:
		return CorePublicItemResult{}, invalidInput("unsupported feedback participation intent")
	}

	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(input.PortalSlug))
	if err != nil {
		return CorePublicItemResult{}, err
	}
	board, err := s.repo.GetBoard(ctx, portal.ID, input.BoardID)
	if err != nil {
		return CorePublicItemResult{}, err
	}
	if board.WorkspaceID != portal.WorkspaceID {
		return CorePublicItemResult{}, ErrNotFound
	}
	similarItems, err := s.repo.ListSimilarItems(ctx, portal.ID, input.Title, input.Description, 1)
	if err != nil {
		return CorePublicItemResult{}, err
	}
	if len(similarItems) > 0 && similarItems[0].Confidence >= duplicateItemConfidence {
		return CorePublicItemResult{}, fmt.Errorf("%w: %s", ErrDuplicateItem, similarItems[0].Title)
	}

	input.Source = strings.ToLower(strings.TrimSpace(input.Source))
	if input.Source != SubmissionSourcePortal && input.Source != SubmissionSourceWidget {
		return CorePublicItemResult{}, invalidInput("public feedback source must be portal or widget")
	}
	itemInput := CoreItemInput{
		WorkspaceID:     portal.WorkspaceID,
		PortalID:        portal.ID,
		BoardID:         board.ID,
		Title:           input.Title,
		Description:     input.Description,
		DescriptionHTML: input.DescriptionHTML,
		Slug:            normalizeSlug(input.Title) + "-" + uuid.NewString()[:8],
		Source:          input.Source,
	}
	if input.ParticipationIntent == ParticipationIntentAccount {
		contributor, err := s.repo.GetOrCreateAccountContributor(ctx, portal.ID, input.AuthorID)
		if err != nil {
			return CorePublicItemResult{}, err
		}
		itemInput.AuthorID = input.AuthorID
		itemInput.ContributorID = contributor.ID
		if s.nextRepo != nil {
			participant, participantErr := s.nextRepo.GetParticipantByUser(ctx, portal.ID, input.AuthorID)
			if participantErr != nil {
				return CorePublicItemResult{}, participantErr
			}
			item, err := s.nextRepo.CreateContributorItemAndFollow(ctx, CoreContributorItemInput{Item: itemInput, Participant: participant})
			return CorePublicItemResult{Item: item, ParticipantKind: ContributorKindAccount, Following: err == nil}, err
		}
		item, err := s.CreateItem(ctx, itemInput)
		return CorePublicItemResult{Item: item, ParticipantKind: ContributorKindAccount}, err
	}
	if input.ParticipationIntent == ParticipationIntentVerifiedGuest || input.ParticipationIntent == ParticipationIntentExternal {
		if portal.ParticipationMode != ParticipationModeVerifiedGuest && portal.ParticipationMode != ParticipationModeAnonymousAllowed {
			return CorePublicItemResult{}, ErrParticipationNotAllowed
		}
		participant := *input.Participant
		if participant.PortalID != portal.ID || participant.BlockedAt != nil {
			return CorePublicItemResult{}, ErrContributorBlocked
		}
		itemInput.ContributorID = participant.ID
		itemInput.AuthorID = participant.UserID
		item, err := s.nextRepo.CreateContributorItemAndFollow(ctx, CoreContributorItemInput{Item: itemInput, Participant: participant})
		return CorePublicItemResult{Item: item, ParticipantKind: participant.Kind, Following: err == nil}, err
	}
	if portal.ParticipationMode != ParticipationModeAnonymousAllowed {
		return CorePublicItemResult{}, ErrParticipationNotAllowed
	}
	item, err := s.repo.CreateAnonymousItem(ctx, itemInput)
	if err != nil {
		return CorePublicItemResult{}, err
	}
	return CorePublicItemResult{Item: item, Anonymous: true, ParticipantKind: ContributorKindAnonymous}, nil
}

func (s *Service) AttachPublicItemFile(
	ctx context.Context,
	portalSlug string,
	itemID, attachmentID, accountID uuid.UUID,
	participant *CoreParticipant,
	participationIntent string,
) (CoreItemAttachment, error) {
	if itemID == uuid.Nil || attachmentID == uuid.Nil {
		return CoreItemAttachment{}, invalidInput("feedback and attachment ids are required")
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreItemAttachment{}, err
	}
	item, err := s.repo.GetItemByPortal(ctx, portal.ID, itemID)
	if err != nil {
		return CoreItemAttachment{}, err
	}
	switch strings.ToLower(strings.TrimSpace(participationIntent)) {
	case ParticipationIntentAccount:
		if accountID == uuid.Nil || item.AuthorID != accountID {
			return CoreItemAttachment{}, ErrAuthenticationRequired
		}
	case ParticipationIntentVerifiedGuest, ParticipationIntentExternal:
		if participant == nil || participant.ID == uuid.Nil || item.ContributorID != participant.ID {
			return CoreItemAttachment{}, ErrAuthenticationRequired
		}
	case ParticipationIntentAnonymous:
		if item.ParticipantKind != ContributorKindAnonymous {
			return CoreItemAttachment{}, ErrAuthenticationRequired
		}
	default:
		return CoreItemAttachment{}, invalidInput("unsupported feedback participation intent")
	}
	return s.repo.LinkItemAttachment(ctx, portal.ID, item.ID, attachmentID)
}

func (s *Service) GetPublicItemAttachment(
	ctx context.Context,
	portalSlug string,
	itemID, attachmentID uuid.UUID,
) (CoreItemAttachment, error) {
	if itemID == uuid.Nil || attachmentID == uuid.Nil {
		return CoreItemAttachment{}, invalidInput("feedback and attachment ids are required")
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreItemAttachment{}, err
	}
	return s.repo.GetItemAttachment(ctx, portal.ID, itemID, attachmentID)
}

func (s *Service) ListPublicSimilarItems(ctx context.Context, portalSlug, title, description string, limit int) ([]CoreSimilarItem, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(title) < minimumSimilarityTitleCharacters {
		return []CoreSimilarItem{}, nil
	}
	if utf8.RuneCountInString(title) > maxPublicFeedbackTitleCharacters {
		return nil, invalidInputf("feedback title must be %d characters or fewer", maxPublicFeedbackTitleCharacters)
	}
	if limit <= 0 {
		limit = defaultSimilarItemsLimit
	}
	if limit > maxSimilarItemsLimit {
		limit = maxSimilarItemsLimit
	}

	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListSimilarItems(ctx, portal.ID, title, description, limit)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].IsDuplicate = items[index].Confidence >= duplicateItemConfidence
	}
	return items, nil
}

func (s *Service) UpdateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, input CoreUpdateItemStatusInput) (CoreItem, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return CoreItem{}, invalidInput("workspace id and feedback id are required")
	}
	if !isValidStatus(input.Status) {
		return CoreItem{}, invalidInput("unsupported feedback status")
	}
	if input.RoadmapSummary != nil {
		trimmed := strings.TrimSpace(*input.RoadmapSummary)
		input.RoadmapSummary = &trimmed
	}
	current, err := s.getInternalItem(ctx, workspaceID, itemID, input.ActorID)
	if err != nil {
		return CoreItem{}, err
	}
	if current.MergedIntoItemID != nil {
		return CoreItem{}, ErrMergeConflict
	}
	item, statusChanged, err := s.repo.UpdateItemStatus(ctx, workspaceID, itemID, input)
	if err != nil {
		return CoreItem{}, err
	}
	if statusChanged && shouldNotifyFeedbackStatusTransition(item.Status, input.RoadmapSummary) {
		s.publishAccountStatusNotifications(ctx, item, input.ActorID, uuid.New(), s.now().UTC())
		s.enqueueStatusDeliveries(ctx, item, input.ActorID, "")
	}
	return item, nil
}

// UpdateItemStatusIfUnchanged applies an external user request only while the
// feedback item still has the version included in the confirmation preview.
// The repository owns the row lock/CAS so a concurrent portal update cannot be
// silently overwritten between the read and write.
func (s *Service) UpdateItemStatusIfUnchanged(
	ctx context.Context,
	workspaceID, itemID uuid.UUID,
	expectedUpdatedAt time.Time,
	input CoreUpdateItemStatusInput,
) (CoreItem, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return CoreItem{}, invalidInput("workspace id and feedback id are required")
	}
	if expectedUpdatedAt.IsZero() {
		return CoreItem{}, invalidInput("expected feedback update time is required")
	}
	if !isValidStatus(input.Status) {
		return CoreItem{}, invalidInput("unsupported feedback status")
	}
	if input.RoadmapSummary != nil {
		trimmed := strings.TrimSpace(*input.RoadmapSummary)
		input.RoadmapSummary = &trimmed
	}
	current, err := s.getInternalItem(ctx, workspaceID, itemID, input.ActorID)
	if err != nil {
		return CoreItem{}, err
	}
	if current.MergedIntoItemID != nil {
		return CoreItem{}, ErrMergeConflict
	}
	item, statusChanged, updated, err := s.repo.UpdateItemStatusIfUnchanged(
		ctx,
		workspaceID,
		itemID,
		expectedUpdatedAt.UTC(),
		input,
	)
	if err != nil {
		return CoreItem{}, err
	}
	if !updated {
		current, getErr := s.getInternalItem(ctx, workspaceID, itemID, input.ActorID)
		if getErr == nil && current.Status == input.Status {
			return current, nil
		}
		return CoreItem{}, ErrVersionConflict
	}
	if statusChanged && shouldNotifyFeedbackStatusTransition(item.Status, input.RoadmapSummary) {
		s.publishAccountStatusNotifications(ctx, item, input.ActorID, uuid.New(), s.now().UTC())
		s.enqueueStatusDeliveries(ctx, item, input.ActorID, "")
	}
	return item, nil
}
