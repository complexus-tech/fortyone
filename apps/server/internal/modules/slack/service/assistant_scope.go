package slack

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type eventAgentSettingsRepository interface {
	GetAgentSettings(ctx context.Context, workspaceID uuid.UUID) (slackAgentSettingsRecord, error)
}

type eventAssistantChannelAudienceRepository interface {
	GetAuthorizedAssistantChannelTeamScope(
		ctx context.Context,
		workspaceID, slackWorkspaceID uuid.UUID,
		slackChannelID string,
		userID uuid.UUID,
	) (assistantChannelTeamScope, error)
}

type eventTeamMembershipRepository interface {
	ListWorkspaceTeamsForUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]slackTeamRecord, error)
}

func (p *EventProcessor) agentSettings(ctx context.Context, workspaceID uuid.UUID) (CoreSlackAgentSettings, error) {
	repository, ok := p.repo.(eventAgentSettingsRepository)
	if !ok {
		// Test and alternate adapters written before workspace-level agent
		// settings retain the secure product defaults. The production Slack
		// repository implements this interface and persists administrator choices.
		return CoreSlackAgentSettings{}, nil
	}
	record, err := repository.GetAgentSettings(ctx, workspaceID)
	if err != nil {
		return CoreSlackAgentSettings{}, err
	}
	return toCoreSlackAgentSettings(record), nil
}

func (p *EventProcessor) authorizedAssistantTeamScope(
	ctx context.Context,
	workspaceID uuid.UUID,
	installation slackWorkspaceRecord,
	userID uuid.UUID,
	event normalizedSlackEvent,
) (assistantChannelTeamScope, error) {
	if event.Kind == slackEventKindDirect {
		repository, ok := p.repo.(eventTeamMembershipRepository)
		if !ok {
			return assistantChannelTeamScope{}, errors.New("slack team membership repository is not configured")
		}
		teams, err := repository.ListWorkspaceTeamsForUser(ctx, workspaceID, userID)
		if err != nil {
			return assistantChannelTeamScope{}, err
		}
		teamIDs := slackTeamRecordIDs(teams)
		return assistantChannelTeamScope{
			AllowedTeamIDs: teamIDs,
			SharedTeamIDs:  append([]uuid.UUID(nil), teamIDs...),
		}, nil
	}
	repository, ok := p.repo.(eventAssistantChannelAudienceRepository)
	if !ok {
		return assistantChannelTeamScope{}, errors.New("slack assistant channel audience repository is not configured")
	}
	return repository.GetAuthorizedAssistantChannelTeamScope(
		ctx,
		workspaceID,
		installation.ID,
		event.ChannelID,
		userID,
	)
}
