package notifications

import (
	"testing"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNotificationRules(t *testing.T) {
	// Mock data
	actorID := uuid.New()
	assigneeID := uuid.New()
	newAssigneeID := uuid.New()
	workspaceID := uuid.New()
	storyID := uuid.New()

	tests := []struct {
		name               string
		payload            events.StoryUpdatedPayload
		actorID            uuid.UUID
		expectedNotifCount int
		description        string
	}{
		{
			name: "Rule 1: New assignment notification",
			payload: events.StoryUpdatedPayload{
				StoryID:     storyID,
				WorkspaceID: workspaceID,
				Updates:     map[string]any{"assignee_id": newAssigneeID},
				AssigneeID:  nil, // No previous assignee
			},
			actorID:            actorID,
			expectedNotifCount: 1,
			description:        "Should notify new assignee",
		},
		{
			name: "Rule 2: Status update notification",
			payload: events.StoryUpdatedPayload{
				StoryID:     storyID,
				WorkspaceID: workspaceID,
				Updates:     map[string]any{"status_id": uuid.New()},
				AssigneeID:  &assigneeID, // Has assignee
			},
			actorID:            actorID,
			expectedNotifCount: 1,
			description:        "Should notify current assignee of status change",
		},
		{
			name: "Rule 3: Unassignment notification",
			payload: events.StoryUpdatedPayload{
				StoryID:     storyID,
				WorkspaceID: workspaceID,
				Updates:     map[string]any{"assignee_id": nil}, // Unassigning
				AssigneeID:  &assigneeID,                        // Had assignee
			},
			actorID:            actorID,
			expectedNotifCount: 1,
			description:        "Should notify original assignee of unassignment",
		},
		{
			name: "Rule 4: No notification for actor",
			payload: events.StoryUpdatedPayload{
				StoryID:     storyID,
				WorkspaceID: workspaceID,
				Updates:     map[string]any{"status_id": uuid.New()},
				AssigneeID:  &actorID, // Actor is the assignee
			},
			actorID:            actorID,
			expectedNotifCount: 0,
			description:        "Should not notify actor of their own changes",
		},
		{
			name: "No relevant updates",
			payload: events.StoryUpdatedPayload{
				StoryID:     storyID,
				WorkspaceID: workspaceID,
				Updates:     map[string]any{"description": "Updated description"},
				AssigneeID:  &assigneeID,
			},
			actorID:            actorID,
			expectedNotifCount: 0,
			description:        "Should not notify for non-relevant updates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test shouldNotify function
			if tt.payload.AssigneeID != nil {
				shouldNotifyResult := shouldNotify(*tt.payload.AssigneeID, tt.actorID)
				if tt.expectedNotifCount > 0 && *tt.payload.AssigneeID != tt.actorID {
					assert.True(t, shouldNotifyResult, "Should notify when recipient is not actor")
				} else if *tt.payload.AssigneeID == tt.actorID {
					assert.False(t, shouldNotifyResult, "Should not notify actor")
				}
			}

			// Test hasRelevantUpdates function
			hasRelevant := hasRelevantUpdates(tt.payload.Updates)
			if tt.name == "No relevant updates" {
				assert.False(t, hasRelevant, "Should not have relevant updates for description change")
			} else if tt.name == "Rule 2: Status update notification" {
				assert.True(t, hasRelevant, "Should have relevant updates for status change")
			}

			// Test getNewAssignee function
			newAssignee := getNewAssignee(tt.payload.Updates)
			if tt.name == "Rule 1: New assignment notification" {
				assert.NotNil(t, newAssignee, "Should detect new assignee")
				assert.Equal(t, newAssigneeID, *newAssignee, "Should return correct new assignee ID")
			}

			// Test isUnassignment function
			isUnassign := isUnassignment(tt.payload.Updates)
			if tt.name == "Rule 3: Unassignment notification" {
				assert.True(t, isUnassign, "Should detect unassignment")
			} else if tt.name == "Rule 1: New assignment notification" {
				assert.False(t, isUnassign, "Should not detect unassignment for new assignment")
			}
		})
	}
}

func TestNotificationTypes(t *testing.T) {
	actorID := uuid.New()
	assigneeID := uuid.New()
	workspaceID := uuid.New()
	storyID := uuid.New()

	// Test assignment notification
	assignmentNotif := createAssignmentNotification(
		assigneeID,
		events.StoryUpdatedPayload{StoryID: storyID, WorkspaceID: workspaceID},
		actorID,
		"Test Story",
		"testuser",
	)

	assert.Equal(t, "story_assignment", assignmentNotif.Type)
	assert.Equal(t, "testuser assigned you a story", assignmentNotif.Description)
	assert.Equal(t, "Test Story", assignmentNotif.Title)

	// Test update notification
	updateNotif := createUpdateNotification(
		assigneeID,
		events.StoryUpdatedPayload{
			StoryID:     storyID,
			WorkspaceID: workspaceID,
			Updates:     map[string]any{"status_id": uuid.New()},
		},
		actorID,
		"Test Story",
		"testuser",
	)

	assert.Equal(t, "story_update", updateNotif.Type)
	assert.Equal(t, "testuser updated the status", updateNotif.Description)

	// Test unassignment notification
	unassignNotif := createUnassignmentNotification(
		assigneeID,
		events.StoryUpdatedPayload{StoryID: storyID, WorkspaceID: workspaceID},
		actorID,
		"Test Story",
		"testuser",
	)

	assert.Equal(t, "story_unassignment", unassignNotif.Type)
	assert.Equal(t, "testuser unassigned your story", unassignNotif.Description)
}
