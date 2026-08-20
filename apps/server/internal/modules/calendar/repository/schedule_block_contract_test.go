package calendarrepository

import (
	"os"
	"strings"
	"testing"
)

func TestScheduleBlockSelectScopesAutoSchedulingMetadataToMayaBlocks(t *testing.T) {
	t.Parallel()

	query := strings.ToLower(scheduleBlockSelect)
	for _, contract := range []string{
		"case when csb.source = 'maya' then s.auto_scheduling_status else null end as auto_scheduling_status",
		"case when csb.source = 'maya' then s.auto_scheduling_reason else null end as auto_scheduling_reason",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("schedule block query is missing Maya metadata contract %q", contract)
		}
	}
}

func TestToCoreScheduleBlockPropagatesAutoSchedulingMetadata(t *testing.T) {
	t.Parallel()

	status := "at_risk"
	reason := "A new calendar conflict displaced this work."
	block := toCoreScheduleBlock(dbScheduleBlock{
		AutoSchedulingStatus: &status,
		AutoSchedulingReason: &reason,
	})

	if block.AutoSchedulingStatus == nil || *block.AutoSchedulingStatus != status {
		t.Fatalf("auto-scheduling status = %v, want %q", block.AutoSchedulingStatus, status)
	}
	if block.AutoSchedulingReason == nil || *block.AutoSchedulingReason != reason {
		t.Fatalf("auto-scheduling reason = %v, want %q", block.AutoSchedulingReason, reason)
	}
}

func TestListSchedulingBlocksForUserRedactsCrossWorkspaceAutoSchedulingMetadata(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("schedule.go")
	if err != nil {
		t.Fatalf("read schedule repository: %v", err)
	}
	source := strings.Join(strings.Fields(string(data)), " ")
	functionStart := strings.Index(source, "func (r *Repo) ListSchedulingBlocksForUser")
	functionEnd := strings.Index(source[functionStart+1:], "func (r *Repo)")
	if functionStart < 0 || functionEnd < 0 {
		t.Fatal("could not locate ListSchedulingBlocksForUser implementation")
	}
	functionSource := source[functionStart : functionStart+1+functionEnd]

	workspaceGuard := strings.Index(functionSource, "if blocks[index].WorkspaceID == workspaceID { continue }")
	statusRedaction := strings.Index(functionSource, "blocks[index].AutoSchedulingStatus = nil")
	reasonRedaction := strings.Index(functionSource, "blocks[index].AutoSchedulingReason = nil")
	if workspaceGuard < 0 || statusRedaction < 0 || reasonRedaction < 0 {
		t.Fatalf("cross-workspace schedule redaction is incomplete: %s", functionSource)
	}
	if workspaceGuard >= statusRedaction || workspaceGuard >= reasonRedaction {
		t.Fatal("auto-scheduling metadata must only be redacted after retaining current-workspace blocks")
	}
}

func TestListSchedulingBlocksForUserUsesContiguousQueryParameters(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("schedule.go")
	if err != nil {
		t.Fatalf("read schedule repository: %v", err)
	}
	source := strings.Join(strings.Fields(string(data)), " ")
	functionStart := strings.Index(source, "func (r *Repo) ListSchedulingBlocksForUser")
	functionEnd := strings.Index(source[functionStart+1:], "func (r *Repo)")
	if functionStart < 0 || functionEnd < 0 {
		t.Fatal("could not locate ListSchedulingBlocksForUser implementation")
	}
	functionSource := source[functionStart : functionStart+1+functionEnd]

	for _, contract := range []string{
		"WHERE csb.user_id = $1",
		"AND csb.start_at < $3",
		"AND csb.end_at > $2",
		"SelectContext(ctx, &rows, query, userID, startAt, endAt)",
	} {
		if !strings.Contains(functionSource, contract) {
			t.Errorf("account-wide schedule query is missing parameter contract %q", contract)
		}
	}
	if strings.Contains(functionSource, "$4") || strings.Contains(functionSource, "query, workspaceID, userID, startAt, endAt") {
		t.Fatal("workspace ID must remain a redaction input and must not be passed as an unused SQL parameter")
	}
}

func TestScheduleBlockConflictsUsesContiguousQueryParameters(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("schedule.go")
	if err != nil {
		t.Fatalf("read schedule repository: %v", err)
	}
	source := strings.Join(strings.Fields(string(data)), " ")
	functionStart := strings.Index(source, "func scheduleBlockConflicts")
	functionEnd := strings.Index(source[functionStart+1:], "func ")
	if functionStart < 0 || functionEnd < 0 {
		t.Fatal("could not locate scheduleBlockConflicts implementation")
	}
	functionSource := source[functionStart : functionStart+1+functionEnd]

	for _, contract := range []string{
		"WHERE csb.user_id = $1",
		"AND csb.start_at < $3",
		"AND csb.end_at > $2",
		"AND ($4 = CAST('00000000-0000-0000-0000-000000000000' AS uuid) OR csb.block_id <> $4)",
		"WHERE cbw.user_id = $1",
		"input.UserID, input.StartAt, input.EndAt, excludeBlockID,",
	} {
		if !strings.Contains(functionSource, contract) {
			t.Errorf("schedule conflict query is missing parameter contract %q", contract)
		}
	}
	if strings.Contains(functionSource, "$5") || strings.Contains(functionSource, "input.WorkspaceID") {
		t.Fatal("schedule conflict query must not include an unused workspace parameter")
	}
}

func TestManualRescheduleAllowsExplicitCalendarOverlaps(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("schedule.go")
	if err != nil {
		t.Fatalf("read schedule repository: %v", err)
	}
	source := strings.Join(strings.Fields(string(data)), " ")
	functionStart := strings.Index(source, "func (r *Repo) ManuallyRescheduleScheduleBlock")
	functionEnd := strings.Index(source[functionStart+1:], "func ")
	if functionStart < 0 || functionEnd < 0 {
		t.Fatal("could not locate ManuallyRescheduleScheduleBlock implementation")
	}
	functionSource := source[functionStart : functionStart+1+functionEnd]

	if strings.Contains(functionSource, "scheduleBlockConflicts") || strings.Contains(functionSource, "ErrCalendarScheduleConflict") {
		t.Fatal("an explicit user reschedule must allow overlaps that the calendar can render")
	}
}
