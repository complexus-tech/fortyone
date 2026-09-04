package googledriverepository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestProviderLifecycleLocksUseSeparateUserAndGlobalSubjectNamespaces(t *testing.T) {
	t.Parallel()

	keys := make([]string, 0, 2)
	repository := &Repository{
		runProviderLock: func(ctx context.Context, key string, operation func(context.Context) error) error {
			keys = append(keys, key)
			return operation(ctx)
		},
	}
	userID := uuid.New()

	err := repository.WithinProviderUserLifecycle(t.Context(), userID, func(userCtx context.Context) error {
		return repository.WithinProviderSubjectLifecycle(userCtx, "google-subject", func(context.Context) error {
			return nil
		})
	})

	require.NoError(t, err)
	require.Equal(t, []string{
		providerUserLifecycleLockPrefix + userID.String(),
		providerSubjectLifecycleLockPrefix + "google-subject",
	}, keys)
}

func TestProviderLifecycleAdmissionLimit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		maxConnections int32
		want           int32
	}{
		{name: "defensive zero", maxConnections: 0, want: 1},
		{name: "single connection", maxConnections: 1, want: 1},
		{name: "one connection of headroom", maxConnections: 2, want: 1},
		{name: "small pool", maxConnections: 4, want: 3},
		{name: "bounded production pool", maxConnections: 25, want: 4},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, testCase.want, providerLifecycleAdmissionLimit(testCase.maxConnections))
		})
	}
}
