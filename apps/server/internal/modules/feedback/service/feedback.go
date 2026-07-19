package feedback

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid feedback input")
	ErrNotFound     = sql.ErrNoRows
)

var nonSlugCharacters = regexp.MustCompile(`[^a-z0-9]+`)

const (
	maxPublicFeedbackTitleCharacters       = 200
	maxPublicFeedbackDescriptionCharacters = 20_000
	maxPublicFeedbackCommentCharacters     = 10_000
)

type Service struct {
	repo    Repository
	stories StoryService
}

func New(repo Repository, stories StoryService) *Service {
	return &Service{repo: repo, stories: stories}
}

func (s *Service) GetPortalSnapshot(ctx context.Context, slug string) (CorePortalSnapshot, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(slug))
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	boards, err := s.repo.ListBoards(ctx, portal.ID)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	itemsPage, err := s.ListItems(ctx, CoreListItemsInput{PortalID: portal.ID, Page: 1, PageSize: 50})
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	comments, err := s.repo.ListComments(ctx, portal.ID)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	links, err := s.repo.ListStoryLinks(ctx, portal.ID)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	return CorePortalSnapshot{Portal: portal, Boards: boards, Items: itemsPage.Items, ItemsHasMore: itemsPage.HasMore, Comments: comments, Links: links}, nil
}

func (s *Service) GetWorkspacePortalSnapshot(ctx context.Context, workspaceSlug, portalSlug string) (CorePortalSnapshot, error) {
	portal, err := s.repo.GetPortalByWorkspaceSlugAndSlug(ctx, strings.TrimSpace(workspaceSlug), strings.TrimSpace(portalSlug))
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	return s.getPortalSnapshot(ctx, portal)
}

func (s *Service) getPortalSnapshot(ctx context.Context, portal CorePortal) (CorePortalSnapshot, error) {
	boards, err := s.repo.ListBoards(ctx, portal.ID)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	itemsPage, err := s.ListItems(ctx, CoreListItemsInput{PortalID: portal.ID, Page: 1, PageSize: 50})
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	comments, err := s.repo.ListComments(ctx, portal.ID)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	links, err := s.repo.ListStoryLinks(ctx, portal.ID)
	if err != nil {
		return CorePortalSnapshot{}, err
	}
	return CorePortalSnapshot{Portal: portal, Boards: boards, Items: itemsPage.Items, ItemsHasMore: itemsPage.HasMore, Comments: comments, Links: links}, nil
}

func (s *Service) ListPortals(ctx context.Context, input CoreWorkspacePortalInput) ([]CorePortal, error) {
	if input.WorkspaceID == uuid.Nil {
		return nil, errors.New("workspace id is required")
	}
	portals, err := s.repo.ListPortals(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(portals) > 0 {
		return portals, nil
	}
	portal, err := s.CreatePortal(ctx, CorePortalInput{
		WorkspaceID: input.WorkspaceID,
		Description: "Collect public feedback, prioritize requests, and publish roadmap progress.",
		IsPublic:    true,
	})
	if err != nil {
		return nil, err
	}
	return []CorePortal{portal}, nil
}

func (s *Service) CreatePortal(ctx context.Context, input CorePortalInput) (CorePortal, error) {
	if input.WorkspaceID == uuid.Nil {
		return CorePortal{}, errors.New("workspace id is required")
	}
	input.Description = strings.TrimSpace(input.Description)
	return s.repo.CreatePortal(ctx, input)
}

func (s *Service) UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input CorePortalInput) (CorePortal, error) {
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return CorePortal{}, errors.New("workspace id and portal id are required")
	}
	input.Description = strings.TrimSpace(input.Description)
	return s.repo.UpdatePortal(ctx, workspaceID, portalID, input)
}

func (s *Service) ListPortalBoards(ctx context.Context, workspaceID, portalID uuid.UUID) ([]CoreBoard, error) {
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return nil, errors.New("workspace id and portal id are required")
	}
	portal, err := s.repo.GetPortal(ctx, workspaceID, portalID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListBoards(ctx, portal.ID)
}

func (s *Service) ListItems(ctx context.Context, input CoreListItemsInput) (CoreItemsPage, error) {
	if input.PortalID == uuid.Nil {
		return CoreItemsPage{}, errors.New("portal id is required")
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
	if input.Status != "" && !isValidStatus(input.Status) {
		return CoreItemsPage{}, errors.New("unsupported feedback status")
	}
	return s.repo.ListItems(ctx, input)
}

func (s *Service) CreateBoard(ctx context.Context, input CoreBoardInput) (CoreBoard, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.TeamID == uuid.Nil {
		return CoreBoard{}, errors.New("workspace, portal, and team are required")
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return CoreBoard{}, errors.New("board name is required")
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

func (s *Service) CreateItem(ctx context.Context, input CoreItemInput) (CoreItem, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.BoardID == uuid.Nil || input.AuthorID == uuid.Nil {
		return CoreItem{}, errors.New("workspace, portal, board, and author are required")
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return CoreItem{}, errors.New("feedback title is required")
	}
	input.Slug = normalizeSlug(input.Slug)
	if input.Slug == "" {
		input.Slug = normalizeSlug(input.Title) + "-" + uuid.NewString()[:8]
	}
	return s.repo.CreateItem(ctx, input)
}

func (s *Service) CreatePublicItem(ctx context.Context, input CorePublicItemInput) (CoreItem, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return CoreItem{}, errors.New("feedback title is required")
	}
	if utf8.RuneCountInString(input.Title) > maxPublicFeedbackTitleCharacters {
		return CoreItem{}, fmt.Errorf("feedback title must be %d characters or fewer", maxPublicFeedbackTitleCharacters)
	}
	if utf8.RuneCountInString(input.Description) > maxPublicFeedbackDescriptionCharacters {
		return CoreItem{}, fmt.Errorf("feedback description must be %d characters or fewer", maxPublicFeedbackDescriptionCharacters)
	}

	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(input.PortalSlug))
	if err != nil {
		return CoreItem{}, err
	}
	board, err := s.repo.GetBoard(ctx, portal.ID, input.BoardID)
	if err != nil {
		return CoreItem{}, err
	}
	if board.WorkspaceID != portal.WorkspaceID {
		return CoreItem{}, ErrNotFound
	}

	return s.CreateItem(ctx, CoreItemInput{
		WorkspaceID: portal.WorkspaceID,
		PortalID:    portal.ID,
		BoardID:     board.ID,
		AuthorID:    input.AuthorID,
		Title:       input.Title,
		Description: input.Description,
	})
}

func (s *Service) UpdateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, input CoreUpdateItemStatusInput) (CoreItem, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return CoreItem{}, errors.New("workspace id and feedback id are required")
	}
	if !isValidStatus(input.Status) {
		return CoreItem{}, errors.New("unsupported feedback status")
	}
	if input.RoadmapSummary != nil {
		trimmed := strings.TrimSpace(*input.RoadmapSummary)
		input.RoadmapSummary = &trimmed
	}
	return s.repo.UpdateItemStatus(ctx, workspaceID, itemID, input)
}

func (s *Service) CreateComment(ctx context.Context, input CoreCommentInput) (CoreComment, error) {
	if input.WorkspaceID == uuid.Nil || input.ItemID == uuid.Nil || input.AuthorID == uuid.Nil {
		return CoreComment{}, errors.New("workspace, feedback, and author are required")
	}
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" {
		return CoreComment{}, errors.New("comment body is required")
	}
	return s.repo.CreateComment(ctx, input)
}

func (s *Service) CreatePublicComment(ctx context.Context, input CorePublicCommentInput) (CoreComment, error) {
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" {
		return CoreComment{}, errors.New("comment body is required")
	}
	if utf8.RuneCountInString(input.Body) > maxPublicFeedbackCommentCharacters {
		return CoreComment{}, fmt.Errorf("comment body must be %d characters or fewer", maxPublicFeedbackCommentCharacters)
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

	return s.CreateComment(ctx, CoreCommentInput{
		WorkspaceID: portal.WorkspaceID,
		ItemID:      item.ID,
		AuthorID:    input.AuthorID,
		Body:        input.Body,
	})
}

func (s *Service) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (CoreVoteResult, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || userID == uuid.Nil {
		return CoreVoteResult{}, errors.New("workspace, feedback, and user are required")
	}
	if vote != -1 && vote != 1 {
		return CoreVoteResult{}, errors.New("feedback vote must be either -1 or 1")
	}
	if _, err := s.repo.GetItem(ctx, workspaceID, itemID); err != nil {
		return CoreVoteResult{}, err
	}
	return s.repo.ToggleVote(ctx, workspaceID, itemID, userID, vote)
}

func (s *Service) TogglePublicVote(ctx context.Context, input CorePublicVoteInput) (CoreVoteResult, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(input.PortalSlug))
	if err != nil {
		return CoreVoteResult{}, err
	}
	item, err := s.repo.GetItemByPortal(ctx, portal.ID, input.ItemID)
	if err != nil {
		return CoreVoteResult{}, err
	}
	if item.WorkspaceID != portal.WorkspaceID {
		return CoreVoteResult{}, ErrNotFound
	}

	return s.ToggleVote(ctx, portal.WorkspaceID, item.ID, input.UserID, input.Vote)
}

func (s *Service) LinkStory(ctx context.Context, input CoreStoryLinkInput) (CoreStoryLink, error) {
	if input.WorkspaceID == uuid.Nil || input.ItemID == uuid.Nil || input.StoryID == uuid.Nil || input.CreatedByUserID == uuid.Nil {
		return CoreStoryLink{}, errors.New("workspace, feedback, story, and actor are required")
	}
	if input.Relationship == "" {
		input.Relationship = RelationshipLinked
	}
	return s.repo.LinkStory(ctx, input)
}

func (s *Service) CreateStoryFromItem(ctx context.Context, workspaceID, itemID, actorID uuid.UUID, input CoreCreateStoryInput) (CoreCreateStoryResult, error) {
	if s.stories == nil {
		return CoreCreateStoryResult{}, errors.New("story service is required")
	}
	if workspaceID == uuid.Nil || itemID == uuid.Nil || actorID == uuid.Nil || input.TeamID == uuid.Nil {
		return CoreCreateStoryResult{}, errors.New("workspace, feedback, actor, and team are required")
	}
	item, err := s.repo.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	statusID := input.StatusID
	if statusID == nil {
		statusID, err = s.repo.FindFirstStatusByCategory(ctx, input.TeamID, "unstarted")
		if err != nil {
			return CoreCreateStoryResult{}, err
		}
		if statusID == nil {
			return CoreCreateStoryResult{}, errors.New("team has no unstarted status configured")
		}
	}
	description := item.Description
	story, err := s.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title:       item.Title,
		Description: &description,
		Status:      statusID,
		Reporter:    &actorID,
		Team:        input.TeamID,
		Priority:    "No Priority",
	}, workspaceID)
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	link, err := s.repo.LinkStory(ctx, CoreStoryLinkInput{
		WorkspaceID:     workspaceID,
		ItemID:          itemID,
		StoryID:         story.ID,
		Relationship:    RelationshipCreatedFrom,
		CreatedByUserID: actorID,
	})
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	return CoreCreateStoryResult{ItemID: itemID, StoryID: story.ID, LinkID: link.ID}, nil
}

func normalizeSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = nonSlugCharacters.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func isValidStatus(status string) bool {
	switch status {
	case StatusPending, StatusReviewing, StatusPlanned, StatusInProgress, StatusCompleted, StatusClosed:
		return true
	default:
		return false
	}
}
