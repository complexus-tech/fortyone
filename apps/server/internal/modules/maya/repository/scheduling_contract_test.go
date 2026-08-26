package mayarepository

import (
	"os"
	"strings"
	"testing"
)

func TestScheduleRecoveryClaimIsBoundedFairAndConcurrencySafe(t *testing.T) {
	t.Parallel()
	query := strings.ToLower(claimScheduleRecoveryStoryRefsQuery)
	if strings.Contains(query, "status.deleted_at") {
		t.Fatal("schedule recovery must only reference columns defined by statuses")
	}
	for _, fragment := range []string{
		"coalesce(ownership.recovery_attempted_at, timestamp 'epoch') <= $2",
		"story.updated_at > ownership.updated_at",
		"team_settings.updated_at > ownership.updated_at",
		"sprint.updated_at > ownership.updated_at",
		"story.assignee_id <> ownership.user_id",
		"run.status = 'running'",
		"order by",
		"limit $1",
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
	query := strings.ToLower(listScheduleStoryRefsForUserQuery)
	if count := strings.Count(query, "inner join stories story"); count != 1 {
		t.Fatalf("user schedule story query must join canonical stories exactly once, got %d: %s", count, query)
	}
}

func TestScheduleOwnershipRetentionExcludesPermanentLifecycleStates(t *testing.T) {
	t.Parallel()
	query := strings.ToLower(storyScheduleOwnershipRetainableQuery)
	for _, fragment := range []string{
		"story.deleted_at is null",
		"story.archived_at is null",
		"story.completed_at is null",
		"status.category not in ('completed', 'cancelled')",
		"owner_user.is_active = true",
		"inner join workspace_members",
		"inner join team_members",
		"story.assignee_id is null or story.assignee_id = $3",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("ownership retention query must contain %q: %s", fragment, query)
		}
	}
}

func TestSchedulingRepositoryDoesNotReferenceMissingStatusDeletionColumn(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("scheduling.go")
	if err != nil {
		t.Fatalf("read scheduling.go: %v", err)
	}
	if strings.Contains(strings.ToLower(string(data)), "status.deleted_at") {
		t.Fatal("Maya scheduling queries must not reference statuses.deleted_at")
	}
}

func TestInterruptedRunRecoveryCannotBeOverwrittenByLateWorker(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("maya.go")
	if err != nil {
		t.Fatalf("read maya.go: %v", err)
	}
	source := strings.ToLower(string(data))
	if !strings.Contains(source, "where run_id = $1 and status = 'running'") {
		t.Fatal("late workers must not overwrite a recovery-terminalized Maya run")
	}
	if strings.Count(source, "where action_id = $1 and status = 'proposed'") < 2 {
		t.Fatal("late workers must not overwrite recovery-terminalized Maya actions")
	}
}
