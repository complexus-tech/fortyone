package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
)

type channelAudienceRepository interface {
	ListChannelTeamAccess(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.ChannelTeamAccessRecord, error)
	ReplaceChannelTeamAccess(
		ctx context.Context,
		workspaceID, slackWorkspaceID uuid.UUID,
		slackChannelID string,
		teamIDs []uuid.UUID,
		actorID uuid.UUID,
	) error
	ListAuthorizedChannelTeamIDs(
		ctx context.Context,
		workspaceID, slackWorkspaceID uuid.UUID,
		slackChannelID string,
		userID uuid.UUID,
	) ([]uuid.UUID, error)
}

type authorizedChannelAudienceRepository interface {
	ListAuthorizedChannelTeamIDs(
		ctx context.Context,
		workspaceID, slackWorkspaceID uuid.UUID,
		slackChannelID string,
		userID uuid.UUID,
	) ([]uuid.UUID, error)
}

type authorizedAssistantChannelAudienceRepository interface {
	GetAuthorizedAssistantChannelTeamScope(
		ctx context.Context,
		workspaceID, slackWorkspaceID uuid.UUID,
		slackChannelID string,
		userID uuid.UUID,
	) (slackrepository.AssistantChannelTeamScope, error)
}

type CoreSlackChannelAudience struct {
	Channel CoreSlackChannel
	TeamIDs []uuid.UUID
}

func (s *Service) ListChannelAudiences(ctx context.Context, workspaceID uuid.UUID) ([]CoreSlackChannelAudience, error) {
	repository, ok := s.repo.(channelAudienceRepository)
	if !ok {
		return nil, errors.New("Slack channel audience repository is not configured")
	}
	channels, err := s.repo.ListChannels(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	access, err := repository.ListChannelTeamAccess(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	teamIDsByChannel := make(map[string][]uuid.UUID, len(channels))
	for _, record := range access {
		channelID := strings.TrimSpace(record.SlackChannelID)
		if channelID == "" || record.TeamID == uuid.Nil {
			continue
		}
		teamIDsByChannel[channelID] = append(teamIDsByChannel[channelID], record.TeamID)
	}

	result := make([]CoreSlackChannelAudience, 0, len(channels))
	for _, channel := range channels {
		result = append(result, CoreSlackChannelAudience{
			Channel: toCoreChannel(channel),
			TeamIDs: append([]uuid.UUID(nil), teamIDsByChannel[channel.SlackChannelID]...),
		})
	}
	return result, nil
}

func (s *Service) UpdateChannelAudience(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
	slackChannelID string,
	teamIDs []uuid.UUID,
) error {
	repository, ok := s.repo.(channelAudienceRepository)
	if !ok {
		return errors.New("Slack channel audience repository is not configured")
	}
	if workspaceID == uuid.Nil || actorID == uuid.Nil {
		return errors.New("workspace and actor are required")
	}
	slackChannelID = strings.TrimSpace(slackChannelID)
	if slackChannelID == "" {
		return errors.New("Slack channel is required")
	}
	installation, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	if err := repository.ReplaceChannelTeamAccess(
		ctx,
		workspaceID,
		installation.ID,
		slackChannelID,
		teamIDs,
		actorID,
	); err != nil {
		return fmt.Errorf("update Slack channel audience: %w", err)
	}
	return nil
}

func (s *Service) authorizedChannelTeamIDs(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	userID uuid.UUID,
) ([]uuid.UUID, error) {
	repository, ok := s.repo.(authorizedChannelAudienceRepository)
	if !ok {
		return nil, errors.New("Slack channel audience repository is not configured")
	}
	return repository.ListAuthorizedChannelTeamIDs(
		ctx,
		workspaceID,
		slackWorkspaceID,
		strings.TrimSpace(slackChannelID),
		userID,
	)
}

func (s *Service) authorizedAssistantChannelTeamScope(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	userID uuid.UUID,
) (slackrepository.AssistantChannelTeamScope, error) {
	repository, ok := s.repo.(authorizedAssistantChannelAudienceRepository)
	if !ok {
		return slackrepository.AssistantChannelTeamScope{}, errors.New("Slack assistant channel audience repository is not configured")
	}
	return repository.GetAuthorizedAssistantChannelTeamScope(
		ctx,
		workspaceID,
		slackWorkspaceID,
		strings.TrimSpace(slackChannelID),
		userID,
	)
}

func (s *Service) availableTeamsForSlackSource(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	source requestSourceContext,
) ([]slackrepository.TeamRecord, error) {
	teams, err := s.repo.ListWorkspaceTeamsForUser(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	channelID := strings.TrimSpace(source.SlackChannelID)
	if channelID == "" || strings.TrimSpace(source.SlackTeamID) == "" || strings.HasPrefix(strings.ToUpper(channelID), "D") {
		return teams, nil
	}
	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, strings.TrimSpace(source.SlackTeamID))
	if err != nil {
		return nil, err
	}
	if installation.WorkspaceID != workspaceID {
		return nil, ErrSlackNoWorkspaceLinked
	}
	allowedTeamIDs, err := s.authorizedChannelTeamIDs(ctx, workspaceID, installation.ID, channelID, userID)
	if err != nil {
		return nil, err
	}
	allowed := make(map[uuid.UUID]struct{}, len(allowedTeamIDs))
	for _, teamID := range allowedTeamIDs {
		if teamID != uuid.Nil {
			allowed[teamID] = struct{}{}
		}
	}
	filtered := make([]slackrepository.TeamRecord, 0, len(teams))
	for _, team := range teams {
		if _, ok := allowed[team.ID]; ok {
			filtered = append(filtered, team)
		}
	}
	return filtered, nil
}

// availableTeamsForSlackCreation returns the actor's teams in the same
// personal order used by FortyOne. Story creation is an explicit, confirmed
// action, so the selected team is constrained by current membership rather
// than by the assistant's channel disclosure boundary.
func (s *Service) availableTeamsForSlackCreation(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
) ([]slackrepository.TeamRecord, error) {
	return s.repo.ListWorkspaceTeamsForUser(ctx, workspaceID, userID)
}

func (s *Service) ensureTeamAvailableForSlackCreation(
	ctx context.Context,
	workspaceID, userID, teamID uuid.UUID,
) error {
	teams, err := s.availableTeamsForSlackCreation(ctx, workspaceID, userID)
	if err != nil {
		return err
	}
	for _, team := range teams {
		if team.ID == teamID {
			return nil
		}
	}
	return ErrSlackTeamNotAvailable
}

func (s *Service) ensureTeamAvailableForSlackSource(
	ctx context.Context,
	workspaceID, userID, teamID uuid.UUID,
	source requestSourceContext,
) error {
	teams, err := s.availableTeamsForSlackSource(ctx, workspaceID, userID, source)
	if err != nil {
		return err
	}
	for _, team := range teams {
		if team.ID == teamID {
			return nil
		}
	}
	return ErrSlackTeamNotAvailable
}

func (s *Service) slackDeliveryAuthorizationCurrent(
	ctx context.Context,
	workspaceID uuid.UUID,
	externalWorkspaceID, channelID, slackUserID string,
	payload SlackProviderPayload,
) (bool, error) {
	authorization := payload.Authorization
	if authorization == nil {
		return true, nil
	}
	if authorization.ActorUserID == nil || *authorization.ActorUserID == uuid.Nil {
		return false, nil
	}
	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, strings.TrimSpace(externalWorkspaceID))
	if err != nil {
		return false, err
	}
	if installation.WorkspaceID != workspaceID {
		return false, nil
	}
	actorTeams, err := s.repo.ListWorkspaceTeamsForUser(ctx, workspaceID, *authorization.ActorUserID)
	if err != nil {
		return false, err
	}
	if !uuidSubset(authorization.AllowedTeamIDs, slackTeamRecordIDs(actorTeams)) {
		return false, nil
	}
	channelID = strings.TrimSpace(channelID)
	if strings.HasPrefix(strings.ToUpper(channelID), "D") {
		if strings.TrimSpace(slackUserID) == "" {
			return false, nil
		}
		linkedUserID, err := s.repo.FindLinkedUserIDBySlackUser(ctx, workspaceID, installation.SlackTeamID, strings.TrimSpace(slackUserID))
		if err != nil {
			return false, err
		}
		if linkedUserID == nil || *linkedUserID == uuid.Nil {
			return false, nil
		}
		recipientTeams, err := s.repo.ListWorkspaceTeamsForUser(ctx, workspaceID, *linkedUserID)
		if err != nil {
			return false, err
		}
		return uuidSubset(authorization.AllowedTeamIDs, slackTeamRecordIDs(recipientTeams)), nil
	}
	if authorization.Scope == slackDeliveryAuthorizationScopeActorMembership {
		return true, nil
	}
	teamIDs, err := s.authorizedChannelTeamIDs(ctx, workspaceID, installation.ID, channelID, *authorization.ActorUserID)
	if err != nil {
		return false, err
	}
	return uuidSubset(authorization.AllowedTeamIDs, teamIDs), nil
}

func slackTeamRecordIDs(teams []slackrepository.TeamRecord) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(teams))
	for _, team := range teams {
		if team.ID != uuid.Nil {
			ids = append(ids, team.ID)
		}
	}
	return ids
}

func toCoreChannel(record slackrepository.SlackChannelRecord) CoreSlackChannel {
	return CoreSlackChannel{
		ID:             record.ID,
		SlackChannelID: record.SlackChannelID,
		Name:           record.Name,
		IsPrivate:      record.IsPrivate,
		IsArchived:     record.IsArchived,
		IsMember:       record.IsMember,
		IsActive:       record.IsActive,
		LastSyncedAt:   record.LastSyncedAt,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}
