package feedback

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
	require.Equal(t, "A marked crossing would make school pickup safer.", storyService.created[0].Description)
	require.Equal(t, &statusID, storyService.created[0].StatusID)
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
	storyService := &storyServiceStub{stories: map[uuid.UUID]StoryPlan{
		storyID: {ID: storyID, WorkspaceID: workspaceID, TeamID: teamID},
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
	storyService := &storyServiceStub{stories: map[uuid.UUID]StoryPlan{
		storyID: {ID: storyID, WorkspaceID: workspaceID, TeamID: teamID, DeletedAt: &deletedAt},
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

	page, err := service.ListTeamItems(context.Background(), workspaceID, teamID, uuid.New(), "all", "", 1, 25)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, teamID, page.Items[0].Board.TeamID)
}

func TestListTeamItemsCarriesTrimmedSearch(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	repo := &repoStub{}
	service := New(repo, nil)

	_, err := service.ListTeamItems(
		context.Background(),
		workspaceID,
		teamID,
		uuid.New(),
		"active",
		"  export filters  ",
		1,
		25,
	)

	require.NoError(t, err)
	require.Len(t, repo.listItemInputs, 1)
	require.Equal(t, "export filters", repo.listItemInputs[0].Search)
}

func TestListTeamItemsScopesTrashToDeletedFeedback(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	deletedAt := time.Now()
	repo := &repoStub{items: []CoreItem{
		{ID: uuid.New(), WorkspaceID: workspaceID, Board: CoreBoard{TeamID: teamID}},
		{ID: uuid.New(), WorkspaceID: workspaceID, DeletedAt: &deletedAt, Board: CoreBoard{TeamID: teamID}},
	}}
	service := New(repo, nil)

	page, err := service.ListTeamItems(
		context.Background(),
		workspaceID,
		teamID,
		uuid.New(),
		ListStatusTrashed,
		"",
		1,
		25,
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.NotNil(t, page.Items[0].DeletedAt)
	require.True(t, repo.listItemInputs[0].DeletedOnly)
	require.Equal(t, "all", repo.listItemInputs[0].Status)
}

func TestListTeamItemsTreatsLegacyAllFilterAsActive(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	repo := &repoStub{}
	service := New(repo, nil)

	_, err := service.ListTeamItems(
		context.Background(),
		workspaceID,
		teamID,
		uuid.New(),
		"all",
		"",
		1,
		25,
	)

	require.NoError(t, err)
	require.Len(t, repo.listItemInputs, 1)
	require.Equal(t, "active", repo.listItemInputs[0].Status)
	require.False(t, repo.listItemInputs[0].DeletedOnly)
}

func TestTrashAndRestoreItemLifecycle(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	itemID := uuid.New()
	repo := &repoStub{items: []CoreItem{{ID: itemID, WorkspaceID: workspaceID}}}
	service := New(repo, nil)

	require.NoError(t, service.TrashItem(context.Background(), workspaceID, itemID))
	require.NotNil(t, repo.items[0].DeletedAt)
	require.NoError(t, service.RestoreItem(context.Background(), workspaceID, itemID))
	require.Nil(t, repo.items[0].DeletedAt)
	require.ErrorIs(t, service.TrashItem(context.Background(), uuid.Nil, itemID), ErrInvalidInput)
	require.ErrorIs(t, service.RestoreItem(context.Background(), workspaceID, uuid.Nil), ErrInvalidInput)
}

func TestTrashItemRejectsFeedbackManagedByPrimaryStory(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	itemID := uuid.New()
	repo := &repoStub{items: []CoreItem{{
		ID:          itemID,
		WorkspaceID: workspaceID,
		StoryLinks:  []CoreStoryLink{{IsPrimary: true}},
	}}}
	service := New(repo, nil)

	require.ErrorIs(t, service.TrashItem(context.Background(), workspaceID, itemID), ErrStoryManaged)
	require.Nil(t, repo.items[0].DeletedAt)
}

func TestTrashItemRejectsCanonicalTargetWithMergedSources(t *testing.T) {
	t.Parallel()

	workspaceID, portalID, sourceID, targetID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo := &repoStub{items: []CoreItem{
		{ID: sourceID, WorkspaceID: workspaceID, PortalID: portalID, MergedIntoItemID: &targetID},
		{ID: targetID, WorkspaceID: workspaceID, PortalID: portalID},
	}}
	service := New(repo, nil)

	require.ErrorIs(t, service.TrashItem(context.Background(), workspaceID, targetID), ErrMergeConflict)
	require.Nil(t, repo.items[1].DeletedAt)
}

func TestMergedSourceRejectsInternalMutationsButRemainsReadable(t *testing.T) {
	workspaceID, itemID, targetID := uuid.New(), uuid.New(), uuid.New()
	updatedAt := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	repo := &repoStub{items: []CoreItem{{
		ID: itemID, WorkspaceID: workspaceID, MergedIntoItemID: &targetID,
		Status: StatusPending, UpdatedAt: updatedAt,
	}}}
	service := New(repo, &storyServiceStub{})

	item, readErr := service.GetItem(context.Background(), workspaceID, itemID)
	_, statusErr := service.UpdateItemStatus(context.Background(), workspaceID, itemID, CoreUpdateItemStatusInput{Status: StatusCompleted, ActorID: uuid.New()})
	_, statusCASErr := service.UpdateItemStatusIfUnchanged(context.Background(), workspaceID, itemID, updatedAt, CoreUpdateItemStatusInput{Status: StatusCompleted, ActorID: uuid.New()})
	trashErr := service.TrashItem(context.Background(), workspaceID, itemID)
	restoreErr := service.RestoreItem(context.Background(), workspaceID, itemID)
	_, linkErr := service.LinkStory(context.Background(), CoreStoryLinkInput{
		WorkspaceID: workspaceID, ItemID: itemID, StoryID: uuid.New(), CreatedByUserID: uuid.New(),
	})
	_, createStoryErr := service.CreateStoryFromItem(context.Background(), workspaceID, itemID, uuid.New(), CoreCreateStoryInput{TeamID: uuid.New()})

	require.NoError(t, readErr)
	require.Equal(t, targetID, *item.MergedIntoItemID)
	for _, err := range []error{statusErr, statusCASErr, trashErr, restoreErr, linkErr, createStoryErr} {
		require.ErrorIs(t, err, ErrMergeConflict)
	}
	require.Equal(t, StatusPending, repo.items[0].Status)
	require.Nil(t, repo.items[0].DeletedAt)
	require.Empty(t, repo.storyLinks)
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
