package maya

import "testing"

func TestInferWorkFocusSkipsManualContext(t *testing.T) {
	result := InferWorkFocus(WorkFocusInferenceInput{
		ManualRoleTitle:       "Platform engineer",
		ManualRoleDescription: "Owns infra and backend reliability.",
		Evidence: []WorkFocusEvidence{
			{Title: "Fix React filters"},
			{Title: "Build issue detail layout"},
			{Title: "Improve dashboard cards"},
			{Title: "Polish story list spacing"},
			{Title: "Update frontend menu"},
			{Title: "Adjust sidebar navigation"},
		},
	})

	if result.ShouldInfer {
		t.Fatalf("expected manual context to prevent inference")
	}
}

func TestInferWorkFocusRequiresEnoughEvidence(t *testing.T) {
	result := InferWorkFocus(WorkFocusInferenceInput{
		Evidence: []WorkFocusEvidence{
			{Title: "Build React filters"},
			{Title: "Fix frontend card layout"},
		},
	})

	if result.ShouldInfer {
		t.Fatalf("expected inference to be skipped with too little evidence")
	}
}

func TestInferWorkFocusDetectsFrontendFocus(t *testing.T) {
	result := InferWorkFocus(WorkFocusInferenceInput{
		Evidence: []WorkFocusEvidence{
			{Title: "Build React filters toolbar", Labels: []string{"Frontend"}},
			{Title: "Polish story detail layout", Description: "Fix spacing, hover states, and responsive UI."},
			{Title: "Update dashboard cards", Description: "Match the summary page card styles."},
			{Title: "Fix sidebar navigation", Labels: []string{"UI"}},
			{Title: "Improve label picker", Description: "Frontend menu behavior and component styling."},
			{Title: "Create calendar grid view", Description: "Responsive client-side calendar layout."},
		},
	})

	if !result.ShouldInfer {
		t.Fatalf("expected frontend work focus inference, got %#v", result)
	}
	if result.RoleTitle != "Frontend engineer" {
		t.Fatalf("expected frontend role title, got %q", result.RoleTitle)
	}
	if result.StoryCount != 6 {
		t.Fatalf("expected story count to be retained, got %d", result.StoryCount)
	}
	if result.Confidence <= 0 {
		t.Fatalf("expected positive confidence, got %f", result.Confidence)
	}
}

func TestInferWorkFocusDetectsBackendFocus(t *testing.T) {
	result := InferWorkFocus(WorkFocusInferenceInput{
		Evidence: []WorkFocusEvidence{
			{Title: "Add API endpoint for workload reports", Labels: []string{"Backend"}},
			{Title: "Fix database migration", Description: "Add SQL indexes and repository query."},
			{Title: "Process webhook events", Description: "Worker should handle integration payloads."},
			{Title: "Improve auth callback", Description: "Server returns better OAuth errors."},
			{Title: "Batch assignment job", Description: "Schedule backend automation task."},
			{Title: "Add repository tests", Description: "Verify SQL query behavior."},
		},
	})

	if !result.ShouldInfer {
		t.Fatalf("expected backend work focus inference, got %#v", result)
	}
	if result.RoleTitle != "Backend engineer" {
		t.Fatalf("expected backend role title, got %q", result.RoleTitle)
	}
}
