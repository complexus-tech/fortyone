package adminrepository

import (
	"testing"
	"time"

	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceIntegrationSummaryNormalizesHealthAndMappingCoverage(t *testing.T) {
	now := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	figmaID := uuid.New()
	expiresAt := now.Add(-time.Minute)
	figmaCreatedAt := now.Add(-24 * time.Hour)

	summary, err := workspaceIntegrationSummaryFromRow(adminsql.ListAdminWorkspaceIntegrationsRow{
		WorkspaceID:             uuid.New(),
		Name:                    "Acme",
		Slug:                    "acme",
		MemberCount:             5,
		GithubConnectionCount:   1,
		GithubAccountLabel:      "acme-inc",
		GithubNeedsAttention:    true,
		GithubMappingCount:      3,
		GithubLastSyncedAt:      now.Add(-time.Hour),
		CalendarConnectionCount: 2,
		CalendarNeedsAttention:  true,
		CalendarLastSyncedAt:    now.Add(-2 * time.Hour),
		FigmaID:                 &figmaID,
		FigmaExpiresAt:          &expiresAt,
		FigmaCreatedAt:          &figmaCreatedAt,
	}, now)

	require.NoError(t, err)
	require.Equal(t, "not_connected", summary.Integrations[0].State)
	require.Equal(t, "suspended", summary.Integrations[1].State)
	require.Equal(t, 2, summary.Integrations[1].UnmappedMemberCount)
	require.Equal(t, "degraded", summary.Integrations[2].State)
	require.Equal(t, 3, summary.Integrations[2].UnmappedMemberCount)
	require.Equal(t, "reauthorization_required", summary.Integrations[3].State)
}
