package agentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
)

type addressedChatFailure struct{ cause error }

func (err *addressedChatFailure) Error() string { return err.cause.Error() }
func (err *addressedChatFailure) Unwrap() error { return err.cause }

func (r *ChatRuntime) CanPublishFailure(run agentruns.Run, err error) bool {
	if run.Proposal != nil && run.Proposal.Ignore {
		return false
	}
	var payload chatPayload
	if len(run.Payload) > 0 && json.Unmarshal(run.Payload, &payload) != nil {
		return false
	}
	if !payload.Followup || run.Proposal != nil {
		return true
	}
	var addressed *addressedChatFailure
	return errors.As(err, &addressed)
}

type ignoredChatReceipt struct {
	AttemptCount int32           `json:"attemptCount"`
	Usage        messaging.Usage `json:"usage"`
}

func ignoredChatProposal(run agentruns.Run, usage messaging.Usage) (agentruns.Proposal, error) {
	proposal := agentruns.Proposal{Ignore: true}
	if usage.InputTokens == 0 && usage.OutputTokens == 0 && usage.TotalTokens == 0 {
		return proposal, nil
	}
	receipt, err := json.Marshal(ignoredChatReceipt{AttemptCount: run.AttemptCount, Usage: usage})
	if err != nil {
		return agentruns.Proposal{}, err
	}
	proposal.IgnoreReceipt = receipt
	return proposal, nil
}

// Bookkeeping uses the worker context after the proposal commits, not the
// classification deadline. Its original attempt key makes accounting retries
// idempotent even when an acknowledgement or completion commit was lost.
func (r *ChatRuntime) FinalizeIgnored(ctx context.Context, run agentruns.Run, proposal agentruns.Proposal) error {
	if !proposal.Ignore || len(proposal.IgnoreReceipt) == 0 {
		return nil
	}
	var receipt ignoredChatReceipt
	if json.Unmarshal(proposal.IgnoreReceipt, &receipt) != nil || receipt.AttemptCount < 1 {
		return agentruns.ErrInvalidInput
	}
	run.AttemptCount = receipt.AttemptCount
	return r.recordUsage(ctx, run, receipt.Usage)
}

func loadChatHistory(ctx context.Context, chats ChatMessages, scope chatdomain.ActorScope, source chatdomain.Message) ([]chatdomain.Message, error) {
	limit := messaging.MaximumConversationTurns + 1 // Includes the current source.
	if source.ThreadRootID != nil {
		limit--
	}
	history, err := chats.BoundedContext(ctx, chatdomain.ContextQuery{Scope: scope, ConversationID: source.ConversationID, ThreadRootID: source.ThreadRootID, BeforeMessageID: &source.ID, Limit: limit})
	if err != nil {
		return nil, err
	}
	if source.ThreadRootID != nil {
		root, err := chats.GetMessage(ctx, scope, *source.ThreadRootID)
		if err != nil {
			return nil, chatAccessError(err)
		}
		history = append([]chatdomain.Message{root}, history...)
	}
	return history, nil
}

// This is a cheap eligibility gate, not an addressing classifier. Only an
// existing Maya thread reaches the worker; ordinary team chatter never does.
func chatFollowupCandidate(history []chatdomain.Message, source chatdomain.Message, mayaID uuid.UUID) bool {
	if source.ThreadRootID == nil || mayaID == uuid.Nil || source.Author.IsSystem || strings.TrimSpace(source.Text) == "" || chatAddressesAnotherPerson(source, mayaID) {
		return false
	}
	for _, message := range history {
		if message.Author.ID != mayaID || message.DeletedAt != nil || message.Sequence >= source.Sequence || message.ConversationID != source.ConversationID {
			continue
		}
		if message.ID == *source.ThreadRootID || (message.ThreadRootID != nil && *message.ThreadRootID == *source.ThreadRootID) {
			return true
		}
	}
	return false
}

// A leading, canonical mention addresses that teammate. Mentions later in a
// request ("assign this to @Alex") remain context for the classifier.
func chatAddressesAnotherPerson(source chatdomain.Message, mayaID uuid.UUID) bool {
	var root chatContentNode
	if json.Unmarshal(source.RichContent, &root) != nil || root.Type != "doc" {
		return false
	}
	var first func(chatContentNode, int) (uuid.UUID, bool)
	first = func(node chatContentNode, depth int) (uuid.UUID, bool) {
		if depth > 32 || node.Type == "blockquote" || node.Type == "codeBlock" {
			return uuid.Nil, true
		}
		if node.Type == "mention" {
			id, _ := node.Attrs["id"].(string)
			parsed, _ := uuid.Parse(id)
			return parsed, true
		}
		if strings.TrimSpace(node.Text) != "" {
			return uuid.Nil, true
		}
		for _, child := range node.Content {
			if id, found := first(child, depth+1); found {
				return id, true
			}
		}
		return uuid.Nil, false
	}
	id, _ := first(root, 0)
	return id != uuid.Nil && id != mayaID && slices.Contains(source.MentionedUserIDs, id)
}
