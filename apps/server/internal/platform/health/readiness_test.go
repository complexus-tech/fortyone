package health

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReadinessRequiresSupervisorAndDependencies(t *testing.T) {
	t.Parallel()

	var databaseHealthy atomic.Bool
	databaseHealthy.Store(true)
	readiness, err := NewReadiness(time.Second,
		Dependency{Name: "postgres", Check: func(context.Context) error {
			if !databaseHealthy.Load() {
				return errors.New("database unavailable")
			}
			return nil
		}},
		Dependency{Name: "redis", Check: func(context.Context) error { return nil }},
	)
	require.NoError(t, err)

	require.Equal(t, Report{
		Status: "not_ready",
		Phase:  PhaseStarting,
		Checks: map[string]string{"postgres": "not_checked", "redis": "not_checked"},
	}, readiness.Report(context.Background()))

	readiness.MarkReady()
	require.Equal(t, "ready", readiness.Report(context.Background()).Status)

	databaseHealthy.Store(false)
	report := readiness.Report(context.Background())
	require.Equal(t, "not_ready", report.Status)
	require.Equal(t, "unavailable", report.Checks["postgres"])
	require.Equal(t, "ready", report.Checks["redis"])

	readiness.MarkDraining()
	report = readiness.Report(context.Background())
	require.Equal(t, PhaseDraining, report.Phase)
	require.Equal(t, "not_checked", report.Checks["postgres"])

	readiness.MarkFailed()
	require.Equal(t, PhaseFailed, readiness.Phase())
}

func TestReadinessBoundsDependencyChecks(t *testing.T) {
	t.Parallel()

	readiness, err := NewReadiness(10*time.Millisecond, Dependency{
		Name: "blocked",
		Check: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	})
	require.NoError(t, err)
	readiness.MarkReady()

	started := time.Now()
	report := readiness.Report(context.Background())
	require.Less(t, time.Since(started), 250*time.Millisecond)
	require.Equal(t, "not_ready", report.Status)
	require.Equal(t, "unavailable", report.Checks["blocked"])
}

func TestNewReadinessRejectsInvalidDependencies(t *testing.T) {
	t.Parallel()

	_, err := NewReadiness(time.Second, Dependency{})
	require.ErrorContains(t, err, "name is required")

	_, err = NewReadiness(time.Second, Dependency{Name: "postgres"})
	require.ErrorContains(t, err, "has no check")

	check := func(context.Context) error { return nil }
	_, err = NewReadiness(time.Second,
		Dependency{Name: "redis", Check: check},
		Dependency{Name: " redis ", Check: check},
	)
	require.ErrorContains(t, err, "registered more than once")
}
