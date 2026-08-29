package github

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) UpdateWorkspaceSettings(ctx context.Context, workspaceID, actorID uuid.UUID, input CoreUpdateWorkspaceSettingsInput) (CoreWorkspaceSettings, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return CoreWorkspaceSettings{}, err
	}
	if input.BranchFormat != nil && strings.TrimSpace(*input.BranchFormat) == "" {
		return CoreWorkspaceSettings{}, errors.New("branch format is required")
	}
	if input.BranchFormat != nil && !slices.Contains([]string{
		BranchFormatUsernameIdentifierTitle,
		BranchFormatIdentifierTitle,
		BranchFormatIdentifierSlashTitle,
	}, *input.BranchFormat) {
		return CoreWorkspaceSettings{}, errors.New("invalid branch format")
	}
	return s.repo.UpdateWorkspaceSettings(ctx, workspaceID, input)
}

func (s *Service) CreateIssueSyncLink(ctx context.Context, workspaceID, userID uuid.UUID, input CoreIssueSyncLinkInput) (CoreIssueSyncLink, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, userID); err != nil {
		return CoreIssueSyncLink{}, err
	}
	if !slices.Contains([]string{SyncDirectionInboundOnly, SyncDirectionBidirectional}, input.SyncDirection) {
		return CoreIssueSyncLink{}, errors.New("invalid sync direction")
	}
	link, err := s.repo.CreateIssueSyncLink(ctx, workspaceID, userID, input)
	if err != nil {
		return CoreIssueSyncLink{}, err
	}
	settings, err := s.repo.GetTeamWorkflowSettings(ctx, workspaceID, input.TeamID)
	if err != nil {
		return CoreIssueSyncLink{}, err
	}
	if len(settings.Rules) == 0 {
		if _, err := s.seedDefaultWorkflowRules(ctx, workspaceID, input.TeamID); err != nil {
			return CoreIssueSyncLink{}, err
		}
	}
	return link, nil
}

func (s *Service) UpdateIssueSyncLink(ctx context.Context, workspaceID, actorID, linkID uuid.UUID, input CoreUpdateIssueSyncLinkInput) (CoreIssueSyncLink, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return CoreIssueSyncLink{}, err
	}
	if input.SyncDirection != nil && !slices.Contains([]string{SyncDirectionInboundOnly, SyncDirectionBidirectional}, *input.SyncDirection) {
		return CoreIssueSyncLink{}, errors.New("invalid sync direction")
	}
	return s.repo.UpdateIssueSyncLink(ctx, workspaceID, linkID, input)
}

func (s *Service) DeleteIssueSyncLink(ctx context.Context, workspaceID, actorID, linkID uuid.UUID) error {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return err
	}
	return s.repo.DeleteIssueSyncLink(ctx, workspaceID, linkID)
}

func (s *Service) GetTeamSettings(ctx context.Context, workspaceID, teamID uuid.UUID) (CoreTeamGitHubSettings, error) {
	settings, err := s.repo.GetTeamWorkflowSettings(ctx, workspaceID, teamID)
	if err != nil {
		return CoreTeamGitHubSettings{}, err
	}
	if len(settings.Rules) == 0 {
		return s.seedDefaultWorkflowRules(ctx, workspaceID, teamID)
	}
	return settings, nil
}

func (s *Service) UpdateTeamSettings(ctx context.Context, workspaceID, actorID, teamID uuid.UUID, input CoreUpdateTeamGitHubSettings) (CoreTeamGitHubSettings, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return CoreTeamGitHubSettings{}, err
	}
	if len(input.Rules) == 0 {
		return CoreTeamGitHubSettings{}, errors.New("at least one rule is required")
	}
	return s.repo.ReplaceTeamWorkflowSettings(ctx, workspaceID, teamID, input.Rules)
}

func (s *Service) GetStoryGitHubLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]StoryGitHubLink, error) {
	return s.repo.GetStoryGitHubLinks(ctx, workspaceID, storyID)
}

func (s *Service) DeleteStoryGitHubLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	return s.repo.DeleteStoryGitHubLink(ctx, workspaceID, linkID)
}
