package mayarepository

import (
	"os"
	"strings"
	"testing"
)

func TestWorkspaceMayaAccessPreservesTrialAndPaidSubscriptionPolicy(t *testing.T) {
	t.Parallel()

	query := readNamedRepositoryQuery(t, "access.sql", "WorkspaceCanUseMaya")
	for _, fragment := range []string{
		"workspace.workspace_id = sqlc.arg(workspace_id)",
		"workspace.deleted_at is null",
		"workspace.trial_ends_on > current_timestamp",
		"subscription.workspace_id = workspace.workspace_id",
		"subscription.subscription_tier <> 'free'",
		"subscription.subscription_status in ('active', 'trialing', 'past_due')",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("Maya entitlement query must contain %q: %s", fragment, query)
		}
	}
}

func TestScheduleRecoveryClaimIsBoundedFairAndConcurrencySafe(t *testing.T) {
	t.Parallel()
	query := readNamedRepositoryQuery(t, "scheduling.sql", "ClaimScheduleRecoveryStoryRefs")
	if strings.Contains(query, "status.deleted_at") {
		t.Fatal("schedule recovery must only reference columns defined by statuses")
	}
	for _, fragment := range []string{
		"coalesce(ownership.recovery_attempted_at, timestamp 'epoch')",
		"<= sqlc.arg(retry_before)",
		"story.updated_at > ownership.updated_at",
		"team_settings.updated_at > ownership.updated_at",
		"sprint.updated_at > ownership.updated_at",
		"story.assignee_id <> ownership.user_id",
		"run.status = 'running'",
		"order by",
		"limit cast(sqlc.arg(row_limit) as integer)",
		"for update of ownership skip locked",
		"set recovery_attempted_at = current_timestamp",
		"from calendar_schedule_blocks elapsed_block",
		"from calendar_schedule_blocks future_block",
		"future_block.end_at > current_timestamp",
		"story.auto_scheduling_status = 'planning'",
		"story.auto_scheduling_status = 'cannot_fit'",
		"current_timestamp - interval '1 hour'",
		"from calendar_schedule_blocks retry_block",
		"story.auto_scheduling_status in ('scheduled', 'locked')",
		"retry_block.completed_at is null",
		"elapsed_block.completed_at is null",
		"future_block.completed_at is null",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("recovery claim must contain %q: %s", fragment, query)
		}
	}
}

func TestUserScheduleStoryQueryJoinsCanonicalStoryOnce(t *testing.T) {
	t.Parallel()
	query := readNamedRepositoryQuery(t, "scheduling.sql", "ListScheduleStoryRefsForUser")
	if count := strings.Count(query, "inner join stories story"); count != 1 {
		t.Fatalf("user schedule story query must join canonical stories exactly once, got %d: %s", count, query)
	}
}

func TestScheduleOwnershipRetentionExcludesPermanentLifecycleStates(t *testing.T) {
	t.Parallel()
	query := readNamedRepositoryQuery(t, "scheduling.sql", "StoryScheduleOwnershipIsRetainable")
	for _, fragment := range []string{
		"story.deleted_at is null",
		"story.archived_at is null",
		"story.completed_at is null",
		"status.category not in ('completed', 'cancelled')",
		"owner_user.is_active = true",
		"inner join workspace_members",
		"inner join team_members",
		"story.assignee_id is null or story.assignee_id = sqlc.arg(user_id)",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("ownership retention query must contain %q: %s", fragment, query)
		}
	}
}

func TestSchedulingRepositoryDoesNotReferenceMissingStatusDeletionColumn(t *testing.T) {
	t.Parallel()
	if strings.Contains(readRepositoryQuery(t, "scheduling.sql"), "status.deleted_at") {
		t.Fatal("Maya scheduling queries must not reference statuses.deleted_at")
	}
}

func TestInterruptedRunRecoveryCannotBeOverwrittenByLateWorker(t *testing.T) {
	t.Parallel()
	runs := readRepositoryQuery(t, "runs.sql")
	if !strings.Contains(runs, "where run_id = sqlc.arg(run_id)\n    and status = 'running'") {
		t.Fatal("late workers must not overwrite a recovery-terminalized Maya run")
	}
	if strings.Count(runs, "and status = 'proposed'") < 2 {
		t.Fatal("late workers must not overwrite recovery-terminalized Maya actions")
	}
}

func readRepositoryQuery(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("queries/" + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return strings.ToLower(string(data))
}

func readNamedRepositoryQuery(t *testing.T, fileName, queryName string) string {
	t.Helper()
	contents := readRepositoryQuery(t, fileName)
	marker := "-- name: " + strings.ToLower(queryName) + " "
	start := strings.Index(contents, marker)
	if start < 0 {
		t.Fatalf("query %s not found in %s", queryName, fileName)
	}
	contents = contents[start:]
	if end := strings.Index(contents[len(marker):], "\n-- name:"); end >= 0 {
		contents = contents[:len(marker)+end]
	}
	return contents
}
