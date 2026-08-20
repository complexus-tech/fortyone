package taskhandlers

import (
	"os"
	"strings"
	"testing"
)

func TestMayaAssignmentRecoveryCandidateQueryContract(t *testing.T) {
	t.Parallel()

	source := readNormalizedMayaHandlerSource(t)
	for name, contract := range map[string]string{
		"auto-scheduling must be explicitly enabled": "WHERE s.auto_scheduling_enabled = TRUE",
		"candidate must have an assignee":            "AND s.assignee_id IS NOT NULL",
		"Maya-assigned stories remain candidates":    "s.assignee_id = $1",
		"human stories require missing ownership":    "OR NOT EXISTS ( SELECT 1 FROM calendar_maya_schedule_ownerships ownership WHERE ownership.workspace_id = s.workspace_id AND ownership.story_id = s.id )",
		"terminal stories remain excluded":           "stat.category IN ('completed', 'cancelled')",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(source, contract) {
				t.Fatalf("Maya assignment recovery query is missing contract %q", contract)
			}
		})
	}
}

func TestMayaAssignmentRecoveryDispatchContract(t *testing.T) {
	t.Parallel()

	source := readNormalizedMayaHandlerSource(t)
	for name, contract := range map[string]string{
		"Maya-assigned stories stay in assignment batches": "if story.AssigneeID == h.systemUserID { mayaAssigned = append(mayaAssigned, story) continue }",
		"direct-human stories reconcile their schedules":   "h.mayaService.ReconcileSchedule(ctx, maya.ReconcileScheduleInput{ WorkspaceID: &workspaceID, StoryID: &storyID, })",
		"assignment batches contain only Maya stories":     "groups := groupMayaAssignmentCandidates(mayaAssigned)",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(source, contract) {
				t.Fatalf("Maya assignment recovery dispatch is missing contract %q", contract)
			}
		})
	}
}

func TestWorkspaceScheduleBatchContract(t *testing.T) {
	t.Parallel()

	source := readNormalizedMayaHandlerSource(t)
	for name, contract := range map[string]string{
		"batch includes every active enabled story": "WHERE s.workspace_id = $1 AND s.auto_scheduling_enabled = TRUE AND s.assignee_id IS NOT NULL",
		"human owners reconcile once":               "humanOwnerIDs[story.AssigneeID] = struct{}{}",
		"Maya stories stay grouped by team":         "groupMayaAssignmentCandidates(mayaAssigned)",
		"large Maya batches remain bounded":         "start += mayaWorkspaceAssignmentBatchSize",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(source, contract) {
				t.Fatalf("workspace schedule batch is missing contract %q", contract)
			}
		})
	}
}

func readNormalizedMayaHandlerSource(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile("maya.go")
	if err != nil {
		t.Fatalf("read Maya task handler: %v", err)
	}
	return strings.Join(strings.Fields(string(data)), " ")
}
