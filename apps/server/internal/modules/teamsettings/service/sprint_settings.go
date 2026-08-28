package teamsettings

import "context"

func (s *Service) UpdateSprintSettings(
	ctx context.Context,
	access Access,
	updates CoreUpdateTeamSprintSettings,
) (CoreTeamSprintSettings, error) {
	if err := s.authorizeWrite(access); err != nil {
		return CoreTeamSprintSettings{}, err
	}
	if err := validateSprintSettingsUpdate(updates); err != nil {
		return CoreTeamSprintSettings{}, err
	}
	result, err := s.repo.UpdateSprintSettings(
		ctx,
		access.TeamID,
		access.WorkspaceID,
		updates,
		UserAuditActor(access.Actor),
	)
	if err != nil {
		return CoreTeamSprintSettings{}, err
	}

	if result.AutoCreateSprints && s.scheduler != nil {
		if err := s.scheduler.ScheduleSprintCreation(); err != nil && s.log != nil {
			s.log.Error(ctx, "failed to schedule sprint automation after settings commit", "error", err)
		}
	}
	return result, nil
}

// ReconcileSprintSchedule is invoked by the internal automation worker. The
// repository reloads and locks canonical settings before changing any sprint.
func (s *Service) ReconcileSprintSchedule(ctx context.Context, settings CoreTeamSprintSettings) (int, error) {
	return s.repo.ReconcileSprintSchedule(ctx, settings, SystemAuditActor())
}
