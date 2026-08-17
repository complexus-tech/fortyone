package calendarrepository

import (
	"os"
	"strings"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

func TestScheduleBlockNeedsProviderUpsertChecksMappingAndHash(t *testing.T) {
	t.Parallel()
	blockID := uuid.New()
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	event := calendar.ExternalScheduleEventInput{
		CalendarID: "primary", EventID: calendar.StableGoogleScheduleEventID(blockID),
		Title: "Focus", StartAt: startAt, EndAt: startAt.Add(time.Hour),
	}
	hash := calendar.ScheduleEventSyncHash(event)
	provider := string(calendar.ProviderGoogle)
	calendarID := event.CalendarID
	eventID := event.EventID
	block := reconciliationBlock{
		ID: blockID, Title: event.Title, StartAt: event.StartAt, EndAt: event.EndAt,
		ExternalProvider: &provider, ExternalCalendarID: &calendarID, ExternalEventID: &eventID, ExternalSyncHash: &hash,
	}
	if scheduleBlockNeedsProviderUpsert(block, true, event, hash) {
		t.Fatal("canonical provider mapping and hash must not enqueue a redundant upsert")
	}
	wrongCalendarID := "secondary"
	block.ExternalCalendarID = &wrongCalendarID
	if !scheduleBlockNeedsProviderUpsert(block, true, event, hash) {
		t.Fatal("a matching hash must not hide a noncanonical provider mapping")
	}
}

func TestMayaScheduleReconciliationPersistsLockedInputWithoutProviderRewrite(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("reconciliation.go")
	if err != nil {
		t.Fatalf("read reconciliation.go: %v", err)
	}
	source := string(data)
	for _, contract := range []string{
		"IsLocked           bool      `db:\"is_locked\"`",
		"SELECT block_id, segment_index, title, start_at, end_at, is_locked,",
		"blockChanged := providerChanged || block.IsLocked != input.Locked",
		"is_locked = $9",
		"is_locked <> $9",
		"$10, 'maya'",
		"IsLocked:           input.Locked",
		"syncHash, providerChanged",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("locked Maya reconciliation is missing contract %q", contract)
		}
	}
	if calls := strings.Count(source, "eventID, input.Locked"); calls != 2 {
		t.Fatalf("locked input must bind both update and insert statements exactly once; got %d bindings", calls)
	}
	if strings.Contains(source, "syncHash, blockChanged") {
		t.Fatal("an internal lock-only change must not reactivate or rewrite the provider event")
	}
}

func TestMayaScheduleEligibilityAllowsAssignmentAfterScheduleCommit(t *testing.T) {
	query := strings.ToLower(mayaScheduleEligibilityQuery)
	if strings.Contains(query, "story.assignee_id") {
		t.Fatal("initial scheduling must allow a selected active team member before the dependent story assignment commits")
	}
	for _, required := range []string{
		"selected_user.is_active = true",
		"inner join workspace_members",
		"inner join team_members",
		"story.deleted_at is null",
		"story.archived_at is null",
		"story.completed_at is null",
		"status.category not in ('completed', 'cancelled')",
		"for update of story",
		"for share of status, selected_user, membership, team_membership",
	} {
		if !strings.Contains(query, required) {
			t.Fatalf("eligibility query must enforce %q", required)
		}
	}
}

func TestMayaScheduleVersionGuardRejectsStalePlans(t *testing.T) {
	query := strings.ToLower(mayaScheduleStoryVersionQuery)
	for _, fragment := range []string{"story.updated_at = $3", "for update of story"} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("story version guard must contain %q: %s", fragment, query)
		}
	}
}

func TestProviderUpsertRequiresCurrentCanonicalStoryState(t *testing.T) {
	query := strings.ToLower(scheduleEventUpsertIsCurrentQuery)
	if strings.Contains(query, "status.deleted_at") {
		t.Fatal("provider upsert must only reference columns defined by statuses")
	}
	for _, fragment := range []string{
		"ownership.user_id = block.user_id",
		"story.assignee_id = block.user_id",
		"story.auto_scheduling_enabled = true",
		"ownership.updated_at >= story.updated_at",
		"ownership.updated_at >= team_settings.updated_at",
		"ownership.updated_at >= sprint.updated_at",
		"owner_user.is_active = true",
		"inner join workspace_members",
		"inner join team_members",
		"status.category not in ('completed', 'cancelled')",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("provider upsert gate must enforce %q: %s", fragment, query)
		}
	}
}

func TestCalendarReconciliationDoesNotReferenceMissingStatusDeletionColumn(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("reconciliation.go")
	if err != nil {
		t.Fatalf("read reconciliation.go: %v", err)
	}
	if strings.Contains(strings.ToLower(string(data)), "status.deleted_at") {
		t.Fatal("calendar reconciliation queries must not reference statuses.deleted_at")
	}
}

func TestSameStateUpsertPreservesDeadLetterWhileDeleteReactivates(t *testing.T) {
	data, err := os.ReadFile("reconciliation.go")
	if err != nil {
		t.Fatalf("read reconciliation.go: %v", err)
	}
	source := string(data)
	for _, contract := range []string{
		"dedupe_key <> $2",
		"WHEN $9 OR calendar_schedule_event_outbox.dead_lettered_at IS NULL THEN NULL",
		"reactivateTerminal bool",
		"ScheduleEventOperationDelete, event, \"\", true",
	} {
		if !strings.Contains(source, contract) {
			t.Fatalf("outbox conflict handling is missing %q", contract)
		}
	}
}

func TestCalendarDisconnectReactivatesUnprocessedDeadLetters(t *testing.T) {
	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read commands.go: %v", err)
	}
	source := string(data)
	for _, contract := range []string{
		"Retrying provider cleanup during calendar disconnect.",
		"SET dead_lettered_at = NULL",
		"WHERE user_id = $1 AND processed_at IS NULL",
	} {
		if !strings.Contains(source, contract) {
			t.Fatalf("disconnect cleanup must reactivate pending terminal rows with %q", contract)
		}
	}
}

func TestReleasedOutboxClaimsRefundUntouchedAttempts(t *testing.T) {
	query := strings.ToLower(releaseScheduleEventOutboxQuery)
	if !strings.Contains(query, "attempt_count = greatest(attempt_count - 1, 0)") {
		t.Fatal("rows released after an earlier provider failure must not consume a delivery attempt")
	}
	if !strings.Contains(query, "available_at = current_timestamp") {
		t.Fatal("released rows must be immediately available to the next serialized batch")
	}
}

func TestOAuthUpsertLocksCleanupPendingCredentialsBeforeReplacement(t *testing.T) {
	query := strings.ToLower(lockExistingCalendarConnectionQuery)
	if !strings.Contains(query, "cleanup_pending_at is not null as cleanup_pending") || !strings.Contains(query, "for update") {
		t.Fatal("OAuth upsert must lock and inspect cleanup-pending credentials before any token replacement")
	}
}

func TestOAuthRefreshReactivatesTerminalScheduleWrites(t *testing.T) {
	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read commands.go: %v", err)
	}
	source := string(data)
	for _, contract := range []string{
		"Retrying after calendar authorization refresh.",
		"AND dead_lettered_at IS NOT NULL",
		"attempt_count = 0",
	} {
		if !strings.Contains(source, contract) {
			t.Fatalf("authorization refresh must reactivate terminal schedule writes with %q", contract)
		}
	}
}

func TestCleanupPendingConnectionsAreIsolatedFromNormalCalendarWork(t *testing.T) {
	for file, required := range map[string][]string{
		"queries.go": {
			"AND cleanup_pending_at IS NULL",
			"GetScheduleEventDispatchConnection",
		},
		"push.go": {
			"revoked_at IS NULL AND cleanup_pending_at IS NULL",
			"AND cleanup_pending_at IS NULL",
		},
		"commands.go": {
			"return calendar.CoreConnection{}, calendar.ErrCalendarCleanupPending",
			"AND cleanup_pending_at IS NULL",
		},
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		source := string(data)
		for _, contract := range required {
			if !strings.Contains(source, contract) {
				t.Errorf("%s is missing cleanup isolation contract %q", file, contract)
			}
		}
	}
}

func TestFullSnapshotInvalidatesManagedMirrorHashes(t *testing.T) {
	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read commands.go: %v", err)
	}
	source := string(data)
	for _, contract := range []string{
		"invalidate Maya schedule mirrors before full calendar sync",
		"external_sync_hash = NULL",
		"external_synced_at = NULL",
		"AND source = 'maya'",
		"AND external_provider = 'google'",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("full snapshot is missing mirror invalidation contract %q", contract)
		}
	}
}

func TestOutboxBackoffDeadLetterAndFairnessContracts(t *testing.T) {
	data, err := os.ReadFile("reconciliation.go")
	if err != nil {
		t.Fatalf("read reconciliation.go: %v", err)
	}
	source := string(data)
	for _, contract := range []string{
		"ORDER BY MIN(outbox.available_at), outbox.user_id",
		"AND outbox.dead_lettered_at IS NULL",
		"WHEN $3 OR attempt_count >= 8 THEN CURRENT_TIMESTAMP",
		"WHEN attempt_count <= 1 THEN INTERVAL '1 minute'",
		"ELSE INTERVAL '1 hour'",
		"AND outbox.dead_lettered_at IS NULL",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("outbox implementation is missing retry/fairness contract %q", contract)
		}
	}
}
