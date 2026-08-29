package slack

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

func (p *EventProcessor) persistAssistantPrompt(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	event normalizedSlackEvent,
	prompt string,
	allowedTeamIDs, sharedTeamIDs []uuid.UUID,
) (uuid.UUID, error) {
	input := assistantConversationInput(workspaceID, userID, event)
	if input.AudienceScope == conversationAudienceChannel {
		input.AudienceFingerprint = assistantAudienceFingerprint(allowedTeamIDs, sharedTeamIDs)
	}
	conversationID, err := p.store.UpsertConversation(ctx, input)
	if err != nil {
		return uuid.Nil, err
	}
	if err := p.store.AppendMessage(ctx, conversationID, event.MessageTS, "user", prompt); err != nil {
		return uuid.Nil, err
	}
	return conversationID, nil
}

func (p *EventProcessor) channelThreadIsSubscribed(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	installation slackWorkspaceRecord,
	event normalizedSlackEvent,
) (bool, error) {
	input := assistantConversationInput(workspaceID, userID, event)
	if input.AudienceScope == conversationAudienceChannel {
		teamScope, err := p.authorizedAssistantTeamScope(ctx, workspaceID, installation, userID, event)
		if err != nil {
			return false, err
		}
		if len(teamScope.AllowedTeamIDs) == 0 {
			return false, nil
		}
		input.AudienceFingerprint = assistantAudienceFingerprint(teamScope.AllowedTeamIDs, teamScope.SharedTeamIDs)
	}
	record, err := findSlackConversation(ctx, p.store, input)
	if err != nil {
		if errors.Is(err, errMessagingRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return slackThreadSubscriptionIsCurrent(record, installation, p.clock.Now()), nil
}

func slackThreadSubscriptionIsCurrent(
	record conversationRecord,
	installation slackWorkspaceRecord,
	now time.Time,
) bool {
	if record.ID == uuid.Nil || record.UpdatedAt.IsZero() {
		return false
	}
	updatedAt := record.UpdatedAt.UTC()
	if installation.AuthorizedAt.IsZero() || updatedAt.Before(installation.AuthorizedAt.UTC()) {
		return false
	}
	return updatedAt.After(now.UTC().Add(-slackThreadSubscriptionTTL))
}

func assistantConversationInput(workspaceID, userID uuid.UUID, event normalizedSlackEvent) conversationInput {
	audienceScope := conversationAudienceActor
	if event.Kind != slackEventKindDirect {
		audienceScope = conversationAudienceChannel
	}
	return conversationInput{
		Provider:            "slack",
		WorkspaceID:         workspaceID,
		ExternalWorkspaceID: event.TeamID,
		ExternalChannelID:   event.ChannelID,
		ExternalThreadID:    conversationThreadID(event),
		UserID:              userID,
		AudienceScope:       audienceScope,
	}
}

func assistantAudienceFingerprint(allowedTeamIDs, sharedTeamIDs []uuid.UUID) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte("slack-channel-audience-v2\x00"))
	for _, scope := range []struct {
		label   string
		teamIDs []uuid.UUID
	}{
		{label: "allowed", teamIDs: allowedTeamIDs},
		{label: "shared", teamIDs: sharedTeamIDs},
	} {
		_, _ = hash.Write([]byte(scope.label))
		_, _ = hash.Write([]byte{0})
		values := uniqueSortedUUIDStrings(scope.teamIDs)
		for _, value := range values {
			_, _ = hash.Write([]byte(value))
			_, _ = hash.Write([]byte{0})
		}
	}
	return "v2:" + hex.EncodeToString(hash.Sum(nil))
}

func uniqueSortedUUIDStrings(teamIDs []uuid.UUID) []string {
	values := make([]string, 0, len(teamIDs))
	seen := make(map[uuid.UUID]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		if teamID == uuid.Nil {
			continue
		}
		if _, exists := seen[teamID]; exists {
			continue
		}
		seen[teamID] = struct{}{}
		values = append(values, teamID.String())
	}
	sort.Strings(values)
	return values
}

func findSlackConversation(
	ctx context.Context,
	store slackConversationFinder,
	input conversationInput,
) (conversationRecord, error) {
	if input.AudienceScope == conversationAudienceChannel {
		if finder, ok := store.(slackChannelConversationFinder); ok {
			return finder.FindChannelConversation(ctx, input)
		}
	}
	return store.FindConversation(ctx, input)
}
