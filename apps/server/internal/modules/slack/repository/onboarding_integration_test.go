package slackrepository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestSlackUserOnboardingReceiptPostgresContract is skipped unless an
// isolated, fully migrated PostgreSQL database is supplied.
func TestSlackUserOnboardingReceiptPostgresContract(t *testing.T) {
	databaseURL := os.Getenv("SLACK_ONBOARDING_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SLACK_ONBOARDING_TEST_DATABASE_URL is not configured")
	}
	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()
	deliveryID := uuid.New()
	suffix := uuid.NewString()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "slack-onboarding-"+suffix, "slack-onboarding-"+suffix+"@example.test")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Slack Onboarding Test', $2, $3)
	`, workspaceID, "slack-onboarding-"+suffix, userID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := db.ExecContext(context.Background(), "DELETE FROM workspaces WHERE workspace_id = $1", workspaceID)
		require.NoError(t, cleanupErr)
		_, cleanupErr = db.ExecContext(context.Background(), "DELETE FROM users WHERE user_id = $1", userID)
		require.NoError(t, cleanupErr)
	})

	_, err = db.ExecContext(ctx, `
		INSERT INTO messaging_outbound_deliveries (
			id, provider, workspace_id, installation_generation,
			external_workspace_id, external_recipient_user_id,
			idempotency_key, external_channel_id, content, purpose,
			status, attempt_count
		) VALUES (
			$1, 'slack', $2, $3,
			'T-onboarding', 'U-onboarding',
			$4, 'U-onboarding', 'Welcome', 'onboarding',
			'delivering', 1
		)
	`, deliveryID, workspaceID, uuid.New(), "slack-onboarding:"+suffix)
	require.NoError(t, err)

	repo := New(nil, db)
	delivered, err := repo.HasSlackUserOnboardingReceipt(ctx, workspaceID, "T-onboarding", "U-onboarding")
	require.NoError(t, err)
	require.False(t, delivered)

	_, err = db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'delivered', external_message_id = '171.100', delivered_at = NOW()
		WHERE id = $1
	`, deliveryID)
	require.NoError(t, err)
	delivered, err = repo.HasSlackUserOnboardingReceipt(ctx, workspaceID, "T-onboarding", "U-onboarding")
	require.NoError(t, err)
	require.True(t, delivered)

	_, err = db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'delivered'
		WHERE id = $1
	`, deliveryID)
	require.NoError(t, err)
	var receiptCount int
	identityDigest := slackOnboardingIdentityDigest("T-onboarding", "U-onboarding")
	require.NoError(t, db.GetContext(ctx, &receiptCount, `
		SELECT COUNT(*)
		FROM slack_user_onboarding_receipts
		WHERE workspace_id = $1
		  AND external_identity_digest = $2
	`, workspaceID, identityDigest[:]))
	require.Equal(t, 1, receiptCount)
	var guideDeliveredAt time.Time
	require.NoError(t, db.GetContext(ctx, &guideDeliveredAt, `
		SELECT guide_delivered_at
		FROM slack_user_onboarding_receipts
		WHERE workspace_id = $1
		  AND external_identity_digest = $2
	`, workspaceID, identityDigest[:]))
	require.False(t, guideDeliveredAt.IsZero())

	_, err = db.ExecContext(ctx, "DELETE FROM messaging_outbound_deliveries WHERE id = $1", deliveryID)
	require.NoError(t, err)
	delivered, err = repo.HasSlackUserOnboardingReceipt(ctx, workspaceID, "T-onboarding", "U-onboarding")
	require.NoError(t, err)
	require.True(t, delivered)
}
