package feedback

import (
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetPublicContributorReturnsPortalScopedStats(t *testing.T) {
	t.Parallel()

	portalID := uuid.New()
	authorID := uuid.New()
	joinedAt := time.Now().AddDate(-1, 0, 0)
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, Slug: "city-roads", IsPublic: true}},
		contributors: []CoreContributor{{
			ID:       authorID,
			Name:     "Ada Lovelace",
			JoinedAt: joinedAt,
			Stats: CoreContributorStats{
				FeedbackCount: 5,
				CommentCount:  8,
				VoteScore:     -2,
			},
		}},
		contributorPortals: map[uuid.UUID]uuid.UUID{authorID: portalID},
	}

	contributor, err := New(repo, nil).GetPublicContributor(context.Background(), " city-roads ", authorID)

	require.NoError(t, err)
	require.Equal(t, authorID, contributor.ID)
	require.Equal(t, "Ada Lovelace", contributor.Name)
	require.Equal(t, 5, contributor.Stats.FeedbackCount)
	require.Equal(t, 8, contributor.Stats.CommentCount)
	require.Equal(t, -2, contributor.Stats.VoteScore)
	require.Equal(t, joinedAt, contributor.JoinedAt)
}

func TestListContributorActivityNormalizesPagination(t *testing.T) {
	t.Parallel()

	activityID := uuid.New()
	repo := &repoStub{
		contributorActivityPage: CoreContributorActivityPage{
			Activities: []CoreContributorActivity{{
				ID:            activityID,
				Type:          "feedback",
				FeedbackTitle: "Safer crossing",
			}},
			FeedbackCount: 1,
		},
	}
	page, err := New(repo, nil).
		ListContributorActivity(context.Background(), uuid.New(), " feedback ", 0, 500)

	require.NoError(t, err)
	require.Len(t, page.Activities, 1)
	require.Equal(t, activityID, page.Activities[0].ID)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 50, page.PageSize)
	require.Len(t, repo.contributorActivityInputs, 1)
	require.Equal(t, 1, repo.contributorActivityInputs[0].Page)
	require.Equal(t, 50, repo.contributorActivityInputs[0].PageSize)
	require.Equal(t, "feedback", repo.contributorActivityInputs[0].ActivityType)
}

func TestListContributorActivityRequiresUser(t *testing.T) {
	t.Parallel()

	_, err := New(&repoStub{}, nil).
		ListContributorActivity(context.Background(), uuid.Nil, "", 1, 20)

	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestListContributorActivityRejectsUnknownType(t *testing.T) {
	t.Parallel()

	_, err := New(&repoStub{}, nil).
		ListContributorActivity(context.Background(), uuid.New(), "vote", 1, 20)

	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestListPublicContributorCommentsNormalizesPagination(t *testing.T) {
	t.Parallel()

	portalID := uuid.New()
	authorID := uuid.New()
	commentID := uuid.New()
	repo := &repoStub{
		portals:            []CorePortal{{ID: portalID, Slug: "city-roads", IsPublic: true}},
		contributors:       []CoreContributor{{ID: authorID, Name: "Ada Lovelace"}},
		contributorPortals: map[uuid.UUID]uuid.UUID{authorID: portalID},
		contributorComments: []CoreContributorComment{{
			ID:            commentID,
			ItemID:        uuid.New(),
			FeedbackTitle: "Repair the crossing",
			FeedbackSlug:  "repair-the-crossing",
			Body:          "This would make the crossing safer.",
		}},
		contributorCommentsHasMore: true,
	}

	page, err := New(repo, nil).ListPublicContributorComments(context.Background(), "city-roads", authorID, 0, 500)

	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 50, page.PageSize)
	require.True(t, page.HasMore)
	require.Len(t, page.Comments, 1)
	require.Equal(t, commentID, page.Comments[0].ID)
	require.Len(t, repo.contributorCommentInputs, 1)
	require.Equal(t, portalID, repo.contributorCommentInputs[0].PortalID)
	require.Equal(t, authorID, repo.contributorCommentInputs[0].AuthorID)
	require.Equal(t, 1, repo.contributorCommentInputs[0].Page)
	require.Equal(t, 50, repo.contributorCommentInputs[0].PageSize)
}

func TestPublicContributorMethodsRejectNilAuthorID(t *testing.T) {
	t.Parallel()

	service := New(&repoStub{}, nil)

	_, profileErr := service.GetPublicContributor(context.Background(), "city-roads", uuid.Nil)
	_, commentsErr := service.ListPublicContributorComments(context.Background(), "city-roads", uuid.Nil, 1, 20)

	require.ErrorIs(t, profileErr, ErrInvalidInput)
	require.ErrorIs(t, commentsErr, ErrInvalidInput)
}

func TestListPublicContributorCommentsRejectsUnknownContributor(t *testing.T) {
	t.Parallel()

	portalID := uuid.New()
	service := New(&repoStub{
		portals: []CorePortal{{ID: portalID, Slug: "city-roads", IsPublic: true}},
	}, nil)

	_, err := service.ListPublicContributorComments(context.Background(), "city-roads", uuid.New(), 1, 20)

	require.ErrorIs(t, err, ErrNotFound)
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
		IsPublic: pointer(false),
	})

	require.NoError(t, err)
	require.Equal(t, "City Roads", portal.Name)
	require.Equal(t, "city-roads", portal.Slug)
	require.False(t, portal.IsPublic)
	require.Empty(t, repo.createdPortals)
}

func TestUpdatePortalParticipationModePreservesAvailability(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	repo := &repoStub{portals: []CorePortal{{
		ID:                portalID,
		WorkspaceID:       workspaceID,
		IsPublic:          true,
		ParticipationMode: ParticipationModeAccountRequired,
	}}}
	service := New(repo, nil)

	portal, err := service.UpdatePortal(context.Background(), workspaceID, portalID, CorePortalInput{
		ParticipationMode: pointer(ParticipationModeAnonymousAllowed),
	})

	require.NoError(t, err)
	require.True(t, portal.IsPublic)
	require.Equal(t, ParticipationModeAnonymousAllowed, portal.ParticipationMode)
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

func TestCreatePublicCommentReplyTargetsTopLevelComment(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	itemID := uuid.New()
	feedbackAuthorID := uuid.New()
	parentAuthorID := uuid.New()
	replierID := uuid.New()
	parentID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		items: []CoreItem{{
			ID:          itemID,
			WorkspaceID: workspaceID,
			PortalID:    portalID,
			AuthorID:    feedbackAuthorID,
			Title:       "Safer school crossing",
			Slug:        "safer-school-crossing",
		}},
		comments: []CoreComment{{
			ID:          parentID,
			WorkspaceID: workspaceID,
			ItemID:      itemID,
			AuthorID:    parentAuthorID,
			Body:        "Traffic is worst after school.",
		}},
	}
	publisher := &eventPublisherStub{}
	service := New(repo, nil, WithEventPublisher(nil, publisher))

	comment, err := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "city-roads",
		ItemID:     itemID,
		AuthorID:   replierID,
		ParentID:   &parentID,
		Body:       "Thanks, this is helpful context.",
	})

	require.NoError(t, err)
	require.Equal(t, &parentID, comment.ParentID)
	require.Len(t, publisher.events, 2)

	recipients := make(map[uuid.UUID]events.FeedbackCommentCreatedPayload, len(publisher.events))
	for _, event := range publisher.events {
		payload := event.Payload.(events.FeedbackCommentCreatedPayload)
		recipients[payload.RecipientID] = payload
	}
	require.False(t, recipients[feedbackAuthorID].IsReply)
	require.True(t, recipients[parentAuthorID].IsReply)
}

func TestCreatePublicCommentRejectsNestedReply(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	itemID := uuid.New()
	topLevelID := uuid.New()
	replyID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		items:   []CoreItem{{ID: itemID, WorkspaceID: workspaceID, PortalID: portalID}},
		comments: []CoreComment{
			{ID: topLevelID, WorkspaceID: workspaceID, ItemID: itemID},
			{ID: replyID, WorkspaceID: workspaceID, ItemID: itemID, ParentID: &topLevelID},
		},
	}
	service := New(repo, nil)

	_, err := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "city-roads",
		ItemID:     itemID,
		AuthorID:   uuid.New(),
		ParentID:   &replyID,
		Body:       "This would be a second-level reply.",
	})

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Len(t, repo.comments, 2)
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

func TestDirectAndCASStatusNotificationPolicy(t *testing.T) {
	t.Parallel()

	text := func(value string) *string { return &value }
	tests := []struct {
		name                string
		initialStatus       string
		targetStatus        string
		existingExplanation *string
		inputExplanation    *string
		wantNotification    bool
	}{
		{name: "pending is internal triage", initialStatus: StatusPlanned, targetStatus: StatusPending},
		{name: "reviewing is internal triage", initialStatus: StatusPending, targetStatus: StatusReviewing},
		{name: "planned is public", initialStatus: StatusPending, targetStatus: StatusPlanned, wantNotification: true},
		{name: "in progress is public", initialStatus: StatusPlanned, targetStatus: StatusInProgress, wantNotification: true},
		{name: "completed is public", initialStatus: StatusInProgress, targetStatus: StatusCompleted, wantNotification: true},
		{
			name:          "closed without an explicit explanation is silent despite stale summary",
			initialStatus: StatusCompleted, targetStatus: StatusClosed,
			existingExplanation: text("An older public update"),
		},
		{
			name:          "closed with whitespace explanation is silent",
			initialStatus: StatusCompleted, targetStatus: StatusClosed,
			inputExplanation: text("   "),
		},
		{
			name:          "closed with transition explanation is public",
			initialStatus: StatusCompleted, targetStatus: StatusClosed,
			inputExplanation: text("We are not pursuing this because the platform changed."), wantNotification: true,
		},
	}
	writers := []struct {
		name string
		call func(*Service, uuid.UUID, uuid.UUID, time.Time, CoreUpdateItemStatusInput) (CoreItem, error)
	}{
		{
			name: "direct",
			call: func(service *Service, workspaceID, itemID uuid.UUID, _ time.Time, input CoreUpdateItemStatusInput) (CoreItem, error) {
				return service.UpdateItemStatus(context.Background(), workspaceID, itemID, input)
			},
		},
		{
			name: "compare and swap",
			call: func(service *Service, workspaceID, itemID uuid.UUID, updatedAt time.Time, input CoreUpdateItemStatusInput) (CoreItem, error) {
				return service.UpdateItemStatusIfUnchanged(context.Background(), workspaceID, itemID, updatedAt, input)
			},
		},
	}

	for _, writer := range writers {
		writer := writer
		t.Run(writer.name, func(t *testing.T) {
			for _, test := range tests {
				test := test
				t.Run(test.name, func(t *testing.T) {
					workspaceID, itemID, authorID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
					updatedAt := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
					repo := &repoStub{items: []CoreItem{{
						ID: itemID, WorkspaceID: workspaceID, AuthorID: authorID,
						Title: "Dark mode", Slug: "dark-mode", Status: test.initialStatus,
						RoadmapSummary: test.existingExplanation, UpdatedAt: updatedAt,
					}}}
					publisher := &eventPublisherStub{}
					service := New(repo, nil, WithEventPublisher(nil, publisher))

					item, err := writer.call(service, workspaceID, itemID, updatedAt, CoreUpdateItemStatusInput{
						Status: test.targetStatus, RoadmapSummary: test.inputExplanation, ActorID: actorID,
					})

					require.NoError(t, err)
					require.Equal(t, test.targetStatus, item.Status, "notification policy must not block status mutation")
					if test.wantNotification {
						require.Len(t, publisher.events, 1)
						require.Equal(t, test.targetStatus, publisher.events[0].Payload.(events.FeedbackStatusUpdatedPayload).Status)
					} else {
						require.Empty(t, publisher.events)
					}
				})
			}
		})
	}
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
