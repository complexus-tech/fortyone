package notifications

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWeeklyDigestStatsQueryRequiresCompleteScopeAndPeriod(t *testing.T) {
	t.Parallel()

	valid := WeeklyDigestStatsQuery{
		UserID: uuid.New(), WorkspaceID: uuid.New(), AsOf: time.Now(),
	}
	require.NoError(t, valid.Validate())

	for name, query := range map[string]WeeklyDigestStatsQuery{
		"user":      {WorkspaceID: valid.WorkspaceID, AsOf: valid.AsOf},
		"workspace": {UserID: valid.UserID, AsOf: valid.AsOf},
		"as of":     {UserID: valid.UserID, WorkspaceID: valid.WorkspaceID},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.ErrorIs(t, query.Validate(), ErrInvalid)
		})
	}
}
