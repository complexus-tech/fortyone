package slack

import (
	"context"
	"errors"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/google/uuid"
)

type agentSettingsRepository interface {
	GetAgentSettingsForAdmin(ctx context.Context, query slackdomain.WorkspaceActorQuery) (slackdomain.AgentSettings, error)
	UpsertAgentSettingsForAdmin(ctx context.Context, command slackdomain.UpdateAgentSettingsCommand) (slackdomain.AgentSettings, error)
}

type CoreSlackAgentSettings struct {
	Guidance string
}

func (s *Service) GetAgentSettings(ctx context.Context, workspaceID, actorID uuid.UUID) (CoreSlackAgentSettings, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return CoreSlackAgentSettings{}, err
	}
	repository, ok := s.repo.(agentSettingsRepository)
	if !ok {
		return CoreSlackAgentSettings{}, errors.New("slack agent settings repository is not configured")
	}
	record, err := repository.GetAgentSettingsForAdmin(ctx, slackdomain.WorkspaceActorQuery{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		return CoreSlackAgentSettings{}, err
	}
	return toCoreSlackAgentSettings(record), nil
}

func (s *Service) UpdateAgentSettings(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
	input CoreSlackAgentSettings,
) (CoreSlackAgentSettings, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return CoreSlackAgentSettings{}, err
	}
	repository, ok := s.repo.(agentSettingsRepository)
	if !ok {
		return CoreSlackAgentSettings{}, errors.New("slack agent settings repository is not configured")
	}
	record, err := repository.UpsertAgentSettingsForAdmin(ctx, slackdomain.UpdateAgentSettingsCommand{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Guidance:    input.Guidance,
		Now:         s.clock.Now().UTC(),
	})
	if err != nil {
		return CoreSlackAgentSettings{}, err
	}
	return toCoreSlackAgentSettings(record), nil
}

func toCoreSlackAgentSettings(record slackdomain.AgentSettings) CoreSlackAgentSettings {
	return CoreSlackAgentSettings{
		Guidance: record.Guidance,
	}
}
