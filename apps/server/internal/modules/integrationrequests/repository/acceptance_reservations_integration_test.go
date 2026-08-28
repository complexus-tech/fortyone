//go:build integration

package integrationrequestsrepository

import (
	"testing"
	"time"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAcceptanceReservationPostgresContract(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	storyID := uuid.New()
	reservationActorID := uuid.New()
	retryActorID := uuid.New()

	_, err := postgres.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3), ($4, $5, $6)
	`,
		reservationActorID, "reservation-"+reservationActorID.String(), "reservation-"+reservationActorID.String()+"@example.com",
		retryActorID, "retry-"+retryActorID.String(), "retry-"+retryActorID.String()+"@example.com",
	)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, 'Acceptance reservation test', $2)
	`, workspaceID, "acceptance-reservation-"+workspaceID.String())
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, 'Reservation team', 'RSV', '#111827')
	`, teamID, workspaceID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2), ($1, $3)
	`, teamID, reservationActorID, retryActorID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO integration_requests (
			id, workspace_id, team_id, provider, source_type,
			source_external_id, title, priority
		) VALUES ($1, $2, $3, 'slack', 'message', $4, 'Crash-safe request', 'High')
	`, requestID, workspaceID, teamID, "event-"+requestID.String())
	require.NoError(t, err)

	repo := New(postgres.Pool)
	reserved, err := repo.ReserveAcceptance(ctx, workspaceID, requestID, reservationActorID)
	require.NoError(t, err)
	require.Equal(t, integrationrequestdomain.AcceptanceStateReserved, reserved.AcceptanceState)
	require.Equal(t, &reservationActorID, reserved.AcceptanceStartedByUserID)
	require.WithinDuration(t, time.Now().UTC(), *reserved.AcceptanceStartedAt, 5*time.Second)

	_, err = repo.MarkDeclined(ctx, workspaceID, requestID, retryActorID)
	require.ErrorIs(t, err, integrationrequestdomain.ErrNotFound)

	_, err = postgres.Pool.Exec(ctx, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, reservationActorID)
	require.NoError(t, err)
	resumed, err := repo.ReserveAcceptance(ctx, workspaceID, requestID, retryActorID)
	require.NoError(t, err)
	require.Equal(t, &reservationActorID, resumed.AcceptanceStartedByUserID)

	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO stories (id, team_id, workspace_id, title, reporter_id, priority)
		VALUES ($1, $2, $3, 'Canonical converted story', $4, 'High')
	`, storyID, teamID, workspaceID, reservationActorID)
	require.NoError(t, err)

	accepted, err := repo.MarkAccepted(ctx, workspaceID, requestID, storyID, reservationActorID)
	require.NoError(t, err)
	require.Equal(t, integrationrequestdomain.StatusAccepted, accepted.Status)
	require.Equal(t, &storyID, accepted.AcceptedStoryID)
	require.Equal(t, &reservationActorID, accepted.AcceptedByUserID)
	require.Equal(t, integrationrequestdomain.AcceptanceStateIdle, accepted.AcceptanceState)
	require.Nil(t, accepted.AcceptanceStartedByUserID)
	require.Nil(t, accepted.AcceptanceStartedAt)
}
