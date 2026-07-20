package feedback

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type repoStub struct {
	portals         []CorePortal
	boards          []CoreBoard
	items           []CoreItem
	storyLinks      []CoreStoryLink
	linkStoryErr    error
	linkStoryWinner *CoreStoryLink
	statusID        uuid.UUID
	createdItems    []CoreItemInput
	createdPortals  []CorePortalInput
	reads           map[string]time.Time
}

func feedbackReadKey(itemID, userID uuid.UUID) string {
	return itemID.String() + ":" + userID.String()
}

func (r *repoStub) GetPortalBySlug(ctx context.Context, slug string) (CorePortal, error) {
	for _, portal := range r.portals {
		if portal.Slug == slug {
			return portal, nil
		}
	}
	return CorePortal{}, sql.ErrNoRows
}

func (r *repoStub) GetPortalByWorkspaceSlugAndSlug(ctx context.Context, workspaceSlug, slug string) (CorePortal, error) {
	return r.GetPortalBySlug(ctx, slug)
}

func (r *repoStub) GetPortal(ctx context.Context, workspaceID, portalID uuid.UUID) (CorePortal, error) {
	for _, portal := range r.portals {
		if portal.WorkspaceID == workspaceID && portal.ID == portalID {
			return portal, nil
		}
	}
	return CorePortal{}, sql.ErrNoRows
}

func (r *repoStub) ListPortals(ctx context.Context, workspaceID uuid.UUID) ([]CorePortal, error) {
	result := make([]CorePortal, 0, len(r.portals))
	for _, portal := range r.portals {
		if portal.WorkspaceID == workspaceID {
			result = append(result, portal)
		}
	}
	return result, nil
}

func (r *repoStub) CreatePortal(ctx context.Context, input CorePortalInput) (CorePortal, error) {
	r.createdPortals = append(r.createdPortals, input)
	portal := CorePortal{ID: uuid.New(), WorkspaceID: input.WorkspaceID, Name: "City Roads Program", Slug: "city-roads", IsPublic: input.IsPublic}
	r.portals = append(r.portals, portal)
	return portal, nil
}

func (r *repoStub) UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input CorePortalInput) (CorePortal, error) {
	for index, portal := range r.portals {
		if portal.WorkspaceID == workspaceID && portal.ID == portalID {
			portal.IsPublic = input.IsPublic
			r.portals[index] = portal
			return portal, nil
		}
	}
	return CorePortal{}, sql.ErrNoRows
}

func (r *repoStub) ListBoards(ctx context.Context, portalID uuid.UUID) ([]CoreBoard, error) {
	result := make([]CoreBoard, 0, len(r.boards))
	for _, board := range r.boards {
		if board.PortalID == portalID {
			result = append(result, board)
		}
	}
	return result, nil
}

func (r *repoStub) GetBoard(ctx context.Context, portalID, boardID uuid.UUID) (CoreBoard, error) {
	for _, board := range r.boards {
		if board.PortalID == portalID && board.ID == boardID {
			return board, nil
		}
	}
	return CoreBoard{}, sql.ErrNoRows
}

func (r *repoStub) CreateBoard(ctx context.Context, input CoreBoardInput) (CoreBoard, error) {
	board := CoreBoard{ID: uuid.New(), WorkspaceID: input.WorkspaceID, PortalID: input.PortalID, TeamID: input.TeamID, Name: input.Name, Slug: input.Slug, Color: input.Color}
	r.boards = append(r.boards, board)
	return board, nil
}

func (r *repoStub) ListItems(ctx context.Context, input CoreListItemsInput) (CoreItemsPage, error) {
	result := make([]CoreItem, 0, len(r.items))
	for _, item := range r.items {
		if input.PortalID != uuid.Nil && item.PortalID != input.PortalID {
			continue
		}
		if input.TeamID != nil && item.Board.TeamID != *input.TeamID {
			continue
		}
		if input.ViewerID != uuid.Nil && r.reads != nil {
			if readAt, ok := r.reads[feedbackReadKey(item.ID, input.ViewerID)]; ok {
				item.ReadAt = &readAt
			}
		}
		result = append(result, item)
	}
	return CoreItemsPage{Items: result, HasMore: false}, nil
}

func (r *repoStub) ListComments(ctx context.Context, portalID uuid.UUID) ([]CoreComment, error) {
	return []CoreComment{}, nil
}

func (r *repoStub) ListItemComments(ctx context.Context, workspaceID, itemID uuid.UUID) ([]CoreComment, error) {
	return []CoreComment{}, nil
}

func (r *repoStub) ListStoryLinks(ctx context.Context, portalID uuid.UUID) ([]CoreStoryLink, error) {
	return r.storyLinks, nil
}

func (r *repoStub) ListItemStoryLinks(ctx context.Context, workspaceID, itemID uuid.UUID) ([]CoreStoryLink, error) {
	result := make([]CoreStoryLink, 0, len(r.storyLinks))
	for _, link := range r.storyLinks {
		if link.WorkspaceID == workspaceID && link.ItemID == itemID {
			result = append(result, link)
		}
	}
	return result, nil
}

func (r *repoStub) ListStoryFeedbackLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]CoreStoryFeedbackLink, error) {
	result := make([]CoreStoryFeedbackLink, 0)
	for _, link := range r.storyLinks {
		if link.WorkspaceID != workspaceID || link.StoryID != storyID || !link.IsPrimary {
			continue
		}
		for _, item := range r.items {
			if item.ID == link.ItemID {
				result = append(result, CoreStoryFeedbackLink{
					ID:            link.ID,
					WorkspaceID:   workspaceID,
					ItemID:        item.ID,
					StoryID:       storyID,
					TeamID:        item.Board.TeamID,
					FeedbackTitle: item.Title,
					Relationship:  link.Relationship,
					IsPrimary:     true,
					CreatedAt:     link.CreatedAt,
				})
			}
		}
	}
	return result, nil
}

func (r *repoStub) GetItem(ctx context.Context, workspaceID, itemID uuid.UUID) (CoreItem, error) {
	for _, item := range r.items {
		if item.WorkspaceID == workspaceID && item.ID == itemID {
			return item, nil
		}
	}
	return CoreItem{}, sql.ErrNoRows
}

func (r *repoStub) GetItemReadAt(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (*time.Time, error) {
	if r.reads == nil {
		return nil, nil
	}
	readAt, ok := r.reads[feedbackReadKey(itemID, userID)]
	if !ok {
		return nil, nil
	}
	return &readAt, nil
}

func (r *repoStub) ListTeamSummaries(ctx context.Context, workspaceID, userID uuid.UUID) ([]CoreTeamSummary, error) {
	byTeam := make(map[uuid.UUID]*CoreTeamSummary)
	for _, board := range r.boards {
		if board.WorkspaceID != workspaceID {
			continue
		}
		byTeam[board.TeamID] = &CoreTeamSummary{TeamID: board.TeamID, Enabled: true}
	}
	for _, item := range r.items {
		summary := byTeam[item.Board.TeamID]
		if summary == nil || item.WorkspaceID != workspaceID {
			continue
		}
		summary.TotalCount++
		if _, read := r.reads[feedbackReadKey(item.ID, userID)]; !read {
			summary.UnreadCount++
		}
	}
	result := make([]CoreTeamSummary, 0, len(byTeam))
	for _, summary := range byTeam {
		result = append(result, *summary)
	}
	return result, nil
}

func (r *repoStub) MarkItemRead(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (time.Time, error) {
	if r.reads == nil {
		r.reads = make(map[string]time.Time)
	}
	key := feedbackReadKey(itemID, userID)
	if readAt, ok := r.reads[key]; ok {
		return readAt, nil
	}
	readAt := time.Now()
	r.reads[key] = readAt
	return readAt, nil
}

func (r *repoStub) MarkItemUnread(ctx context.Context, workspaceID, itemID, userID uuid.UUID) error {
	delete(r.reads, feedbackReadKey(itemID, userID))
	return nil
}

func (r *repoStub) GetItemByPortal(ctx context.Context, portalID, itemID uuid.UUID) (CoreItem, error) {
	for _, item := range r.items {
		if item.PortalID == portalID && item.ID == itemID {
			return item, nil
		}
	}
	return CoreItem{}, sql.ErrNoRows
}

func (r *repoStub) CreateItem(ctx context.Context, input CoreItemInput) (CoreItem, error) {
	r.createdItems = append(r.createdItems, input)
	item := CoreItem{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		PortalID:    input.PortalID,
		BoardID:     input.BoardID,
		AuthorID:    input.AuthorID,
		Title:       input.Title,
		Description: input.Description,
		Slug:        input.Slug,
		Status:      StatusPending,
	}
	r.items = append(r.items, item)
	return item, nil
}

func (r *repoStub) UpdateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, input CoreUpdateItemStatusInput) (CoreItem, bool, error) {
	item, err := r.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return CoreItem{}, false, err
	}
	if !input.AllowLinked {
		for _, link := range item.StoryLinks {
			if link.IsPrimary {
				return CoreItem{}, false, ErrStoryManaged
			}
		}
	}
	statusChanged := item.Status != input.Status
	item.Status = input.Status
	if input.RoadmapSummary != nil {
		item.RoadmapSummary = input.RoadmapSummary
	}
	for index := range r.items {
		if r.items[index].ID == itemID {
			r.items[index] = item
			break
		}
	}
	return item, statusChanged, nil
}

func (r *repoStub) CreateComment(ctx context.Context, input CoreCommentInput) (CoreComment, error) {
	return CoreComment{ID: uuid.New(), WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, AuthorID: input.AuthorID, Body: input.Body}, nil
}

func (r *repoStub) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (CoreVoteResult, error) {
	return CoreVoteResult{Vote: vote, VoteCount: vote}, nil
}

func (r *repoStub) LinkStory(ctx context.Context, input CoreStoryLinkInput) (CoreStoryLink, error) {
	if r.linkStoryErr != nil {
		err := r.linkStoryErr
		r.linkStoryErr = nil
		if r.linkStoryWinner != nil {
			r.storyLinks = append(r.storyLinks, *r.linkStoryWinner)
		}
		return CoreStoryLink{}, err
	}
	link := CoreStoryLink{ID: uuid.New(), WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, StoryID: input.StoryID, Relationship: input.Relationship, IsPrimary: input.IsPrimary, CreatedByUserID: input.CreatedByUserID}
	r.storyLinks = append(r.storyLinks, link)
	return link, nil
}

func (r *repoStub) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	return &r.statusID, nil
}

func (r *repoStub) GetStatusCategory(ctx context.Context, teamID, statusID uuid.UUID) (string, error) {
	if statusID != r.statusID {
		return "", sql.ErrNoRows
	}
	return "unstarted", nil
}

type storyServiceStub struct {
	created []stories.CoreNewStory
	stories map[uuid.UUID]stories.CoreSingleStory
	deleted []uuid.UUID
}

type eventPublisherStub struct {
	events []events.Event
}

func (p *eventPublisherStub) Publish(_ context.Context, event events.Event) error {
	p.events = append(p.events, event)
	return nil
}

func (s *storyServiceStub) CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	s.created = append(s.created, ns)
	story := stories.CoreSingleStory{ID: uuid.New(), Title: ns.Title, Team: ns.Team, Workspace: workspaceID, Status: ns.Status}
	if s.stories == nil {
		s.stories = make(map[uuid.UUID]stories.CoreSingleStory)
	}
	s.stories[story.ID] = story
	return story, nil
}

func (s *storyServiceStub) Get(ctx context.Context, id uuid.UUID, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	story, ok := s.stories[id]
	if !ok || story.Workspace != workspaceID {
		return stories.CoreSingleStory{}, sql.ErrNoRows
	}
	return story, nil
}

func (s *storyServiceStub) Delete(ctx context.Context, id uuid.UUID, workspaceID uuid.UUID) error {
	s.deleted = append(s.deleted, id)
	delete(s.stories, id)
	return nil
}

func TestListPortalsCreatesWorkspaceDefaultPortalWhenMissing(t *testing.T) {
	workspaceID := uuid.New()
	repo := &repoStub{}
	service := New(repo, nil)

	portals, err := service.ListPortals(context.Background(), CoreWorkspacePortalInput{
		WorkspaceID:   workspaceID,
		WorkspaceName: "City Roads Program",
		WorkspaceSlug: "city-roads",
	})

	require.NoError(t, err)
	require.Len(t, portals, 1)
	require.Equal(t, "City Roads Program", portals[0].Name)
	require.Equal(t, "city-roads", portals[0].Slug)
	require.True(t, portals[0].IsPublic)
	require.Len(t, repo.createdPortals, 1)
}

func TestUpdatePortalAvailability(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{
			ID:          portalID,
			WorkspaceID: workspaceID,
			Name:        "City Roads",
			Slug:        "city-roads",
			IsPublic:    true,
		}},
	}
	service := New(repo, nil)

	portal, err := service.UpdatePortal(context.Background(), workspaceID, portalID, CorePortalInput{
		IsPublic: false,
	})

	require.NoError(t, err)
	require.Equal(t, "City Roads", portal.Name)
	require.Equal(t, "city-roads", portal.Slug)
	require.False(t, portal.IsPublic)
	require.Empty(t, repo.createdPortals)
}

func TestCreatePublicCommentPublishesAuthorNotificationEvent(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	itemID := uuid.New()
	authorID := uuid.New()
	commenterID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		items: []CoreItem{{
			ID:          itemID,
			WorkspaceID: workspaceID,
			PortalID:    portalID,
			AuthorID:    authorID,
			Title:       "Safer school crossing",
			Slug:        "safer-school-crossing",
		}},
	}
	publisher := &eventPublisherStub{}
	service := New(repo, nil, WithEventPublisher(nil, publisher))

	_, err := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "city-roads",
		ItemID:     itemID,
		AuthorID:   commenterID,
		Body:       "This is now under review.",
	})

	require.NoError(t, err)
	require.Len(t, publisher.events, 1)
	require.Equal(t, events.FeedbackCommentCreated, publisher.events[0].Type)
	payload := publisher.events[0].Payload.(events.FeedbackCommentCreatedPayload)
	require.Equal(t, authorID, payload.RecipientID)
	require.Equal(t, itemID, payload.FeedbackID)
	require.Equal(t, commenterID, publisher.events[0].ActorID)
}

func TestCreatePublicCommentDoesNotPublishSelfNotification(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	itemID := uuid.New()
	authorID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		items:   []CoreItem{{ID: itemID, WorkspaceID: workspaceID, PortalID: portalID, AuthorID: authorID}},
	}
	publisher := &eventPublisherStub{}
	service := New(repo, nil, WithEventPublisher(nil, publisher))

	_, err := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "city-roads",
		ItemID:     itemID,
		AuthorID:   authorID,
		Body:       "One more detail.",
	})

	require.NoError(t, err)
	require.Empty(t, publisher.events)
}

func TestUpdateItemStatusPublishesAuthorNotificationEvent(t *testing.T) {
	workspaceID := uuid.New()
	itemID := uuid.New()
	authorID := uuid.New()
	actorID := uuid.New()
	repo := &repoStub{items: []CoreItem{{
		ID:          itemID,
		WorkspaceID: workspaceID,
		AuthorID:    authorID,
		Title:       "Safer school crossing",
		Slug:        "safer-school-crossing",
		Status:      StatusPending,
	}}}
	publisher := &eventPublisherStub{}
	service := New(repo, nil, WithEventPublisher(nil, publisher))

	_, err := service.UpdateItemStatus(context.Background(), workspaceID, itemID, CoreUpdateItemStatusInput{
		Status:  StatusPlanned,
		ActorID: actorID,
	})

	require.NoError(t, err)
	require.Len(t, publisher.events, 1)
	require.Equal(t, events.FeedbackStatusUpdated, publisher.events[0].Type)
	payload := publisher.events[0].Payload.(events.FeedbackStatusUpdatedPayload)
	require.Equal(t, authorID, payload.RecipientID)
	require.Equal(t, StatusPlanned, payload.Status)
}

func TestUpdateItemStatusRejectsFeedbackManagedByPrimaryStory(t *testing.T) {
	workspaceID := uuid.New()
	itemID := uuid.New()
	repo := &repoStub{items: []CoreItem{{
		ID:          itemID,
		WorkspaceID: workspaceID,
		Status:      StatusPlanned,
		StoryLinks:  []CoreStoryLink{{ID: uuid.New(), StoryID: uuid.New(), IsPrimary: true}},
	}}}
	service := New(repo, nil)

	_, err := service.UpdateItemStatus(context.Background(), workspaceID, itemID, CoreUpdateItemStatusInput{
		Status:  StatusClosed,
		ActorID: uuid.New(),
	})

	require.ErrorIs(t, err, ErrStoryManaged)
	require.Equal(t, StatusPlanned, repo.items[0].Status)
}

func TestUpdateItemStatusDoesNotPublishWhenStatusIsUnchanged(t *testing.T) {
	workspaceID := uuid.New()
	itemID := uuid.New()
	repo := &repoStub{items: []CoreItem{{
		ID:          itemID,
		WorkspaceID: workspaceID,
		AuthorID:    uuid.New(),
		Title:       "Safer school crossing",
		Slug:        "safer-school-crossing",
		Status:      StatusPlanned,
	}}}
	publisher := &eventPublisherStub{}
	service := New(repo, nil, WithEventPublisher(nil, publisher))

	_, err := service.UpdateItemStatus(context.Background(), workspaceID, itemID, CoreUpdateItemStatusInput{
		Status:  StatusPlanned,
		ActorID: uuid.New(),
	})

	require.NoError(t, err)
	require.Empty(t, publisher.events)
}

func TestListPortalBoardsAllowsManagingDisabledPortal(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	boardID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{
			ID:          portalID,
			WorkspaceID: workspaceID,
			Name:        "City Roads",
			Slug:        "city-roads",
			IsPublic:    false,
		}},
		boards: []CoreBoard{{
			ID:          boardID,
			WorkspaceID: workspaceID,
			PortalID:    portalID,
			Name:        "Traffic lights",
			Slug:        "traffic-lights",
		}},
	}
	service := New(repo, nil)

	boards, err := service.ListPortalBoards(context.Background(), workspaceID, portalID)

	require.NoError(t, err)
	require.Len(t, boards, 1)
	require.Equal(t, boardID, boards[0].ID)
}

func TestCreateItemDefaultsToPendingAndGeneratesSlug(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	boardID := uuid.New()
	authorID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Name: "City Roads", Slug: "city-roads"}},
		boards:  []CoreBoard{{ID: boardID, WorkspaceID: workspaceID, PortalID: portalID, Name: "Traffic lights", Slug: "traffic-lights"}},
	}
	service := New(repo, nil)

	item, err := service.CreateItem(context.Background(), CoreItemInput{
		WorkspaceID: workspaceID,
		PortalID:    portalID,
		BoardID:     boardID,
		AuthorID:    authorID,
		Title:       "Repair school-zone signal timing",
		Description: "Morning crossing phase is too short.",
	})

	require.NoError(t, err)
	require.Equal(t, StatusPending, item.Status)
	require.Regexp(t, `^repair-school-zone-signal-timing-[a-f0-9]{8}$`, item.Slug)
	require.Equal(t, item.Slug, repo.createdItems[0].Slug)
}

func TestCreateItemPreservesProvidedSlug(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	boardID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID}},
		boards:  []CoreBoard{{ID: boardID, WorkspaceID: workspaceID, PortalID: portalID}},
	}
	service := New(repo, nil)

	item, err := service.CreateItem(context.Background(), CoreItemInput{
		WorkspaceID: workspaceID,
		PortalID:    portalID,
		BoardID:     boardID,
		AuthorID:    uuid.New(),
		Title:       "Repair school-zone signal timing",
		Description: "Morning crossing phase is too short.",
		Slug:        "custom-feedback-slug",
	})

	require.NoError(t, err)
	require.Equal(t, "custom-feedback-slug", item.Slug)
}

func TestPublicFeedbackWritesDeriveWorkspaceFromPortal(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	boardID := uuid.New()
	authorID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		boards:  []CoreBoard{{ID: boardID, WorkspaceID: workspaceID, PortalID: portalID}},
	}
	service := New(repo, nil)

	item, err := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug: "city-roads",
		BoardID:    boardID,
		AuthorID:   authorID,
		Title:      "Repair the crossing signal",
	})

	require.NoError(t, err)
	require.Equal(t, workspaceID, item.WorkspaceID)
	require.Equal(t, portalID, item.PortalID)
	require.Equal(t, authorID, item.AuthorID)
	require.Equal(t, workspaceID, repo.createdItems[0].WorkspaceID)
}

func TestPublicFeedbackRejectsBoardFromAnotherPortal(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		boards:  []CoreBoard{{ID: uuid.New(), WorkspaceID: workspaceID, PortalID: uuid.New()}},
	}
	service := New(repo, nil)

	_, err := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug: "city-roads",
		BoardID:    repo.boards[0].ID,
		AuthorID:   uuid.New(),
		Title:      "Cross-portal feedback",
	})

	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Empty(t, repo.createdItems)
}

func TestPublicFeedbackRejectsOversizedContent(t *testing.T) {
	service := New(&repoStub{}, nil)

	_, titleErr := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		Title: strings.Repeat("a", maxPublicFeedbackTitleCharacters+1),
	})
	_, descriptionErr := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		Title:       "Valid title",
		Description: strings.Repeat("a", maxPublicFeedbackDescriptionCharacters+1),
	})
	_, commentErr := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		Body: strings.Repeat("a", maxPublicFeedbackCommentCharacters+1),
	})

	require.ErrorContains(t, titleErr, "200 characters or fewer")
	require.ErrorContains(t, descriptionErr, "20000 characters or fewer")
	require.ErrorContains(t, commentErr, "10000 characters or fewer")
}

func TestPublicCommentAndVoteRejectItemFromAnotherPortal(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	itemID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		items:   []CoreItem{{ID: itemID, WorkspaceID: workspaceID, PortalID: uuid.New()}},
	}
	service := New(repo, nil)

	_, commentErr := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "city-roads",
		ItemID:     itemID,
		AuthorID:   uuid.New(),
		Body:       "This should not be accepted.",
	})
	_, voteErr := service.TogglePublicVote(context.Background(), CorePublicVoteInput{
		PortalSlug: "city-roads",
		ItemID:     itemID,
		UserID:     uuid.New(),
		Vote:       1,
	})

	require.ErrorIs(t, commentErr, sql.ErrNoRows)
	require.ErrorIs(t, voteErr, sql.ErrNoRows)
}

func TestPublicCommentAndVoteRejectItemFromAnotherWorkspace(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	itemID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		items:   []CoreItem{{ID: itemID, WorkspaceID: uuid.New(), PortalID: portalID}},
	}
	service := New(repo, nil)

	_, commentErr := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "city-roads",
		ItemID:     itemID,
		AuthorID:   uuid.New(),
		Body:       "This should not be accepted.",
	})
	_, voteErr := service.TogglePublicVote(context.Background(), CorePublicVoteInput{
		PortalSlug: "city-roads",
		ItemID:     itemID,
		UserID:     uuid.New(),
		Vote:       1,
	})

	require.ErrorIs(t, commentErr, sql.ErrNoRows)
	require.ErrorIs(t, voteErr, sql.ErrNoRows)
}

func TestToggleVoteSupportsUpvotesAndDownvotes(t *testing.T) {
	workspaceID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()
	repo := &repoStub{items: []CoreItem{{ID: itemID, WorkspaceID: workspaceID}}}
	service := New(repo, nil)

	upvote, err := service.ToggleVote(context.Background(), workspaceID, itemID, userID, 1)
	require.NoError(t, err)
	require.Equal(t, 1, upvote.Vote)

	downvote, err := service.ToggleVote(context.Background(), workspaceID, itemID, userID, -1)
	require.NoError(t, err)
	require.Equal(t, -1, downvote.Vote)

	_, err = service.ToggleVote(context.Background(), workspaceID, itemID, userID, 0)
	require.ErrorContains(t, err, "either -1 or 1")
}

func TestCreateStoryFromFeedbackPlansAndLinksInternalStory(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	boardID := uuid.New()
	itemID := uuid.New()
	actorID := uuid.New()
	teamID := uuid.New()
	statusID := uuid.New()
	repo := &repoStub{
		statusID: statusID,
		items: []CoreItem{{
			ID:          itemID,
			WorkspaceID: workspaceID,
			PortalID:    portalID,
			BoardID:     boardID,
			AuthorID:    actorID,
			Title:       "Add pedestrian crossing",
			Description: "A marked crossing would make school pickup safer.",
			Status:      StatusPending,
			Board:       CoreBoard{ID: boardID, WorkspaceID: workspaceID, PortalID: portalID, TeamID: teamID},
		}},
	}
	storyService := &storyServiceStub{}
	service := New(repo, storyService)

	result, err := service.CreateStoryFromItem(context.Background(), workspaceID, itemID, actorID, CoreCreateStoryInput{
		TeamID: teamID,
	})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, result.StoryID)
	require.Len(t, storyService.created, 1)
	require.Equal(t, "Add pedestrian crossing", storyService.created[0].Title)
	require.Equal(t, "A marked crossing would make school pickup safer.", *storyService.created[0].Description)
	require.Equal(t, &statusID, storyService.created[0].Status)
	require.Equal(t, RelationshipCreatedFrom, repo.storyLinks[0].Relationship)
	require.True(t, repo.storyLinks[0].IsPrimary)
	require.Equal(t, itemID, repo.storyLinks[0].ItemID)
	require.True(t, result.Created)
	require.Equal(t, StatusPlanned, repo.items[0].Status)
}

func TestCreateStoryFromFeedbackLinksExistingStoryInSameTeam(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	itemID := uuid.New()
	storyID := uuid.New()
	repo := &repoStub{items: []CoreItem{{
		ID:          itemID,
		WorkspaceID: workspaceID,
		Status:      StatusReviewing,
		Board:       CoreBoard{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID},
	}}}
	storyService := &storyServiceStub{stories: map[uuid.UUID]stories.CoreSingleStory{
		storyID: {ID: storyID, Workspace: workspaceID, Team: teamID},
	}}
	service := New(repo, storyService)

	result, err := service.CreateStoryFromItem(context.Background(), workspaceID, itemID, uuid.New(), CoreCreateStoryInput{
		TeamID:  teamID,
		StoryID: &storyID,
	})

	require.NoError(t, err)
	require.False(t, result.Created)
	require.Equal(t, storyID, result.StoryID)
	require.Len(t, repo.storyLinks, 1)
	require.Equal(t, RelationshipSolves, repo.storyLinks[0].Relationship)
	require.True(t, repo.storyLinks[0].IsPrimary)
	require.Equal(t, StatusPlanned, repo.items[0].Status)
}

func TestCreateStoryFromFeedbackRejectsCrossTeamPlanning(t *testing.T) {
	workspaceID := uuid.New()
	itemID := uuid.New()
	itemTeamID := uuid.New()
	repo := &repoStub{items: []CoreItem{{
		ID:          itemID,
		WorkspaceID: workspaceID,
		Board:       CoreBoard{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: itemTeamID},
	}}}
	service := New(repo, &storyServiceStub{})

	_, err := service.CreateStoryFromItem(context.Background(), workspaceID, itemID, uuid.New(), CoreCreateStoryInput{
		TeamID: uuid.New(),
	})

	require.ErrorIs(t, err, ErrTeamMismatch)
}

func TestCreateStoryFromFeedbackRejectsDeletedExistingStory(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	itemID := uuid.New()
	storyID := uuid.New()
	deletedAt := time.Now()
	repo := &repoStub{items: []CoreItem{{
		ID:          itemID,
		WorkspaceID: workspaceID,
		Board:       CoreBoard{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID},
	}}}
	storyService := &storyServiceStub{stories: map[uuid.UUID]stories.CoreSingleStory{
		storyID: {ID: storyID, Workspace: workspaceID, Team: teamID, DeletedAt: &deletedAt},
	}}
	service := New(repo, storyService)

	_, err := service.CreateStoryFromItem(context.Background(), workspaceID, itemID, uuid.New(), CoreCreateStoryInput{
		TeamID:  teamID,
		StoryID: &storyID,
	})

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Empty(t, repo.storyLinks)
}

func TestCreateStoryFromFeedbackReturnsExistingPrimaryLinkOnRetry(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	itemID := uuid.New()
	storyID := uuid.New()
	linkID := uuid.New()
	repo := &repoStub{
		items: []CoreItem{{
			ID:          itemID,
			WorkspaceID: workspaceID,
			Status:      StatusPending,
			Board:       CoreBoard{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID},
		}},
		storyLinks: []CoreStoryLink{{
			ID:           linkID,
			WorkspaceID:  workspaceID,
			ItemID:       itemID,
			StoryID:      storyID,
			Relationship: RelationshipCreatedFrom,
			IsPrimary:    true,
		}},
	}
	storyService := &storyServiceStub{}
	service := New(repo, storyService)

	result, err := service.CreateStoryFromItem(context.Background(), workspaceID, itemID, uuid.New(), CoreCreateStoryInput{
		TeamID: teamID,
	})

	require.NoError(t, err)
	require.False(t, result.Created)
	require.Equal(t, storyID, result.StoryID)
	require.Equal(t, linkID, result.LinkID)
	require.Empty(t, storyService.created)
	require.Equal(t, StatusPending, repo.items[0].Status)
}

func TestCreateStoryFromFeedbackCompensatesConcurrentDuplicate(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	itemID := uuid.New()
	winnerStoryID := uuid.New()
	winnerLinkID := uuid.New()
	repo := &repoStub{
		statusID: uuid.New(),
		items: []CoreItem{{
			ID:          itemID,
			WorkspaceID: workspaceID,
			Title:       "Add export filters",
			Status:      StatusPending,
			Board:       CoreBoard{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID},
		}},
		linkStoryErr: ErrAlreadyPlanned,
		linkStoryWinner: &CoreStoryLink{
			ID:           winnerLinkID,
			WorkspaceID:  workspaceID,
			ItemID:       itemID,
			StoryID:      winnerStoryID,
			Relationship: RelationshipCreatedFrom,
			IsPrimary:    true,
		},
	}
	storyService := &storyServiceStub{}
	service := New(repo, storyService)

	result, err := service.CreateStoryFromItem(context.Background(), workspaceID, itemID, uuid.New(), CoreCreateStoryInput{
		TeamID: teamID,
	})

	require.NoError(t, err)
	require.False(t, result.Created)
	require.Equal(t, winnerStoryID, result.StoryID)
	require.Equal(t, winnerLinkID, result.LinkID)
	require.Len(t, storyService.created, 1)
	require.Len(t, storyService.deleted, 1)
}

func TestListTeamItemsScopesFeedbackToBoardTeam(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	otherTeamID := uuid.New()
	repo := &repoStub{items: []CoreItem{
		{ID: uuid.New(), WorkspaceID: workspaceID, Board: CoreBoard{TeamID: teamID}},
		{ID: uuid.New(), WorkspaceID: workspaceID, Board: CoreBoard{TeamID: otherTeamID}},
	}}
	service := New(repo, nil)

	page, err := service.ListTeamItems(context.Background(), workspaceID, teamID, uuid.New(), "all", 1, 25)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, teamID, page.Items[0].Board.TeamID)
}

func TestFeedbackReadStateIsPerUserAndDrivesTeamSummary(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	boardID := uuid.New()
	itemID := uuid.New()
	firstUserID := uuid.New()
	secondUserID := uuid.New()
	repo := &repoStub{
		boards: []CoreBoard{{ID: boardID, WorkspaceID: workspaceID, TeamID: teamID}},
		items: []CoreItem{{
			ID:          itemID,
			WorkspaceID: workspaceID,
			BoardID:     boardID,
			Board:       CoreBoard{ID: boardID, WorkspaceID: workspaceID, TeamID: teamID},
		}},
	}
	service := New(repo, nil)

	readAt, err := service.MarkItemRead(context.Background(), workspaceID, itemID, firstUserID)
	require.NoError(t, err)
	require.NotNil(t, readAt)
	secondReadAt, err := service.MarkItemRead(context.Background(), workspaceID, itemID, firstUserID)
	require.NoError(t, err)
	require.Equal(t, *readAt, *secondReadAt)

	firstDetails, err := service.GetItemDetails(context.Background(), workspaceID, itemID, firstUserID)
	require.NoError(t, err)
	require.Equal(t, *readAt, *firstDetails.Item.ReadAt)

	secondDetails, err := service.GetItemDetails(context.Background(), workspaceID, itemID, secondUserID)
	require.NoError(t, err)
	require.Nil(t, secondDetails.Item.ReadAt)

	firstSummaries, err := service.ListTeamSummaries(context.Background(), workspaceID, firstUserID)
	require.NoError(t, err)
	require.Equal(t, 1, firstSummaries[0].TotalCount)
	require.Zero(t, firstSummaries[0].UnreadCount)

	secondSummaries, err := service.ListTeamSummaries(context.Background(), workspaceID, secondUserID)
	require.NoError(t, err)
	require.Equal(t, 1, secondSummaries[0].UnreadCount)

	require.NoError(t, service.MarkItemUnread(context.Background(), workspaceID, itemID, firstUserID))
	firstSummaries, err = service.ListTeamSummaries(context.Background(), workspaceID, firstUserID)
	require.NoError(t, err)
	require.Equal(t, 1, firstSummaries[0].UnreadCount)
}

func TestListStoryFeedbackLinksReturnsPrimaryFeedbackMetadata(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	itemID := uuid.New()
	storyID := uuid.New()
	repo := &repoStub{
		items: []CoreItem{{
			ID:          itemID,
			WorkspaceID: workspaceID,
			Title:       "Add monthly CSV exports",
			Board:       CoreBoard{TeamID: teamID},
		}},
		storyLinks: []CoreStoryLink{{
			ID:           uuid.New(),
			WorkspaceID:  workspaceID,
			ItemID:       itemID,
			StoryID:      storyID,
			Relationship: RelationshipCreatedFrom,
			IsPrimary:    true,
		}},
	}
	service := New(repo, nil)

	links, err := service.ListStoryFeedbackLinks(context.Background(), workspaceID, storyID)

	require.NoError(t, err)
	require.Len(t, links, 1)
	require.Equal(t, teamID, links[0].TeamID)
	require.Equal(t, "Add monthly CSV exports", links[0].FeedbackTitle)
}

func TestFeedbackStatusForStoryCategory(t *testing.T) {
	tests := map[string]string{
		"backlog":   StatusReviewing,
		"unstarted": StatusPlanned,
		"started":   StatusInProgress,
		"paused":    StatusPlanned,
		"completed": StatusCompleted,
		"cancelled": StatusClosed,
	}

	for category, expected := range tests {
		t.Run(category, func(t *testing.T) {
			require.Equal(t, expected, feedbackStatusForStoryCategory(category))
		})
	}
}
