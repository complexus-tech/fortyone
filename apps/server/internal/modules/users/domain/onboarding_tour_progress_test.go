package usersdomain

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestUpdateOnboardingTourProgressNormalizesStableIdentifiers(t *testing.T) {
	t.Parallel()

	status := OnboardingTourStatusCompleted
	update := UpdateOnboardingTourProgress{
		OnboardingTourScope: OnboardingTourScope{
			TourKey:     " workspace-getting-started ",
			TourVersion: " 2026-08-31 ",
		},
		CompletedStepIDs:   []string{" task-created ", "maya-requested", "task-created"},
		CompletedActionIDs: []string{" task-created ", "task-created"},
		Status:             &status,
	}

	if err := update.NormalizeAndValidate(); err != nil {
		t.Fatalf("normalize onboarding progress: %v", err)
	}
	if update.TourKey != "workspace-getting-started" || update.TourVersion != "2026-08-31" {
		t.Fatalf("normalized scope = %#v", update.OnboardingTourScope)
	}
	if !reflect.DeepEqual(update.CompletedStepIDs, []string{"maya-requested", "task-created"}) {
		t.Fatalf("completed step IDs = %#v", update.CompletedStepIDs)
	}
	if !reflect.DeepEqual(update.CompletedActionIDs, []string{"task-created"}) {
		t.Fatalf("completed action IDs = %#v", update.CompletedActionIDs)
	}
	if update.Status == nil || *update.Status != OnboardingTourStatusCompleted {
		t.Fatalf("status = %#v", update.Status)
	}
}

func TestUpdateOnboardingTourProgressRejectsUnsafeOrUnboundedInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		update UpdateOnboardingTourProgress
	}{
		{
			name:   "missing scope",
			update: UpdateOnboardingTourProgress{},
		},
		{
			name: "invalid status",
			update: UpdateOnboardingTourProgress{
				OnboardingTourScope: OnboardingTourScope{TourKey: "workspace-getting-started", TourVersion: "v1"},
				Status: func() *OnboardingTourStatus {
					status := OnboardingTourStatus("waiting")
					return &status
				}(),
			},
		},
		{
			name: "oversized step set",
			update: UpdateOnboardingTourProgress{
				OnboardingTourScope: OnboardingTourScope{TourKey: "workspace-getting-started", TourVersion: "v1"},
				CompletedStepIDs: func() []string {
					values := make([]string, MaximumOnboardingTourProgressIDs+1)
					for index := range values {
						values[index] = "step"
					}
					return values
				}(),
			},
		},
		{
			name: "invalid identifier",
			update: UpdateOnboardingTourProgress{
				OnboardingTourScope: OnboardingTourScope{TourKey: "workspace-getting-started", TourVersion: "v1"},
				CompletedActionIDs: []string{
					strings.Repeat("a", MaximumOnboardingTourProgressIDRunes+1),
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.update.NormalizeAndValidate(); !errors.Is(err, ErrInvalidOnboardingTour) {
				t.Fatalf("NormalizeAndValidate() error = %v, want ErrInvalidOnboardingTour", err)
			}
		})
	}
}
