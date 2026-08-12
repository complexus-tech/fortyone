package slack

import (
	"context"
	"errors"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
)

type agentSettingsRepository interface {
	GetAgentSettings(ctx context.Context, workspaceID uuid.UUID) (slackrepository.AgentSettingsRecord, error)
	UpsertAgentSettings(ctx context.Context, workspaceID uuid.UUID, input slackrepository.AgentSettingsInput) (slackrepository.AgentSettingsRecord, error)
}

type agentSettingsReader interface {
	GetAgentSettings(ctx context.Context, workspaceID uuid.UUID) (slackrepository.AgentSettingsRecord, error)
}

type CoreSlackAgentSettings struct {
	Guidance string
}

func (s *Service) GetAgentSettings(ctx context.Context, workspaceID uuid.UUID) (CoreSlackAgentSettings, error) {
	repository, ok := s.repo.(agentSettingsReader)
	if !ok {
		return CoreSlackAgentSettings{}, errors.New("Slack agent settings repository is not configured")
	}
	record, err := repository.GetAgentSettings(ctx, workspaceID)
	if err != nil {
		return CoreSlackAgentSettings{}, err
	}
	return toCoreSlackAgentSettings(record), nil
}

func (s *Service) UpdateAgentSettings(
	ctx context.Context,
	workspaceID uuid.UUID,
	input CoreSlackAgentSettings,
) (CoreSlackAgentSettings, error) {
	repository, ok := s.repo.(agentSettingsRepository)
	if !ok {
		return CoreSlackAgentSettings{}, errors.New("Slack agent settings repository is not configured")
	}
	record, err := repository.UpsertAgentSettings(ctx, workspaceID, slackrepository.AgentSettingsInput{
		Guidance: input.Guidance,
	})
	if err != nil {
		return CoreSlackAgentSettings{}, err
	}
	return toCoreSlackAgentSettings(record), nil
}

func toCoreSlackAgentSettings(record slackrepository.AgentSettingsRecord) CoreSlackAgentSettings {
	return CoreSlackAgentSettings{
		Guidance: record.Guidance,
	}
}
