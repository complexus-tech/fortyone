package mayahttp

import (
	"fmt"
	"net/http"
	"testing"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

func TestStatusCodeMapsAutoSchedulingContractErrors(t *testing.T) {
	handler := &Handlers{}
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "locked schedule", err: stories.ErrAutoSchedulingOwnerLocked, status: http.StatusConflict},
		{name: "lock without blocks", err: stories.ErrAutoSchedulingLockEmpty, status: http.StatusConflict},
		{name: "stale story", err: stories.ErrStoryChanged, status: http.StatusConflict},
		{name: "lost entitlement", err: maya.ErrMayaAccessDenied, status: http.StatusPaymentRequired},
		{name: "invalid plan", err: maya.ErrInvalidPlanInput, status: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if status := handler.statusCode(fmt.Errorf("wrapped: %w", test.err)); status != test.status {
				t.Fatalf("statusCode() = %d, want %d", status, test.status)
			}
		})
	}
}
