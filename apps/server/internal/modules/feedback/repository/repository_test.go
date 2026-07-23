package feedbackrepository

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestIsPrimaryStoryConflictRecognizesPGXConstraintError(t *testing.T) {
	err := fmt.Errorf("insert primary feedback story: %w", &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "feedback_story_links_one_primary_per_item",
	})

	require.True(t, isPrimaryStoryConflict(err))
	require.False(t, isPrimaryStoryConflict(errors.New("database unavailable")))
}

func TestIsBoardTeamConflictRecognizesPGXConstraintError(t *testing.T) {
	err := fmt.Errorf("insert feedback board: %w", &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "feedback_boards_workspace_team_unique",
	})

	require.True(t, isBoardTeamConflict(err))
	require.False(t, isBoardTeamConflict(errors.New("database unavailable")))
}

func TestItemSelectQuerySupportsNamedPaginationParameters(t *testing.T) {
	query := fmt.Sprintf(
		"%s WHERE fi.portal_id = :portal_id ORDER BY fi.created_at DESC LIMIT :limit OFFSET :offset",
		itemSelectQuery(),
	)

	boundQuery, args, err := sqlx.BindNamed(sqlx.DOLLAR, query, map[string]any{
		"portal_id": uuid.New(),
		"limit":     21,
		"offset":    0,
	})

	require.NoError(t, err)
	require.Len(t, args, 3)
	require.False(t, strings.Contains(boundQuery, ":"), "bound query contains an uncompiled named parameter")
}

func TestBuildListContributorCommentsQueryScopesAndPaginates(t *testing.T) {
	portalID := uuid.New()
	authorID := uuid.New()
	query, params := buildListContributorCommentsQuery(feedback.CoreListContributorCommentsInput{
		PortalID: portalID,
		AuthorID: authorID,
		Page:     3,
		PageSize: 20,
	})

	require.Contains(t, query, "fi.portal_id = :portal_id")
	require.Contains(t, query, "fc.author_id = :author_id")
	require.Contains(t, query, "ORDER BY fc.created_at DESC, fc.id DESC")
	require.Contains(t, query, "LIMIT :limit OFFSET :offset")
	require.Equal(t, portalID, params["portal_id"])
	require.Equal(t, authorID, params["author_id"])
	require.Equal(t, 21, params["limit"])
	require.Equal(t, 40, params["offset"])

	boundQuery, args, err := sqlx.BindNamed(sqlx.DOLLAR, query, params)
	require.NoError(t, err)
	require.Len(t, args, 4)
	require.False(t, strings.Contains(boundQuery, ":"), "bound query contains an uncompiled named parameter")
}

func TestBuildListItemsQueryUsesFullTextSearchAndFilters(t *testing.T) {
	portalID := uuid.New()
	boardID := uuid.New()
	authorID := uuid.New()
	query, params := buildListItemsQuery(feedback.CoreListItemsInput{
		PortalID: portalID,
		BoardID:  &boardID,
		AuthorID: authorID,
		Status:   feedback.StatusReviewing,
		Search:   `traffic lights -closed`,
		Sort:     "top",
		Page:     2,
		PageSize: 20,
	})

	require.Contains(t, query, feedbackItemSearchVector+" @@ websearch_to_tsquery('english', :search)")
	require.Contains(t, query, projectedFeedbackStatus+" = :status")
	require.Contains(t, query, "fi.board_id = :board_id")
	require.Contains(t, query, "fi.author_id = :author_id")
	require.Contains(t, query, "ORDER BY vote_count DESC, fi.created_at DESC")
	require.Equal(t, portalID, params["portal_id"])
	require.Equal(t, boardID, params["board_id"])
	require.Equal(t, authorID, params["author_id"])
	require.Equal(t, feedback.StatusReviewing, params["status"])
	require.Equal(t, `traffic lights -closed`, params["search"])
	require.Equal(t, 21, params["limit"])
	require.Equal(t, 20, params["offset"])
}

func TestBuildListItemsQueryScopesTeamFeedbackAcrossBoards(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	viewerID := uuid.New()
	query, params := buildListItemsQuery(feedback.CoreListItemsInput{
		WorkspaceID: workspaceID,
		TeamID:      &teamID,
		ViewerID:    viewerID,
		Status:      "active",
		Sort:        "newest",
		Page:        1,
		PageSize:    25,
	})

	require.Contains(t, query, "fi.workspace_id = :workspace_id")
	require.Contains(t, query, "fb.team_id = :team_id")
	require.Contains(t, query, "feedback_read.user_id = :viewer_id")
	require.Contains(t, query, "feedback_read.read_at")
	require.Contains(t, query, projectedFeedbackStatus+" IN ('pending', 'reviewing')")
	require.NotContains(t, query, "fi.portal_id = :portal_id")
	require.Equal(t, workspaceID, params["workspace_id"])
	require.Equal(t, teamID, params["team_id"])
	require.Equal(t, viewerID, params["viewer_id"])
}

func TestProjectedFeedbackStatusCoversEveryStoryCategory(t *testing.T) {
	expected := map[string]string{
		"backlog":   "reviewing",
		"unstarted": "planned",
		"started":   "in_progress",
		"paused":    "planned",
		"completed": "completed",
		"cancelled": "closed",
	}

	for category, status := range expected {
		require.Contains(t, projectedFeedbackStatus, "projected_state.category = '"+category+"'")
		require.Contains(t, projectedFeedbackStatus, "THEN '"+status+"'")
	}
	query := itemSelectQuery()
	require.Contains(t, query, "fsl.is_primary = true")
	require.Contains(t, query, "fv.direction = 1) AS integer) AS upvote_count")
	require.Contains(t, query, "fv.direction = -1) AS integer) AS downvote_count")
}

func TestToCoreItemIncludesPrimaryStoryLink(t *testing.T) {
	itemID := uuid.New()
	workspaceID := uuid.New()
	linkID := uuid.New()
	storyID := uuid.New()
	relationship := feedback.RelationshipCreatedFrom
	storyTitle := "Repair the traffic signals"
	readAt := time.Now().Add(-time.Minute)
	createdAt := time.Now()

	item := toCoreItem(itemRow{
		ID:                itemID,
		WorkspaceID:       workspaceID,
		PrimaryLinkID:     &linkID,
		PrimaryStoryID:    &storyID,
		PrimaryStoryTitle: &storyTitle,
		PrimaryRelation:   &relationship,
		PrimaryCreatedAt:  &createdAt,
		ReadAt:            &readAt,
		UpvoteCount:       8,
		DownvoteCount:     3,
	})

	require.Len(t, item.StoryLinks, 1)
	require.Equal(t, linkID, item.StoryLinks[0].ID)
	require.Equal(t, storyID, item.StoryLinks[0].StoryID)
	require.Equal(t, storyTitle, item.StoryLinks[0].StoryTitle)
	require.True(t, item.StoryLinks[0].IsPrimary)
	require.Equal(t, readAt, *item.ReadAt)
	require.Equal(t, 8, item.UpvoteCount)
	require.Equal(t, 3, item.DownvoteCount)
}

func TestToCoreBoardReviewerMapsDigestPreference(t *testing.T) {
	userID := uuid.New()
	reviewer := toCoreBoardReviewer(boardReviewerRow{
		UserID:         userID,
		Name:           "Ada Lovelace",
		Email:          "ada@example.com",
		Role:           "admin",
		EmailFrequency: feedback.EmailFrequencyDaily,
	})

	require.Equal(t, userID, reviewer.UserID)
	require.Equal(t, "Ada Lovelace", reviewer.Name)
	require.Equal(t, "ada@example.com", reviewer.Email)
	require.Equal(t, "admin", reviewer.Role)
	require.Equal(t, feedback.EmailFrequencyDaily, reviewer.EmailFrequency)
}

func TestToCoreContributorMapsProfileAndNetVoteScore(t *testing.T) {
	authorID := uuid.New()
	joinedAt := time.Now().AddDate(-1, 0, 0)
	avatar := "profiles/ada.webp"

	contributor := toCoreContributor(contributorRow{
		ID:            authorID,
		Name:          "Ada Lovelace",
		AvatarURL:     &avatar,
		JoinedAt:      joinedAt,
		FeedbackCount: 4,
		CommentCount:  7,
		VoteScore:     -3,
	})

	require.Equal(t, authorID, contributor.ID)
	require.Equal(t, "Ada Lovelace", contributor.Name)
	require.Equal(t, avatar, *contributor.AvatarURL)
	require.Equal(t, joinedAt, contributor.JoinedAt)
	require.Equal(t, 4, contributor.Stats.FeedbackCount)
	require.Equal(t, 7, contributor.Stats.CommentCount)
	require.Equal(t, -3, contributor.Stats.VoteScore)
}

func TestBuildListItemsQuerySortsFeedback(t *testing.T) {
	tests := []struct {
		name    string
		sort    string
		orderBy string
	}{
		{name: "top", sort: "top", orderBy: "vote_count DESC, fi.created_at DESC"},
		{name: "newest", sort: "newest", orderBy: "fi.created_at DESC"},
		{name: "oldest", sort: "oldest", orderBy: "fi.created_at ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, _ := buildListItemsQuery(feedback.CoreListItemsInput{
				PortalID: uuid.New(),
				Sort:     tt.sort,
				Page:     1,
				PageSize: 20,
			})

			require.Contains(t, query, "ORDER BY "+tt.orderBy)
		})
	}
}
