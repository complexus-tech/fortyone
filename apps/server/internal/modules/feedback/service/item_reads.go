package feedback

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) ListItems(ctx context.Context, input CoreListItemsInput) (CoreItemsPage, error) {
	if input.PortalID == uuid.Nil && (input.WorkspaceID == uuid.Nil || input.TeamID == nil || *input.TeamID == uuid.Nil) {
		return CoreItemsPage{}, invalidInput("portal id or workspace and team ids are required")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 50 {
		input.PageSize = 50
	}
	input.Search = strings.TrimSpace(input.Search)
	input.Sort = strings.TrimSpace(input.Sort)
	if input.Sort == "" {
		input.Sort = "top"
	}
	if input.Status != "" && input.Status != "active" && input.Status != "all" && !isValidStatus(input.Status) {
		return CoreItemsPage{}, invalidInput("unsupported feedback status")
	}
	return s.repo.ListItems(ctx, input)
}

func (s *Service) GetPublicContributor(ctx context.Context, portalSlug string, authorID uuid.UUID) (CoreContributor, error) {
	if authorID == uuid.Nil {
		return CoreContributor{}, invalidInput("contributor id is required")
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributor{}, err
	}
	return s.repo.GetContributor(ctx, portal.ID, authorID)
}

func (s *Service) ListContributorActivity(ctx context.Context, userID uuid.UUID, activityType string, page, pageSize int) (CoreContributorActivityPage, error) {
	if userID == uuid.Nil {
		return CoreContributorActivityPage{}, invalidInput("user id is required")
	}
	activityType = strings.TrimSpace(activityType)
	if activityType != "" && activityType != "feedback" && activityType != "comment" {
		return CoreContributorActivityPage{}, invalidInput("activity type must be feedback or comment")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultContributorPageSize
	}
	if pageSize > maxContributorPageSize {
		pageSize = maxContributorPageSize
	}
	return s.repo.ListContributorActivity(ctx, CoreListContributorActivityInput{
		UserID:       userID,
		ActivityType: activityType,
		Page:         page,
		PageSize:     pageSize,
	})
}

func (s *Service) GetWorkspacePublicContributor(ctx context.Context, workspaceSlug, portalSlug string, authorID uuid.UUID) (CoreContributor, error) {
	if authorID == uuid.Nil {
		return CoreContributor{}, invalidInput("contributor id is required")
	}
	portal, err := s.repo.GetPortalByWorkspaceSlugAndSlug(ctx, strings.TrimSpace(workspaceSlug), strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributor{}, err
	}
	return s.repo.GetContributor(ctx, portal.ID, authorID)
}

func (s *Service) ListPublicContributorComments(ctx context.Context, portalSlug string, authorID uuid.UUID, page, pageSize int) (CoreContributorCommentsPage, error) {
	if authorID == uuid.Nil {
		return CoreContributorCommentsPage{}, invalidInput("contributor id is required")
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorCommentsPage{}, err
	}
	return s.listContributorComments(ctx, portal.ID, authorID, page, pageSize)
}

func (s *Service) ListWorkspacePublicContributorComments(ctx context.Context, workspaceSlug, portalSlug string, authorID uuid.UUID, page, pageSize int) (CoreContributorCommentsPage, error) {
	if authorID == uuid.Nil {
		return CoreContributorCommentsPage{}, invalidInput("contributor id is required")
	}
	portal, err := s.repo.GetPortalByWorkspaceSlugAndSlug(ctx, strings.TrimSpace(workspaceSlug), strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorCommentsPage{}, err
	}
	return s.listContributorComments(ctx, portal.ID, authorID, page, pageSize)
}

func (s *Service) listContributorComments(ctx context.Context, portalID, authorID uuid.UUID, page, pageSize int) (CoreContributorCommentsPage, error) {
	exists, err := s.repo.ContributorExists(ctx, portalID, authorID)
	if err != nil {
		return CoreContributorCommentsPage{}, err
	}
	if !exists {
		return CoreContributorCommentsPage{}, ErrNotFound
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultContributorPageSize
	}
	if pageSize > maxContributorPageSize {
		pageSize = maxContributorPageSize
	}
	return s.repo.ListContributorComments(ctx, CoreListContributorCommentsInput{
		PortalID: portalID,
		AuthorID: authorID,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *Service) ListTeamItems(ctx context.Context, workspaceID, teamID, viewerID uuid.UUID, status, search string, page, pageSize int) (CoreItemsPage, error) {
	deletedOnly := status == ListStatusTrashed
	if deletedOnly {
		status = "all"
	} else if status == "all" {
		// Preserve old team-feedback URLs without exposing terminal feedback in
		// the broad list. Completed and closed items have dedicated filters.
		status = "active"
	}
	input := CoreListItemsInput{
		WorkspaceID: workspaceID,
		TeamID:      &teamID,
		ViewerID:    viewerID,
		Status:      status,
		Search:      search,
		Sort:        "newest",
		Page:        page,
		PageSize:    pageSize,
		DeletedOnly: deletedOnly,
	}
	if scope, scoped, err := s.coreScope(ctx, workspaceID, viewerID); err != nil {
		return CoreItemsPage{}, err
	} else if scoped {
		return s.scopedCoreRepo.ListItemsScoped(ctx, scope, input)
	}
	return s.ListItems(ctx, input)
}

func (s *Service) GetItemDetails(ctx context.Context, workspaceID, itemID, viewerID uuid.UUID) (CoreItemDetails, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || viewerID == uuid.Nil {
		return CoreItemDetails{}, invalidInput("workspace, feedback, and viewer ids are required")
	}
	item, err := s.getInternalItem(ctx, workspaceID, itemID, viewerID)
	if err != nil {
		return CoreItemDetails{}, err
	}
	if scope, scoped, scopeErr := s.coreScope(ctx, workspaceID, viewerID); scopeErr != nil {
		return CoreItemDetails{}, scopeErr
	} else if scoped {
		item.ReadAt, err = s.scopedCoreRepo.GetItemReadAtScoped(ctx, scope, itemID)
	} else {
		item.ReadAt, err = s.repo.GetItemReadAt(ctx, workspaceID, itemID, viewerID)
	}
	if err != nil {
		return CoreItemDetails{}, err
	}
	var comments []CoreComment
	var links []CoreStoryLink
	if scope, scoped, scopeErr := s.coreScope(ctx, workspaceID, viewerID); scopeErr != nil {
		return CoreItemDetails{}, scopeErr
	} else if scoped {
		comments, err = s.scopedCoreRepo.ListItemCommentsScoped(ctx, scope, itemID)
		if err == nil {
			links, err = s.scopedCoreRepo.ListItemStoryLinksScoped(ctx, scope, itemID)
		}
	} else {
		comments, err = s.repo.ListItemComments(ctx, workspaceID, itemID)
		if err == nil {
			links, err = s.repo.ListItemStoryLinks(ctx, workspaceID, itemID)
		}
	}
	if err != nil {
		return CoreItemDetails{}, err
	}
	return CoreItemDetails{Item: item, Comments: comments, StoryLinks: links}, nil
}

func (s *Service) GetPrivateAuthor(ctx context.Context, workspaceID, itemID uuid.UUID) (CorePrivateAuthor, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return CorePrivateAuthor{}, invalidInput("workspace and feedback ids are required")
	}
	scope, scoped, err := s.coreScope(ctx, workspaceID, uuid.Nil)
	if err != nil {
		return CorePrivateAuthor{}, err
	}
	if scoped {
		return s.scopedCoreRepo.GetPrivateAuthorScoped(ctx, scope, itemID)
	}
	return s.repo.GetPrivateAuthor(ctx, workspaceID, itemID)
}

func (s *Service) ResolveCanonicalItem(ctx context.Context, portalSlug, itemReference string) (CoreCanonicalItem, error) {
	itemReference = strings.TrimSpace(itemReference)
	if itemReference == "" || len(itemReference) > 255 {
		return CoreCanonicalItem{}, invalidInput("feedback id or slug is required")
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreCanonicalItem{}, err
	}
	return s.repo.ResolveCanonicalItem(ctx, portal.ID, itemReference)
}

func (s *Service) ListTeamSummaries(ctx context.Context, workspaceID, userID uuid.UUID) ([]CoreTeamSummary, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil {
		return nil, invalidInput("workspace and user ids are required")
	}
	scope, scoped, err := s.coreScope(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if scoped {
		return s.scopedCoreRepo.ListTeamSummariesScoped(ctx, scope)
	}
	return s.repo.ListTeamSummaries(ctx, workspaceID, userID)
}

func (s *Service) MarkItemRead(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (*time.Time, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || userID == uuid.Nil {
		return nil, invalidInput("workspace, feedback, and user ids are required")
	}
	readAt, err := s.markInternalItemRead(ctx, workspaceID, itemID, userID)
	if err != nil {
		return nil, err
	}
	return &readAt, nil
}

func (s *Service) MarkItemUnread(ctx context.Context, workspaceID, itemID, userID uuid.UUID) error {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || userID == uuid.Nil {
		return invalidInput("workspace, feedback, and user ids are required")
	}
	scope, scoped, err := s.coreScope(ctx, workspaceID, userID)
	if err != nil {
		return err
	}
	if scoped {
		return s.scopedCoreRepo.MarkItemUnreadScoped(ctx, scope, itemID)
	}
	return s.repo.MarkItemUnread(ctx, workspaceID, itemID, userID)
}

func (s *Service) ListStoryFeedbackLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]CoreStoryFeedbackLink, error) {
	if workspaceID == uuid.Nil || storyID == uuid.Nil {
		return nil, invalidInput("workspace and story ids are required")
	}
	scope, scoped, err := s.coreScope(ctx, workspaceID, uuid.Nil)
	if err != nil {
		return nil, err
	}
	if scoped {
		return s.scopedCoreRepo.ListStoryFeedbackLinksScoped(ctx, scope, storyID)
	}
	return s.repo.ListStoryFeedbackLinks(ctx, workspaceID, storyID)
}

func (s *Service) GetItem(ctx context.Context, workspaceID, itemID uuid.UUID) (CoreItem, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return CoreItem{}, invalidInput("workspace id and feedback id are required")
	}
	return s.getInternalItem(ctx, workspaceID, itemID, uuid.Nil)
}

func (s *Service) TrashItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return invalidInput("workspace id and feedback id are required")
	}
	item, err := s.getInternalItem(ctx, workspaceID, itemID, uuid.Nil)
	if err != nil {
		return err
	}
	if item.MergedIntoItemID != nil {
		return ErrMergeConflict
	}
	scope, scoped, err := s.coreScope(ctx, workspaceID, uuid.Nil)
	if err != nil {
		return err
	}
	if scoped {
		return s.scopedCoreRepo.TrashItemScoped(ctx, scope, itemID)
	}
	return s.repo.TrashItem(ctx, workspaceID, itemID)
}

func (s *Service) RestoreItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return invalidInput("workspace id and feedback id are required")
	}
	item, err := s.getInternalItem(ctx, workspaceID, itemID, uuid.Nil)
	if err == nil && item.MergedIntoItemID != nil {
		return ErrMergeConflict
	}
	scope, scoped, scopeErr := s.coreScope(ctx, workspaceID, uuid.Nil)
	if scopeErr != nil {
		return scopeErr
	}
	if scoped {
		return s.scopedCoreRepo.RestoreItemScoped(ctx, scope, itemID)
	}
	return s.repo.RestoreItem(ctx, workspaceID, itemID)
}
