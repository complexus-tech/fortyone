package agentadapter

import (
	"context"
	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestChatRuntimeCannotPerformAnyStepWhenRolloutIsDisabled(t *testing.T) {
	runtime, run, store := chatRuntimeFixture()
	runtime.Rollout = chatrollout.Policy{}
	now := time.Now()
	run.ApprovedAt = &now
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
	_, err := runtime.Prepare(t.Context(), run)
	require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
	_, err = runtime.Apply(t.Context(), run, agentruns.Proposal{})
	require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
	_, err = runtime.Publish(t.Context(), run, agentruns.Outcome{})
	require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
	_, err = runtime.PublishProposal(t.Context(), run, agentruns.Proposal{})
	require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
	_, err = runtime.PublishFailure(t.Context(), run, "Unavailable")
	require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
	require.Empty(t, store.replies)
	runtime.Rollout = chatrollout.New(chatrollout.ReaderFunc(func(_ context.Context, id uuid.UUID) (bool, error) { return true && id == uuid.New(), nil }))
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
}

func TestChatRunControlsRejectDisabledWorkspaceBeforeAccessLookup(t *testing.T) {
	access := ChatAccess{Configured: true}
	workspace := uuid.New()
	require.ErrorIs(t, access.CanRead(t.Context(), uuid.New(), workspace), chatdomain.ErrForbidden)
	enabled, err := access.Enabled(t.Context(), workspace)
	require.NoError(t, err)
	require.False(t, enabled)
}
