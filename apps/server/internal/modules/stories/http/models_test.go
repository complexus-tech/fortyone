package storieshttp

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestStoryMediaUsesStableResolverURL(t *testing.T) {
	storyID := uuid.MustParse("66db0798-2eef-4dad-bb35-413612ab0fd1")
	attachmentID := uuid.MustParse("f124a762-a767-446c-bbd1-0b3f43dce115")
	stableURL := storyMediaURL("product and design", storyID, attachmentID)
	media := toAppStoryMedia(attachments.FileInfo{
		ID:       attachmentID,
		Filename: "brief.png",
		MimeType: "image/png",
		URL:      "https://storage.test/temporary-signature",
	}, stableURL)

	want := "/workspaces/product%20and%20design/stories/66db0798-2eef-4dad-bb35-413612ab0fd1/media/f124a762-a767-446c-bbd1-0b3f43dce115"
	if media.URL != want {
		t.Fatalf("stable media URL = %q, want %q", media.URL, want)
	}
	if media.URL == "https://storage.test/temporary-signature" {
		t.Fatal("story media response exposed a temporary storage URL")
	}
}

func TestAppBulkDeleteRequestValidate(t *testing.T) {
	validStoryIDs := make([]uuid.UUID, 50)
	for index := range validStoryIDs {
		validStoryIDs[index] = uuid.New()
	}

	tests := []struct {
		name      string
		storyIDs  []uuid.UUID
		wantError bool
	}{
		{name: "allows the 50 story workflow", storyIDs: validStoryIDs},
		{name: "rejects an empty request", wantError: true},
		{name: "rejects more than 50 stories", storyIDs: append(append([]uuid.UUID{}, validStoryIDs...), uuid.New()), wantError: true},
		{name: "rejects a nil story ID", storyIDs: []uuid.UUID{uuid.Nil}, wantError: true},
		{name: "rejects duplicate story IDs", storyIDs: []uuid.UUID{validStoryIDs[0], validStoryIDs[0]}, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (AppBulkDeleteRequest{StoryIDs: tt.storyIDs}).Validate()
			if tt.wantError && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("expected request to validate, got %v", err)
			}
		})
	}
}

func TestToCoreNewStoryMapsLabelIDs(t *testing.T) {
	userID := uuid.New()
	idempotencyKey := "maya:tool-call-1"
	labelIDs := []uuid.UUID{uuid.New(), uuid.New()}
	estimatedDurationMinutes := 180
	minimumFocusBlockMinutes := 45
	autoSchedulingReason := "Waiting for an owner"
	autoSchedulingUpdatedAt := time.Date(2026, time.August, 15, 9, 30, 0, 0, time.UTC)

	coreStory := toCoreNewStory(AppNewStory{
		Title:                    "Add reporting filters",
		LabelIDs:                 labelIDs,
		Team:                     uuid.New(),
		Priority:                 "High",
		EstimatedDurationMinutes: &estimatedDurationMinutes,
		MinimumFocusBlockMinutes: &minimumFocusBlockMinutes,
		AutoSchedulingEnabled:    true,
		IdempotencyKey:           &idempotencyKey,
	}, userID)

	if len(coreStory.LabelIDs) != len(labelIDs) {
		t.Fatalf("expected %d labels, got %d", len(labelIDs), len(coreStory.LabelIDs))
	}

	for i, labelID := range labelIDs {
		if coreStory.LabelIDs[i] != labelID {
			t.Fatalf("expected label %d to be %s, got %s", i, labelID, coreStory.LabelIDs[i])
		}
	}
	if coreStory.EstimatedDurationMinutes != &estimatedDurationMinutes {
		t.Fatalf("expected estimated duration pointer to be preserved, got %v", coreStory.EstimatedDurationMinutes)
	}
	if coreStory.MinimumFocusBlockMinutes != &minimumFocusBlockMinutes {
		t.Fatalf("expected minimum focus block pointer to be preserved, got %v", coreStory.MinimumFocusBlockMinutes)
	}
	if !coreStory.AutoSchedulingEnabled || coreStory.AutoSchedulingLocked {
		t.Fatalf("expected auto-scheduling preferences to be preserved, got enabled=%t locked=%t", coreStory.AutoSchedulingEnabled, coreStory.AutoSchedulingLocked)
	}
	if coreStory.CreationKey == nil || *coreStory.CreationKey != "app:"+userID.String()+":"+idempotencyKey {
		t.Fatalf("expected a user-scoped idempotency key, got %v", coreStory.CreationKey)
	}

	appStory := toAppStory(stories.CoreSingleStory{
		AutoSchedulingEnabled:   true,
		AutoSchedulingLocked:    true,
		AutoSchedulingStatus:    stories.AutoSchedulingStatusNeedsOwner,
		AutoSchedulingReason:    &autoSchedulingReason,
		AutoSchedulingUpdatedAt: &autoSchedulingUpdatedAt,
	}, nil)
	if !appStory.AutoSchedulingEnabled || !appStory.AutoSchedulingLocked || appStory.AutoSchedulingStatus != stories.AutoSchedulingStatusNeedsOwner {
		t.Fatalf("expected auto-scheduling state in story response, got %#v", appStory)
	}
	if appStory.AutoSchedulingReason != &autoSchedulingReason || appStory.AutoSchedulingUpdatedAt != &autoSchedulingUpdatedAt {
		t.Fatalf("expected auto-scheduling metadata pointers to be preserved, got %#v", appStory)
	}
}

func TestToAppStorySerializesEmptyCollaboratorIDsAsArray(t *testing.T) {
	payload, err := json.Marshal(toAppStory(stories.CoreSingleStory{}, nil))
	if err != nil {
		t.Fatalf("marshal app story: %v", err)
	}

	var response struct {
		CollaboratorIDs json.RawMessage `json:"collaboratorIds"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("unmarshal app story: %v", err)
	}
	if string(response.CollaboratorIDs) != "[]" {
		t.Fatalf("collaboratorIds = %s, want []", response.CollaboratorIDs)
	}
}

func TestParseStoryQueryMapsEstimateValues(t *testing.T) {
	request := httptest.NewRequest(
		"GET",
		"/stories?estimateValues=1,5&estimateValues=8",
		nil,
	)

	query, err := parseStoryQuery(request)
	if err != nil {
		t.Fatalf("expected query to parse, got error: %v", err)
	}

	expected := []int16{1, 5, 8}
	if len(query.Filters.EstimateValues) != len(expected) {
		t.Fatalf("expected %d estimates, got %d", len(expected), len(query.Filters.EstimateValues))
	}

	for i, estimateValue := range expected {
		if query.Filters.EstimateValues[i] != estimateValue {
			t.Fatalf("expected estimate %d to be %d, got %d", i, estimateValue, query.Filters.EstimateValues[i])
		}
	}
}

func TestParseStoryQueryMapsNegatedFilters(t *testing.T) {
	statusID := uuid.New()
	assigneeID := uuid.New()
	objectiveID := uuid.New()
	request := httptest.NewRequest(
		"GET",
		"/stories?excludedStatusIds="+statusID.String()+"&excludedAssigneeIds="+assigneeID.String()+"&titleNotContains=deprecated&excludedObjectiveId="+objectiveID.String()+"&hasAssignee=true&deadlineNot=2026-08-31&orderDirection=asc",
		nil,
	)

	query, err := parseStoryQuery(request)
	if err != nil {
		t.Fatalf("expected query to parse, got error: %v", err)
	}

	if len(query.Filters.ExcludedStatusIDs) != 1 || query.Filters.ExcludedStatusIDs[0] != statusID {
		t.Fatalf("expected excluded status %s, got %v", statusID, query.Filters.ExcludedStatusIDs)
	}
	if len(query.Filters.ExcludedAssigneeIDs) != 1 || query.Filters.ExcludedAssigneeIDs[0] != assigneeID {
		t.Fatalf("expected excluded assignee %s, got %v", assigneeID, query.Filters.ExcludedAssigneeIDs)
	}
	if query.Filters.TitleNotContains == nil || *query.Filters.TitleNotContains != "deprecated" {
		t.Fatalf("expected titleNotContains to be parsed, got %v", query.Filters.TitleNotContains)
	}
	if query.Filters.ExcludedObjective == nil || *query.Filters.ExcludedObjective != objectiveID {
		t.Fatalf("expected excluded objective %s, got %v", objectiveID, query.Filters.ExcludedObjective)
	}
	if query.Filters.HasAssignee == nil || !*query.Filters.HasAssignee {
		t.Fatalf("expected hasAssignee to be parsed, got %v", query.Filters.HasAssignee)
	}
	if query.Filters.DeadlineNot == nil {
		t.Fatal("expected deadlineNot to be parsed")
	}
	if query.OrderDirection != "asc" {
		t.Fatalf("expected ascending order, got %q", query.OrderDirection)
	}
}

func TestParseStoryQueryMapsCollaborationFilters(t *testing.T) {
	collaboratorID := uuid.New()
	request := httptest.NewRequest(
		"GET",
		"/stories?collaboratorIds="+collaboratorID.String()+"&collaboratingWithMe=true",
		nil,
	)

	query, err := parseStoryQuery(request)
	if err != nil {
		t.Fatalf("expected query to parse, got error: %v", err)
	}

	if len(query.Filters.CollaboratorIDs) != 1 || query.Filters.CollaboratorIDs[0] != collaboratorID {
		t.Fatalf("expected collaborator %s, got %v", collaboratorID, query.Filters.CollaboratorIDs)
	}
	if query.Filters.CollaboratingWithMe == nil || !*query.Filters.CollaboratingWithMe {
		t.Fatalf("expected collaboratingWithMe to be parsed, got %v", query.Filters.CollaboratingWithMe)
	}
}

func TestToAppStoryListItemIncludesEmbeddedSummaries(t *testing.T) {
	teamID := uuid.New()
	objectiveID := uuid.New()
	sprintID := uuid.New()
	workspaceID := uuid.New()
	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC)
	description := "Improve conversion from trial teams"
	goal := "Ship the onboarding pass"

	story := stories.CoreStoryList{
		ID:        uuid.New(),
		Title:     "Add activation checklist",
		Objective: &objectiveID,
		Sprint:    &sprintID,
		Team:      teamID,
		Workspace: workspaceID,
		TeamSummary: &stories.CoreTeamSummary{
			ID:   teamID,
			Name: "Growth",
			Code: "GRO",
		},
		ObjectiveSummary: &stories.CoreObjectiveSummary{
			ID:          objectiveID,
			Name:        "Increase activation",
			Description: &description,
		},
		SprintSummary: &stories.CoreSprintSummary{
			ID:        sprintID,
			Name:      "June hardening",
			Goal:      &goal,
			StartDate: startDate,
			EndDate:   endDate,
		},
	}

	appStory := toAppStoryListItem(story, nil)

	if appStory.TeamSummary == nil {
		t.Fatal("expected team summary to be embedded")
	}
	if appStory.TeamSummary.Code != "GRO" {
		t.Fatalf("expected team code GRO, got %q", appStory.TeamSummary.Code)
	}
	if appStory.ObjectiveSummary == nil {
		t.Fatal("expected objective summary to be embedded")
	}
	if appStory.ObjectiveSummary.Name != "Increase activation" {
		t.Fatalf("expected objective name to be embedded, got %q", appStory.ObjectiveSummary.Name)
	}
	if appStory.ObjectiveSummary.Description == nil || *appStory.ObjectiveSummary.Description != description {
		t.Fatalf("expected objective description %q, got %#v", description, appStory.ObjectiveSummary.Description)
	}
	if appStory.SprintSummary == nil {
		t.Fatal("expected sprint summary to be embedded")
	}
	if appStory.SprintSummary.Name != "June hardening" {
		t.Fatalf("expected sprint name to be embedded, got %q", appStory.SprintSummary.Name)
	}
	if appStory.SprintSummary.Goal == nil || *appStory.SprintSummary.Goal != goal {
		t.Fatalf("expected sprint goal %q, got %#v", goal, appStory.SprintSummary.Goal)
	}
}
