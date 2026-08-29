package jobs

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWeeklyDigestStatsRequireActionableSignal(t *testing.T) {
	for name, stats := range map[string]WeeklyDigestStats{
		"unread administrative updates only": {UnreadNotifications: 3},
		"team comments only":                 {TeamComments: 2},
		"unread mention":                     {UnreadPriorityNotifications: 1},
		"overdue work":                       {OverdueStories: 1},
		"strategy risk":                      {ObjectiveRisks: 1},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, name != "unread administrative updates only" && name != "team comments only", hasWeeklyDigestSignal(stats))
		})
	}
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
