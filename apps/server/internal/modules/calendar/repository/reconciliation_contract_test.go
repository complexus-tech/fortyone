package calendarrepository

import (
	"os"
	"strings"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
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
	provider, calendarID, eventID := string(calendar.ProviderGoogle), event.CalendarID, event.EventID
	block := reconciliationBlock{
		ID: blockID, Title: event.Title, StartAt: event.StartAt, EndAt: event.EndAt,
		ExternalProvider: &provider, ExternalCalendarID: &calendarID,
		ExternalEventID: &eventID, ExternalSyncHash: &hash,
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
	query := normalizedNamedQuery(t, "queries/reconciliation.sql", "ListCalendarWriteDestinations")
	if !strings.Contains(query, "order by connection.is_primary desc, connection.created_at, connection.connection_id") {
		t.Fatal("calendar write destinations must put the selected primary first")
	}
	source := readRepositorySource(t, "reconciliation.go")
	for _, contract := range []string{"if connection.IsPrimary", "if connection.CanWrite"} {
		if !strings.Contains(source, contract) {
			t.Fatalf("primary provider selection is missing %q", contract)
		}
	}
}

func TestMayaScheduleReconciliationPersistsLockedInputWithoutProviderRewrite(t *testing.T) {
	t.Parallel()
	listQuery := normalizedNamedQuery(t, "queries/reconciliation.sql", "ListExistingMayaScheduleSegments")
	updateQuery := normalizedNamedQuery(t, "queries/reconciliation.sql", "UpdateMayaScheduleSegment")
	createQuery := normalizedNamedQuery(t, "queries/reconciliation.sql", "CreateMayaScheduleSegment")
	for _, contract := range []string{"block.is_locked", "block.external_provider", "for update"} {
		if !strings.Contains(listQuery, contract) {
			t.Errorf("locked reconciliation read is missing %q", contract)
		}
	}
	for _, contract := range []string{
		"is_locked = sqlc.arg(is_locked)",
		"or is_locked <> sqlc.arg(is_locked)",
		"then current_timestamp else updated_at",
	} {
		if !strings.Contains(updateQuery, contract) {
			t.Errorf("locked reconciliation update is missing %q", contract)
		}
	}
	if !strings.Contains(createQuery, "sqlc.arg(is_locked)") {
		t.Fatal("new Maya segments must persist the requested lock")
	}
	source := readRepositorySource(t, "reconciliation.go")
	for _, contract := range []string{
		"blockChanged := providerChanged || block.IsLocked != input.Locked",
		"syncHash, providerChanged",
		"if input.AllowConflicts",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("locked Maya reconciliation is missing %q", contract)
		}
	}
	if strings.Contains(source, "syncHash, blockChanged") {
		t.Fatal("an internal lock-only change must not reactivate or rewrite the provider event")
	}
}

func TestMayaScheduleEligibilityAllowsAssignmentAfterScheduleCommit(t *testing.T) {
	query := normalizedNamedQuery(t, "queries/reconciliation.sql", "LockEligibleMayaScheduleStory")
	if strings.Contains(query, "story.assignee_id") {
		t.Fatal("initial scheduling must allow a selected active team member before the dependent story assignment commits")
	}
	for _, required := range []string{
		"selected_user.is_active = true", "inner join workspace_members", "inner join team_members",
		"story.deleted_at is null", "story.archived_at is null", "story.completed_at is null",
		"status.category not in ('completed', 'cancelled')", "for update of story",
		"for share of status, selected_user, membership, team_membership",
	} {
		if !strings.Contains(query, required) {
			t.Fatalf("eligibility query must enforce %q", required)
		}
	}
}

func TestMayaScheduleVersionGuardRejectsStalePlans(t *testing.T) {
	query := normalizedNamedQuery(t, "queries/reconciliation.sql", "LockMayaScheduleStoryVersion")
	for _, fragment := range []string{"story.updated_at = sqlc.arg(expected_updated_at)", "for update of story"} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("story version guard must contain %q: %s", fragment, query)
		}
	}
}

func TestProviderUpsertRequiresCurrentCanonicalStoryState(t *testing.T) {
	query := normalizedNamedQuery(t, "queries/schedule_outbox.sql", "ScheduleEventUpsertIsCurrent")
	if strings.Contains(query, "status.deleted_at") {
		t.Fatal("provider upsert must only reference columns defined by statuses")
	}
	for _, fragment := range []string{
		"ownership.user_id = block.user_id", "story.assignee_id = block.user_id",
		"story.auto_scheduling_enabled = true", "block.completed_at is null",
		"ownership.updated_at >= story.updated_at", "ownership.updated_at >= team_settings.updated_at",
		"ownership.updated_at >= sprint.updated_at", "owner_user.is_active = true",
		"inner join workspace_members", "inner join team_members",
		"status.category not in ('completed', 'cancelled')",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("provider upsert gate must enforce %q: %s", fragment, query)
		}
	}
}

func TestSameStateUpsertPreservesDeadLetterWhileDeleteReactivates(t *testing.T) {
	supersede := normalizedNamedQuery(t, "queries/schedule_outbox.sql", "SupersedeStaleScheduleEventOutbox")
	enqueue := normalizedNamedQuery(t, "queries/schedule_outbox.sql", "EnqueueScheduleEventOutbox")
	for _, contract := range []string{
		"dedupe_key <> sqlc.arg(dedupe_key)",
		"provider = sqlc.arg(provider) or operation = 'upsert'",
	} {
		if !strings.Contains(supersede, contract) {
			t.Fatalf("outbox supersession is missing %q", contract)
		}
	}
	for _, contract := range []string{
		"cast(sqlc.arg(reactivate_terminal) as boolean)",
		"calendar_schedule_event_outbox.dead_lettered_at is null",
	} {
		if !strings.Contains(enqueue, contract) {
			t.Fatalf("outbox reactivation is missing %q", contract)
		}
	}
	if !strings.Contains(readRepositorySource(t, "reconciliation.go"), "ScheduleEventOperationDelete, event, \"\", true") {
		t.Fatal("provider deletes must reactivate terminal cleanup work")
	}
}

func TestCalendarDisconnectAndOAuthRefreshReactivateOutbox(t *testing.T) {
	cleanup := normalizedNamedQuery(t, "queries/connections.sql", "ReactivateCalendarOutboxForCleanup")
	refresh := normalizedNamedQuery(t, "queries/connections.sql", "ReactivateCalendarOutboxAfterAuthorizationRefresh")
	for _, contract := range []string{
		"retrying provider cleanup during calendar disconnect.",
		"dead_lettered_at = null", "processed_at is null",
	} {
		if !strings.Contains(cleanup, contract) {
			t.Fatalf("disconnect cleanup must enforce %q", contract)
		}
	}
	for _, contract := range []string{
		"retrying after calendar authorization refresh.",
		"dead_lettered_at = null", "dead_lettered_at is not null", "attempt_count = 0",
	} {
		if !strings.Contains(refresh, contract) {
			t.Fatalf("authorization refresh must enforce %q", contract)
		}
	}
}

func TestReleasedOutboxClaimsRefundUntouchedAttempts(t *testing.T) {
	query := normalizedNamedQuery(t, "queries/schedule_outbox.sql", "ReleaseScheduleEventOutbox")
	if !strings.Contains(query, "attempt_count = greatest(attempt_count - 1, 0)") {
		t.Fatal("released rows after an earlier provider failure must not consume a delivery attempt")
	}
	if !strings.Contains(query, "available_at = current_timestamp") {
		t.Fatal("released rows must be immediately available to the next serialized batch")
	}
}

func TestOAuthUpsertLocksCleanupPendingCredentialsBeforeReplacement(t *testing.T) {
	query := normalizedNamedQuery(t, "queries/connections.sql", "LockExistingCalendarConnection")
	if !strings.Contains(query, "cleanup_pending_at is not null as boolean) as cleanup_pending") || !strings.Contains(query, "for update") {
		t.Fatal("OAuth upsert must lock and inspect cleanup-pending credentials before token replacement")
	}
}

func TestCleanupPendingConnectionsAreIsolatedFromNormalCalendarWork(t *testing.T) {
	for _, query := range []struct{ path, name string }{
		{"queries/connections.sql", "GetCalendarConnection"},
		{"queries/events.sql", "LockCalendarConnectionForIncrementalSync"},
		{"queries/events.sql", "ListCalendarEvents"},
	} {
		source := normalizedNamedQuery(t, query.path, query.name)
		if !strings.Contains(source, "cleanup_pending_at is null") {
			t.Errorf("%s must isolate cleanup-pending connections", query.name)
		}
	}
}

func TestScheduleEventDispatchIdentifiesCleanupPendingConnections(t *testing.T) {
	query := normalizedNamedQuery(t, "queries/connections.sql", "GetScheduleEventDispatchConnection")
	for _, contract := range []string{
		"cast(connection.cleanup_pending_at is not null as boolean) as cleanup_pending",
		"connection.revoked_at is null",
		"order by connection.cleanup_pending_at nulls first",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("schedule event dispatch is missing cleanup contract %q", contract)
		}
	}
}

func TestProviderMirrorBookkeepingPreservesScheduleBlockConcurrencyVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path, name string
		fields     []string
	}{
		{"queries/schedule_outbox.sql", "MarkScheduleBlockMirrored", []string{"external_sync_hash", "external_synced_at"}},
		{"queries/connections.sql", "InvalidateMayaScheduleMirrorHashes", []string{"external_sync_hash", "external_synced_at"}},
		{"queries/events.sql", "InvalidateDeletedManagedScheduleEvent", []string{"external_sync_hash", "external_synced_at"}},
		{"queries/events.sql", "InvalidateChangedManagedScheduleEvent", []string{"external_sync_hash", "external_synced_at"}},
		{"queries/connections.sql", "DetachMayaScheduleMirrors", []string{"external_provider", "external_calendar_id", "external_event_id", "external_sync_hash", "external_synced_at"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			query := normalizedNamedQuery(t, test.path, test.name)
			if !strings.HasPrefix(query, ":exec update calendar_schedule_blocks set ") {
				t.Fatalf("query must update calendar schedule blocks: %s", query)
			}
			if strings.Contains(query, "updated_at") {
				t.Fatalf("provider metadata write must preserve updated_at: %s", query)
			}
			for _, field := range test.fields {
				if !strings.Contains(query, field+" = ") {
					t.Errorf("query must update provider metadata field %q", field)
				}
			}
		})
	}
}

func TestUserVisibleScheduleWritesAdvanceScheduleBlockConcurrencyVersion(t *testing.T) {
	t.Parallel()
	for _, query := range []struct{ path, name string }{
		{"queries/schedule_mutations.sql", "UpdateCalendarScheduleBlock"},
		{"queries/schedule_mutations.sql", "ManuallyRescheduleCalendarBlock"},
		{"queries/reconciliation.sql", "UpdateMayaScheduleSegment"},
	} {
		if source := normalizedNamedQuery(t, query.path, query.name); !strings.Contains(source, "updated_at") {
			t.Errorf("%s must advance the concurrency version for visible changes", query.name)
		}
	}
}

func TestOutboxBackoffDeadLetterAndFairnessContracts(t *testing.T) {
	ready := normalizedNamedQuery(t, "queries/schedule_outbox.sql", "ListReadyScheduleEventOutboxUsers")
	failed := normalizedNamedQuery(t, "queries/schedule_outbox.sql", "MarkScheduleEventOutboxFailed")
	for _, contract := range []string{"order by min(outbox.available_at), outbox.user_id", "outbox.dead_lettered_at is null"} {
		if !strings.Contains(ready, contract) {
			t.Errorf("ready outbox query is missing fairness contract %q", contract)
		}
	}
	for _, contract := range []string{
		"attempt_count >= 8 then current_timestamp", "attempt_count <= 1 then interval '1 minute'", "else interval '1 hour'",
	} {
		if !strings.Contains(failed, contract) {
			t.Errorf("outbox failure query is missing retry contract %q", contract)
		}
	}
}

func readRepositorySource(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.Join(strings.Fields(string(data)), " ")
}
