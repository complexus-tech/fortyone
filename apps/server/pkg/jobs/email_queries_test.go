package jobs

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOverdueStoriesDetailQueryIsWorkspaceScoped(t *testing.T) {
	query := overdueStoriesForAssigneeQuery()
	assigneeID := uuid.New()
	workspaceID := uuid.New()
	params := overdueStoriesForAssigneeParams(assigneeID, workspaceID)

	require.Contains(t, query, "s.workspace_id = :workspace_id")
	require.Contains(t, query, "w.deleted_at IS NULL")
	require.Contains(t, query, "JOIN workspace_members wm")
	require.Contains(t, query, "wm.role IN ('admin', 'member', 'guest')")
	require.Contains(t, query, "wm.role = 'admin'")
	require.Contains(t, query, "FROM team_members tm")
	require.Contains(t, query, "tm.team_id = s.team_id")
	require.NotContains(t, query, "JOIN team_members tm")
	require.Equal(t, assigneeID, params["assignee_id"])
	require.Equal(t, workspaceID, params["workspace_id"])
}

func TestOverdueObjectivesDetailQueryIsWorkspaceScoped(t *testing.T) {
	query := overdueObjectivesForLeadQuery()
	leadID := uuid.New()
	workspaceID := uuid.New()
	params := overdueObjectivesForLeadParams(leadID, workspaceID)

	require.Contains(t, query, "o.workspace_id = :workspace_id")
	require.Contains(t, query, "w.deleted_at IS NULL")
	require.Contains(t, query, "JOIN workspace_members wm")
	require.Contains(t, query, "wm.role IN ('admin', 'member', 'guest')")
	require.Contains(t, query, "wm.role = 'admin'")
	require.Contains(t, query, "FROM team_members tm")
	require.Contains(t, query, "tm.team_id = o.team_id")
	require.NotContains(t, query, "JOIN team_members tm")
	require.Equal(t, leadID, params["lead_id"])
	require.Equal(t, workspaceID, params["workspace_id"])
}

func TestScheduledGuidanceRecipientQueriesAllowAdminsOrCurrentTeamMembersIncludingGuests(t *testing.T) {
	queries := map[string]struct {
		query        string
		teamIDClause string
	}{
		"overdue tasks": {
			query:        overdueStoryRecipientsQuery(),
			teamIDClause: "tm.team_id = s.team_id",
		},
		"overdue objectives": {
			query:        overdueObjectiveRecipientsQuery(),
			teamIDClause: "tm.team_id = o.team_id",
		},
	}

	for name, expectation := range queries {
		t.Run(name, func(t *testing.T) {
			require.Contains(t, expectation.query, "wm.role IN ('admin', 'member', 'guest')")
			require.Contains(t, expectation.query, "wm.role = 'admin'")
			require.Contains(t, expectation.query, "OR EXISTS (")
			require.Contains(t, expectation.query, "FROM team_members tm")
			require.Contains(t, expectation.query, expectation.teamIDClause)
			require.NotContains(t, expectation.query, "JOIN team_members tm")
		})
	}
}

func TestScheduledGuidanceRecipientQueriesExcludeDeletedWorkspaces(t *testing.T) {
	queries := map[string]string{
		"weekly digest":      weeklyDigestRecipientsQuery(),
		"overdue tasks":      overdueStoryRecipientsQuery(),
		"overdue objectives": overdueObjectiveRecipientsQuery(),
	}

	for name, query := range queries {
		t.Run(name, func(t *testing.T) {
			require.Contains(t, query, "w.deleted_at IS NULL")
		})
	}
}

func TestOverdueObjectiveQueriesUseDirectionAwareKeyResultCompletion(t *testing.T) {
	for name, query := range map[string]string{
		"recipient query": overdueObjectiveRecipientsQuery(),
		"detail query":    overdueObjectivesForLeadQuery(),
	} {
		t.Run(name, func(t *testing.T) {
			require.Contains(t, query, "kr.target_value >= kr.start_value")
			require.Contains(t, query, "kr.current_value >= kr.target_value")
			require.Contains(t, query, "kr.target_value < kr.start_value")
			require.Contains(t, query, "kr.current_value <= kr.target_value")
		})
	}
}

func TestScheduledGuidanceRecipientPaginationOrdersByUserAndWorkspace(t *testing.T) {
	tests := map[string]struct {
		query   string
		orderBy string
	}{
		"overdue tasks": {
			query:   overdueStoryRecipientsQuery(),
			orderBy: "ORDER BY s.assignee_id, w.workspace_id",
		},
		"overdue objectives": {
			query:   overdueObjectiveRecipientsQuery(),
			orderBy: "ORDER BY o.lead_user_id, w.workspace_id",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Contains(t, test.query, test.orderBy)
			require.Less(t, strings.Index(test.query, test.orderBy), strings.Index(test.query, "LIMIT :batch_size OFFSET :offset"))
		})
	}
}

func TestWeeklyDigestStatsEnforceCurrentWorkspaceAndTeamAccess(t *testing.T) {
	query := weeklyDigestStatsQuery()

	require.Contains(t, query, "WITH recipient_access AS")
	require.Contains(t, query, "w.deleted_at IS NULL")
	require.Contains(t, query, "EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')")
	require.GreaterOrEqual(t, strings.Count(query, "EXISTS (SELECT 1 FROM recipient_access)"), 5)
	require.Contains(t, query, "OR s.team_id IN (SELECT team_id FROM visible_teams)")
	require.Contains(t, query, "OR o.team_id IN (SELECT team_id FROM visible_teams)")
}

func TestWeeklyDigestUnreadNotificationsUseCurrentEntityAccess(t *testing.T) {
	query := weeklyDigestStatsQuery()

	require.Contains(t, query, "accessible_notifications AS")
	require.Equal(t, 1, strings.Count(query, "FROM notifications n"))
	require.Equal(t, 2, strings.Count(query, "FROM accessible_notifications"))
	for _, entityType := range []string{"feedback", "story", "comment", "objective", "key_result", "strategy"} {
		require.Contains(t, query, "CAST(n.entity_type AS TEXT) = '"+entityType+"'")
	}
	require.Contains(t, query, "feedback.deleted_at IS NULL")
	require.Contains(t, query, "story.deleted_at IS NULL")
	require.Contains(t, query, "OR story.team_id IN (SELECT team_id FROM visible_teams)")
	require.Contains(t, query, "OR objective.team_id IN (SELECT team_id FROM visible_teams)")
	require.Contains(t, query, "OR n.message -> 'strategy' ->> 'kind' = 'weekly_check_in'")
	require.Contains(t, query, "notification.type IN ('mention', 'comment_reply')")
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

func TestFormatWeeklyDigestEmailContentUsesCompactNotificationRows(t *testing.T) {
	stats := WeeklyDigestStats{
		UnreadNotifications:         2,
		UnreadPriorityNotifications: 1,
		DueThisWeekStories:          4,
		TeamComments:                3,
	}

	rendered := formatWeeklyDigestEmailContent(stats)

	require.Contains(t, rendered, "Here is what needs attention this week:")
	require.Contains(t, rendered, "You have <strong")
	require.Contains(t, rendered, "2 unread updates, including 1 mention or reply")
	require.Contains(t, rendered, "4 assigned tasks due this week")
	require.Contains(t, rendered, "3 new team comments")
	require.Contains(t, rendered, "border-top: 0")
	require.Contains(t, rendered, "border-top: 1px solid #e5e4e2")
	require.NotContains(t, rendered, "<ul")
	require.NotContains(t, rendered, "<li")
	require.NotContains(t, rendered, "<h3")
}

func TestFormatOverdueStoriesEmailContentUsesCompactNotificationRows(t *testing.T) {
	dueSoon := OverdueStory{
		ID:         uuid.New(),
		Title:      "Design launch",
		EndDate:    time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC),
		TeamCode:   "PRD",
		SequenceID: 571,
	}
	dueToday := OverdueStory{
		ID:    uuid.New(),
		Title: "Prepare metrics",
	}
	overdue := OverdueStory{
		ID:             uuid.New(),
		Title:          "Review pricing",
		DaysDifference: 3,
	}

	rendered := formatOverdueStoriesEmailContent(dueSoon, []OverdueStory{dueSoon}, []OverdueStory{dueToday}, []OverdueStory{overdue}, "https://product.fortyone.app")

	require.Contains(t, rendered, "You have <strong")
	require.Contains(t, rendered, "3 tasks</strong> that need attention")
	require.Contains(t, rendered, ">Design launch</a> is due July 8, 2026")
	require.Contains(t, rendered, `href="https://product.fortyone.app/work/PRD-571"`)
	require.Contains(t, rendered, ">Prepare metrics</a> is due today")
	require.Contains(t, rendered, ">Review pricing</a> is <strong")
	require.Contains(t, rendered, "3 days</strong> overdue")
	require.Contains(t, rendered, "border-top: 0")
	require.Contains(t, rendered, "border-top: 1px solid #e5e4e2")
	require.NotContains(t, rendered, "<ul")
	require.NotContains(t, rendered, "<li")
	require.NotContains(t, rendered, "<h3")
	require.NotContains(t, rendered, " - ")
}

func TestFormatOverdueStoriesEmailContentCapsRowsAndAccountsForRemainder(t *testing.T) {
	stories := make([]OverdueStory, 15)
	for index := range stories {
		stories[index] = OverdueStory{
			ID:             uuid.New(),
			Title:          "Task " + strings.Repeat("x", index+1),
			DeadlineStatus: "overdue",
			DaysDifference: 2,
		}
	}

	rendered := formatOverdueStoriesEmailContent(stories[0], nil, nil, stories, "https://product.fortyone.app")

	require.Equal(t, maxGuidanceEmailRows-1, strings.Count(rendered, "<a href="))
	require.Contains(t, rendered, "15 tasks</strong> that need attention")
	require.Contains(t, rendered, "This email includes 11 of them; 4 more are available in assigned work")
	require.Contains(t, rendered, ">Task xxxxxxxxxxx</a>")
	require.NotContains(t, rendered, ">Task xxxxxxxxxxxx</a>")
}

func TestFormatOverdueObjectivesEmailContentUsesCompactNotificationRows(t *testing.T) {
	teamID := uuid.New()
	dueToday := OverdueObjective{
		ID:             uuid.New(),
		Name:           "Launch reporting",
		TeamID:         teamID,
		DeadlineStatus: "due_today",
	}
	keyResults, err := json.Marshal([]OverdueKeyResult{
		{
			ID:             uuid.New(),
			Name:           "Raise activation",
			EndDate:        "2026-07-01",
			DeadlineStatus: "overdue",
			DaysDifference: 2,
		},
	})
	require.NoError(t, err)
	needsAttention := OverdueObjective{
		ID:             uuid.New(),
		Name:           "Improve onboarding",
		TeamID:         teamID,
		DeadlineStatus: "future",
		KeyResults:     string(keyResults),
	}

	rendered := formatObjectiveOverdueEmailContent(dueToday, []OverdueObjective{needsAttention}, []OverdueObjective{dueToday}, nil, "https://product.fortyone.app")

	require.Contains(t, rendered, "You have <strong")
	require.Contains(t, rendered, "2 objectives</strong> that need attention")
	require.Contains(t, rendered, ">Improve onboarding</a> is on schedule, but key results need attention")
	require.Contains(t, rendered, ">Launch reporting</a> is due today")
	require.Contains(t, rendered, "Key result <a")
	require.Contains(t, rendered, ">Raise activation</a> is <strong")
	require.Contains(t, rendered, "2 days</strong> overdue")
	require.Contains(t, rendered, "border-top: 0")
	require.Contains(t, rendered, "border-top: 1px solid #e5e4e2")
	require.NotContains(t, rendered, "<ul")
	require.NotContains(t, rendered, "<li")
	require.NotContains(t, rendered, "<h3")
	require.NotContains(t, rendered, " - ")
}

func TestFormatOverdueObjectivesEmailContentCapsSignalsAndAccountsForRemainder(t *testing.T) {
	objectives := make([]OverdueObjective, 12)
	for index := range objectives {
		objectives[index] = OverdueObjective{
			ID:             uuid.New(),
			Name:           "Objective " + strings.Repeat("x", index+1),
			TeamID:         uuid.New(),
			DeadlineStatus: "overdue",
			DaysDifference: 2,
		}
	}

	rendered := formatObjectiveOverdueEmailContent(objectives[0], nil, nil, objectives, "https://product.fortyone.app")

	require.Equal(t, maxGuidanceEmailRows-1, strings.Count(rendered, "<a href="))
	require.Contains(t, rendered, "12 objectives</strong> that need attention")
	require.Contains(t, rendered, "They represent 12 objective and key-result signals")
	require.Contains(t, rendered, "This email includes 11 signals; 1 more is available in objectives")
	require.Contains(t, rendered, ">Objective xxxxxxxxxxx</a>")
	require.NotContains(t, rendered, ">Objective xxxxxxxxxxxx</a>")
}
