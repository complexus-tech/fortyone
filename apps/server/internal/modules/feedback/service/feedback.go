package feedback

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrInvalidInput   = errors.New("invalid feedback input")
	ErrNotFound       = sql.ErrNoRows
	ErrAlreadyPlanned = errors.New("feedback is already linked to a primary story")
	ErrTeamMismatch   = errors.New("feedback and story must belong to the same team")
	ErrStoryManaged   = errors.New("feedback status is managed by its linked story")
)

var nonSlugCharacters = regexp.MustCompile(`[^a-z0-9]+`)

const (
	maxPublicFeedbackTitleCharacters       = 200
	maxPublicFeedbackDescriptionCharacters = 20_000
	maxPublicFeedbackCommentCharacters     = 10_000
)

type Service struct {
	repo      Repository
	stories   StoryService
	publisher EventPublisher
	log       *logger.Logger
}

type Option func(*Service)

func WithEventPublisher(log *logger.Logger, publisher EventPublisher) Option {
	return func(service *Service) {
		service.log = log
		service.publisher = publisher
	}
}

func New(repo Repository, stories StoryService, options ...Option) *Service {
	service := &Service{repo: repo, stories: stories}
	for _, option := range options {
		option(service)
	}
	return service
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
		return nil, invalidInput("workspace id is required")
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
		IsPublic:    true,
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
	return s.repo.CreatePortal(ctx, input)
}

func (s *Service) UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input CorePortalInput) (CorePortal, error) {
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return CorePortal{}, invalidInput("workspace id and portal id are required")
	}
	return s.repo.UpdatePortal(ctx, workspaceID, portalID, input)
}

func (s *Service) ListPortalBoards(ctx context.Context, workspaceID, portalID uuid.UUID) ([]CoreBoard, error) {
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return nil, invalidInput("workspace id and portal id are required")
	}
	portal, err := s.repo.GetPortal(ctx, workspaceID, portalID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListBoards(ctx, portal.ID)
}

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

func (s *Service) ListTeamItems(ctx context.Context, workspaceID, teamID, viewerID uuid.UUID, status string, page, pageSize int) (CoreItemsPage, error) {
	return s.ListItems(ctx, CoreListItemsInput{
		WorkspaceID: workspaceID,
		TeamID:      &teamID,
		ViewerID:    viewerID,
		Status:      status,
		Sort:        "newest",
		Page:        page,
		PageSize:    pageSize,
	})
}

func (s *Service) GetItemDetails(ctx context.Context, workspaceID, itemID, viewerID uuid.UUID) (CoreItemDetails, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || viewerID == uuid.Nil {
		return CoreItemDetails{}, invalidInput("workspace, feedback, and viewer ids are required")
	}
	item, err := s.repo.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return CoreItemDetails{}, err
	}
	item.ReadAt, err = s.repo.GetItemReadAt(ctx, workspaceID, itemID, viewerID)
	if err != nil {
		return CoreItemDetails{}, err
	}
	comments, err := s.repo.ListItemComments(ctx, workspaceID, itemID)
	if err != nil {
		return CoreItemDetails{}, err
	}
	links, err := s.repo.ListItemStoryLinks(ctx, workspaceID, itemID)
	if err != nil {
		return CoreItemDetails{}, err
	}
	return CoreItemDetails{Item: item, Comments: comments, StoryLinks: links}, nil
}

func (s *Service) ListTeamSummaries(ctx context.Context, workspaceID, userID uuid.UUID) ([]CoreTeamSummary, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil {
		return nil, invalidInput("workspace and user ids are required")
	}
	return s.repo.ListTeamSummaries(ctx, workspaceID, userID)
}

func (s *Service) MarkItemRead(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (*time.Time, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || userID == uuid.Nil {
		return nil, invalidInput("workspace, feedback, and user ids are required")
	}
	readAt, err := s.repo.MarkItemRead(ctx, workspaceID, itemID, userID)
	if err != nil {
		return nil, err
	}
	return &readAt, nil
}

func (s *Service) MarkItemUnread(ctx context.Context, workspaceID, itemID, userID uuid.UUID) error {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || userID == uuid.Nil {
		return invalidInput("workspace, feedback, and user ids are required")
	}
	return s.repo.MarkItemUnread(ctx, workspaceID, itemID, userID)
}

func (s *Service) ListStoryFeedbackLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]CoreStoryFeedbackLink, error) {
	if workspaceID == uuid.Nil || storyID == uuid.Nil {
		return nil, invalidInput("workspace and story ids are required")
	}
	return s.repo.ListStoryFeedbackLinks(ctx, workspaceID, storyID)
}

func (s *Service) GetItem(ctx context.Context, workspaceID, itemID uuid.UUID) (CoreItem, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return CoreItem{}, invalidInput("workspace id and feedback id are required")
	}
	return s.repo.GetItem(ctx, workspaceID, itemID)
}

func (s *Service) CreateBoard(ctx context.Context, input CoreBoardInput) (CoreBoard, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.TeamID == uuid.Nil {
		return CoreBoard{}, invalidInput("workspace, portal, and team are required")
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

func (s *Service) CreateItem(ctx context.Context, input CoreItemInput) (CoreItem, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.BoardID == uuid.Nil || input.AuthorID == uuid.Nil {
		return CoreItem{}, invalidInput("workspace, portal, board, and author are required")
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return CoreItem{}, invalidInput("feedback title is required")
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
		return CoreItem{}, invalidInput("feedback title is required")
	}
	if utf8.RuneCountInString(input.Title) > maxPublicFeedbackTitleCharacters {
		return CoreItem{}, invalidInputf("feedback title must be %d characters or fewer", maxPublicFeedbackTitleCharacters)
	}
	if utf8.RuneCountInString(input.Description) > maxPublicFeedbackDescriptionCharacters {
		return CoreItem{}, invalidInputf("feedback description must be %d characters or fewer", maxPublicFeedbackDescriptionCharacters)
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
		return CoreItem{}, invalidInput("workspace id and feedback id are required")
	}
	if !isValidStatus(input.Status) {
		return CoreItem{}, invalidInput("unsupported feedback status")
	}
	if input.RoadmapSummary != nil {
		trimmed := strings.TrimSpace(*input.RoadmapSummary)
		input.RoadmapSummary = &trimmed
	}
	item, statusChanged, err := s.repo.UpdateItemStatus(ctx, workspaceID, itemID, input)
	if err != nil {
		return CoreItem{}, err
	}
	if statusChanged && shouldNotify(item.AuthorID, input.ActorID) {
		s.publish(ctx, events.Event{
			Type: events.FeedbackStatusUpdated,
			Payload: events.FeedbackStatusUpdatedPayload{
				EventID:       uuid.New(),
				FeedbackID:    item.ID,
				FeedbackTitle: item.Title,
				FeedbackSlug:  item.Slug,
				WorkspaceID:   item.WorkspaceID,
				RecipientID:   item.AuthorID,
				Status:        item.Status,
			},
			Timestamp: time.Now(),
			ActorID:   input.ActorID,
		})
	}
	return item, nil
}

func (s *Service) CreateComment(ctx context.Context, input CoreCommentInput) (CoreComment, error) {
	if input.WorkspaceID == uuid.Nil || input.ItemID == uuid.Nil || input.AuthorID == uuid.Nil {
		return CoreComment{}, invalidInput("workspace, feedback, and author are required")
	}
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" {
		return CoreComment{}, invalidInput("comment body is required")
	}
	item, err := s.repo.GetItem(ctx, input.WorkspaceID, input.ItemID)
	if err != nil {
		return CoreComment{}, err
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

	return s.createComment(ctx, CoreCommentInput{
		WorkspaceID: portal.WorkspaceID,
		ItemID:      item.ID,
		AuthorID:    input.AuthorID,
		Body:        input.Body,
	}, item)
}

func (s *Service) createComment(ctx context.Context, input CoreCommentInput, item CoreItem) (CoreComment, error) {
	comment, err := s.repo.CreateComment(ctx, input)
	if err != nil {
		return CoreComment{}, err
	}
	if shouldNotify(item.AuthorID, input.AuthorID) {
		s.publish(ctx, events.Event{
			Type: events.FeedbackCommentCreated,
			Payload: events.FeedbackCommentCreatedPayload{
				CommentID:     comment.ID,
				FeedbackID:    item.ID,
				FeedbackTitle: item.Title,
				FeedbackSlug:  item.Slug,
				WorkspaceID:   item.WorkspaceID,
				RecipientID:   item.AuthorID,
				Content:       comment.Body,
			},
			Timestamp: time.Now(),
			ActorID:   input.AuthorID,
		})
	}
	return comment, nil
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

func (s *Service) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (CoreVoteResult, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || userID == uuid.Nil {
		return CoreVoteResult{}, invalidInput("workspace, feedback, and user are required")
	}
	if vote != -1 && vote != 1 {
		return CoreVoteResult{}, invalidInput("feedback vote must be either -1 or 1")
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
		return CoreStoryLink{}, invalidInput("workspace, feedback, story, and actor are required")
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
		return CoreCreateStoryResult{}, invalidInput("workspace, feedback, actor, and team are required")
	}
	item, err := s.repo.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	if item.Board.TeamID == uuid.Nil || item.Board.TeamID != input.TeamID {
		return CoreCreateStoryResult{}, ErrTeamMismatch
	}
	existingLinks, err := s.repo.ListItemStoryLinks(ctx, workspaceID, itemID)
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	for _, existingLink := range existingLinks {
		if !existingLink.IsPrimary {
			continue
		}
		if input.StoryID != nil && existingLink.StoryID != *input.StoryID {
			return CoreCreateStoryResult{}, ErrAlreadyPlanned
		}
		if _, err := s.UpdateItemStatus(ctx, workspaceID, itemID, CoreUpdateItemStatusInput{Status: item.Status, ActorID: actorID, AllowLinked: true}); err != nil {
			return CoreCreateStoryResult{}, err
		}
		return CoreCreateStoryResult{
			ItemID:  itemID,
			StoryID: existingLink.StoryID,
			LinkID:  existingLink.ID,
			Created: false,
		}, nil
	}

	if input.StoryID != nil {
		story, err := s.stories.Get(ctx, *input.StoryID, workspaceID)
		if err != nil {
			return CoreCreateStoryResult{}, err
		}
		if story.Team != item.Board.TeamID {
			return CoreCreateStoryResult{}, ErrTeamMismatch
		}
		if story.DeletedAt != nil {
			return CoreCreateStoryResult{}, fmt.Errorf("%w: deleted stories cannot be linked to feedback", ErrInvalidInput)
		}
		plannedStatus := StatusPlanned
		if story.Status != nil {
			category, err := s.repo.GetStatusCategory(ctx, item.Board.TeamID, *story.Status)
			if err != nil {
				return CoreCreateStoryResult{}, err
			}
			plannedStatus = feedbackStatusForStoryCategory(category)
		}
		link, err := s.repo.LinkStory(ctx, CoreStoryLinkInput{
			WorkspaceID:     workspaceID,
			ItemID:          itemID,
			StoryID:         story.ID,
			Relationship:    RelationshipSolves,
			IsPrimary:       true,
			CreatedByUserID: actorID,
		})
		if err != nil {
			if errors.Is(err, ErrAlreadyPlanned) {
				return s.resolveExistingPlan(ctx, workspaceID, itemID, input.StoryID)
			}
			return CoreCreateStoryResult{}, err
		}
		if _, err := s.UpdateItemStatus(ctx, workspaceID, itemID, CoreUpdateItemStatusInput{Status: plannedStatus, ActorID: actorID, AllowLinked: true}); err != nil {
			return CoreCreateStoryResult{}, err
		}
		return CoreCreateStoryResult{ItemID: itemID, StoryID: story.ID, LinkID: link.ID, Created: false}, nil
	}

	statusID := input.StatusID
	statusCategory := "unstarted"
	if statusID == nil {
		statusID, err = s.repo.FindFirstStatusByCategory(ctx, input.TeamID, "unstarted")
		if err != nil {
			return CoreCreateStoryResult{}, err
		}
		if statusID == nil {
			return CoreCreateStoryResult{}, invalidInput("team has no unstarted status configured")
		}
	} else {
		statusCategory, err = s.repo.GetStatusCategory(ctx, input.TeamID, *statusID)
		if err != nil {
			return CoreCreateStoryResult{}, invalidInput("story status does not belong to the feedback team")
		}
	}
	description := item.Description
	story, err := s.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title:       item.Title,
		Description: &description,
		Status:      statusID,
		Reporter:    &actorID,
		Team:        item.Board.TeamID,
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
		IsPrimary:       true,
		CreatedByUserID: actorID,
	})
	if err != nil {
		if deleteErr := s.stories.Delete(context.WithoutCancel(ctx), story.ID, workspaceID); deleteErr != nil {
			return CoreCreateStoryResult{}, errors.Join(err, fmt.Errorf("compensating story delete: %w", deleteErr))
		}
		if errors.Is(err, ErrAlreadyPlanned) {
			return s.resolveExistingPlan(ctx, workspaceID, itemID, nil)
		}
		return CoreCreateStoryResult{}, err
	}
	if _, err := s.UpdateItemStatus(ctx, workspaceID, itemID, CoreUpdateItemStatusInput{Status: feedbackStatusForStoryCategory(statusCategory), ActorID: actorID, AllowLinked: true}); err != nil {
		return CoreCreateStoryResult{}, err
	}
	return CoreCreateStoryResult{ItemID: itemID, StoryID: story.ID, LinkID: link.ID, Created: true}, nil
}

func (s *Service) resolveExistingPlan(ctx context.Context, workspaceID, itemID uuid.UUID, requestedStoryID *uuid.UUID) (CoreCreateStoryResult, error) {
	links, err := s.repo.ListItemStoryLinks(ctx, workspaceID, itemID)
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	for _, link := range links {
		if !link.IsPrimary {
			continue
		}
		if requestedStoryID != nil && link.StoryID != *requestedStoryID {
			return CoreCreateStoryResult{}, ErrAlreadyPlanned
		}
		return CoreCreateStoryResult{
			ItemID:  itemID,
			StoryID: link.StoryID,
			LinkID:  link.ID,
			Created: false,
		}, nil
	}
	return CoreCreateStoryResult{}, ErrAlreadyPlanned
}

func feedbackStatusForStoryCategory(category string) string {
	switch category {
	case "backlog":
		return StatusReviewing
	case "started":
		return StatusInProgress
	case "completed":
		return StatusCompleted
	case "cancelled":
		return StatusClosed
	case "unstarted", "paused":
		return StatusPlanned
	default:
		return StatusPlanned
	}
}

func invalidInput(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, message)
}

func invalidInputf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
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
