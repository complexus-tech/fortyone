package stories

import (
	"testing"

	"github.com/google/uuid"
)

func TestShouldTriggerMayaAssignmentWhenCreatedForMaya(t *testing.T) {
	mayaUserID := uuid.New()

	if !shouldTriggerMayaAssignment(nil, &mayaUserID, mayaUserID) {
		t.Fatal("expected Maya assignment to trigger for a newly assigned story")
	}
}

func TestShouldTriggerMayaAssignmentWhenReassignedToMaya(t *testing.T) {
	previousAssigneeID := uuid.New()
	mayaUserID := uuid.New()

	if !shouldTriggerMayaAssignment(&previousAssigneeID, &mayaUserID, mayaUserID) {
		t.Fatal("expected reassignment to Maya to trigger assignment planning")
	}
}

func TestShouldTriggerMayaAssignmentWhenMayaIsExplicitlySelectedAgain(t *testing.T) {
	mayaUserID := uuid.New()

	if !shouldTriggerMayaAssignment(&mayaUserID, &mayaUserID, mayaUserID) {
		t.Fatal("expected explicit Maya selection to trigger assignment planning")
	}
}

func TestMayaAssignmentUpdateAssigneeParsesUUIDValues(t *testing.T) {
	mayaUserID := uuid.New()
	updates := map[string]any{"assignee_id": mayaUserID}

	assigneeID, ok := mayaAssignmentUpdateAssignee(updates)
	if !ok {
		t.Fatal("expected assignee update to be detected")
	}
	if assigneeID == nil || *assigneeID != mayaUserID {
		t.Fatalf("expected assignee %s, got %v", mayaUserID, assigneeID)
	}
}

func TestMayaAssignmentUpdateAssigneeParsesNilValues(t *testing.T) {
	updates := map[string]any{"assignee_id": nil}

	assigneeID, ok := mayaAssignmentUpdateAssignee(updates)
	if !ok {
		t.Fatal("expected assignee update to be detected")
	}
	if assigneeID != nil {
		t.Fatalf("expected nil assignee, got %v", assigneeID)
	}
}
