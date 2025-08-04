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

			// Test hasNonAssignmentUpdates function
			rules := &Rules{}
			hasRelevant := rules.hasNonAssignmentUpdates(tt.payload)
			switch tt.name {
			case "No relevant updates":
				assert.False(t, hasRelevant, "Should not have relevant updates for description change")
			case "Rule 2: Status update notification":
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
			switch tt.name {
			case "Rule 3: Unassignment notification":
				assert.True(t, isUnassign, "Should detect unassignment")
			case "Rule 1: New assignment notification":
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

	rules := &Rules{}
	assignmentNotif := rules.createNotification(
		assigneeID,
		events.StoryUpdatedPayload{StoryID: storyID, WorkspaceID: workspaceID},
		actorID,
		"story_update",
		"Test Story",
		assignmentMessage,
	)

	assert.Equal(t, "story_update", assignmentNotif.Type)
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

	rules = &Rules{}
	updateNotif := rules.createNotification(
		assigneeID,
		events.StoryUpdatedPayload{
			StoryID:     storyID,
			WorkspaceID: workspaceID,
			Updates:     map[string]any{"priority": "High"},
		},
		actorID,
		"story_update",
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

	rules = &Rules{}
	unassignNotif := rules.createNotification(
		assigneeID,
		events.StoryUpdatedPayload{StoryID: storyID, WorkspaceID: workspaceID},
		actorID,
		"story_update",
		"Test Story",
		unassignMessage,
	)

	assert.Equal(t, "story_update", unassignNotif.Type)
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
					"field": {Value: "Priority", Type: "field"},
					"value": {Value: "High", Type: "value"},
				},
			},
			description: "Should generate structured priority update message",
		},
		{
			name:    "Status update",
			updates: map[string]any{"status_id": uuid.New().String()},
			expected: NotificationMessage{
				Template: "{actor} updated {field}",
				Variables: map[string]Variable{
					"actor": {Value: "jack", Type: "actor"},
					"field": {Value: "Status", Type: "field"},
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
					"field": {Value: "Deadline", Type: "field"},
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

			result := rules.generateNonAssignmentUpdateMessage("jack", tt.updates)

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

func TestProcessCommentCreated(t *testing.T) {
	actorID := uuid.New()
	assigneeID := uuid.New()
	workspaceID := uuid.New()

	tests := []struct {
		name              string
		payload           events.CommentCreatedPayload
		actorID           uuid.UUID
		expectedCount     int
		expectedType      string
		expectedRecipient uuid.UUID
	}{
		{
			name: "should notify assignee when someone comments on their story",
			payload: events.CommentCreatedPayload{
				CommentID:   uuid.New(),
				StoryID:     uuid.New(),
				StoryTitle:  "Test Story",
				AssigneeID:  &assigneeID,
				WorkspaceID: workspaceID,
				Content:     "Great work!",
				Mentions:    []uuid.UUID{},
			},
			actorID:           actorID,
			expectedCount:     1,
			expectedType:      "story_comment",
			expectedRecipient: assigneeID,
		},
		{
			name: "should not notify when assignee comments on their own story",
			payload: events.CommentCreatedPayload{
				CommentID:   uuid.New(),
				StoryID:     uuid.New(),
				StoryTitle:  "Test Story",
				AssigneeID:  &assigneeID,
				WorkspaceID: workspaceID,
				Content:     "Working on this",
				Mentions:    []uuid.UUID{},
			},
			actorID:       assigneeID, // Same as assignee
			expectedCount: 0,
		},
		{
			name: "should not notify when story has no assignee",
			payload: events.CommentCreatedPayload{
				CommentID:   uuid.New(),
				StoryID:     uuid.New(),
				StoryTitle:  "Test Story",
				AssigneeID:  nil,
				WorkspaceID: workspaceID,
				Content:     "Unassigned story comment",
				Mentions:    []uuid.UUID{},
			},
			actorID:       actorID,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := NewRules(nil, nil, nil)

			notifications, err := rules.ProcessCommentCreated(context.Background(), tt.payload, tt.actorID)
			assert.NoError(t, err)
			assert.Len(t, notifications, tt.expectedCount)

			if tt.expectedCount > 0 {
				notification := notifications[0]
				assert.Equal(t, tt.expectedRecipient, notification.RecipientID)
				assert.Equal(t, tt.expectedType, notification.Type)
				assert.Equal(t, "story", notification.EntityType)
				assert.Equal(t, tt.payload.StoryID, notification.EntityID)
				assert.Equal(t, tt.payload.StoryTitle, notification.Title)
				assert.Equal(t, tt.actorID, notification.ActorID)
				assert.Equal(t, tt.payload.WorkspaceID, notification.WorkspaceID)
			}
		})
	}
}

func TestProcessCommentReplied(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	parentAuthorID := uuid.New()

	tests := []struct {
		name              string
		payload           events.CommentRepliedPayload
		actorID           uuid.UUID
		expectedCount     int
		expectedType      string
		expectedRecipient uuid.UUID
	}{
		{
			name: "should notify parent comment author when someone replies",
			payload: events.CommentRepliedPayload{
				CommentID:       uuid.New(),
				ParentCommentID: uuid.New(),
				ParentAuthorID:  parentAuthorID,
				StoryID:         uuid.New(),
				StoryTitle:      "Test Story",
				WorkspaceID:     workspaceID,
				Content:         "Thanks for the feedback!",
				Mentions:        []uuid.UUID{},
			},
			actorID:           actorID,
			expectedCount:     1,
			expectedType:      "comment_reply",
			expectedRecipient: parentAuthorID,
		},
		{
			name: "should not notify when replying to own comment",
			payload: events.CommentRepliedPayload{
				CommentID:       uuid.New(),
				ParentCommentID: uuid.New(),
				ParentAuthorID:  actorID, // Same as actor
				StoryID:         uuid.New(),
				StoryTitle:      "Test Story",
				WorkspaceID:     workspaceID,
				Content:         "Adding more details",
				Mentions:        []uuid.UUID{},
			},
			actorID:       actorID,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := NewRules(nil, nil, nil)

			notifications, err := rules.ProcessCommentReplied(context.Background(), tt.payload, tt.actorID)
			assert.NoError(t, err)
			assert.Len(t, notifications, tt.expectedCount)

			if tt.expectedCount > 0 {
				notification := notifications[0]
				assert.Equal(t, tt.expectedRecipient, notification.RecipientID)
				assert.Equal(t, tt.expectedType, notification.Type)
				assert.Equal(t, "story", notification.EntityType)
				assert.Equal(t, tt.payload.StoryID, notification.EntityID)
				assert.Equal(t, tt.payload.StoryTitle, notification.Title)
				assert.Equal(t, tt.actorID, notification.ActorID)
				assert.Equal(t, tt.payload.WorkspaceID, notification.WorkspaceID)
			}
		})
	}
}

func TestProcessUserMentioned(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	mentionedUserID := uuid.New()

	tests := []struct {
		name              string
		payload           events.UserMentionedPayload
		actorID           uuid.UUID
		expectedCount     int
		expectedType      string
		expectedRecipient uuid.UUID
	}{
		{
			name: "should notify mentioned user",
			payload: events.UserMentionedPayload{
				CommentID:     uuid.New(),
				StoryID:       uuid.New(),
				StoryTitle:    "Test Story",
				WorkspaceID:   workspaceID,
				MentionedUser: mentionedUserID,
				Content:       "Hey @user, check this out!",
			},
			actorID:           actorID,
			expectedCount:     1,
			expectedType:      "mention",
			expectedRecipient: mentionedUserID,
		},
		{
			name: "should not notify when mentioning yourself",
			payload: events.UserMentionedPayload{
				CommentID:     uuid.New(),
				StoryID:       uuid.New(),
				StoryTitle:    "Test Story",
				WorkspaceID:   workspaceID,
				MentionedUser: actorID, // Same as actor
				Content:       "Mentioning myself @me",
			},
			actorID:       actorID,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := NewRules(nil, nil, nil)

			notifications, err := rules.ProcessUserMentioned(context.Background(), tt.payload, tt.actorID)
			assert.NoError(t, err)
			assert.Len(t, notifications, tt.expectedCount)

			if tt.expectedCount > 0 {
				notification := notifications[0]
				assert.Equal(t, tt.expectedRecipient, notification.RecipientID)
				assert.Equal(t, tt.expectedType, notification.Type)
				assert.Equal(t, "story", notification.EntityType)
				assert.Equal(t, tt.payload.StoryID, notification.EntityID)
				assert.Equal(t, tt.payload.StoryTitle, notification.Title)
				assert.Equal(t, tt.actorID, notification.ActorID)
				assert.Equal(t, tt.payload.WorkspaceID, notification.WorkspaceID)
			}
		})
	}
}

func TestPreventDuplicateNotifications(t *testing.T) {
	actorID := uuid.New()
	assigneeID := uuid.New()
	workspaceID := uuid.New()
	storyID := uuid.New()

	// Create a rules instance with nil services for this test
	// We'll test the logic by checking if the story lookup would prevent duplicates
	rules := NewRules(nil, nil, nil)

	// Test that when the story assignee is mentioned, we would need to check for duplicates
	// This test verifies the logic exists, but we can't easily mock the full stories service
	// In a real scenario, the ProcessUserMentioned would check the story assignee
	payload := events.UserMentionedPayload{
		CommentID:     uuid.New(),
		StoryID:       storyID,
		StoryTitle:    "Test Story",
		WorkspaceID:   workspaceID,
		MentionedUser: assigneeID,
		Content:       "Hey @assignee, check this out!",
	}

	// With nil stories service, it should still create the notification
	// (the duplicate prevention only works when stories service is available)
	notifications, err := rules.ProcessUserMentioned(context.Background(), payload, actorID)
	assert.NoError(t, err)
	assert.Len(t, notifications, 1, "Should create mention notification when stories service is nil")
}
