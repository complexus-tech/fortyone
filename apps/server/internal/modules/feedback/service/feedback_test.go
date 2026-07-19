package feedback

import (
	"context"
	"database/sql"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type repoStub struct {
	portals        []CorePortal
	boards         []CoreBoard
	items          []CoreItem
	storyLinks     []CoreStoryLink
	statusID       uuid.UUID
	createdItems   []CoreItemInput
	createdPortals []CorePortalInput
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
	portal := CorePortal{ID: uuid.New(), WorkspaceID: input.WorkspaceID, Name: "City Roads Program", Slug: "city-roads", Description: input.Description, IsPublic: input.IsPublic}
	r.portals = append(r.portals, portal)
	return portal, nil
}

func (r *repoStub) UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input CorePortalInput) (CorePortal, error) {
	for index, portal := range r.portals {
		if portal.WorkspaceID == workspaceID && portal.ID == portalID {
			portal.Description = input.Description
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

func (r *repoStub) CreateBoard(ctx context.Context, input CoreBoardInput) (CoreBoard, error) {
	board := CoreBoard{ID: uuid.New(), WorkspaceID: input.WorkspaceID, PortalID: input.PortalID, TeamID: input.TeamID, Name: input.Name, Slug: input.Slug, Color: input.Color}
	r.boards = append(r.boards, board)
	return board, nil
}

func (r *repoStub) ListItems(ctx context.Context, input CoreListItemsInput) (CoreItemsPage, error) {
	result := make([]CoreItem, 0, len(r.items))
	for _, item := range r.items {
		if item.PortalID == input.PortalID {
			result = append(result, item)
		}
	}
	return CoreItemsPage{Items: result, HasMore: false}, nil
}

func (r *repoStub) ListComments(ctx context.Context, portalID uuid.UUID) ([]CoreComment, error) {
	return []CoreComment{}, nil
}

func (r *repoStub) ListStoryLinks(ctx context.Context, portalID uuid.UUID) ([]CoreStoryLink, error) {
	return r.storyLinks, nil
}

func (r *repoStub) GetItem(ctx context.Context, workspaceID, itemID uuid.UUID) (CoreItem, error) {
	for _, item := range r.items {
		if item.WorkspaceID == workspaceID && item.ID == itemID {
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

func (r *repoStub) UpdateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, input CoreUpdateItemStatusInput) (CoreItem, error) {
	item, err := r.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return CoreItem{}, err
	}
	item.Status = input.Status
	item.RoadmapSummary = input.RoadmapSummary
	return item, nil
}

func (r *repoStub) CreateComment(ctx context.Context, input CoreCommentInput) (CoreComment, error) {
	return CoreComment{ID: uuid.New(), WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, AuthorID: input.AuthorID, Body: input.Body}, nil
}

func (r *repoStub) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (CoreVoteResult, error) {
	return CoreVoteResult{Voted: true, VoteCount: 1}, nil
}

func (r *repoStub) LinkStory(ctx context.Context, input CoreStoryLinkInput) (CoreStoryLink, error) {
	link := CoreStoryLink{ID: uuid.New(), WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, StoryID: input.StoryID, Relationship: input.Relationship, CreatedByUserID: input.CreatedByUserID}
	r.storyLinks = append(r.storyLinks, link)
	return link, nil
}

func (r *repoStub) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	return &r.statusID, nil
}

type storyServiceStub struct {
	created []stories.CoreNewStory
}

func (s *storyServiceStub) CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	s.created = append(s.created, ns)
	return stories.CoreSingleStory{ID: uuid.New(), Title: ns.Title}, nil
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

func TestUpdatePortalCustomizesExistingWorkspacePortal(t *testing.T) {
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
		Description: "Share progress on public works requests.",
		IsPublic:    false,
	})

	require.NoError(t, err)
	require.Equal(t, "City Roads", portal.Name)
	require.Equal(t, "city-roads", portal.Slug)
	require.Equal(t, "Share progress on public works requests.", portal.Description)
	require.False(t, portal.IsPublic)
	require.Empty(t, repo.createdPortals)
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
	require.Equal(t, "repair-school-zone-signal-timing", item.Slug)
	require.Equal(t, "repair-school-zone-signal-timing", repo.createdItems[0].Slug)
}

func TestCreateStoryFromFeedbackLinksInternalStoryWithoutCompletingFeedback(t *testing.T) {
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
			Status:      StatusPlanned,
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
	require.Equal(t, itemID, repo.storyLinks[0].ItemID)
}
