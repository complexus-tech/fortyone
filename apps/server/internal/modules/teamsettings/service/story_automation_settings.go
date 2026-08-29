package teamsettings

import "context"

func (s *Service) UpdateStoryAutomationSettings(
	ctx context.Context,
	access Access,
	updates CoreUpdateTeamStoryAutomationSettings,
) (CoreTeamStoryAutomationSettings, error) {
	if err := s.authorizeWrite(access); err != nil {
		return CoreTeamStoryAutomationSettings{}, err
	}
	if err := validateStoryAutomationSettingsUpdate(updates); err != nil {
		return CoreTeamStoryAutomationSettings{}, err
	}
	result, err := s.repo.UpdateStoryAutomationSettings(
		ctx,
		access.TeamID,
		access.WorkspaceID,
		updates,
		UserAuditActor(access.Actor),
	)
	if err != nil {
		return CoreTeamStoryAutomationSettings{}, err
	}

	if s.scheduler == nil {
		return result, nil
	}
	if result.AutoCloseInactiveEnabled {
		if err := s.scheduler.ScheduleStoryAutoClose(); err != nil && s.log != nil {
			s.log.Error(ctx, "failed to schedule story auto-close after settings commit", "error", err)
		}
	}
	if result.AutoArchiveEnabled {
		if err := s.scheduler.ScheduleStoryAutoArchive(); err != nil && s.log != nil {
			s.log.Error(ctx, "failed to schedule story auto-archive after settings commit", "error", err)
		}
	}
	return result, nil
}
