package feedback

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
	require.Equal(t, SubmissionSourceInternal, repo.createdItems[0].Source)
}

func TestCreateItemRejectsUnsupportedSubmissionSource(t *testing.T) {
	service := New(&repoStub{}, nil)

	_, err := service.CreateItem(context.Background(), CoreItemInput{
		WorkspaceID: uuid.New(),
		PortalID:    uuid.New(),
		BoardID:     uuid.New(),
		AuthorID:    uuid.New(),
		Title:       "Repair school-zone signal timing",
		Source:      "email",
	})

	require.ErrorContains(t, err, "unsupported feedback submission source")
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

func TestCreateItemRejectsReservedPublicRoadmapSlug(t *testing.T) {
	service := New(&repoStub{}, nil)

	_, err := service.CreateItem(context.Background(), CoreItemInput{
		WorkspaceID: uuid.New(),
		PortalID:    uuid.New(),
		BoardID:     uuid.New(),
		AuthorID:    uuid.New(),
		Title:       "Roadmap",
		Slug:        "roadmap",
	})

	require.ErrorContains(t, err, "feedback slug roadmap is reserved")
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

	result, err := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug:          "city-roads",
		BoardID:             boardID,
		AuthorID:            authorID,
		Title:               "Repair the crossing signal",
		Source:              SubmissionSourcePortal,
		ParticipationIntent: ParticipationIntentAccount,
	})

	require.NoError(t, err)
	item := result.Item
	require.Equal(t, workspaceID, item.WorkspaceID)
	require.Equal(t, portalID, item.PortalID)
	require.Equal(t, authorID, item.AuthorID)
	require.Equal(t, workspaceID, repo.createdItems[0].WorkspaceID)
	require.Equal(t, SubmissionSourcePortal, repo.createdItems[0].Source)
}

func TestAnonymousPublicFeedbackCreatesUnidentifiedContributorItem(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	boardID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{
			ID:                portalID,
			WorkspaceID:       workspaceID,
			Slug:              "city-roads",
			IsPublic:          true,
			ParticipationMode: ParticipationModeAnonymousAllowed,
		}},
		boards: []CoreBoard{{ID: boardID, WorkspaceID: workspaceID, PortalID: portalID}},
	}
	service := New(repo, nil)

	result, err := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug:          "city-roads",
		BoardID:             boardID,
		AuthorID:            uuid.New(),
		Title:               "Repair the crossing signal",
		Source:              SubmissionSourceWidget,
		ParticipationIntent: ParticipationIntentAnonymous,
	})

	require.NoError(t, err)
	require.True(t, result.Anonymous)
	require.Equal(t, uuid.Nil, result.Item.AuthorID)
	require.Len(t, repo.createdAnonymousItems, 1)
	require.Len(t, repo.createdItems, 1)
	require.Equal(t, uuid.Nil, repo.createdItems[0].AuthorID)
	stored := repo.createdAnonymousItems[0]
	require.Regexp(t, `^repair-the-crossing-signal-[a-f0-9]{8}$`, stored.Slug)
	require.Equal(t, SubmissionSourceWidget, stored.Source)
}

func TestAnonymousPublicFeedbackHonorsAccountRequiredMode(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	boardID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{
			ID:                portalID,
			WorkspaceID:       workspaceID,
			Slug:              "city-roads",
			IsPublic:          true,
			ParticipationMode: ParticipationModeAccountRequired,
		}},
		boards: []CoreBoard{{ID: boardID, WorkspaceID: workspaceID, PortalID: portalID}},
	}
	service := New(repo, nil)

	_, err := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug:          "city-roads",
		BoardID:             boardID,
		Title:               "Repair the crossing signal",
		Source:              SubmissionSourcePortal,
		ParticipationIntent: ParticipationIntentAnonymous,
	})

	require.ErrorIs(t, err, ErrParticipationNotAllowed)
	require.Empty(t, repo.createdAnonymousItems)
	require.Empty(t, repo.createdItems)
}

func TestAccountPublicFeedbackRequiresAnAuthenticatedAuthor(t *testing.T) {
	repo := &repoStub{}
	service := New(repo, nil)

	_, err := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug:          "city-roads",
		BoardID:             uuid.New(),
		Title:               "Repair the crossing signal",
		Source:              SubmissionSourcePortal,
		ParticipationIntent: ParticipationIntentAccount,
	})

	require.ErrorIs(t, err, ErrAuthenticationRequired)
	require.Empty(t, repo.createdAnonymousItems)
	require.Empty(t, repo.createdItems)
}

func TestPublicFeedbackRejectsHighConfidenceDuplicate(t *testing.T) {
	workspaceID := uuid.New()
	portalID := uuid.New()
	boardID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		boards:  []CoreBoard{{ID: boardID, WorkspaceID: workspaceID, PortalID: portalID}},
		similarItems: []CoreSimilarItem{{
			ID:         uuid.New(),
			Title:      "Repair the crossing signal",
			Confidence: duplicateItemConfidence,
		}},
	}
	service := New(repo, nil)

	_, err := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug:          "city-roads",
		BoardID:             boardID,
		AuthorID:            uuid.New(),
		Title:               "Repair crossing signal",
		Source:              SubmissionSourcePortal,
		ParticipationIntent: ParticipationIntentAccount,
	})

	require.ErrorIs(t, err, ErrDuplicateItem)
	require.Empty(t, repo.createdItems)
}

func TestListPublicSimilarItemsMarksOnlyBlockingMatchesAsDuplicates(t *testing.T) {
	portalID := uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, Slug: "city-roads", IsPublic: true}},
		similarItems: []CoreSimilarItem{
			{ID: uuid.New(), Title: "Exact match", Confidence: duplicateItemConfidence},
			{ID: uuid.New(), Title: "Possible match", Confidence: duplicateItemConfidence - 0.01},
		},
	}
	service := New(repo, nil)

	items, err := service.ListPublicSimilarItems(context.Background(), "city-roads", "Signal timing", "", 10)

	require.NoError(t, err)
	require.Len(t, items, 2)
	require.True(t, items[0].IsDuplicate)
	require.False(t, items[1].IsDuplicate)
}

func TestListPublicSimilarItemsSkipsTitlesBelowTheSuggestionThreshold(t *testing.T) {
	service := New(&repoStub{}, nil)

	items, err := service.ListPublicSimilarItems(context.Background(), "city-roads", "create", "", 3)

	require.NoError(t, err)
	require.Empty(t, items)
}

func TestSetBoardReviewerNormalizesEmailFrequency(t *testing.T) {
	repo := &repoStub{}
	service := New(repo, nil)

	reviewer, err := service.SetBoardReviewer(context.Background(), CoreBoardReviewerInput{
		WorkspaceID:    uuid.New(),
		BoardID:        uuid.New(),
		UserID:         uuid.New(),
		EmailFrequency: " Weekly ",
	})

	require.NoError(t, err)
	require.Equal(t, EmailFrequencyWeekly, reviewer.EmailFrequency)
	require.Equal(t, EmailFrequencyWeekly, repo.reviewerInputs[0].EmailFrequency)
}

func TestSetBoardReviewerRejectsUnsupportedEmailFrequency(t *testing.T) {
	repo := &repoStub{}
	service := New(repo, nil)

	_, err := service.SetBoardReviewer(context.Background(), CoreBoardReviewerInput{
		WorkspaceID:    uuid.New(),
		BoardID:        uuid.New(),
		UserID:         uuid.New(),
		EmailFrequency: "hourly",
	})

	require.ErrorContains(t, err, "email frequency must be off, daily, or weekly")
	require.Empty(t, repo.reviewerInputs)
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
		PortalSlug:          "city-roads",
		BoardID:             repo.boards[0].ID,
		AuthorID:            uuid.New(),
		Title:               "Cross-portal feedback",
		Source:              SubmissionSourcePortal,
		ParticipationIntent: ParticipationIntentAccount,
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

	require.ErrorIs(t, commentErr, ErrNotFound)
	require.ErrorIs(t, voteErr, ErrNotFound)
}

func TestMergedSourceRejectsPublicAndAccountParticipation(t *testing.T) {
	workspaceID, portalID, itemID, targetID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "city-roads", IsPublic: true}},
		items: []CoreItem{{
			ID: itemID, WorkspaceID: workspaceID, PortalID: portalID, MergedIntoItemID: &targetID,
		}},
	}
	service := New(repo, nil)

	_, publicCommentErr := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "city-roads", ItemID: itemID, AuthorID: uuid.New(), Body: "Stale comment",
	})
	_, publicVoteErr := service.TogglePublicVote(context.Background(), CorePublicVoteInput{
		PortalSlug: "city-roads", ItemID: itemID, UserID: uuid.New(), Vote: 1,
	})
	_, accountCommentErr := service.CreateComment(context.Background(), CoreCommentInput{
		WorkspaceID: workspaceID, ItemID: itemID, AuthorID: uuid.New(), Body: "Stale comment",
	})
	_, accountVoteErr := service.ToggleVote(context.Background(), workspaceID, itemID, uuid.New(), 1)

	require.ErrorIs(t, publicCommentErr, ErrMergeConflict)
	require.ErrorIs(t, publicVoteErr, ErrMergeConflict)
	require.ErrorIs(t, accountCommentErr, ErrMergeConflict)
	require.ErrorIs(t, accountVoteErr, ErrMergeConflict)
	require.Empty(t, repo.comments)
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
