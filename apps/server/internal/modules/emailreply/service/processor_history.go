package emailreply

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func (processor *Processor) listUnsummarizedMessages(ctx context.Context, thread Thread) ([]Message, error) {
	messages := make([]Message, 0, max(0, int(thread.NextMessageSequence-thread.SummaryThroughSequence-1)))
	after := thread.SummaryThroughSequence
	for {
		page, err := processor.store.ListMessages(ctx, MessagePageInput{
			ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
			AfterSequence: after, Limit: emailReplyHistoryPageSize,
		})
		if err != nil {
			return nil, err
		}
		messages = append(messages, page.Messages...)
		if !page.HasMore {
			return messages, nil
		}
		if page.NextSequence <= after {
			return nil, errors.New("email conversation pagination did not advance")
		}
		after = page.NextSequence
	}
}

func (processor *Processor) refreshSummary(
	ctx context.Context,
	thread Thread,
	messages []Message,
) (Thread, error) {
	cutoffIndex := len(messages) - emailReplyRecentTurnCount
	if cutoffIndex <= 0 || messages[cutoffIndex-1].Sequence <= thread.SummaryThroughSequence {
		return thread, nil
	}
	if processor.summarizer == nil {
		return thread, errors.New("email conversation summarizer is required once history exceeds the recent context window")
	}
	toSummarize := make([]HistoryTurn, 0, cutoffIndex)
	sequences := make([]int64, 0, cutoffIndex)
	for _, message := range messages[:cutoffIndex] {
		if message.Sequence <= thread.SummaryThroughSequence || message.Role == MessageRoleSystem {
			continue
		}
		role, ok := emailAgentRole(message.Role)
		if !ok || strings.TrimSpace(message.Content) == "" {
			continue
		}
		toSummarize = append(toSummarize, HistoryTurn{
			Role: role, Text: truncateRunes(strings.TrimSpace(message.Content), maximumSummaryTurnRunes), SentAt: message.CreatedAt,
		})
		sequences = append(sequences, message.Sequence)
	}
	for len(toSummarize) > 0 {
		batchSize := summaryBatchSize(toSummarize)
		generation, err := processor.summarizer.Summarize(ctx, SummaryRequest{
			SafetyIdentifier: thread.UserID.String(), PreviousSummary: thread.Summary,
			OmittedTurns: toSummarize[:batchSize],
		})
		if err != nil {
			return thread, err
		}
		updated, err := processor.store.UpdateThreadSummary(ctx, ThreadSummaryUpdate{
			ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
			ExpectedSummaryThroughSequence: thread.SummaryThroughSequence,
			Summary:                        generation.Summary, ThroughSequence: sequences[batchSize-1],
		})
		if err != nil {
			return thread, err
		}
		thread = updated
		toSummarize = toSummarize[batchSize:]
		sequences = sequences[batchSize:]
	}
	return thread, nil
}

func (processor *Processor) refreshSummaryForReply(
	ctx context.Context,
	currentReply string,
	thread Thread,
	messages []Message,
) (Thread, error) {
	if _, isControl := parseControlCommand(currentReply); isControl {
		return thread, nil
	}
	return processor.refreshSummary(ctx, thread, messages)
}

func summaryBatchSize(turns []HistoryTurn) int {
	count, runes := 0, 0
	for count < len(turns) && count < maximumSummaryBatchTurnCount {
		next := len([]rune(turns[count].Text))
		if count > 0 && runes+next > maximumSummaryBatchRunes {
			break
		}
		runes += next
		count++
	}
	if count == 0 {
		return 1
	}
	return count
}

func conversationHistory(messages []Message, afterSequence int64, currentMessageID uuid.UUID) []HistoryTurn {
	history := make([]HistoryTurn, 0, len(messages))
	for _, message := range messages {
		if message.Sequence <= afterSequence || message.ID == currentMessageID || message.Role == MessageRoleSystem {
			continue
		}
		role, ok := emailAgentRole(message.Role)
		if !ok || strings.TrimSpace(message.Content) == "" {
			continue
		}
		history = append(history, HistoryTurn{Role: role, Text: message.Content, SentAt: message.CreatedAt})
	}
	return history
}

func emailAgentRole(role string) (ConversationRole, bool) {
	switch role {
	case MessageRoleUser:
		return ConversationRoleUser, true
	case MessageRoleAssistant:
		return ConversationRoleAssistant, true
	default:
		return "", false
	}
}

func pendingProposalPreviews(records []Proposal) []PendingProposal {
	result := make([]PendingProposal, 0, len(records))
	for _, record := range records {
		result = append(result, PendingProposal{ID: record.ID, Summary: proposalSummary(record)})
	}
	return result
}

func proposalSummary(record Proposal) string {
	var proposal ActionProposal
	if json.Unmarshal(record.ProposedDiff, &proposal) == nil && strings.TrimSpace(proposal.Summary) != "" {
		return proposal.Summary
	}
	return strings.ReplaceAll(strings.TrimSpace(record.ActionKind), "_", " ")
}

func proposalTarget(proposal ActionProposal) (TargetSnapshot, error) {
	switch proposal.Kind {
	case ActionObjectiveUpdate:
		if proposal.Objective != nil {
			return proposal.Objective.Target, nil
		}
	case ActionKeyResultUpdate:
		if proposal.KeyResult != nil {
			return proposal.KeyResult.Target, nil
		}
	case ActionStoryUpdate:
		if proposal.Story != nil {
			return proposal.Story.Target, nil
		}
	case ActionFeedbackStatus:
		if proposal.Feedback != nil {
			return proposal.Feedback.Target, nil
		}
	}
	return TargetSnapshot{}, errors.New("email action proposal has no matching target")
}

func proposalEntityType(kind ActionKind) string {
	switch kind {
	case ActionObjectiveUpdate:
		return string(TargetObjective)
	case ActionKeyResultUpdate:
		return string(TargetKeyResult)
	case ActionStoryUpdate:
		return string(TargetStory)
	case ActionFeedbackStatus:
		return string(TargetFeedback)
	default:
		return "unknown"
	}
}
