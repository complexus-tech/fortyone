package notifications

import (
	"context"
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
	assignmentMessage := NotificationMessage{
		Template: "{actor} assigned you a story",
		Variables: map[string]Variable{
			"actor": {Value: "testuser", Type: "actor"},
		},
	}

	assignmentNotif := createAssignmentNotification(
		assigneeID,
		events.StoryUpdatedPayload{StoryID: storyID, WorkspaceID: workspaceID},
		actorID,
		"Test Story",
		assignmentMessage,
	)

	assert.Equal(t, "story_assignment", assignmentNotif.Type)
	assert.Equal(t, "{actor} assigned you a story", assignmentNotif.Message.Template)
	assert.Equal(t, "testuser", assignmentNotif.Message.Variables["actor"].Value)
	assert.Equal(t, "actor", assignmentNotif.Message.Variables["actor"].Type)
	assert.Equal(t, "Test Story", assignmentNotif.Title)

	// Test update notification with priority
	updateMessage := NotificationMessage{
		Template: "{actor} set the {field} to {value}",
		Variables: map[string]Variable{
			"actor": {Value: "testuser", Type: "actor"},
			"field": {Value: "priority", Type: "field"},
			"value": {Value: "High", Type: "value"},
		},
	}

	updateNotif := createUpdateNotification(
		assigneeID,
		events.StoryUpdatedPayload{
			StoryID:     storyID,
			WorkspaceID: workspaceID,
			Updates:     map[string]any{"priority": "High"},
		},
		actorID,
		"Test Story",
		updateMessage,
	)

	assert.Equal(t, "story_update", updateNotif.Type)
	assert.Equal(t, "{actor} set the {field} to {value}", updateNotif.Message.Template)
	assert.Equal(t, "testuser", updateNotif.Message.Variables["actor"].Value)
	assert.Equal(t, "priority", updateNotif.Message.Variables["field"].Value)
	assert.Equal(t, "High", updateNotif.Message.Variables["value"].Value)

	// Test unassignment notification
	unassignMessage := NotificationMessage{
		Template: "{actor} reassigned story to {assignee}",
		Variables: map[string]Variable{
			"actor":    {Value: "testuser", Type: "actor"},
			"assignee": {Value: "tom", Type: "assignee"},
		},
	}

	unassignNotif := createUnassignmentNotification(
		assigneeID,
		events.StoryUpdatedPayload{StoryID: storyID, WorkspaceID: workspaceID},
		actorID,
		"Test Story",
		unassignMessage,
	)

	assert.Equal(t, "story_unassignment", unassignNotif.Type)
	assert.Equal(t, "{actor} reassigned story to {assignee}", unassignNotif.Message.Template)
	assert.Equal(t, "testuser", unassignNotif.Message.Variables["actor"].Value)
	assert.Equal(t, "tom", unassignNotif.Message.Variables["assignee"].Value)
}

func TestGenerateUpdateMessage(t *testing.T) {
	tests := []struct {
		name        string
		updates     map[string]any
		expected    NotificationMessage
		description string
	}{
		{
			name:    "Priority update",
			updates: map[string]any{"priority": "High"},
			expected: NotificationMessage{
				Template: "{actor} set the {field} to {value}",
				Variables: map[string]Variable{
					"actor": {Value: "jack", Type: "actor"},
					"field": {Value: "priority", Type: "field"},
					"value": {Value: "High", Type: "value"},
				},
			},
			description: "Should generate structured priority update message",
		},
		{
			name:    "Status update",
			updates: map[string]any{"status_id": uuid.New().String()},
			expected: NotificationMessage{
				Template: "{actor} changed the {field} to {value}",
				Variables: map[string]Variable{
					"actor": {Value: "jack", Type: "actor"},
					"field": {Value: "status", Type: "field"},
					"value": {Value: "In Progress", Type: "value"},
				},
			},
			description: "Should generate structured status update message",
		},
		{
			name:    "Due date update",
			updates: map[string]any{"end_date": "2024-06-03"},
			expected: NotificationMessage{
				Template: "{actor} set the {field} to {value}",
				Variables: map[string]Variable{
					"actor": {Value: "jack", Type: "actor"},
					"field": {Value: "deadline", Type: "field"},
					"value": {Value: "3 Jun", Type: "date"},
				},
			},
			description: "Should generate structured due date update message",
		},
		{
			name:    "Due date removal",
			updates: map[string]any{"end_date": nil},
			expected: NotificationMessage{
				Template: "{actor} removed the {field}",
				Variables: map[string]Variable{
					"actor": {Value: "jack", Type: "actor"},
					"field": {Value: "deadline", Type: "field"},
				},
			},
			description: "Should generate structured due date removal message",
		},
		{
			name:    "Unknown update",
			updates: map[string]any{"description": "Updated description"},
			expected: NotificationMessage{
				Template: "{actor} updated the story",
				Variables: map[string]Variable{
					"actor": {Value: "jack", Type: "actor"},
				},
			},
			description: "Should generate structured default update message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock rules instance
			rules := &Rules{}

			result := generateUpdateMessage("jack", tt.updates, context.Background(), rules)

			assert.Equal(t, tt.expected.Template, result.Template, tt.description)
			assert.Equal(t, len(tt.expected.Variables), len(result.Variables), "Should have same number of variables")

			for key, expectedVar := range tt.expected.Variables {
				actualVar, exists := result.Variables[key]
				assert.True(t, exists, "Variable %s should exist", key)
				assert.Equal(t, expectedVar.Value, actualVar.Value, "Variable %s value should match", key)
				assert.Equal(t, expectedVar.Type, actualVar.Type, "Variable %s type should match", key)
			}
		})
	}
}
