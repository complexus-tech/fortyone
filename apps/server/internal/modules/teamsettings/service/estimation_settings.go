package teamsettings

import "context"

func (s *Service) UpdateEstimationSettings(
	ctx context.Context,
	access Access,
	updates CoreUpdateTeamEstimationSettings,
) (CoreTeamEstimationSettings, error) {
	if err := s.authorizeWrite(access); err != nil {
		return CoreTeamEstimationSettings{}, err
	}
	if err := validateEstimationSettingsUpdate(updates); err != nil {
		return CoreTeamEstimationSettings{}, err
	}
	return s.repo.UpdateEstimationSettings(
		ctx,
		access.TeamID,
		access.WorkspaceID,
		updates,
		UserAuditActor(access.Actor),
	)
}
