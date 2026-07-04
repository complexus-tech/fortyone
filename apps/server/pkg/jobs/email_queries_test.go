package jobs

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOverdueStoriesDetailQueryIsWorkspaceScoped(t *testing.T) {
	query := overdueStoriesForAssigneeQuery()
	assigneeID := uuid.New()
	workspaceID := uuid.New()
	params := overdueStoriesForAssigneeParams(assigneeID, workspaceID)

	require.Contains(t, query, "s.workspace_id = :workspace_id")
	require.Equal(t, assigneeID, params["assignee_id"])
	require.Equal(t, workspaceID, params["workspace_id"])
}

func TestOverdueObjectivesDetailQueryIsWorkspaceScoped(t *testing.T) {
	query := overdueObjectivesForLeadQuery()
	leadID := uuid.New()
	workspaceID := uuid.New()
	params := overdueObjectivesForLeadParams(leadID, workspaceID)

	require.Contains(t, query, "o.workspace_id = :workspace_id")
	require.Equal(t, leadID, params["lead_id"])
	require.Equal(t, workspaceID, params["workspace_id"])
}

func TestFormatWeeklyDigestEmailContentSkipsZeroSections(t *testing.T) {
	stats := WeeklyDigestStats{
		UnreadNotifications:         2,
		UnreadPriorityNotifications: 1,
		OverdueStories:              0,
		DueThisWeekStories:          4,
		ObjectiveRisks:              0,
		TeamComments:                3,
	}

	rendered := formatWeeklyDigestEmailContent(stats)

	require.Contains(t, rendered, "2 unread updates")
	require.Contains(t, rendered, "4 assigned tasks due this week")
	require.Contains(t, rendered, "3 new team comments")
	require.False(t, strings.Contains(rendered, "0 overdue"))
	require.False(t, strings.Contains(rendered, "0 objectives"))
}
