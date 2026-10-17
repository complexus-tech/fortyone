package chatrollout

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPolicyReadsLiveSettingAndFailsClosed(t *testing.T) {
	id := uuid.New()
	enabled := true
	var failure error
	calls := 0
	policy := New(ReaderFunc(func(_ context.Context, workspace uuid.UUID) (bool, error) {
		calls++
		require.Equal(t, id, workspace)
		return enabled, failure
	}))
	require.True(t, policy.Allows(t.Context(), id))
	enabled = false
	require.False(t, policy.Allows(t.Context(), id))
	enabled = true
	require.True(t, policy.Allows(t.Context(), id))
	failure = errors.New("unavailable")
	require.False(t, policy.Allows(t.Context(), id))
	require.Equal(t, 4, calls)
	require.False(t, Policy{}.Allows(t.Context(), id))
	require.False(t, policy.Allows(t.Context(), uuid.Nil))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.False(t, policy.Allows(ctx, id))
	require.Equal(t, 4, calls)
}
