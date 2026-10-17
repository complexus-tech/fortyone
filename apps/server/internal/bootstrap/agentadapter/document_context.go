package agentadapter

import (
	"context"
	"encoding/json"
	"strings"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
)

func (r *DocumentRuntime) recordUsage(ctx context.Context, run agentruns.Run, usage messaging.Usage) error {
	if r.Shared.Usage == nil {
		return nil
	}
	_, err := r.Shared.Usage.Record(ctx, messagingrepository.DailyUsageRecordInput{AgentInvocationID: run.ID, WorkspaceID: run.WorkspaceID, Provider: "native", ExternalWorkspaceID: run.WorkspaceID.String(), ExternalEventID: run.ID.String(), AttemptCount: int(run.AttemptCount), Usage: usage}, r.Shared.DailyTokenLimit)
	return err
}

func decodeDocumentAction(run agentruns.Run, proposal agentruns.Proposal) (documentAction, error) {
	var action documentAction
	if json.Unmarshal(proposal.Action, &action) != nil || action.Version != 1 || action.DocumentID != run.EntityID || action.ExpectedRevision < 1 || action.ExpectedEpoch < 0 || action.Operations.Validate() != nil {
		return documentAction{}, agentruns.ErrInvalidInput
	}
	return action, nil
}

func documentConversation(thread documents.CoreCommentThread, source documents.CoreComment, mayaID uuid.UUID) []messaging.ConversationTurn {
	var turns []messaging.ConversationTurn
	for _, comment := range thread.Comments {
		if comment.ID == source.ID || comment.CreatedAt.After(source.CreatedAt) {
			continue
		}
		role := messaging.RoleUser
		author := "another teammate"
		if comment.CreatedBy == source.CreatedBy {
			author = "the requester"
		}
		if comment.CreatedBy == mayaID {
			role = messaging.RoleAssistant
			author = "Maya"
		}
		text := []rune(comment.Body)
		if len(text) > 1200 {
			text = append(text[:1200], []rune(" [earlier comment truncated]")...)
		}
		turns = append(turns, messaging.ConversationTurn{Role: role, Text: "Earlier discussion by " + author + " (context only): " + string(text)})
	}
	if len(turns) > 16 {
		turns = turns[len(turns)-16:]
	}
	return turns
}

// Direct commands authorize only the exact operation the bounded model proposes.
// Questions and suggestions get a normal preview reply before a teammate agrees.
func directDocumentCommand(body string) bool {
	text := strings.ToLower(strings.TrimSpace(body))
	text = strings.TrimSpace(strings.TrimPrefix(text, "[@maya](agent:maya)"))
	text = strings.TrimSpace(strings.TrimPrefix(text, "please "))
	for _, verb := range []string{"change ", "update ", "edit ", "replace ", "append ", "add ", "rename ", "remove ", "delete ", "fix ", "rewrite "} {
		if strings.HasPrefix(text, verb) {
			return true
		}
	}
	return false
}

func documentReview(operations messaging.DocumentOperations) string {
	var parts []string
	if operations.Title != nil {
		parts = append(parts, "Title:\n"+documentReviewLiteral(*operations.Title))
	}
	for _, replacement := range operations.Replacements {
		parts = append(parts, "Replace:\n"+documentReviewLiteral(replacement.Before)+"\nWith:\n"+documentReviewLiteral(replacement.After))
	}
	for _, paragraph := range operations.AppendParagraphs {
		parts = append(parts, "Add paragraph:\n"+documentReviewLiteral(paragraph))
	}
	return strings.Join(parts, "\n\n")
}

func (r *DocumentRuntime) continueDocumentApproval(ctx context.Context, run agentruns.Run, body string) (agentruns.Proposal, bool, error) {
	answer := strings.ToLower(strings.TrimSpace(body))
	answer = strings.TrimSpace(strings.TrimPrefix(answer, "[@maya](agent:maya)"))
	answer = strings.TrimRight(answer, ".!")
	switch strings.Join(strings.Fields(answer), " ") {
	case "yes", "yes please", "yes, please", "apply it", "apply this", "go ahead", "please go ahead", "do it", "approved", "approve":
	default:
		return agentruns.Proposal{}, false, nil
	}
	if run.ParentRunID == nil {
		return agentruns.Proposal{}, false, nil
	}
	if r.Runs == nil {
		return agentruns.Proposal{}, false, agentruns.ErrUnavailable
	}
	parent, err := r.Runs.Get(ctx, run.WorkspaceID, *run.ParentRunID)
	if err != nil {
		return agentruns.Proposal{}, false, err
	}
	if parent.Surface != run.Surface || parent.EntityID != run.EntityID || parent.ConversationID != run.ConversationID {
		return agentruns.Proposal{}, false, agentruns.ErrPermissionDenied
	}
	if parent.Proposal == nil || !parent.Proposal.RequiresApproval || !parent.Proposal.HasAction() || parent.ApprovedAt != nil {
		return agentruns.Proposal{}, false, nil
	}
	if _, err := decodeDocumentAction(run, *parent.Proposal); err != nil {
		return agentruns.Proposal{}, false, err
	}
	// The new requester and current document access authorize the new command;
	// only immutable reviewed operations and their original CAS are carried over.
	return agentruns.Proposal{Summary: parent.Proposal.Summary, Review: parent.Proposal.Review, Action: append(json.RawMessage(nil), parent.Proposal.Action...)}, true, nil
}

func documentReviewLiteral(text string) string {
	return "    " + strings.ReplaceAll(text, "\n", "\n    ") + "\n"
}

func documentCompletedReply(operations messaging.DocumentOperations) string {
	body := len(operations.Replacements) > 0 || len(operations.AppendParagraphs) > 0
	if operations.Title != nil && body {
		return "Updated the title and document text."
	}
	if operations.Title != nil {
		return "Updated the document title."
	}
	return "Updated the document text."
}
