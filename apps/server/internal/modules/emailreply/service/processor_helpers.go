package emailreply

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func deterministicReply(subject string, paragraphs ...string) resolvedReply {
	blocks := make([]CopyBlock, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		blocks = append(blocks, CopyBlock{Kind: CopyBlockParagraph, Text: paragraph})
	}
	copy := EmailCopy{
		Subject: replyEmailSubject(subject), Blocks: blocks, PlainText: renderPlainText(blocks),
	}
	return resolvedReply{Copy: copy, Kind: MessageKindReceipt, Key: "control"}
}

func resumedProposalReply(subject string, proposal Proposal, eventID string) resolvedReply {
	blocks := []CopyBlock{
		{Kind: CopyBlockParagraph, Text: "Here’s the change I’m ready to make:"},
		{Kind: CopyBlockCallout, Text: proposalSummary(proposal)},
		{Kind: CopyBlockParagraph, Text: "Reply CONFIRM to apply it, or CANCEL to leave everything unchanged."},
	}
	copy := EmailCopy{
		Subject: replyEmailSubject(subject), Blocks: blocks, PlainText: renderPlainText(blocks),
	}
	return resolvedReply{Copy: copy, Kind: MessageKindProposal, Key: "proposal:" + eventID}
}

func deterministicReplyMessageID(threadID uuid.UUID, eventID, key string) string {
	digest := sha256.Sum256([]byte(threadID.String() + ":" + strings.TrimSpace(eventID) + ":" + key))
	return "<maya-email-" + hex.EncodeToString(digest[:16]) + "@fortyone.app>"
}

func boundedProviderMessageID(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" || strings.ContainsAny(value, "\r\n\x00") {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		value = "missing"
	}
	if len(value) <= 998 {
		return value
	}
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func emailReplyReferences(thread Thread) []string {
	values := []string{thread.RootInternetMessageID, thread.LatestInternetMessageID}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (processor *Processor) failUnreadableEvent(ctx context.Context, scope, eventID string, cause error) error {
	receipt, claimed, err := processor.store.StartInboundEvent(ctx, Provider, scope, eventID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	stateCtx, cancel := processorStateContext(ctx)
	defer cancel()
	if err := processor.store.CompleteInboundEvent(stateCtx, receipt.ID, "failed", truncateProcessorError(cause)); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func terminalInboundStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "completed", "ignored", "cancelled":
		return true
	default:
		return false
	}
}

func truncateProcessorError(err error) string {
	if err == nil {
		return ""
	}
	return truncateRunes(err.Error(), maximumStoredErrorRunes)
}

func truncateRunes(value string, maximum int) string {
	runes := []rune(value)
	if len(runes) <= maximum {
		return value
	}
	if maximum <= 1 {
		return "…"
	}
	return string(runes[:maximum-1]) + "…"
}
