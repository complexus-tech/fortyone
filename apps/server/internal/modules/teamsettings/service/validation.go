package teamsettings

func validateSprintSettingsUpdate(updates CoreUpdateTeamSprintSettings) error {
	if updates.Empty() {
		return ErrNoSettingsChanges
	}
	validDays := map[string]struct{}{
		"Monday": {}, "Tuesday": {}, "Wednesday": {}, "Thursday": {},
		"Friday": {}, "Saturday": {}, "Sunday": {},
	}
	if updates.SprintStartDay.Present {
		if _, valid := validDays[updates.SprintStartDay.Value]; !valid {
			return ErrInvalidSprintStartDay
		}
	}
	if updates.SprintDurationWeeks.Present &&
		(updates.SprintDurationWeeks.Value < 1 || updates.SprintDurationWeeks.Value > 8) {
		return ErrInvalidSprintDuration
	}
	if updates.UpcomingSprintsCount.Present &&
		(updates.UpcomingSprintsCount.Value < 0 || updates.UpcomingSprintsCount.Value > 10) {
		return ErrInvalidUpcomingCount
	}
	if updates.WorkingDays.Present {
		workingDays := updates.WorkingDays.Value
		if len(workingDays) == 0 || len(workingDays) > 7 {
			return ErrInvalidWorkingDays
		}
		seen := make(map[int]struct{}, len(workingDays))
		for _, day := range workingDays {
			if day < 1 || day > 7 {
				return ErrInvalidWorkingDays
			}
			if _, duplicate := seen[day]; duplicate {
				return ErrInvalidWorkingDays
			}
			seen[day] = struct{}{}
		}
	}
	if updates.NextAutoSprintNumber.Present &&
		(updates.NextAutoSprintNumber.Value < 1 || updates.NextAutoSprintNumber.Value > 10000) {
		return ErrInvalidNextAutoNumber
	}
	return nil
}

func validateStoryAutomationSettingsUpdate(updates CoreUpdateTeamStoryAutomationSettings) error {
	if updates.Empty() {
		return ErrNoSettingsChanges
	}
	if updates.AutoCloseInactiveMonths.Present &&
		(updates.AutoCloseInactiveMonths.Value < 1 || updates.AutoCloseInactiveMonths.Value > 24) {
		return ErrInvalidCloseMonths
	}
	if updates.AutoArchiveMonths.Present &&
		(updates.AutoArchiveMonths.Value < 1 || updates.AutoArchiveMonths.Value > 24) {
		return ErrInvalidArchiveMonths
	}
	return nil
}

func validateEstimationSettingsUpdate(updates CoreUpdateTeamEstimationSettings) error {
	if updates.Empty() {
		return ErrNoSettingsChanges
	}
	switch updates.Scheme.Value {
	case "points", "tshirt":
		return nil
	default:
		return ErrInvalidEstimateScheme
	}
}
