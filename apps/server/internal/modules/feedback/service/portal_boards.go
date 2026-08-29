package feedback

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) GetPortalSnapshot(ctx context.Context, slug string, input CorePortalSnapshotInput) (CorePortalSnapshot, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(slug))
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	return s.getPortalSnapshot(ctx, portal, input)
}

func (s *Service) GetWorkspacePortalSnapshot(ctx context.Context, workspaceSlug, portalSlug string, input CorePortalSnapshotInput) (CorePortalSnapshot, error) {
	portal, err := s.repo.GetPortalByWorkspaceSlugAndSlug(ctx, strings.TrimSpace(workspaceSlug), strings.TrimSpace(portalSlug))
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	return s.getPortalSnapshot(ctx, portal, input)
}

func (s *Service) getPortalSnapshot(ctx context.Context, portal CorePortal, input CorePortalSnapshotInput) (CorePortalSnapshot, error) {
	boards, err := s.repo.ListBoards(ctx, portal.ID)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	itemsPage, err := s.ListItems(ctx, CoreListItemsInput{
		PortalID: portal.ID,
		AuthorID: input.AuthorID,
		ItemID:   input.ItemID,
		Status:   input.Status,
		BoardID:  input.BoardID,
		Search:   input.Search,
		Sort:     input.Sort,
		Page:     input.Page,
		PageSize: input.PageSize,
	})
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	if input.SummaryOnly {
		return CorePortalSnapshot{Portal: portal, Boards: boards, Items: itemsPage.Items, ItemsHasMore: itemsPage.HasMore}, nil
	}
	visibleItemIDs := coreItemIDs(itemsPage.Items)
	comments, err := s.repo.ListComments(ctx, portal.ID, visibleItemIDs)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	links, err := s.repo.ListStoryLinks(ctx, portal.ID, visibleItemIDs)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	return CorePortalSnapshot{Portal: portal, Boards: boards, Items: itemsPage.Items, ItemsHasMore: itemsPage.HasMore, Comments: comments, Links: links}, nil
}

func (s *Service) ListPortals(ctx context.Context, input CoreWorkspacePortalInput) ([]CorePortal, error) {
	if input.WorkspaceID == uuid.Nil {
		return nil, invalidInput("workspace id is required")
	}
	if err := s.authorizeInternalAccess(ctx, input.WorkspaceID, uuid.Nil); err != nil {
		return nil, err
	}
	portals, err := s.repo.ListPortals(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(portals) > 0 {
		return portals, nil
	}
	portal, err := s.CreatePortal(ctx, CorePortalInput{
		WorkspaceID:       input.WorkspaceID,
		IsPublic:          pointer(true),
		ParticipationMode: pointer(ParticipationModeAccountRequired),
	})
	if err != nil {
		return nil, err
	}
	return []CorePortal{portal}, nil
}

func (s *Service) CreatePortal(ctx context.Context, input CorePortalInput) (CorePortal, error) {
	if input.WorkspaceID == uuid.Nil {
		return CorePortal{}, invalidInput("workspace id is required")
	}
	if err := s.authorizeInternalAccess(ctx, input.WorkspaceID, uuid.Nil); err != nil {
		return CorePortal{}, err
	}
	if input.IsPublic == nil {
		input.IsPublic = pointer(true)
	}
	if input.ParticipationMode == nil {
		input.ParticipationMode = pointer(ParticipationModeAccountRequired)
	}
	if input.GuestIdentityPolicy == nil {
		input.GuestIdentityPolicy = pointer(GuestIdentityPolicyShowIdentity)
	}
	if !isValidParticipationMode(*input.ParticipationMode) {
		return CorePortal{}, invalidInput("unsupported feedback participation mode")
	}
	if !isValidGuestIdentityPolicy(*input.GuestIdentityPolicy) {
		return CorePortal{}, invalidInput("unsupported feedback guest identity policy")
	}
	return s.repo.CreatePortal(ctx, input)
}

func (s *Service) UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input CorePortalInput) (CorePortal, error) {
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return CorePortal{}, invalidInput("workspace id and portal id are required")
	}
	if err := s.authorizeInternalAccess(ctx, workspaceID, uuid.Nil); err != nil {
		return CorePortal{}, err
	}
	if input.IsPublic == nil && input.ParticipationMode == nil && input.GuestIdentityPolicy == nil {
		return CorePortal{}, invalidInput("at least one feedback portal setting is required")
	}
	if input.GuestIdentityPolicy != nil {
		policy := strings.ToLower(strings.TrimSpace(*input.GuestIdentityPolicy))
		if !isValidGuestIdentityPolicy(policy) {
			return CorePortal{}, invalidInput("unsupported feedback guest identity policy")
		}
		input.GuestIdentityPolicy = &policy
	}
	if input.ParticipationMode != nil {
		mode := strings.ToLower(strings.TrimSpace(*input.ParticipationMode))
		if !isValidParticipationMode(mode) {
			return CorePortal{}, invalidInput("unsupported feedback participation mode")
		}
		input.ParticipationMode = &mode
	}
	return s.repo.UpdatePortal(ctx, workspaceID, portalID, input)
}

func (s *Service) ListPortalBoards(ctx context.Context, workspaceID, portalID uuid.UUID) ([]CoreBoard, error) {
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return nil, invalidInput("workspace id and portal id are required")
	}
	if err := s.authorizeInternalAccess(ctx, workspaceID, uuid.Nil); err != nil {
		return nil, err
	}
	portal, err := s.repo.GetPortal(ctx, workspaceID, portalID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListBoards(ctx, portal.ID)
}

func (s *Service) CreateBoard(ctx context.Context, input CoreBoardInput) (CoreBoard, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.TeamID == uuid.Nil || input.CreatorID == uuid.Nil {
		return CoreBoard{}, invalidInput("workspace, portal, team, and creator are required")
	}
	if err := s.authorizeInternalAccess(ctx, input.WorkspaceID, input.CreatorID); err != nil {
		return CoreBoard{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return CoreBoard{}, invalidInput("board name is required")
	}
	input.Slug = normalizeSlug(input.Slug)
	if input.Slug == "" {
		input.Slug = normalizeSlug(input.Name)
	}
	input.Color = strings.TrimSpace(input.Color)
	if input.Color == "" {
		input.Color = "green"
	}
	return s.repo.CreateBoard(ctx, input)
}

func (s *Service) DeleteBoard(ctx context.Context, workspaceID, boardID uuid.UUID) error {
	if workspaceID == uuid.Nil || boardID == uuid.Nil {
		return invalidInput("workspace and board ids are required")
	}
	if err := s.authorizeInternalAccess(ctx, workspaceID, uuid.Nil); err != nil {
		return err
	}
	return s.repo.DeleteBoard(ctx, workspaceID, boardID)
}

func (s *Service) ListBoardReviewers(ctx context.Context, workspaceID, boardID uuid.UUID) ([]CoreBoardReviewer, error) {
	if workspaceID == uuid.Nil || boardID == uuid.Nil {
		return nil, invalidInput("workspace and board ids are required")
	}
	if err := s.authorizeInternalAccess(ctx, workspaceID, uuid.Nil); err != nil {
		return nil, err
	}
	return s.repo.ListBoardReviewers(ctx, workspaceID, boardID)
}

func (s *Service) SetBoardReviewer(ctx context.Context, input CoreBoardReviewerInput) (CoreBoardReviewer, error) {
	if input.WorkspaceID == uuid.Nil || input.BoardID == uuid.Nil || input.UserID == uuid.Nil {
		return CoreBoardReviewer{}, invalidInput("workspace, board, and user ids are required")
	}
	if err := s.authorizeInternalAccess(ctx, input.WorkspaceID, uuid.Nil); err != nil {
		return CoreBoardReviewer{}, err
	}
	input.EmailFrequency = strings.ToLower(strings.TrimSpace(input.EmailFrequency))
	if !isValidReviewerEmailFrequency(input.EmailFrequency) {
		return CoreBoardReviewer{}, invalidInput("email frequency must be off, daily, or weekly")
	}
	return s.repo.SetBoardReviewer(ctx, input)
}
