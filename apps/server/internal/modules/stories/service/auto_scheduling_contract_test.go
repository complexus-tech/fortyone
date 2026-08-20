package stories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateStoryAutoSchedulingContract(t *testing.T) {
	t.Parallel()

	for _, status := range []string{
		AutoSchedulingStatusOff,
		AutoSchedulingStatusNeedsOwner,
		AutoSchedulingStatusNeedsTime,
		AutoSchedulingStatusPlanning,
		AutoSchedulingStatusScheduled,
		AutoSchedulingStatusAtRisk,
		AutoSchedulingStatusCannotFit,
		AutoSchedulingStatusLocked,
	} {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			if err := ValidateStoryAutoSchedulingContract(true, false, status); err != nil {
				t.Fatalf("validate status %q: %v", status, err)
			}
		})
	}
}

func TestValidateStoryAutoSchedulingContractRejectsInvalidState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enabled bool
		locked  bool
		status  string
		wantErr error
	}{
		{
			name:    "locked while disabled",
			enabled: false,
			locked:  true,
			status:  AutoSchedulingStatusLocked,
			wantErr: ErrLockedAutoSchedulingOff,
		},
		{
			name:    "unknown status",
			enabled: true,
			locked:  false,
			status:  "queued",
			wantErr: ErrInvalidAutoSchedulingStatus,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateStoryAutoSchedulingContract(test.enabled, test.locked, test.status)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("validate auto-scheduling contract: got %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestToCoreSingleStoryDefaultsAutoSchedulingStatus(t *testing.T) {
	t.Parallel()

	story := toCoreSingleStory(CoreNewStory{}, [16]byte{})
	if story.AutoSchedulingStatus != AutoSchedulingStatusOff {
		t.Fatalf("auto-scheduling status = %q, want %q", story.AutoSchedulingStatus, AutoSchedulingStatusOff)
	}
}

func TestPrepareAutoSchedulingCreateForMayaEnablesDurableIntent(t *testing.T) {
	mayaActorID := uuid.New()
	service := &Service{mayaAssignment: &mayaAssignmentPolicy{assigneeID: mayaActorID}}
	story := CoreSingleStory{Assignee: &mayaActorID}

	if err := service.prepareAutoSchedulingCreate(&story); err != nil {
		t.Fatalf("prepareAutoSchedulingCreate returned error: %v", err)
	}
	if !story.AutoSchedulingEnabled || story.AutoSchedulingLocked || story.AutoSchedulingStatus != AutoSchedulingStatusNeedsOwner {
		t.Fatalf("Maya assignment must enable intake without creating a lock: %#v", story)
	}
	if story.AutoSchedulingReason == nil || *story.AutoSchedulingReason != autoSchedulingSelectingReason {
		t.Fatalf("unexpected Maya intake reason: %v", story.AutoSchedulingReason)
	}
}

func TestPrepareAutoSchedulingCreateRejectsLockWithoutCommittedBlocks(t *testing.T) {
	service := &Service{}
	story := CoreSingleStory{AutoSchedulingEnabled: true, AutoSchedulingLocked: true}

	err := service.prepareAutoSchedulingCreate(&story)
	if !errors.Is(err, ErrAutoSchedulingLockEmpty) {
		t.Fatalf("prepareAutoSchedulingCreate error = %v, want %v", err, ErrAutoSchedulingLockEmpty)
	}
}

func TestPrepareAutoSchedulingUpdateAllowsPausingExistingMayaAssignment(t *testing.T) {
	mayaActorID := uuid.New()
	service := &Service{mayaAssignment: &mayaAssignmentPolicy{assigneeID: mayaActorID}}
	story := CoreSingleStory{
		Assignee: &mayaActorID, AutoSchedulingEnabled: true, AutoSchedulingLocked: true,
		AutoSchedulingStatus: AutoSchedulingStatusLocked,
	}
	updates := map[string]any{"auto_scheduling_enabled": false}

	reconcile, err := service.prepareAutoSchedulingUpdate(context.Background(), story, updates)
	if err != nil {
		t.Fatalf("prepareAutoSchedulingUpdate returned error: %v", err)
	}
	if !reconcile || updates["auto_scheduling_enabled"] != false || updates["auto_scheduling_locked"] != false || updates["auto_scheduling_status"] != AutoSchedulingStatusOff {
		t.Fatalf("pause must atomically unlock and retire Maya scheduling: %#v", updates)
	}
}

func TestPrepareAutoSchedulingUpdateAllowsPauseAndReassignFromLockedStory(t *testing.T) {
	previousOwnerID := uuid.New()
	newOwnerID := uuid.New()
	service := &Service{}
	story := CoreSingleStory{
		Assignee: &previousOwnerID, AutoSchedulingEnabled: true, AutoSchedulingLocked: true,
		AutoSchedulingStatus: AutoSchedulingStatusLocked,
	}
	updates := map[string]any{
		"assignee_id":             newOwnerID,
		"auto_scheduling_enabled": false,
	}

	reconcile, err := service.prepareAutoSchedulingUpdate(context.Background(), story, updates)
	if err != nil {
		t.Fatalf("pause plus reassignment returned error: %v", err)
	}
	if !reconcile || updates["auto_scheduling_locked"] != false || updates["auto_scheduling_status"] != AutoSchedulingStatusOff {
		t.Fatalf("pause plus reassignment must retire the lock in one mutation: %#v", updates)
	}
}

func TestPrepareAutoSchedulingUpdateRejectsLockedOwnerChangeWithoutUnlock(t *testing.T) {
	previousOwnerID := uuid.New()
	newOwnerID := uuid.New()
	service := &Service{}
	story := CoreSingleStory{
		Assignee: &previousOwnerID, AutoSchedulingEnabled: true, AutoSchedulingLocked: true,
		AutoSchedulingStatus: AutoSchedulingStatusLocked,
	}

	_, err := service.prepareAutoSchedulingUpdate(context.Background(), story, map[string]any{"assignee_id": newOwnerID})
	if !errors.Is(err, ErrAutoSchedulingOwnerLocked) {
		t.Fatalf("locked owner change error = %v, want %v", err, ErrAutoSchedulingOwnerLocked)
	}
}

func TestPrepareAutoSchedulingUpdatePreservesEnabledIntentForTerminalStory(t *testing.T) {
	ownerID := uuid.New()
	now := time.Now().UTC()
	service := &Service{}
	story := CoreSingleStory{
		Assignee: &ownerID, AutoSchedulingEnabled: true, AutoSchedulingLocked: true,
		AutoSchedulingStatus: AutoSchedulingStatusLocked,
	}
	updates := map[string]any{"completed_at": now}

	reconcile, err := service.prepareAutoSchedulingUpdate(context.Background(), story, updates)
	if err != nil {
		t.Fatalf("terminal update returned error: %v", err)
	}
	if !reconcile || updates["auto_scheduling_locked"] != false || updates["auto_scheduling_status"] != AutoSchedulingStatusOff {
		t.Fatalf("terminal story must unlock and stop while preserving enabled intent: %#v", updates)
	}
	if _, changed := updates["auto_scheduling_enabled"]; changed {
		t.Fatalf("terminal lifecycle must not clear explicit user intent: %#v", updates)
	}
}

func TestPrepareAutoSchedulingUpdateRequestsCleanupForTerminalLifecycle(t *testing.T) {
	service := &Service{}
	updates := map[string]any{"completed_at": time.Now().UTC()}

	reconcile, err := service.prepareAutoSchedulingUpdate(context.Background(), CoreSingleStory{
		AutoSchedulingStatus: AutoSchedulingStatusOff,
	}, updates)
	if err != nil {
		t.Fatalf("terminal update returned error: %v", err)
	}
	if !reconcile {
		t.Fatal("terminal lifecycle update must enqueue schedule cleanup even when auto-scheduling is already off")
	}
}

func TestPrepareAutoSchedulingUpdateDoesNotRewriteAlreadyOffStory(t *testing.T) {
	service := &Service{}
	updates := map[string]any{"title": "Updated title"}

	reconcile, err := service.prepareAutoSchedulingUpdate(context.Background(), CoreSingleStory{AutoSchedulingStatus: AutoSchedulingStatusOff}, updates)
	if err != nil {
		t.Fatalf("disabled title update returned error: %v", err)
	}
	if reconcile {
		t.Fatal("disabled title update must not enqueue auto-scheduling")
	}
	for _, field := range []string{"auto_scheduling_locked", "auto_scheduling_status", "auto_scheduling_reason", "auto_scheduling_updated_at"} {
		if _, exists := updates[field]; exists {
			t.Fatalf("disabled title update unexpectedly rewrote %s: %#v", field, updates)
		}
	}
}
