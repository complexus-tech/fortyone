package agentadapter

import (
	"context"
	"encoding/json"
	"strings"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	"github.com/complexus-tech/projects-api/internal/platform/agentmentions"
	"golang.org/x/net/html"
)

// A clear answer to Maya's persisted proposal confirms those exact values. A
// different teammate may answer, but the replacement proposal is authorized and
// recorded as that teammate, with the original story-version precondition.
func (r *CommentRuntime) continueApproval(ctx context.Context, run agentruns.Run, content string) (agentruns.Proposal, bool, error) {
	if run.ParentRunID == nil || !acceptsProposal(content) {
		return agentruns.Proposal{}, false, nil
	}
	if r.Runs == nil {
		return agentruns.Proposal{}, false, agentruns.ErrUnavailable
	}
	parent, err := r.Runs.Get(ctx, run.WorkspaceID, *run.ParentRunID)
	if err != nil {
		return agentruns.Proposal{}, false, err
	}
	if parent.Surface != run.Surface || parent.EntityID != run.EntityID || parent.TeamID != run.TeamID || parent.ConversationID != run.ConversationID {
		return agentruns.Proposal{}, false, agentruns.ErrPermissionDenied
	}
	if parent.Proposal == nil || !parent.Proposal.RequiresApproval || !parent.Proposal.HasAction() || parent.ApprovedAt != nil {
		return agentruns.Proposal{}, false, nil
	}
	var previous commentAction
	if err := json.Unmarshal(parent.Proposal.Action, &previous); err != nil {
		return agentruns.Proposal{}, false, agentruns.ErrInvalidInput
	}
	confirmation, err := r.Tools.RebindStoryMutation(ctx, r.toolScope(parent), r.toolScope(run), previous.Token)
	if err != nil {
		return agentruns.Proposal{}, false, mutationFailure(err)
	}
	inspection, err := r.Tools.InspectStoryMutation(ctx, r.toolScope(run), confirmation.Token)
	if err != nil {
		return agentruns.Proposal{}, false, mutationFailure(err)
	}
	if !inspection.DirectEligible {
		return agentruns.Proposal{}, false, agentruns.ErrPermissionDenied
	}
	action, err := json.Marshal(commentAction{Token: confirmation.Token, Fields: inspection.ChangedFields})
	if err != nil {
		return agentruns.Proposal{}, false, err
	}
	return agentruns.Proposal{Summary: confirmation.Prompt, Review: inspection.Review, Action: action}, true, nil
}

func acceptsProposal(content string) bool {
	document, err := html.Parse(strings.NewReader(content))
	if err != nil || containsQuotation(document) {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(agentmentions.Text(content)))
	answer = strings.TrimSpace(strings.TrimPrefix(answer, "@maya"))
	answer = strings.TrimRight(answer, ".!")
	switch strings.Join(strings.Fields(answer), " ") {
	case "yes", "yes please", "yes, please", "yes apply it", "yes, apply it", "apply it", "apply this", "go ahead", "please go ahead", "do it", "please do", "approved", "approve":
		return true
	default:
		return false
	}
}

func containsQuotation(node *html.Node) bool {
	if node.Type == html.ElementNode && (node.Data == "blockquote" || node.Data == "code" || node.Data == "pre") {
		return true
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if containsQuotation(child) {
			return true
		}
	}
	return false
}
