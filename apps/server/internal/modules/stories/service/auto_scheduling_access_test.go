package stories

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

func newAutoSchedulingContractService(repo Repository) *Service {
	return New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil, nil)
}

func TestCreateRejectsIncompleteMayaSchedulingIntentBeforePersistence(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	mayaID := uuid.New()
	durationMinutes := 60
	deliveryDate := time.Now().UTC().Add(24 * time.Hour)

	tests := []struct {
		name    string
		story   CoreNewStory
		wantErr error
	}{
		{
			name: "omitted scheduling flag",
			story: CoreNewStory{
				Assignee: &mayaID,
			},
			wantErr: ErrMayaAssignmentRequiresScheduling,
		},
		{
			name: "explicit false scheduling flag",
			story: CoreNewStory{
				Assignee:              &mayaID,
				AutoSchedulingEnabled: false,
			},
			wantErr: ErrMayaAssignmentRequiresScheduling,
		},
		{
			name: "missing duration",
			story: CoreNewStory{
				Assignee:              &mayaID,
				AutoSchedulingEnabled: true,
				EndDate:               &deliveryDate,
			},
			wantErr: ErrMayaAssignmentRequiresDuration,
		},
		{
			name: "missing delivery date or sprint",
			story: CoreNewStory{
				Assignee:                 &mayaID,
				AutoSchedulingEnabled:    true,
				EstimatedDurationMinutes: &durationMinutes,
			},
			wantErr: ErrMayaAssignmentRequiresDeliveryDate,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &atomicCreateRepo{}
			service := newAutoSchedulingContractService(repo)
			service.ConfigureMayaActor(mayaID)
			service.ConfigureAutoSchedulingEligibility(func(context.Context, uuid.UUID) (bool, error) {
				return true, nil
			})

			story := test.story
			story.Title = "Maya-planned work"
			story.Team = teamID
			story.Reporter = &actorID
			_, err := service.createWithOptions(context.Background(), story, workspaceID, actorID, createOptions{})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("create error = %v, want %v", err, test.wantErr)
			}
			if repo.createCalls != 0 || repo.createdStory != nil {
				t.Fatalf("incomplete Maya intent reached persistence: calls=%d story=%v", repo.createCalls, repo.createdStory)
			}
		})
	}
}

func TestCreateAllowsCompleteEligibleMayaSchedulingIntent(t *testing.T) {
	repo := &atomicCreateRepo{}
	service := newAutoSchedulingContractService(repo)
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	mayaID := uuid.New()
	durationMinutes := 60
	deliveryDate := time.Now().UTC().Add(24 * time.Hour)
	service.ConfigureMayaActor(mayaID)
	service.ConfigureAutoSchedulingEligibility(func(context.Context, uuid.UUID) (bool, error) {
		return true, nil
	})

	created, err := service.createWithOptions(context.Background(), CoreNewStory{
		Title:                    "Maya-planned work",
		Team:                     teamID,
		Reporter:                 &actorID,
		Assignee:                 &mayaID,
		AutoSchedulingEnabled:    true,
		EstimatedDurationMinutes: &durationMinutes,
		EndDate:                  &deliveryDate,
	}, workspaceID, actorID, createOptions{})
	if err != nil {
		t.Fatalf("create complete Maya story: %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("create calls = %d, want 1", repo.createCalls)
	}
	if !created.AutoSchedulingEnabled || created.AutoSchedulingStatus != AutoSchedulingStatusNeedsOwner {
		t.Fatalf("created Maya scheduling state = %#v", created)
	}
}

func TestCreateRejectsIneligibleHumanAutoScheduling(t *testing.T) {
	repo := &atomicCreateRepo{}
	service := newAutoSchedulingContractService(repo)
	service.ConfigureAutoSchedulingEligibility(func(context.Context, uuid.UUID) (bool, error) {
		return false, nil
	})
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	assigneeID := uuid.New()

	_, err := service.createWithOptions(context.Background(), CoreNewStory{
		Title:                 "Human-planned work",
		Team:                  teamID,
		Reporter:              &actorID,
		Assignee:              &assigneeID,
		AutoSchedulingEnabled: true,
	}, workspaceID, actorID, createOptions{})
	if !errors.Is(err, ErrAutoSchedulingUnavailable) {
		t.Fatalf("create error = %v, want %v", err, ErrAutoSchedulingUnavailable)
	}
	if repo.createCalls != 0 {
		t.Fatalf("ineligible scheduling reached persistence: calls=%d", repo.createCalls)
	}
}

func TestUpdateRechecksEligibilityWhenApprovalEnablesScheduling(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	assigneeID := uuid.New()
	repo := &activityRecordingRepo{story: CoreSingleStory{
		ID:                       storyID,
		Workspace:                workspaceID,
		Assignee:                 &assigneeID,
		AutoSchedulingEnabled:    false,
		AutoSchedulingStatus:     AutoSchedulingStatusOff,
		EstimatedDurationMinutes: intPointer(60),
		EndDate:                  timePointer(time.Now().UTC().Add(24 * time.Hour)),
	}}
	service := newAutoSchedulingContractService(repo)
	eligibilityChecks := 0
	eligible := true
	service.ConfigureAutoSchedulingEligibility(func(context.Context, uuid.UUID) (bool, error) {
		eligibilityChecks++
		return eligible, nil
	})

	// The proposal was prepared while eligible; access lapses before the
	// approved mutation reaches the persistence boundary.
	eligible = false
	err := service.Update(context.Background(), storyID, workspaceID, map[string]any{
		"auto_scheduling_enabled": true,
	})
	if !errors.Is(err, ErrAutoSchedulingUnavailable) {
		t.Fatalf("update error = %v, want %v", err, ErrAutoSchedulingUnavailable)
	}
	if eligibilityChecks != 1 {
		t.Fatalf("eligibility checks = %d, want 1", eligibilityChecks)
	}
	if repo.story.AutoSchedulingEnabled {
		t.Fatal("ineligible approved update changed persisted scheduling intent")
	}
}

func TestUpdateRechecksEligibilityForExplicitEnableOnAlreadyEnabledStory(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	assigneeID := uuid.New()
	repo := &activityRecordingRepo{story: CoreSingleStory{
		ID:                    storyID,
		Workspace:             workspaceID,
		Assignee:              &assigneeID,
		AutoSchedulingEnabled: true,
		AutoSchedulingStatus:  AutoSchedulingStatusPlanning,
	}}
	service := newAutoSchedulingContractService(repo)
	eligibilityChecks := 0
	service.ConfigureAutoSchedulingEligibility(func(context.Context, uuid.UUID) (bool, error) {
		eligibilityChecks++
		return false, nil
	})

	err := service.Update(context.Background(), storyID, workspaceID, map[string]any{
		"auto_scheduling_enabled": true,
	})
	if !errors.Is(err, ErrAutoSchedulingUnavailable) {
		t.Fatalf("update error = %v, want %v", err, ErrAutoSchedulingUnavailable)
	}
	if eligibilityChecks != 1 {
		t.Fatalf("eligibility checks = %d, want 1", eligibilityChecks)
	}
}

func TestUpdateDoesNotGateUnrelatedEditAfterEligibilityLapse(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	assigneeID := uuid.New()
	repo := &activityRecordingRepo{story: CoreSingleStory{
		ID:                    storyID,
		Workspace:             workspaceID,
		Assignee:              &assigneeID,
		AutoSchedulingEnabled: true,
		AutoSchedulingStatus:  AutoSchedulingStatusPlanning,
	}}
	service := newAutoSchedulingContractService(repo)
	eligibilityChecks := 0
	service.ConfigureAutoSchedulingEligibility(func(context.Context, uuid.UUID) (bool, error) {
		eligibilityChecks++
		return false, nil
	})

	if err := service.Update(context.Background(), storyID, workspaceID, map[string]any{"title": "Updated title"}); err != nil {
		t.Fatalf("unrelated update after plan lapse: %v", err)
	}
	if eligibilityChecks != 0 {
		t.Fatalf("unrelated update performed %d eligibility checks, want 0", eligibilityChecks)
	}
}

func TestUpdateRejectsIncompleteNewMayaAssignment(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	previousAssigneeID := uuid.New()
	mayaID := uuid.New()
	deliveryDate := time.Now().UTC().Add(24 * time.Hour)

	tests := []struct {
		name                   string
		updates                map[string]any
		existingAutoScheduling bool
		wantErr                error
	}{
		{
			name:                   "omitted scheduling flag on already scheduled story",
			updates:                map[string]any{"assignee_id": mayaID},
			existingAutoScheduling: true,
			wantErr:                ErrMayaAssignmentRequiresScheduling,
		},
		{
			name: "explicit false scheduling flag",
			updates: map[string]any{
				"assignee_id":             mayaID,
				"auto_scheduling_enabled": false,
			},
			wantErr: ErrMayaAssignmentRequiresScheduling,
		},
		{
			name: "missing duration",
			updates: map[string]any{
				"assignee_id":             mayaID,
				"auto_scheduling_enabled": true,
				"end_date":                &deliveryDate,
			},
			wantErr: ErrMayaAssignmentRequiresDuration,
		},
		{
			name: "missing delivery date or sprint",
			updates: map[string]any{
				"assignee_id":                mayaID,
				"auto_scheduling_enabled":    true,
				"estimated_duration_minutes": intPointer(60),
			},
			wantErr: ErrMayaAssignmentRequiresDeliveryDate,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := AutoSchedulingStatusOff
			if test.existingAutoScheduling {
				status = AutoSchedulingStatusPlanning
			}
			repo := &activityRecordingRepo{story: CoreSingleStory{
				ID:                    storyID,
				Workspace:             workspaceID,
				Assignee:              &previousAssigneeID,
				AutoSchedulingEnabled: test.existingAutoScheduling,
				AutoSchedulingStatus:  status,
			}}
			service := newAutoSchedulingContractService(repo)
			service.ConfigureMayaActor(mayaID)
			service.ConfigureAutoSchedulingEligibility(func(context.Context, uuid.UUID) (bool, error) {
				return true, nil
			})

			err := service.Update(context.Background(), storyID, workspaceID, test.updates)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("update error = %v, want %v", err, test.wantErr)
			}
			if repo.story.Assignee == nil || *repo.story.Assignee != previousAssigneeID {
				t.Fatal("incomplete Maya assignment changed persisted assignee")
			}
		})
	}
}

func intPointer(value int) *int {
	return &value
}

func timePointer(value time.Time) *time.Time {
	return &value
}
