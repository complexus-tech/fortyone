package feedbackrepository

import (
	"fmt"
	"strings"
	"testing"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

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
	require.Contains(t, query, "fi.status = :status")
	require.Contains(t, query, "fi.board_id = :board_id")
	require.Contains(t, query, "ORDER BY vote_count DESC, fi.created_at DESC")
	require.Equal(t, portalID, params["portal_id"])
	require.Equal(t, boardID, params["board_id"])
	require.Equal(t, feedback.StatusReviewing, params["status"])
	require.Equal(t, `traffic lights -closed`, params["search"])
	require.Equal(t, 21, params["limit"])
	require.Equal(t, 20, params["offset"])
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
