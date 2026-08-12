package notificationsrepository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkspaceNotificationQueriesShareCurrentAccessPredicate(t *testing.T) {
	predicate := workspaceNotificationAccessPredicate("notification")
	listQuery := listWorkspaceNotificationsQuery()
	unreadQuery := unreadWorkspaceNotificationsQuery()

	require.Contains(t, listQuery, predicate)
	require.Contains(t, unreadQuery, predicate)
	for _, query := range []string{listQuery, unreadQuery} {
		require.Contains(t, query, "notification.recipient_id = :user_id")
		require.Contains(t, query, "notification.workspace_id = :workspace_id")
		require.Contains(t, query, "CAST(notification.entity_type AS text) <> 'feedback'")
	}
	require.Contains(t, unreadQuery, "notification.read_at IS NULL")
	require.Contains(t, listQuery, "CAST(notification.entity_type AS text) <> 'strategy'")
	require.Contains(t, listQuery, "AND CAST(message AS text) ILIKE")
	require.NotContains(t, listQuery, "message - 'strategy'")
}

func TestWorkspaceNotificationAccessPredicateRechecksEntityAndTeamAccess(t *testing.T) {
	predicate := workspaceNotificationAccessPredicate("notification")
	normalized := strings.Join(strings.Fields(predicate), " ")

	require.Contains(t, normalized, "notification_member.workspace_id = notification.workspace_id")
	require.Contains(t, normalized, "notification_member.user_id = notification.recipient_id")
	require.Contains(t, normalized, "notification_member.role IN ('admin', 'member', 'guest')")

	require.Contains(t, normalized, "notification_story.id = notification.entity_id")
	require.Contains(t, normalized, "notification_story.workspace_id = notification.workspace_id")
	require.Contains(t, normalized, "notification_story.deleted_at IS NULL")
	require.Contains(t, normalized, "story_member.team_id = notification_story.team_id")

	require.Contains(t, normalized, "notification_comment.comment_id = notification.entity_id")
	require.Contains(t, normalized, "comment_member.team_id = notification_comment_story.team_id")

	require.Contains(t, normalized, "notification_objective.objective_id = notification.entity_id")
	require.Contains(t, normalized, "notification_objective.workspace_id = notification.workspace_id")
	require.Contains(t, normalized, "objective_member.team_id = notification_objective.team_id")

	require.Contains(t, normalized, "notification_key_result.id = notification.entity_id")
	require.Contains(t, normalized, "notification_key_result_objective.workspace_id = notification.workspace_id")
	require.Contains(t, normalized, "key_result_member.team_id = notification_key_result_objective.team_id")

	require.NotContains(t, predicate, "%!")
	require.NotContains(t, normalized, "CAST(notification.entity_type AS TEXT) <> 'strategy'")
}

func TestWorkspaceNotificationAccessPredicateKeepsPlanningAndMonthlyStrategyAdminOnly(t *testing.T) {
	predicate := strings.Join(strings.Fields(workspaceNotificationAccessPredicate("notification")), " ")
	strategyClause := "CAST(notification.entity_type AS TEXT) = 'strategy' AND ( notification_member.role = 'admin' OR notification.message -> 'strategy' ->> 'kind' = 'weekly_check_in' )"

	require.Contains(t, predicate, strategyClause)
	require.NotContains(t, predicate, "planning_reminder")
	require.NotContains(t, predicate, "monthly_summary")
	require.NotContains(t, predicate, "feedback")
}
