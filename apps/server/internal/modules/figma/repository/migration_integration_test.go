//go:build integration

package figmarepository

import (
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/migrations"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFigmaVaultMigrationBackfillsLegacyConnectionsAndRollsBackBeforeUse(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	down := readFigmaMigration(t, "000169_figma_vault_installation_generation.down.sql")
	up := readFigmaMigration(t, "000169_figma_vault_installation_generation.up.sql")
	_, err := postgres.Pool.Exec(ctx, down)
	require.NoError(t, err)

	workspaceID, userID, _ := insertFigmaOAuthStateOwner(t, postgres, ctx)
	connectionID := uuid.New()
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO figma_connections (
			id, workspace_id, figma_user_id, token_payload, scopes,
			expires_at, connected_by_user_id
		)
		VALUES ($1, $2, 'legacy-user', 'legacy-ciphertext', '{}', $3, $4)
	`, connectionID, workspaceID, time.Now().UTC().Add(time.Hour), userID)
	require.NoError(t, err)

	_, err = postgres.Pool.Exec(ctx, up)
	require.NoError(t, err)
	var version int16
	var generation uuid.UUID
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT credential_key_version, installation_generation
		FROM figma_connections
		WHERE id = $1
	`, connectionID).Scan(&version, &generation))
	require.Equal(t, int16(1), version)
	require.NotEqual(t, uuid.Nil, generation)

	_, err = postgres.Pool.Exec(ctx, `
		UPDATE figma_connections
		SET credential_key_version = 0
		WHERE id = $1
	`, connectionID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "figma_connections_credential_key_version_check")

	_, err = postgres.Pool.Exec(ctx, down)
	require.NoError(t, err)
	var columnsRemain bool
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'figma_connections'
			  AND column_name IN ('credential_key_version', 'installation_generation')
		)
	`).Scan(&columnsRemain))
	require.False(t, columnsRemain)
}

func TestFigmaVaultMigrationRefusesRollbackAfterSharedVaultUse(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, userID, _ := insertFigmaOAuthStateOwner(t, postgres, ctx)
	connectionID := uuid.New()
	_, err := postgres.Pool.Exec(ctx, `
		INSERT INTO figma_connections (
			id, workspace_id, figma_user_id, token_payload,
			credential_key_version, installation_generation, scopes,
			expires_at, connected_by_user_id
		)
		VALUES ($1, $2, 'vault-user', 'vault.v2.redacted', 2, $3, '{}', $4, $5)
	`, connectionID, workspaceID, uuid.New(), time.Now().UTC().Add(time.Hour), userID)
	require.NoError(t, err)

	_, err = postgres.Pool.Exec(
		ctx,
		readFigmaMigration(t, "000169_figma_vault_installation_generation.down.sql"),
	)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "cannot be reversed after Figma credentials use the shared vault"))

	var version int16
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT credential_key_version
		FROM figma_connections
		WHERE id = $1
	`, connectionID).Scan(&version))
	require.Equal(t, int16(2), version)
}

func readFigmaMigration(t testing.TB, name string) string {
	t.Helper()
	script, err := migrations.FS.ReadFile(name)
	require.NoError(t, err)
	return string(script)
}
