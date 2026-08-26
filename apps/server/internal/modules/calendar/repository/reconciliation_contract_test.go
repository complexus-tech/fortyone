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
	if scheduleBlockNeedsProviderUpsert(block, true, calendar.ProviderGoogle, event, hash) {
		t.Fatal("canonical provider mapping and hash must not enqueue a redundant upsert")
	}
	wrongCalendarID := "secondary"
	block.ExternalCalendarID = &wrongCalendarID
	if !scheduleBlockNeedsProviderUpsert(block, true, calendar.ProviderGoogle, event, hash) {
		t.Fatal("a matching hash must not hide a noncanonical provider mapping")
	}
}

func TestNewScheduleBlocksPreferTheSelectedPrimaryProvider(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("reconciliation.go")
	if err != nil {
		t.Fatalf("read reconciliation.go: %v", err)
	}
	source := string(data)
	for _, contract := range []string{
		"ORDER BY is_primary DESC, created_at, connection_id",
		"if connection.IsPrimary",
		"if connection.CanWrite",
	} {
		if !strings.Contains(source, contract) {
			t.Fatalf("primary provider selection is missing %q", contract)
		}
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
		"IsLocked           bool       `db:\"is_locked\"`",
		"SELECT block_id, segment_index, title, start_at, end_at, is_locked,",
		"blockChanged := providerChanged || block.IsLocked != input.Locked",
		"title = CAST($5 AS text)",
		"title <> CAST($5 AS text)",
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
	if !strings.Contains(source, "if input.AllowConflicts") {
		t.Fatal("explicit user placement must have a narrow conflict-override branch")
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
		"block.completed_at is null",
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
		"AND (provider = $3 OR operation = 'upsert')",
		"WHEN $10 OR calendar_schedule_event_outbox.dead_lettered_at IS NULL THEN NULL",
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
		"AND provider = $2",
		"AND processed_at IS NULL",
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
		"AND external_provider = $2",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("full snapshot is missing mirror invalidation contract %q", contract)
		}
	}
}

func TestProviderMirrorBookkeepingPreservesScheduleBlockConcurrencyVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		query          string
		metadataFields []string
	}{
		{
			name:           "outbox completion",
			query:          markScheduleBlockMirroredQuery,
			metadataFields: []string{"external_sync_hash", "external_synced_at"},
		},
		{
			name:           "full snapshot invalidation",
			query:          invalidateMayaScheduleMirrorHashesQuery,
			metadataFields: []string{"external_sync_hash", "external_synced_at"},
		},
		{
			name:           "incremental deleted event invalidation",
			query:          invalidateDeletedManagedScheduleEventQuery,
			metadataFields: []string{"external_sync_hash", "external_synced_at"},
		},
		{
			name:           "incremental changed event invalidation",
			query:          invalidateChangedManagedScheduleEventQuery,
			metadataFields: []string{"external_sync_hash", "external_synced_at"},
		},
		{
			name:  "disconnect cleanup",
			query: detachMayaScheduleMirrorsQuery,
			metadataFields: []string{
				"external_provider",
				"external_calendar_id",
				"external_event_id",
				"external_sync_hash",
				"external_synced_at",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			query := strings.ToLower(strings.Join(strings.Fields(tt.query), " "))
			if !strings.HasPrefix(query, "update calendar_schedule_blocks set ") {
				t.Fatalf("query must update calendar schedule blocks: %s", query)
			}
			if strings.Contains(query, "updated_at") {
				t.Fatalf("provider metadata write must preserve updated_at: %s", query)
			}
			for _, field := range tt.metadataFields {
				if !strings.Contains(query, field+" = ") {
					t.Errorf("query must update provider metadata field %q: %s", field, query)
				}
			}
		})
	}
}

func TestUserVisibleScheduleWritesAdvanceScheduleBlockConcurrencyVersion(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("schedule.go")
	if err != nil {
		t.Fatalf("read schedule repository: %v", err)
	}
	scheduleSource := strings.Join(strings.Fields(string(data)), " ")
	for _, contract := range []string{
		"SET story_id = $4, block_type = $5, title = $6, start_at = $7, end_at = $8, is_locked = $9, source = $10, updated_at = CURRENT_TIMESTAMP",
		"SET start_at = $4, end_at = $5, is_locked = TRUE, manual_override_at = CURRENT_TIMESTAMP, manual_override_by = $6, updated_at = CURRENT_TIMESTAMP",
	} {
		if !strings.Contains(scheduleSource, contract) {
			t.Errorf("user-visible schedule write must advance updated_at; missing %q", contract)
		}
	}

	reconciliationData, err := os.ReadFile("reconciliation.go")
	if err != nil {
		t.Fatalf("read reconciliation repository: %v", err)
	}
	reconciliationSource := strings.Join(strings.Fields(string(reconciliationData)), " ")
	const visibleChangeVersionContract = "updated_at = CASE WHEN title <> CAST($5 AS text) OR start_at <> $6 OR end_at <> $7 OR is_locked <> $9 THEN CURRENT_TIMESTAMP ELSE updated_at END"
	if !strings.Contains(reconciliationSource, visibleChangeVersionContract) {
		t.Errorf("Maya reconciliation must advance updated_at only for user-visible content changes")
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
