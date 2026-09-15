package notifications

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReassignmentSnapshotIdentifiesAssigneeDespiteDuplicateNames(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		maya     bool
		self     bool
		template string
	}{
		{name: "different users with same username", template: "{actor} reassigned task to {assignee}"},
		{name: "actual self reassignment", self: true, template: "{actor} reassigned task to themself"},
		{name: "Maya reassignment", maya: true, template: "Maya reassigned this task to {assignee}: {reason}"},
	} {
		t.Run(test.name, func(t *testing.T) {
			actorID, oldAssigneeID, newAssigneeID := uuid.New(), uuid.New(), uuid.New()
			if test.self {
				newAssigneeID = actorID
			}
			rules := NewRules(nil, &scheduleRulesStories{}, scheduleRulesUsers{
				actorID:       {Username: "alex"},
				newAssigneeID: {Username: "alex"},
			}, nil)
			payload := events.StoryUpdatedPayload{
				StoryID: uuid.New(), WorkspaceID: uuid.New(),
				AssigneeID: &oldAssigneeID, Updates: map[string]any{"assignee_id": newAssigneeID},
			}
			if test.maya {
				payload.Source = events.StoryUpdateSourceMaya
				payload.Reason = "Fits the task requirements."
			}

			notifications := rules.handleReassignment(context.Background(), payload, actorID)
			require.NotEmpty(t, notifications)
			message := notifications[0].Message
			require.Equal(t, oldAssigneeID, notifications[0].RecipientID)
			require.Equal(t, test.template, message.Template)
			require.Equal(t, "alex", message.Variables["assignee"].Value)
			require.Equal(t, map[string]uuid.UUID{"assignee": newAssigneeID}, message.IdentityReferences)
			require.Nil(t, message.Public().IdentityReferences)
			if !test.self {
				require.Len(t, notifications, 2)
				require.Equal(t, newAssigneeID, notifications[1].RecipientID)
				require.Empty(t, notifications[1].Message.IdentityReferences, "recipient-only messages do not need a redundant identity snapshot")
			}
		})
	}
}
