package notifications

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProcessFeedbackCommentCreated(t *testing.T) {
	actorID := uuid.New()
	recipientID := uuid.New()
	payload := events.FeedbackCommentCreatedPayload{
		CommentID:     uuid.New(),
		FeedbackID:    uuid.New(),
		FeedbackTitle: "Export roadmap to PDF",
		FeedbackSlug:  "export-roadmap-to-pdf",
		WorkspaceID:   uuid.New(),
		RecipientID:   recipientID,
		Content:       "We are working on this.",
	}

	rules := NewRules(nil, &scheduleRulesStories{}, nil, nil)
	result := rules.ProcessFeedbackCommentCreated(context.Background(), payload, actorID)

	assert.Len(t, result, 1)
	assert.Equal(t, recipientID, result[0].RecipientID)
	assert.Equal(t, "feedback_comment", string(result[0].Type))
	assert.Equal(t, "feedback", string(result[0].EntityType))
	assert.Equal(t, payload.FeedbackID, result[0].EntityID)
	assert.Equal(t, payload.FeedbackTitle, result[0].Title)
	assert.Equal(t, "{actor} commented on your feedback", result[0].Message.Template)

	assert.Empty(t, rules.ProcessFeedbackCommentCreated(context.Background(), payload, recipientID))

	payload.IsReply = true
	replyResult := rules.ProcessFeedbackCommentCreated(context.Background(), payload, actorID)
	assert.Len(t, replyResult, 1)
	assert.Equal(t, "{actor} replied to your comment", replyResult[0].Message.Template)
}

func TestProcessFeedbackStatusUpdated(t *testing.T) {
	actorID := uuid.New()
	recipientID := uuid.New()
	payload := events.FeedbackStatusUpdatedPayload{
		EventID:       uuid.New(),
		FeedbackID:    uuid.New(),
		FeedbackTitle: "Export roadmap to PDF",
		FeedbackSlug:  "export-roadmap-to-pdf",
		WorkspaceID:   uuid.New(),
		RecipientID:   recipientID,
		Status:        "in_progress",
	}

	rules := NewRules(nil, &scheduleRulesStories{}, nil, nil)
	result := rules.ProcessFeedbackStatusUpdated(context.Background(), payload, actorID)

	assert.Len(t, result, 1)
	assert.Equal(t, "feedback_status_update", string(result[0].Type))
	assert.Equal(t, "in progress", result[0].Message.Variables["status"].Value)
	assert.Empty(t, rules.ProcessFeedbackStatusUpdated(context.Background(), payload, recipientID))
}

func TestProcessFeedbackUpdatePublishedUsesRecipientDedupeKey(t *testing.T) {
	actorID, recipientID := uuid.New(), uuid.New()
	payload := events.FeedbackUpdatePublishedPayload{
		PublicationEventID: uuid.New(), UpdateID: uuid.New(), LinkedItemID: uuid.New(), WorkspaceID: uuid.New(),
		RecipientID: recipientID, UpdateTitle: "Dark mode shipped", UpdateSlug: "dark-mode-shipped",
	}
	rules := NewRules(nil, &scheduleRulesStories{}, nil, nil)

	result := rules.ProcessFeedbackUpdatePublished(context.Background(), payload, actorID)

	assert.Len(t, result, 1)
	assert.Equal(t, "feedback_update_published", string(result[0].Type))
	assert.Equal(t, "feedback", string(result[0].EntityType))
	assert.Equal(t, payload.LinkedItemID, result[0].EntityID)
	assert.Equal(t, "feedback-update:"+payload.PublicationEventID.String()+":"+recipientID.String(), result[0].DedupeKey)
	assert.Empty(t, rules.ProcessFeedbackUpdatePublished(context.Background(), payload, recipientID))
}

func TestProcessFeedbackItemMergedUsesStableRecipientDedupeKey(t *testing.T) {
	actorID, recipientID := uuid.New(), uuid.New()
	payload := events.FeedbackItemMergedPayload{
		MergeEventID: uuid.New(), SourceItemID: uuid.New(), TargetItemID: uuid.New(),
		TargetItemTitle: "Canonical dark mode request", TargetItemSlug: "canonical-dark-mode-request",
		WorkspaceID: uuid.New(), RecipientID: recipientID,
	}
	rules := NewRules(nil, &scheduleRulesStories{}, nil, nil)

	result := rules.ProcessFeedbackItemMerged(context.Background(), payload, actorID)

	assert.Len(t, result, 1)
	assert.Equal(t, "feedback_item_merged", string(result[0].Type))
	assert.Equal(t, "feedback", string(result[0].EntityType))
	assert.Equal(t, payload.TargetItemID, result[0].EntityID)
	assert.Equal(t, "feedback-merge:"+payload.MergeEventID.String()+":"+recipientID.String(), result[0].DedupeKey)
	assert.Equal(t, "{actor} merged feedback into {feedback}", result[0].Message.Template)
	assert.Empty(t, rules.ProcessFeedbackItemMerged(context.Background(), payload, recipientID))
}
