package feedback

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
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
	ErrInvalidInput              = errors.New("invalid feedback input")
	ErrNotFound                  = sql.ErrNoRows
	ErrBoardExists               = errors.New("a feedback board already exists for this team")
	ErrAlreadyPlanned            = errors.New("feedback is already linked to a primary story")
	ErrTeamMismatch              = errors.New("feedback and story must belong to the same team")
	ErrStoryManaged              = errors.New("feedback status is managed by its linked story")
	ErrDuplicateItem             = errors.New("similar feedback has already been reported")
	ErrVersionConflict           = errors.New("feedback changed since it was reviewed")
	ErrParticipationNotAllowed   = errors.New("anonymous feedback is not allowed for this portal")
	ErrAuthenticationRequired    = errors.New("feedback account authentication is required")
	ErrContributorSessionInvalid = errors.New("feedback contributor session is invalid")
	ErrContributorBlocked        = errors.New("feedback contributor is blocked")
	ErrVerificationExpired       = errors.New("feedback verification has expired")
	ErrVerificationConsumed      = errors.New("feedback verification has already been used")
	ErrVerificationAttempts      = errors.New("feedback verification attempt limit reached")
	ErrWidgetOriginNotAllowed    = errors.New("feedback widget origin is not allowed")
	ErrWidgetAssertionInvalid    = errors.New("feedback widget identity assertion is invalid")
	ErrWidgetAssertionReplayed   = errors.New("feedback widget identity assertion was already used")
	ErrFeatureUnavailable        = errors.New("feedback contributor feature is unavailable")
	ErrMergeConflict             = errors.New("feedback cannot be merged into the requested target")
)

var nonSlugCharacters = regexp.MustCompile(`[^a-z0-9]+`)

const (
	maxPublicFeedbackTitleCharacters       = 200
	maxPublicFeedbackDescriptionCharacters = 20_000
	maxPublicFeedbackCommentCharacters     = 10_000
	defaultContributorPageSize             = 20
	maxContributorPageSize                 = 50
	publicRoadmapSlug                      = "roadmap"
	defaultSimilarItemsLimit               = 3
	maxSimilarItemsLimit                   = 5
	minimumSimilarityTitleCharacters       = 10
	duplicateItemConfidence                = 0.82
)

type Service struct {
	repo                     Repository
	nextRepo                 NextPhaseRepository
	stories                  StoryService
	publisher                EventPublisher
	log                      *logger.Logger
	security                 *contributorSecurity
	websiteURL               string
	tasks                    ContributorTasks
	guestNotificationActorID uuid.UUID
	now                      func() time.Time
	random                   io.Reader
}

type Option func(*Service)

func WithEventPublisher(log *logger.Logger, publisher EventPublisher) Option {
	return func(service *Service) {
		service.log = log
		service.publisher = publisher
	}
}

// WithGuestNotificationActor supplies the existing system actor used as the
// database actor for account notifications initiated by a contributor who has
// no global user row. The payload still carries the contributor's public name.
func WithGuestNotificationActor(actorID uuid.UUID) Option {
	return func(service *Service) {
		service.guestNotificationActorID = actorID
	}
}

func New(repo Repository, stories StoryService, options ...Option) *Service {
	service := &Service{repo: repo, stories: stories, now: time.Now, random: rand.Reader}
	if nextRepo, ok := repo.(NextPhaseRepository); ok {
		service.nextRepo = nextRepo
	}
	for _, option := range options {
		option(service)
	}
	return service
}

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
	return s.ListItems(ctx, CoreListItemsInput{
		WorkspaceID: workspaceID,
		TeamID:      &teamID,
		ViewerID:    viewerID,
		Status:      status,
		Search:      search,
		Sort:        "newest",
		Page:        page,
		PageSize:    pageSize,
		DeletedOnly: deletedOnly,
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

func (s *Service) GetPrivateAuthor(ctx context.Context, workspaceID, itemID uuid.UUID) (CorePrivateAuthor, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return CorePrivateAuthor{}, invalidInput("workspace and feedback ids are required")
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

func (s *Service) TrashItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return invalidInput("workspace id and feedback id are required")
	}
	item, err := s.repo.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return err
	}
	if item.MergedIntoItemID != nil {
		return ErrMergeConflict
	}
	return s.repo.TrashItem(ctx, workspaceID, itemID)
}

func (s *Service) RestoreItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	if workspaceID == uuid.Nil || itemID == uuid.Nil {
		return invalidInput("workspace id and feedback id are required")
	}
	item, err := s.repo.GetItem(ctx, workspaceID, itemID)
	if err == nil && item.MergedIntoItemID != nil {
		return ErrMergeConflict
	}
	return s.repo.RestoreItem(ctx, workspaceID, itemID)
}

func (s *Service) CreateBoard(ctx context.Context, input CoreBoardInput) (CoreBoard, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.TeamID == uuid.Nil || input.CreatorID == uuid.Nil {
		return CoreBoard{}, invalidInput("workspace, portal, team, and creator are required")
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
	return s.repo.DeleteBoard(ctx, workspaceID, boardID)
}

func (s *Service) ListBoardReviewers(ctx context.Context, workspaceID, boardID uuid.UUID) ([]CoreBoardReviewer, error) {
	if workspaceID == uuid.Nil || boardID == uuid.Nil {
		return nil, invalidInput("workspace and board ids are required")
	}
	return s.repo.ListBoardReviewers(ctx, workspaceID, boardID)
}

func (s *Service) SetBoardReviewer(ctx context.Context, input CoreBoardReviewerInput) (CoreBoardReviewer, error) {
	if input.WorkspaceID == uuid.Nil || input.BoardID == uuid.Nil || input.UserID == uuid.Nil {
		return CoreBoardReviewer{}, invalidInput("workspace, board, and user ids are required")
	}
	input.EmailFrequency = strings.ToLower(strings.TrimSpace(input.EmailFrequency))
	if !isValidReviewerEmailFrequency(input.EmailFrequency) {
		return CoreBoardReviewer{}, invalidInput("email frequency must be off, daily, or weekly")
	}
	return s.repo.SetBoardReviewer(ctx, input)
}

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
		WorkspaceID: portal.WorkspaceID,
		PortalID:    portal.ID,
		BoardID:     board.ID,
		Title:       input.Title,
		Description: input.Description,
		Slug:        normalizeSlug(input.Title) + "-" + uuid.NewString()[:8],
		Source:      input.Source,
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
	current, err := s.repo.GetItem(ctx, workspaceID, itemID)
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
	current, err := s.repo.GetItem(ctx, workspaceID, itemID)
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
		current, getErr := s.repo.GetItem(ctx, workspaceID, itemID)
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
			parentComment, err := s.repo.GetComment(ctx, portal.WorkspaceID, item.ID, *input.ParentID)
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
		parentComment, err := s.repo.GetComment(ctx, input.WorkspaceID, input.ItemID, *input.ParentID)
		if err != nil {
			return CoreComment{}, err
		}
		if parentComment.ParentID != nil {
			return CoreComment{}, invalidInput("replies can only be one level deep")
		}
		parent = &parentComment
	}

	comment, err := s.repo.CreateComment(ctx, input)
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

func (s *Service) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (CoreVoteResult, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || userID == uuid.Nil {
		return CoreVoteResult{}, invalidInput("workspace, feedback, and user are required")
	}
	if vote != -1 && vote != 1 {
		return CoreVoteResult{}, invalidInput("feedback vote must be either -1 or 1")
	}
	item, err := s.repo.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return CoreVoteResult{}, err
	}
	if item.MergedIntoItemID != nil {
		return CoreVoteResult{}, ErrMergeConflict
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
	if item.MergedIntoItemID != nil {
		return CoreVoteResult{}, ErrMergeConflict
	}
	if input.Participant != nil {
		participant := *input.Participant
		if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.BlockedAt != nil || participant.Kind == ContributorKindAnonymous {
			return CoreVoteResult{}, ErrAuthenticationRequired
		}
		if portal.ParticipationMode == ParticipationModeAccountRequired {
			return CoreVoteResult{}, ErrParticipationNotAllowed
		}
		if input.Vote != -1 && input.Vote != 1 {
			return CoreVoteResult{}, invalidInput("feedback vote must be either -1 or 1")
		}
		return s.nextRepo.ToggleContributorVote(ctx, CoreContributorVoteInput{
			WorkspaceID: portal.WorkspaceID,
			ItemID:      item.ID,
			Participant: participant,
			Vote:        input.Vote,
		})
	}
	if input.UserID == uuid.Nil {
		return CoreVoteResult{}, ErrAuthenticationRequired
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
	item, err := s.repo.GetItem(ctx, input.WorkspaceID, input.ItemID)
	if err != nil {
		return CoreStoryLink{}, err
	}
	if item.MergedIntoItemID != nil {
		return CoreStoryLink{}, ErrMergeConflict
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
	if item.MergedIntoItemID != nil {
		return CoreCreateStoryResult{}, ErrMergeConflict
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

func isValidSubmissionSource(source string) bool {
	switch source {
	case SubmissionSourceInternal, SubmissionSourcePortal, SubmissionSourceWidget, SubmissionSourceIntegration:
		return true
	default:
		return false
	}
}

func isValidReviewerEmailFrequency(frequency string) bool {
	switch frequency {
	case EmailFrequencyOff, EmailFrequencyDaily, EmailFrequencyWeekly:
		return true
	default:
		return false
	}
}

func isValidParticipationMode(mode string) bool {
	return mode == ParticipationModeAccountRequired || mode == ParticipationModeVerifiedGuest || mode == ParticipationModeAnonymousAllowed
}

func isValidGuestIdentityPolicy(policy string) bool {
	switch policy {
	case GuestIdentityPolicyShowIdentity, GuestIdentityPolicyAllowPublicMasking, GuestIdentityPolicyAlwaysMaskGuests:
		return true
	default:
		return false
	}
}

func pointer[T any](value T) *T {
	return &value
}

func coreItemIDs(items []CoreItem) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}
