package agentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

const DocumentCommentSurface = "document_comment"

type DocumentRuntime struct {
	Documents      *documents.Service
	Assistant      messaging.DocumentAssistant
	Shared         *CommentRuntime
	Runs           RunReader
	EditsAvailable func(context.Context) (bool, error)
}

type documentAction struct {
	Version          int                          `json:"version"`
	DocumentID       uuid.UUID                    `json:"documentId"`
	ExpectedRevision int64                        `json:"expectedRevision"`
	ExpectedEpoch    int64                        `json:"expectedEpoch"`
	Operations       messaging.DocumentOperations `json:"operations"`
}

func (r *DocumentRuntime) Authorize(ctx context.Context, run agentruns.Run) error {
	if run.Surface != DocumentCommentSurface || run.TeamID != uuid.Nil || r.Shared == nil || !r.Shared.Enabled || r.Shared.Access == nil || r.Shared.Memberships == nil || r.Documents == nil {
		return agentruns.ErrPermissionDenied
	}
	workspace, err := r.Shared.Memberships.Get(ctx, run.WorkspaceID, run.RequesterID)
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return agentruns.ErrPermissionDenied
		}
		return err
	}
	if workspace.DeletedAt != nil || authorization.RequireMinimumWorkspaceRole(authorization.WorkspaceRole(workspace.UserRole), authorization.WorkspaceRoleMember) != nil {
		return agentruns.ErrPermissionDenied
	}
	allowed, err := r.Shared.Access.WorkspaceCanUseMaya(ctx, run.WorkspaceID)
	if err != nil {
		return err
	}
	if !allowed {
		return agentruns.ErrPermissionDenied
	}
	source, err := r.Documents.GetComment(ctx, run.WorkspaceID, run.RequesterID, run.EntityID, run.SourceID)
	if err != nil {
		return documentAccessError(err)
	}
	if source.Comment.CreatedBy != run.RequesterID || source.Thread.ID != run.ConversationID || source.Document.ArchivedAt != nil {
		return agentruns.ErrSourceChanged
	}
	if run.Outcome == nil && (SourceHash(source.Comment.Body) != run.SourceHash || source.Comment.UpdatedAt.UTC().Format(time.RFC3339Nano) != run.SourceRevision) {
		return agentruns.ErrSourceChanged
	}
	return nil
}

func (r *DocumentRuntime) Prepare(ctx context.Context, run agentruns.Run) (agentruns.Proposal, error) {
	source, err := r.Documents.GetComment(ctx, run.WorkspaceID, run.RequesterID, run.EntityID, run.SourceID)
	if err != nil {
		return agentruns.Proposal{}, documentAccessError(err)
	}
	if proposal, accepted, err := r.continueDocumentApproval(ctx, run, source.Comment.Body); accepted || err != nil {
		return proposal, err
	}
	if r.Assistant == nil {
		return agentruns.Proposal{}, agentruns.ErrUnavailable
	}
	if err := r.Shared.admit(ctx, run); err != nil {
		return agentruns.Proposal{}, err
	}
	members, err := r.Documents.ListCommentMentions(ctx, run.WorkspaceID, run.RequesterID, run.EntityID)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	targets := make([]messaging.DocumentMentionTarget, 0, min(len(members), 500))
	for _, member := range members {
		if len(targets) == 500 {
			break
		}
		targets = append(targets, messaging.DocumentMentionTarget{ID: member.ID, Name: member.Name})
	}
	available, err := r.editsAvailable(ctx)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	content := source.Document.ContentText
	if len(content) > 120000 {
		return agentruns.Proposal{}, &agentruns.Failure{Code: "document_too_large", Summary: "This document is too long for me to review safely in one request. Please ask about a smaller selection."}
	}
	response, responseErr := r.Assistant.RespondDocument(ctx, messaging.DocumentRequest{
		UserID: run.RequesterID, Title: source.Document.Title, Content: content, Selection: source.Thread.Quote, Prompt: source.Comment.Body,
		CanEdit: available && source.Document.CanEdit && source.Document.Collaborative, Members: targets,
		Conversation: documentConversation(source.Thread, source.Comment, r.Shared.MayaID),
	})
	if err := r.recordUsage(ctx, run, response.Usage); err != nil {
		return agentruns.Proposal{}, errors.Join(responseErr, err)
	}
	if responseErr != nil {
		return agentruns.Proposal{}, responseErr
	}
	if response.Operations == nil {
		return agentruns.Proposal{Reply: response.Reply, NeedsInput: response.NeedsInput}, nil
	}
	if !available || !source.Document.CanEdit || !source.Document.Collaborative {
		return agentruns.Proposal{}, agentruns.ErrPermissionDenied
	}
	action := documentAction{Version: 1, DocumentID: run.EntityID, ExpectedRevision: source.Document.Revision, ExpectedEpoch: source.Document.CollaborationEpoch, Operations: *response.Operations}
	encoded, err := json.Marshal(action)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	review := documentReview(action.Operations)
	if len([]rune(review)) > 8500 {
		return agentruns.Proposal{Reply: "That change is too large to review safely in one comment. Please ask me for a smaller edit.", NeedsInput: true}, nil
	}
	proposal := agentruns.Proposal{Summary: "Update this document", Review: review, Action: encoded, RequiresApproval: !directDocumentCommand(source.Comment.Body)}
	if proposal.RequiresApproval {
		proposal.Reply = "Here’s the change I suggest:\n\n" + review + "\n\nWould you like me to apply it?"
	}
	return proposal, nil
}

func (r *DocumentRuntime) Apply(context.Context, agentruns.Run, agentruns.Proposal) (agentruns.Outcome, error) {
	return agentruns.Outcome{}, agentruns.ErrConflict // Document writes require the authoritative collaboration receipt.
}

func (r *DocumentRuntime) UsesDeferredAction(run agentruns.Run) bool {
	return run.Surface == DocumentCommentSurface
}

func (r *DocumentRuntime) BeginApply(ctx context.Context, run agentruns.Run, proposal agentruns.Proposal) error {
	action, err := decodeDocumentAction(run, proposal)
	if err != nil {
		return err
	}
	available, err := r.editsAvailable(ctx)
	if err != nil {
		return err
	}
	if !available {
		return &agentruns.Failure{Code: "editing_unavailable", Summary: "Document editing is temporarily unavailable. Please try again shortly."}
	}
	document, err := r.Documents.Get(ctx, run.WorkspaceID, run.RequesterID, run.EntityID)
	if err != nil {
		return documentAccessError(err)
	}
	if !document.CanEdit || !document.Collaborative {
		return agentruns.ErrPermissionDenied
	}
	operations, err := json.Marshal(action.Operations)
	if err != nil {
		return err
	}
	_, err = r.Documents.EnqueueAgentEdit(ctx, documents.CoreAgentEditInput{
		InvocationID: run.ID, WorkspaceID: run.WorkspaceID, RequesterID: run.RequesterID, DocumentID: run.EntityID,
		ExpectedRevision: action.ExpectedRevision, ExpectedEpoch: action.ExpectedEpoch, Operations: operations,
	})
	return documentAccessError(err)
}

func (r *DocumentRuntime) PollApply(ctx context.Context, run agentruns.Run, proposal agentruns.Proposal) (agentruns.Outcome, bool, error) {
	edit, err := r.Documents.GetAgentEdit(ctx, run.WorkspaceID, run.ID)
	if err != nil {
		return agentruns.Outcome{}, false, err
	}
	switch edit.Status {
	case "pending":
		return agentruns.Outcome{}, false, nil
	case "applied":
		receipt, err := json.Marshal(struct {
			EditID   uuid.UUID `json:"editId"`
			Revision *int64    `json:"revision"`
		}{edit.ID, edit.ResultingRevision})
		action, actionErr := decodeDocumentAction(run, proposal)
		if actionErr != nil {
			return agentruns.Outcome{}, false, actionErr
		}
		return agentruns.Outcome{Reply: documentCompletedReply(action.Operations), Receipt: receipt}, true, err
	case "conflict":
		return agentruns.Outcome{}, false, &agentruns.Failure{Code: "document_changed", Summary: "The document changed while I was preparing this edit. Please ask me again so I can use the latest version."}
	case "forbidden":
		return agentruns.Outcome{}, false, agentruns.ErrPermissionDenied
	default:
		return agentruns.Outcome{}, false, &agentruns.Failure{Code: "document_edit_failed", Summary: "I couldn’t safely apply that edit. The document has been left unchanged; please try a smaller change."}
	}
}

func (r *DocumentRuntime) Publish(ctx context.Context, run agentruns.Run, outcome agentruns.Outcome) (uuid.UUID, error) {
	return r.publish(ctx, run, "reply", outcome.Reply)
}
func (r *DocumentRuntime) PublishProposal(ctx context.Context, run agentruns.Run, proposal agentruns.Proposal) (uuid.UUID, error) {
	return r.publish(ctx, run, "proposal", proposal.Reply)
}
func (r *DocumentRuntime) PublishFailure(ctx context.Context, run agentruns.Run, summary string) (uuid.UUID, error) {
	return r.publish(ctx, run, "failure", strings.NewReplacer("Maya has", "I have", "Maya is", "I am", "Maya", "I").Replace(summary))
}

func (r *DocumentRuntime) publish(ctx context.Context, run agentruns.Run, kind, body string) (uuid.UUID, error) {
	comment, err := r.Documents.CreateAgentReply(ctx, documents.CoreAgentReplyInput{
		WorkspaceID: run.WorkspaceID, DocumentID: run.EntityID, RequesterID: run.RequesterID, AuthorID: r.Shared.MayaID,
		ThreadID: run.ConversationID, ReplyToCommentID: &run.SourceID, CreationID: uuid.NewSHA1(run.ID, []byte(kind)), Body: body,
	})
	return comment.ID, err
}

func (r *DocumentRuntime) editsAvailable(ctx context.Context) (bool, error) {
	if r.EditsAvailable == nil {
		return false, nil
	}
	return r.EditsAvailable(ctx)
}

func documentAccessError(err error) error {
	if errors.Is(err, documents.ErrForbidden) || errors.Is(err, documents.ErrNotFound) {
		return errors.Join(agentruns.ErrPermissionDenied, err)
	}
	if errors.Is(err, documents.ErrConflict) {
		return &agentruns.Failure{Code: "document_changed", Summary: "The document changed while I was preparing this edit. Please ask me again using the latest version.", Cause: err}
	}
	return err
}
