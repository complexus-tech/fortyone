package api

import "testing"

func TestTeamSettingsSchedulerIsOptionalWhenTaskRuntimeIsDisabled(t *testing.T) {
	t.Parallel()

	if scheduler := newTeamSettingsAutomationScheduler(nil); scheduler != nil {
		t.Fatalf("scheduler = %#v, want nil", scheduler)
	}
}
