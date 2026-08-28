package feedback

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type repoStub struct {
	portals                    []CorePortal
	boards                     []CoreBoard
	items                      []CoreItem
	similarItems               []CoreSimilarItem
	comments                   []CoreComment
	storyLinks                 []CoreStoryLink
	linkStoryErr               error
	linkStoryWinner            *CoreStoryLink
	statusID                   uuid.UUID
	createdItems               []CoreItemInput
	createdAnonymousItems      []CoreItemInput
	createdPortals             []CorePortalInput
	createdBoards              []CoreBoardInput
	listItemInputs             []CoreListItemsInput
	listCommentItemIDs         [][]uuid.UUID
	listStoryLinkItemIDs       [][]uuid.UUID
	reviewers                  []CoreBoardReviewer
	reviewerInputs             []CoreBoardReviewerInput
	reads                      map[string]time.Time
	contributors               []CoreContributor
	contributorPortals         map[uuid.UUID]uuid.UUID
	contributorComments        []CoreContributorComment
	contributorCommentsHasMore bool
	contributorCommentInputs   []CoreListContributorCommentsInput
	contributorActivityPage    CoreContributorActivityPage
	contributorActivityInputs  []CoreListContributorActivityInput
}

func feedbackReadKey(itemID, userID uuid.UUID) string {
	return itemID.String() + ":" + userID.String()
}

func TestGetPortalSnapshotHydratesOnlyRequestedItems(t *testing.T) {
	portalID := uuid.New()
	workspaceID := uuid.New()
	itemID := uuid.New()
	unrelatedItemID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{
			ID:          portalID,
			WorkspaceID: workspaceID,
			Slug:        "city-roads",
			IsPublic:    true,
		}},
		items: []CoreItem{{
			ID:          itemID,
			PortalID:    portalID,
			WorkspaceID: workspaceID,
			Status:      StatusPlanned,
		}},
		comments: []CoreComment{
			{ID: uuid.New(), ItemID: itemID},
			{ID: uuid.New(), ItemID: unrelatedItemID},
		},
		storyLinks: []CoreStoryLink{
			{ID: uuid.New(), ItemID: itemID},
			{ID: uuid.New(), ItemID: unrelatedItemID},
		},
	}
	service := New(repo, nil)

	snapshot, err := service.GetPortalSnapshot(context.Background(), "city-roads", CorePortalSnapshotInput{
		Page:     2,
		PageSize: 7,
		Sort:     "newest",
		Status:   StatusPlanned,
	})

	require.NoError(t, err)
	require.Len(t, snapshot.Items, 1)
	require.Equal(t, itemID, snapshot.Items[0].ID)
	require.Len(t, snapshot.Comments, 1)
	require.Equal(t, itemID, snapshot.Comments[0].ItemID)
	require.Len(t, snapshot.Links, 1)
	require.Equal(t, itemID, snapshot.Links[0].ItemID)
	require.Len(t, repo.listItemInputs, 1)
	require.Equal(t, portalID, repo.listItemInputs[0].PortalID)
	require.Equal(t, StatusPlanned, repo.listItemInputs[0].Status)
	require.Equal(t, 2, repo.listItemInputs[0].Page)
	require.Equal(t, 7, repo.listItemInputs[0].PageSize)
	require.Equal(t, [][]uuid.UUID{{itemID}}, repo.listCommentItemIDs)
	require.Equal(t, [][]uuid.UUID{{itemID}}, repo.listStoryLinkItemIDs)
}

func TestGetPortalSnapshotSummarySkipsItemHydration(t *testing.T) {
	portalID := uuid.New()
	itemID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, Slug: "city-roads", IsPublic: true}},
		items:   []CoreItem{{ID: itemID, PortalID: portalID}},
	}
	service := New(repo, nil)

	snapshot, err := service.GetPortalSnapshot(context.Background(), "city-roads", CorePortalSnapshotInput{
		PageSize:    20,
		SummaryOnly: true,
	})

	require.NoError(t, err)
	require.Len(t, snapshot.Items, 1)
	require.Empty(t, snapshot.Comments)
	require.Empty(t, snapshot.Links)
	require.Empty(t, repo.listCommentItemIDs)
	require.Empty(t, repo.listStoryLinkItemIDs)
}

func TestGetPortalSnapshotFiltersByExactItemID(t *testing.T) {
	t.Parallel()
	portalID, requestedID, otherID := uuid.New(), uuid.New(), uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, Slug: "city-roads", IsPublic: true}},
		items: []CoreItem{
			{ID: otherID, PortalID: portalID, Title: "Dark mode for reports"},
			{ID: requestedID, PortalID: portalID, Title: "Dark mode"},
		},
	}
	service := New(repo, nil)

	snapshot, err := service.GetPortalSnapshot(context.Background(), "city-roads", CorePortalSnapshotInput{
		ItemID: requestedID, PageSize: 20, SummaryOnly: true,
	})

	require.NoError(t, err)
	require.Len(t, snapshot.Items, 1)
	require.Equal(t, requestedID, snapshot.Items[0].ID)
	require.Equal(t, requestedID, repo.listItemInputs[0].ItemID)
}

func (r *repoStub) GetPortalBySlug(ctx context.Context, slug string) (CorePortal, error) {
	for _, portal := range r.portals {
		if portal.Slug == slug {
			return portal, nil
		}
	}
	return CorePortal{}, sql.ErrNoRows
}

func (r *repoStub) GetPrivateAuthor(_ context.Context, workspaceID, itemID uuid.UUID) (CorePrivateAuthor, error) {
	for _, item := range r.items {
		if item.WorkspaceID == workspaceID && item.ID == itemID {
			email := item.AuthorEmail
			return CorePrivateAuthor{
				ContributorID: item.ContributorID,
				Kind:          item.ParticipantKind,
				DisplayName:   item.AuthorName,
				Email:         &email,
				AvatarURL:     item.AuthorAvatar,
				PublicMasked:  item.AuthorMasked,
			}, nil
		}
	}
	return CorePrivateAuthor{}, sql.ErrNoRows
}

func (r *repoStub) ResolveCanonicalItem(_ context.Context, portalID uuid.UUID, itemReference string) (CoreCanonicalItem, error) {
	for _, item := range r.items {
		if item.PortalID == portalID && (item.ID.String() == itemReference || item.Slug == itemReference) {
			if item.MergedIntoItemID != nil {
				for _, target := range r.items {
					if target.ID == *item.MergedIntoItemID {
						return CoreCanonicalItem{ItemID: target.ID, ItemSlug: target.Slug, Merged: true}, nil
					}
				}
			}
			return CoreCanonicalItem{ItemID: item.ID, ItemSlug: item.Slug}, nil
		}
	}
	return CoreCanonicalItem{}, sql.ErrNoRows
}

func (r *repoStub) ListContributorActivity(ctx context.Context, input CoreListContributorActivityInput) (CoreContributorActivityPage, error) {
	r.contributorActivityInputs = append(r.contributorActivityInputs, input)
	page := r.contributorActivityPage
	page.Page = input.Page
	page.PageSize = input.PageSize
	return page, nil
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
	portal := CorePortal{ID: uuid.New(), WorkspaceID: input.WorkspaceID, Name: "City Roads Program", Slug: "city-roads"}
	if input.IsPublic != nil {
		portal.IsPublic = *input.IsPublic
	}
	if input.ParticipationMode != nil {
		portal.ParticipationMode = *input.ParticipationMode
	}
	r.portals = append(r.portals, portal)
	return portal, nil
}

func (r *repoStub) UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input CorePortalInput) (CorePortal, error) {
	for index, portal := range r.portals {
		if portal.WorkspaceID == workspaceID && portal.ID == portalID {
			if input.IsPublic != nil {
				portal.IsPublic = *input.IsPublic
			}
			if input.ParticipationMode != nil {
				portal.ParticipationMode = *input.ParticipationMode
			}
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
	r.createdBoards = append(r.createdBoards, input)
	board := CoreBoard{ID: uuid.New(), WorkspaceID: input.WorkspaceID, PortalID: input.PortalID, TeamID: input.TeamID, Name: input.Name, Slug: input.Slug, Color: input.Color}
	r.boards = append(r.boards, board)
	return board, nil
}

func (r *repoStub) DeleteBoard(ctx context.Context, workspaceID, boardID uuid.UUID) error {
	for index, board := range r.boards {
		if board.WorkspaceID == workspaceID && board.ID == boardID {
			r.boards = append(r.boards[:index], r.boards[index+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func TestCreateBoardRequiresAndPreservesCreator(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	portalID := uuid.New()
	teamID := uuid.New()
	creatorID := uuid.New()
	repo := &repoStub{}
	service := New(repo, nil)

	_, err := service.CreateBoard(context.Background(), CoreBoardInput{
		WorkspaceID: workspaceID,
		PortalID:    portalID,
		TeamID:      teamID,
		CreatorID:   creatorID,
		Name:        "Product feedback",
		Color:       "blue",
	})

	require.NoError(t, err)
	require.Len(t, repo.createdBoards, 1)
	require.Equal(t, creatorID, repo.createdBoards[0].CreatorID)

	_, err = service.CreateBoard(context.Background(), CoreBoardInput{
		WorkspaceID: workspaceID,
		PortalID:    portalID,
		TeamID:      teamID,
		Name:        "Missing creator",
	})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestDeleteBoardRequiresWorkspaceAndBoard(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	boardID := uuid.New()
	repo := &repoStub{boards: []CoreBoard{{ID: boardID, WorkspaceID: workspaceID}}}
	service := New(repo, nil)

	require.ErrorIs(t, service.DeleteBoard(context.Background(), uuid.Nil, boardID), ErrInvalidInput)
	require.ErrorIs(t, service.DeleteBoard(context.Background(), workspaceID, uuid.Nil), ErrInvalidInput)
	require.NoError(t, service.DeleteBoard(context.Background(), workspaceID, boardID))
	require.Empty(t, repo.boards)
}
