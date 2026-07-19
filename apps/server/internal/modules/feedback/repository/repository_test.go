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

func TestBuildListItemsQueryUsesFullTextSearchAndFilters(t *testing.T) {
	portalID := uuid.New()
	boardID := uuid.New()
	query, params := buildListItemsQuery(feedback.CoreListItemsInput{
		PortalID: portalID,
		BoardID:  &boardID,
		Status:   feedback.StatusReviewing,
		Search:   `traffic lights -closed`,
		Sort:     "top",
		Page:     2,
		PageSize: 20,
	})

	require.Contains(t, query, feedbackItemSearchVector+" @@ websearch_to_tsquery('english', :search)")
	require.Contains(t, query, projectedFeedbackStatus+" = :status")
	require.Contains(t, query, "fi.board_id = :board_id")
	require.Contains(t, query, "ORDER BY vote_count DESC, fi.created_at DESC")
	require.Equal(t, portalID, params["portal_id"])
	require.Equal(t, boardID, params["board_id"])
	require.Equal(t, feedback.StatusReviewing, params["status"])
	require.Equal(t, `traffic lights -closed`, params["search"])
	require.Equal(t, 21, params["limit"])
	require.Equal(t, 20, params["offset"])
}

func TestBuildListItemsQueryScopesTeamFeedbackAcrossBoards(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	query, params := buildListItemsQuery(feedback.CoreListItemsInput{
		WorkspaceID: workspaceID,
		TeamID:      &teamID,
		Status:      "active",
		Sort:        "newest",
		Page:        1,
		PageSize:    25,
	})

	require.Contains(t, query, "fi.workspace_id = :workspace_id")
	require.Contains(t, query, "fb.team_id = :team_id")
	require.Contains(t, query, projectedFeedbackStatus+" IN ('pending', 'reviewing')")
	require.NotContains(t, query, "fi.portal_id = :portal_id")
	require.Equal(t, workspaceID, params["workspace_id"])
	require.Equal(t, teamID, params["team_id"])
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
	require.Contains(t, itemSelectQuery(), "fsl.is_primary = true")
}

func TestToCoreItemIncludesPrimaryStoryLink(t *testing.T) {
	itemID := uuid.New()
	workspaceID := uuid.New()
	linkID := uuid.New()
	storyID := uuid.New()
	relationship := feedback.RelationshipCreatedFrom
	createdAt := time.Now()

	item := toCoreItem(itemRow{
		ID:               itemID,
		WorkspaceID:      workspaceID,
		PrimaryLinkID:    &linkID,
		PrimaryStoryID:   &storyID,
		PrimaryRelation:  &relationship,
		PrimaryCreatedAt: &createdAt,
	})

	require.Len(t, item.StoryLinks, 1)
	require.Equal(t, linkID, item.StoryLinks[0].ID)
	require.Equal(t, storyID, item.StoryLinks[0].StoryID)
	require.True(t, item.StoryLinks[0].IsPrimary)
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
