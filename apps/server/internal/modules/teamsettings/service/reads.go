package teamsettings

import "context"

func (s *Service) GetSettings(ctx context.Context, access Access) (CoreTeamSettings, error) {
	if err := s.authorizeRead(ctx, access); err != nil {
		return CoreTeamSettings{}, err
	}
	sprintSettings, err := s.repo.GetSprintSettings(ctx, access.TeamID, access.WorkspaceID)
	if err != nil {
		return CoreTeamSettings{}, err
	}
	storySettings, err := s.repo.GetStoryAutomationSettings(ctx, access.TeamID, access.WorkspaceID)
	if err != nil {
		return CoreTeamSettings{}, err
	}
	estimationSettings, err := s.repo.GetEstimationSettings(ctx, access.TeamID, access.WorkspaceID)
	if err != nil {
		return CoreTeamSettings{}, err
	}
	return CoreTeamSettings{
		SprintSettings:          sprintSettings,
		StoryAutomationSettings: storySettings,
		EstimationSettings:      estimationSettings,
	}, nil
}

func (s *Service) GetSprintSettings(ctx context.Context, access Access) (CoreTeamSprintSettings, error) {
	if err := s.authorizeRead(ctx, access); err != nil {
		return CoreTeamSprintSettings{}, err
	}
	return s.repo.GetSprintSettings(ctx, access.TeamID, access.WorkspaceID)
}

func (s *Service) GetStoryAutomationSettings(ctx context.Context, access Access) (CoreTeamStoryAutomationSettings, error) {
	if err := s.authorizeRead(ctx, access); err != nil {
		return CoreTeamStoryAutomationSettings{}, err
	}
	return s.repo.GetStoryAutomationSettings(ctx, access.TeamID, access.WorkspaceID)
}

func (s *Service) GetEstimationSettings(ctx context.Context, access Access) (CoreTeamEstimationSettings, error) {
	if err := s.authorizeRead(ctx, access); err != nil {
		return CoreTeamEstimationSettings{}, err
	}
	return s.repo.GetEstimationSettings(ctx, access.TeamID, access.WorkspaceID)
}
