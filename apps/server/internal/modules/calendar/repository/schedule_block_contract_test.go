package calendarrepository

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	"github.com/google/uuid"
)

func TestScheduleBlockSelectScopesAutoSchedulingMetadataToMayaBlocks(t *testing.T) {
	t.Parallel()

	query := normalizedNamedQuery(t, "queries/schedule_reads.sql", "ListCalendarScheduleBlocks")
	for _, contract := range []string{
		"coalesce(case when block.source = 'maya' then story.auto_scheduling_status end, '') as text) as auto_scheduling_status",
		"coalesce(case when block.source = 'maya' then story.auto_scheduling_reason end, '') as text) as auto_scheduling_reason",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("schedule block query is missing Maya metadata contract %q", contract)
		}
	}
}

func TestScheduleBlockSelectIncludesStoryStatusColor(t *testing.T) {
	t.Parallel()

	query := normalizedNamedQuery(t, "queries/schedule_reads.sql", "ListCalendarScheduleBlocks")
	for _, contract := range []string{
		"status.color as story_status_color",
		"left join statuses status on status.status_id = story.status_id and status.team_id = story.team_id",
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
	block := toCoreScheduleBlock(scheduleBlockRecord{
		AutoSchedulingStatus: status,
		AutoSchedulingReason: reason,
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
	block := toCoreScheduleBlock(scheduleBlockRecord{StoryStatusColor: &color})

	if block.StoryStatusColor == nil || *block.StoryStatusColor != color {
		t.Fatalf("story status color = %v, want %q", block.StoryStatusColor, color)
	}
}

func TestToCoreScheduleBlockPropagatesCompletion(t *testing.T) {
	t.Parallel()

	completedAt := time.Date(2026, 8, 26, 8, 30, 0, 0, time.UTC)
	block := toCoreScheduleBlock(scheduleBlockRecord{CompletedAt: &completedAt})

	if block.CompletedAt == nil || !block.CompletedAt.Equal(completedAt) {
		t.Fatalf("completed at = %v, want %v", block.CompletedAt, completedAt)
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

	query := normalizedNamedQuery(t, "queries/schedule_reads.sql", "ListSchedulingBlocksForUser")

	for _, contract := range []string{
		"where block.user_id = sqlc.arg(user_id)",
		"and block.start_at < sqlc.arg(end_at)",
		"and block.end_at > sqlc.arg(start_at)",
		"and block.completed_at is null",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("account-wide schedule query is missing parameter contract %q", contract)
		}
	}
	if strings.Contains(query, "sqlc.arg(workspace_id)") {
		t.Fatal("workspace ID must remain a redaction input and must not be passed as an unused SQL parameter")
	}
}

func TestScheduleBlockConflictsUsesContiguousQueryParameters(t *testing.T) {
	t.Parallel()

	query := normalizedNamedQuery(t, "queries/schedule_reads.sql", "CalendarScheduleBlockConflicts")

	for _, contract := range []string{
		"where schedule_block.user_id = sqlc.arg(user_id)",
		"and schedule_block.completed_at is null",
		"and schedule_block.start_at < sqlc.arg(end_at)",
		"and schedule_block.end_at > sqlc.arg(start_at)",
		"cast(sqlc.arg(exclude_block_id) as uuid)",
		"where busy.user_id = sqlc.arg(user_id)",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("schedule conflict query is missing parameter contract %q", contract)
		}
	}
	if strings.Contains(query, "sqlc.arg(workspace_id)") {
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

func TestResizedStoryEstimateMinutes(t *testing.T) {
	t.Parallel()

	minutes := func(value int) *int { return &value }
	startAt := time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name             string
		currentEstimate  *int
		minimumFocus     *int
		previousDuration time.Duration
		nextDuration     time.Duration
		want             int
		wantError        string
	}{
		{
			name:             "extends existing estimate by block delta",
			currentEstimate:  minutes(180),
			minimumFocus:     minutes(30),
			previousDuration: 60 * time.Minute,
			nextDuration:     90 * time.Minute,
			want:             210,
		},
		{
			name:             "shrinks existing estimate by block delta",
			currentEstimate:  minutes(180),
			minimumFocus:     minutes(30),
			previousDuration: 90 * time.Minute,
			nextDuration:     30 * time.Minute,
			want:             120,
		},
		{
			name:             "initializes missing estimate from resized block",
			minimumFocus:     minutes(45),
			previousDuration: 60 * time.Minute,
			nextDuration:     90 * time.Minute,
			want:             90,
		},
		{
			name:             "accepts one minute lower boundary",
			currentEstimate:  minutes(5),
			previousDuration: 5 * time.Minute,
			nextDuration:     time.Minute,
			want:             1,
		},
		{
			name:             "accepts forty hour upper boundary",
			currentEstimate:  minutes(2390),
			previousDuration: 60 * time.Minute,
			nextDuration:     70 * time.Minute,
			want:             2400,
		},
		{
			name:             "rejects non-positive estimate without clamping",
			currentEstimate:  minutes(5),
			previousDuration: 10 * time.Minute,
			nextDuration:     5 * time.Minute,
			wantError:        "estimated duration minutes must be greater than zero",
		},
		{
			name:             "rejects estimate above forty hours without clamping",
			currentEstimate:  minutes(2390),
			previousDuration: 60 * time.Minute,
			nextDuration:     75 * time.Minute,
			wantError:        "estimated duration minutes must not exceed 2400",
		},
		{
			name:             "rejects estimate below minimum focus block",
			currentEstimate:  minutes(60),
			minimumFocus:     minutes(45),
			previousDuration: 60 * time.Minute,
			nextDuration:     30 * time.Minute,
			wantError:        "minimum focus block minutes must not exceed estimated duration minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := resizedStoryEstimateMinutes(
				tt.currentEstimate,
				tt.minimumFocus,
				startAt,
				startAt.Add(tt.previousDuration),
				startAt,
				startAt.Add(tt.nextDuration),
			)
			if tt.wantError != "" {
				if !errors.Is(err, calendar.ErrInvalidScheduleBlock) {
					t.Fatalf("error = %v, want ErrInvalidScheduleBlock", err)
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %q, want text %q", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("resizedStoryEstimateMinutes returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("estimate = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestManualResizePersistsEstimateBlockAuditAndOutboxAtomically(t *testing.T) {
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

	contracts := []string{
		"r.withinTransaction(ctx",
		"GetManualScheduleRescheduleBlockID",
		"if input.Change == calendar.ManualScheduleBlockChangeResize && current.StoryID != nil",
		"LockScheduleStoryTime",
		"UpdateScheduleStoryEstimate",
		"if storyTime.AutoSchedulingEnabled",
		"ManuallyRescheduleCalendarBlock",
		"RecordManualCalendarReschedule",
		"enqueueScheduleEventOutbox(ctx, queries",
	}
	positions := make([]int, 0, len(contracts))
	for _, contract := range contracts {
		position := strings.Index(functionSource, contract)
		if position < 0 {
			t.Fatalf("manual resize transaction is missing contract %q", contract)
		}
		positions = append(positions, position)
	}
	for index := 1; index < len(positions); index++ {
		if positions[index] <= positions[index-1] {
			t.Fatalf("manual resize transaction contract %q occurs out of order", contracts[index])
		}
	}
	if idempotentReturn := strings.Index(functionSource, "result = calendar.ManualScheduleBlockResult{Block: block}"); idempotentReturn < 0 || idempotentReturn >= positions[2] {
		t.Fatal("idempotent retries must return before locking or updating the story")
	}
}

func normalizedNamedQuery(t *testing.T, path, name string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	marker := "-- name: " + name + " "
	start := strings.Index(string(data), marker)
	if start < 0 {
		t.Fatalf("query %s is missing from %s", name, path)
	}
	remainder := string(data)[start+len(marker):]
	if end := strings.Index(remainder, "-- name: "); end >= 0 {
		remainder = remainder[:end]
	}
	return strings.Join(strings.Fields(strings.ToLower(remainder)), " ")
}
