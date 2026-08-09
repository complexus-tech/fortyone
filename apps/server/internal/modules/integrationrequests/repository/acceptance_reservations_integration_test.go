package integrationrequestsrepository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestAcceptanceReservationPostgresContract(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_REQUEST_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_REQUEST_TEST_DATABASE_URL is not configured")
	}
	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	storyID := uuid.New()
	reservationActorID := uuid.New()
	retryActorID := uuid.New()

	_, err = db.ExecContext(context.Background(), `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3), ($4, $5, $6)
	`,
		reservationActorID, "reservation-"+reservationActorID.String(), "reservation-"+reservationActorID.String()+"@example.com",
		retryActorID, "retry-"+retryActorID.String(), "retry-"+retryActorID.String()+"@example.com",
	)
	require.NoError(t, err)
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, 'Acceptance reservation test', $2)
	`, workspaceID, "acceptance-reservation-"+workspaceID.String())
	require.NoError(t, err)
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, 'Reservation team', 'RSV', '#111827')
	`, teamID, workspaceID)
	require.NoError(t, err)
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2), ($1, $3)
	`, teamID, reservationActorID, retryActorID)
	require.NoError(t, err)
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO integration_requests (
			id, workspace_id, team_id, provider, source_type,
			source_external_id, title, priority
		) VALUES ($1, $2, $3, 'slack', 'message', $4, 'Crash-safe request', 'High')
	`, requestID, workspaceID, teamID, "event-"+requestID.String())
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := db.Exec(`DELETE FROM workspaces WHERE workspace_id = $1`, workspaceID)
		require.NoError(t, cleanupErr)
		_, cleanupErr = db.Exec(`DELETE FROM users WHERE user_id IN ($1, $2)`, reservationActorID, retryActorID)
		require.NoError(t, cleanupErr)
	})

	repo := New(nil, db)
	reserved, err := repo.ReserveAcceptance(context.Background(), workspaceID, requestID, reservationActorID)
	require.NoError(t, err)
	require.Equal(t, integrationrequests.AcceptanceStateReserved, reserved.AcceptanceState)
	require.Equal(t, &reservationActorID, reserved.AcceptanceStartedByUserID)

	_, err = repo.MarkDeclined(context.Background(), workspaceID, requestID, retryActorID)
	require.ErrorIs(t, err, sql.ErrNoRows)

	_, err = db.Exec(`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, reservationActorID)
	require.NoError(t, err)
	resumed, err := repo.ReserveAcceptance(context.Background(), workspaceID, requestID, retryActorID)
	require.NoError(t, err)
	require.Equal(t, &reservationActorID, resumed.AcceptanceStartedByUserID)

	_, err = db.Exec(`
		INSERT INTO stories (id, team_id, workspace_id, title, reporter_id, priority)
		VALUES ($1, $2, $3, 'Canonical converted story', $4, 'High')
	`, storyID, teamID, workspaceID, reservationActorID)
	require.NoError(t, err)

	accepted, err := repo.MarkAccepted(context.Background(), workspaceID, requestID, storyID, reservationActorID)
	require.NoError(t, err)
	require.Equal(t, integrationrequests.StatusAccepted, accepted.Status)
	require.Equal(t, &storyID, accepted.AcceptedStoryID)
	require.Equal(t, &reservationActorID, accepted.AcceptedByUserID)
	require.Equal(t, integrationrequests.AcceptanceStateIdle, accepted.AcceptanceState)
	require.Nil(t, accepted.AcceptanceStartedByUserID)
	require.Nil(t, accepted.AcceptanceStartedAt)
}
