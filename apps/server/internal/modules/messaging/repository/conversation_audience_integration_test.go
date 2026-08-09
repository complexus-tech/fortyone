package messagingrepository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestConversationAudiencePostgresContract proves that one Slack thread uses
// separate retained histories when its effective team audience changes. It is
// skipped unless an isolated, fully migrated PostgreSQL database is supplied.
func TestConversationAudiencePostgresContract(t *testing.T) {
	databaseURL := os.Getenv("MESSAGING_AUDIENCE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("MESSAGING_AUDIENCE_TEST_DATABASE_URL is not configured")
	}
	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()
	suffix := uuid.NewString()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "audience-"+suffix, "audience-"+suffix+"@example.test")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Audience Test', $2, $3)
	`, workspaceID, "audience-"+suffix, userID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := db.ExecContext(context.Background(), "DELETE FROM workspaces WHERE workspace_id = $1", workspaceID)
		require.NoError(t, cleanupErr)
		_, cleanupErr = db.ExecContext(context.Background(), "DELETE FROM users WHERE user_id = $1", userID)
		require.NoError(t, cleanupErr)
	})

	repo := New(db)
	base := ConversationInput{
		Provider:            "slack",
		WorkspaceID:         workspaceID,
		ExternalWorkspaceID: "T-audience",
		ExternalChannelID:   "C-audience",
		ExternalThreadID:    "1710000000.001",
		UserID:              userID,
		AudienceScope:       ConversationAudienceChannel,
	}
	first := base
	first.AudienceFingerprint = "v1:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	second := base
	second.AudienceFingerprint = "v1:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	firstID, err := repo.UpsertConversation(ctx, first)
	require.NoError(t, err)
	require.NoError(t, repo.AppendMessage(ctx, firstID, "1710000000.002", "assistant", "Team A only"))
	secondID, err := repo.UpsertConversation(ctx, second)
	require.NoError(t, err)
	require.NotEqual(t, firstID, secondID)
	require.NoError(t, repo.AppendMessage(ctx, secondID, "1710000000.003", "assistant", "Team B only"))

	firstHistory, err := repo.ListRecentMessages(ctx, firstID, 20)
	require.NoError(t, err)
	secondHistory, err := repo.ListRecentMessages(ctx, secondID, 20)
	require.NoError(t, err)
	require.Len(t, firstHistory, 1)
	require.Equal(t, "Team A only", firstHistory[0].Content)
	require.Len(t, secondHistory, 1)
	require.Equal(t, "Team B only", secondHistory[0].Content)
}
