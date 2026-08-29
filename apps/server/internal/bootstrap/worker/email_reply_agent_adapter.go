package workerbootstrap

import (
	"context"

	emailagent "github.com/complexus-tech/projects-api/internal/modules/emailagent/service"
	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/google/uuid"
)

type emailDecisionBackend interface {
	Decide(context.Context, emailagent.Request) (emailagent.Decision, error)
}

type emailDecisionAdapter struct {
	backend emailDecisionBackend
}

var (
	_ emailreply.DecisionPort    = emailDecisionAdapter{}
	_ emailreply.SummaryPort     = emailSummaryAdapter{}
	_ emailreply.CopyRenderer    = emailCopyRenderer{}
	_ emailreply.ReplyThreadPort = emailReplyThreadAdapter{}
)

func (adapter emailDecisionAdapter) Decide(
	ctx context.Context,
	request emailreply.AgentRequest,
) (emailreply.AgentDecision, error) {
	decision, err := adapter.backend.Decide(ctx, toEmailAgentRequest(request))
	if err != nil {
		return emailreply.AgentDecision{}, err
	}
	return toEmailReplyDecision(decision), nil
}

type emailSummaryBackend interface {
	Summarize(context.Context, emailagent.SummaryRequest) (emailagent.SummaryGeneration, error)
}

type emailSummaryAdapter struct {
	backend emailSummaryBackend
}

func (adapter emailSummaryAdapter) Summarize(
	ctx context.Context,
	request emailreply.SummaryRequest,
) (emailreply.SummaryGeneration, error) {
	turns := make([]emailagent.HistoryTurn, 0, len(request.OmittedTurns))
	for _, turn := range request.OmittedTurns {
		turns = append(turns, toEmailAgentHistoryTurn(turn))
	}
	generation, err := adapter.backend.Summarize(ctx, emailagent.SummaryRequest{
		SafetyIdentifier: request.SafetyIdentifier,
		PreviousSummary:  request.PreviousSummary,
		OmittedTurns:     turns,
	})
	return emailreply.SummaryGeneration{Summary: generation.Summary}, err
}

type emailCopyRenderer struct{}

func (emailCopyRenderer) RenderHTML(copy emailreply.EmailCopy) (string, error) {
	return emailagent.RenderHTML(toEmailAgentCopy(copy))
}

type emailReplyThreadAdapter struct {
	service *emailthread.Service
}

func (adapter emailReplyThreadAdapter) NewReplyToken(
	ctx context.Context,
	thread emailreply.Thread,
) (string, error) {
	return adapter.service.NewReplyToken(ctx, toMessagingThread(thread))
}

func (adapter emailReplyThreadAdapter) PrepareReply(
	ctx context.Context,
	input emailreply.ReplyPreparation,
) (string, error) {
	prepared, err := adapter.service.PrepareReply(ctx, emailthread.ReplyInput{
		Thread: toMessagingThread(input.Thread), ReplyToken: input.ReplyToken,
		InternetMessageID: input.InternetMessageID, InReplyTo: input.InReplyTo,
		Subject: input.Subject, Content: input.Content, Kind: input.Kind,
		IdempotencyKey: input.IdempotencyKey, Context: input.Context,
	})
	return prepared.ReplyTo, err
}

func toMessagingThread(thread emailreply.Thread) messaging.EmailThreadRecord {
	return messaging.EmailThreadRecord{
		ID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
		RecipientEmail: thread.RecipientEmail, ExternalThreadID: thread.ExternalThreadID,
		RootInternetMessageID:   thread.RootInternetMessageID,
		LatestInternetMessageID: thread.LatestInternetMessageID,
		Context:                 append([]byte(nil), thread.Context...), Summary: thread.Summary,
		SummaryThroughSequence: thread.SummaryThroughSequence,
		NextMessageSequence:    thread.NextMessageSequence,
	}
}

func toEmailAgentRequest(request emailreply.AgentRequest) emailagent.Request {
	history := make([]emailagent.HistoryTurn, 0, len(request.History))
	for _, turn := range request.History {
		history = append(history, toEmailAgentHistoryTurn(turn))
	}
	facts := make([]emailagent.GroundedFact, 0, len(request.Facts))
	for _, fact := range request.Facts {
		facts = append(facts, emailagent.GroundedFact{
			Reference: fact.Reference, Text: fact.Text,
			ProtectedTokens: append([]string(nil), fact.ProtectedTokens...),
		})
	}
	targets := make([]emailagent.AuthorizedTarget, 0, len(request.Targets))
	for _, target := range request.Targets {
		targets = append(targets, emailagent.AuthorizedTarget{
			Reference: target.Reference, Kind: emailagent.TargetKind(target.Kind),
			DisplayName: target.DisplayName, CurrentState: target.CurrentState,
			ID: target.ID, TeamID: target.TeamID, ExpectedUpdatedAt: target.ExpectedUpdatedAt,
		})
	}
	choices := make([]emailagent.AuthorizedChoice, 0, len(request.Choices))
	for _, choice := range request.Choices {
		choices = append(choices, emailagent.AuthorizedChoice{
			Reference: choice.Reference, Kind: emailagent.ChoiceKind(choice.Kind),
			DisplayName: choice.DisplayName, ID: choice.ID, TeamID: choice.TeamID,
		})
	}
	pending := make([]emailagent.PendingProposal, 0, len(request.PendingProposals))
	for _, proposal := range request.PendingProposals {
		pending = append(pending, emailagent.PendingProposal{ID: proposal.ID, Summary: proposal.Summary})
	}
	return emailagent.Request{
		WorkspaceID: request.WorkspaceID, ActorID: request.ActorID,
		SafetyIdentifier: request.SafetyIdentifier,
		AllowedTeamIDs:   append([]uuid.UUID(nil), request.AllowedTeamIDs...),
		Subject:          request.Subject, Message: request.Message, Summary: request.Summary,
		History: history, Facts: facts, Targets: targets, Choices: choices,
		PendingProposals: pending,
	}
}

func toEmailAgentHistoryTurn(turn emailreply.HistoryTurn) emailagent.HistoryTurn {
	return emailagent.HistoryTurn{
		Role: emailagent.ConversationRole(turn.Role), Text: turn.Text, SentAt: turn.SentAt,
	}
}

func toEmailReplyDecision(decision emailagent.Decision) emailreply.AgentDecision {
	result := emailreply.AgentDecision{
		Intent: emailreply.AgentIntent(decision.Intent),
		Source: emailreply.AgentDecisionSource(decision.Source),
	}
	if decision.Copy != nil {
		copy := toEmailReplyCopy(*decision.Copy)
		result.Copy = &copy
	}
	if decision.Proposal != nil {
		proposal := toEmailReplyProposalValue(*decision.Proposal)
		result.Proposal = &proposal
	}
	if decision.Command != nil {
		result.Command = &emailreply.ControlCommand{ProposalID: decision.Command.ProposalID()}
	}
	return result
}

func toEmailReplyCopy(copy emailagent.EmailCopy) emailreply.EmailCopy {
	blocks := make([]emailreply.CopyBlock, 0, len(copy.Blocks))
	for _, block := range copy.Blocks {
		blocks = append(blocks, emailreply.CopyBlock{
			Kind: emailreply.CopyBlockKind(block.Kind), Text: block.Text,
			Items:      append([]string(nil), block.Items...),
			References: append([]string(nil), block.References...),
		})
	}
	return emailreply.EmailCopy{Subject: copy.Subject, PlainText: copy.PlainText, Blocks: blocks}
}

func toEmailAgentCopy(copy emailreply.EmailCopy) emailagent.EmailCopy {
	blocks := make([]emailagent.CopyBlock, 0, len(copy.Blocks))
	for _, block := range copy.Blocks {
		blocks = append(blocks, emailagent.CopyBlock{
			Kind: emailagent.CopyBlockKind(block.Kind), Text: block.Text,
			Items:      append([]string(nil), block.Items...),
			References: append([]string(nil), block.References...),
		})
	}
	return emailagent.EmailCopy{Subject: copy.Subject, PlainText: copy.PlainText, Blocks: blocks}
}

func toEmailReplyProposalValue(proposal emailagent.ActionProposal) emailreply.ActionProposal {
	result := emailreply.ActionProposal{
		WorkspaceID: proposal.WorkspaceID, ActorID: proposal.ActorID,
		Kind: emailreply.ActionKind(proposal.Kind), Summary: proposal.Summary,
	}
	if proposal.Objective != nil {
		action := emailreply.ObjectiveAction{
			Target: toEmailReplyTarget(proposal.Objective.Target), CheckIn: proposal.Objective.CheckIn,
		}
		if proposal.Objective.Health != nil {
			health := emailreply.ObjectiveHealth(*proposal.Objective.Health)
			action.Health = &health
		}
		result.Objective = &action
	}
	if proposal.KeyResult != nil {
		result.KeyResult = &emailreply.KeyResultAction{
			Target:       toEmailReplyTarget(proposal.KeyResult.Target),
			CurrentValue: proposal.KeyResult.CurrentValue, CheckIn: proposal.KeyResult.CheckIn,
		}
	}
	if proposal.Story != nil {
		result.Story = toEmailReplyStoryAction(*proposal.Story)
	}
	if proposal.Feedback != nil {
		result.Feedback = &emailreply.FeedbackStatusAction{
			Target: toEmailReplyTarget(proposal.Feedback.Target),
			Status: emailreply.FeedbackStatus(proposal.Feedback.Status),
		}
	}
	return result
}

func toEmailReplyStoryAction(action emailagent.StoryAction) *emailreply.StoryAction {
	result := &emailreply.StoryAction{Target: toEmailReplyTarget(action.Target)}
	if action.DueDate != nil {
		result.DueDate = &emailreply.DateChange{
			Operation: emailreply.DateOperation(action.DueDate.Operation), Date: action.DueDate.Date,
		}
	}
	if action.Status != nil {
		result.Status = &emailreply.StatusChange{
			StatusID: action.Status.StatusID, StatusName: action.Status.StatusName,
		}
	}
	if action.Assignee != nil {
		result.Assignee = &emailreply.AssigneeChange{
			Operation:  emailreply.AssigneeOperation(action.Assignee.Operation),
			AssigneeID: action.Assignee.AssigneeID, AssigneeName: action.Assignee.AssigneeName,
		}
	}
	return result
}

func toEmailReplyTarget(target emailagent.TargetSnapshot) emailreply.TargetSnapshot {
	return emailreply.TargetSnapshot{
		ID: target.ID, TeamID: target.TeamID, DisplayName: target.DisplayName,
		ExpectedUpdatedAt: target.ExpectedUpdatedAt,
	}
}
