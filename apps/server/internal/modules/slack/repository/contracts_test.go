package slackrepository

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHumanQueriesKeepAuthorizationInPersistenceStatements(t *testing.T) {
	t.Parallel()

	authorization := readQuerySource(t, "queries/authorization.sql")
	for _, clause := range []string{
		"actor.is_active = TRUE",
		"workspace.deleted_at IS NULL",
		"membership.role IN ('admin', 'member')",
		"membership.role = 'admin'",
		"ORDER BY request.created_at DESC, request.id DESC",
		"ORDER BY LOWER(channel.name), channel.name, channel.id",
	} {
		require.Contains(t, authorization, clause)
	}

	settings := readQuerySource(t, "queries/settings.sql")
	require.Contains(t, settings, "membership.role = 'admin'")
	require.Contains(t, settings, "actor.is_active = TRUE")
	require.Contains(t, settings, "workspace.deleted_at IS NULL")
}

func TestMutationQueriesFenceActorAndInstallationGeneration(t *testing.T) {
	t.Parallel()

	installations := readQuerySource(t, "queries/installations.sql")
	for _, clause := range []string{
		"FOR UPDATE",
		"installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)",
		"installation_generation = CAST(sqlc.arg(previous_generation) AS uuid)",
		"ORDER BY COALESCE(next_attempt_at, created_at), created_at, id",
	} {
		require.Contains(t, installations+readQuerySource(t, "queries/uninstalls.sql"), clause)
	}

	audience := readQuerySource(t, "queries/channel_audience.sql")
	require.Contains(t, audience, "installation.installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)")
	require.Contains(t, audience, "team.is_private = FALSE")
}

func TestWebhookPayloadCutoverUsesStableScanAndExactCAS(t *testing.T) {
	t.Parallel()

	credentials := readQuerySource(t, "queries/credentials.sql")
	for _, clause := range []string{
		"event.id > CAST(sqlc.arg(after_id) AS uuid)",
		"ORDER BY event.id",
		"event.installation_generation IS NOT NULL",
		"AND payload_encrypted = CAST(sqlc.arg(previous_payload) AS text)",
	} {
		require.Contains(t, credentials, clause)
	}
}

func TestUpsertAgentSettingsRejectsOversizedGuidanceBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()

	repository := &Repo{}
	_, err := repository.UpsertAgentSettingsForAdmin(context.Background(), slackdomain.UpdateAgentSettingsCommand{
		WorkspaceID: uuid.New(),
		ActorID:     uuid.New(),
		Guidance:    strings.Repeat("a", MaxSlackAgentGuidanceRunes+1),
		Now:         time.Now(),
	})
	require.ErrorContains(t, err, "guidance is too long")
}

func readQuerySource(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}
