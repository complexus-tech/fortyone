package calendarrepository

import (
	"os"
	"strings"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

func TestScheduleBlockSelectScopesAutoSchedulingMetadataToMayaBlocks(t *testing.T) {
	t.Parallel()

	query := strings.Join(strings.Fields(strings.ToLower(scheduleBlockSelect)), " ")
	for _, contract := range []string{
		"case when csb.source = 'maya' then s.auto_scheduling_status else null end as auto_scheduling_status",
		"case when csb.source = 'maya' then s.auto_scheduling_reason else null end as auto_scheduling_reason",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("schedule block query is missing Maya metadata contract %q", contract)
		}
	}
}

func TestScheduleBlockSelectIncludesStoryStatusColor(t *testing.T) {
	t.Parallel()

	query := strings.Join(strings.Fields(strings.ToLower(scheduleBlockSelect)), " ")
	for _, contract := range []string{
		"status.color as story_status_color",
		"left join statuses status on status.status_id = s.status_id and status.team_id = s.team_id",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("schedule block query is missing story status contract %q", contract)
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

func TestToCoreScheduleBlockPropagatesStoryStatusColor(t *testing.T) {
	t.Parallel()

	color := "#3c90ff"
	block := toCoreScheduleBlock(dbScheduleBlock{StoryStatusColor: &color})

	if block.StoryStatusColor == nil || *block.StoryStatusColor != color {
		t.Fatalf("story status color = %v, want %q", block.StoryStatusColor, color)
	}
}

func TestRedactCrossWorkspaceScheduleBlocksHidesTaskDetails(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	otherWorkspaceID := uuid.New()
	storyID := uuid.New()
	teamID := uuid.New()
	manualOverrideBy := uuid.New()
	manualOverrideAt := time.Now()
	storyTitle := "Private roadmap task"
	storyCode := "SEC-41"
	teamName := "Secret team"
	teamCode := "SEC"
	storyStatusColor := "#3c90ff"
	status := "scheduled"
	reason := "Private scheduling reason"
	blocks := []calendar.CoreScheduleBlock{
		{WorkspaceID: workspaceID, Title: "Visible current task"},
		{
			WorkspaceID:          otherWorkspaceID,
			StoryID:              &storyID,
			StoryTitle:           &storyTitle,
			StoryCode:            &storyCode,
			StoryStatusColor:     &storyStatusColor,
			StoryPriority:        "high",
			TeamID:               &teamID,
			TeamName:             &teamName,
			TeamCode:             &teamCode,
			Title:                storyTitle,
			AutoSchedulingStatus: &status,
			AutoSchedulingReason: &reason,
			ManualOverrideBy:     &manualOverrideBy,
			ManualOverrideAt:     &manualOverrideAt,
			Source:               calendar.ScheduleBlockSourceMaya,
		},
	}

	redactCrossWorkspaceScheduleBlocks(blocks, workspaceID)

	if blocks[0].Title != "Visible current task" || blocks[0].IsCrossWorkspace {
		t.Fatalf("current-workspace block was changed: %#v", blocks[0])
	}
	redacted := blocks[1]
	if redacted.Title != "Scheduled elsewhere" || !redacted.IsCrossWorkspace {
		t.Fatalf("cross-workspace presentation state is incomplete: %#v", redacted)
	}
	if redacted.StoryID != nil || redacted.StoryTitle != nil || redacted.StoryCode != nil || redacted.TeamID != nil || redacted.TeamName != nil || redacted.TeamCode != nil {
		t.Fatalf("cross-workspace task identity leaked: %#v", redacted)
	}
	if redacted.StoryStatusColor != nil {
		t.Fatalf("cross-workspace story status color leaked: %#v", redacted)
	}
	if redacted.StoryPriority != "" || redacted.AutoSchedulingStatus != nil || redacted.AutoSchedulingReason != nil || redacted.ManualOverrideBy != nil || redacted.ManualOverrideAt != nil {
		t.Fatalf("cross-workspace planning metadata leaked: %#v", redacted)
	}
	if redacted.Source != calendar.ScheduleBlockSourceMaya {
		t.Fatal("internal planning must retain the block's scheduling mechanics")
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
