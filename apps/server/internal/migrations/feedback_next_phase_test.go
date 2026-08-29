package migrations

import (
	"strings"
	"testing"
)

const feedbackNextPhaseMigration = "000122_feedback_contributor_delivery_updates_widget_identity"

func TestFeedbackNextPhaseForwardMigrationContracts(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackNextPhaseMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	assertMigrationContains(t, migration,
		"participation_mode IN ('account_required', 'verified_guest', 'anonymous_allowed')",
		"ADD VALUE IF NOT EXISTS 'feedback_update_published'",
		"guest_identity_policy IN ('show_identity', 'allow_public_masking', 'always_mask_guests')",
		"kind IN ('account', 'verified_guest', 'anonymous', 'external')",
		"CHECK (NOT public_masked OR kind = 'verified_guest')",
		"feedback_contributors_portal_email_unique",
		"WHERE email IS NOT NULL AND kind <> 'external'",
		"feedback_contributors_portal_external_id_unique",
		"octet_length(token_hash) = 32",
		"CREATE TABLE public.feedback_contributor_verifications",
		"CREATE TABLE public.feedback_contributor_sessions",
		"source IN ('portal', 'widget', 'preferences')",
		"ALTER COLUMN contributor_id SET NOT NULL",
		"feedback_votes_pkey PRIMARY KEY (item_id, contributor_id)",
		"REFERENCES public.users(user_id) ON DELETE SET NULL",
		"CREATE TABLE public.feedback_item_followers",
		"CREATE TABLE public.feedback_portal_followers",
		"CREATE TABLE public.feedback_contributor_preferences",
		"CREATE TABLE public.feedback_contributor_unsubscribe_tokens",
		"purpose IN ('unsubscribe_item', 'unsubscribe_portal', 'all_email', 'manage_preferences')",
		"CREATE TABLE public.feedback_contributor_deliveries",
		"subject text NOT NULL",
		"destination_url text NOT NULL",
		"feedback_contributor_deliveries_event_dedupe_key",
		"feedback_updates_portal_slug_unique",
		"feedback_updates_publication_check",
		"CREATE TABLE public.feedback_widget_settings",
		"feedback_allowed_origins_are_valid",
		"CREATE TABLE public.feedback_widget_signing_secret_rotations",
		"CREATE TABLE public.feedback_widget_assertion_nonces",
	)

	if strings.Contains(migration, "nonce text") || strings.Contains(migration, "token text") {
		t.Fatal("forward migration must store hashes, not raw tokens or nonces")
	}
}

func TestFeedbackNextPhaseRollbackRestoresLegacyContracts(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackNextPhaseMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	assertMigrationContains(t, migration,
		"cannot be rolled back while verified guest or external contributors exist",
		"cannot be rolled back while contributor-only votes exist",
		"cannot be rolled back while contributor-only comments exist",
		"DROP TABLE IF EXISTS public.feedback_widget_assertion_nonces",
		"DROP TABLE IF EXISTS public.feedback_contributor_deliveries",
		"ALTER COLUMN user_id SET NOT NULL",
		"feedback_votes_pkey PRIMARY KEY (item_id, user_id)",
		"REFERENCES public.users(user_id) ON DELETE CASCADE",
		"kind IN ('account', 'anonymous')",
		"participation_mode IN ('account_required', 'anonymous_allowed')",
	)
}

func assertMigrationContains(t *testing.T, migration string, contracts ...string) {
	t.Helper()

	for _, contract := range contracts {
		if !strings.Contains(migration, contract) {
			t.Errorf("migration is missing contract %q", contract)
		}
	}
}
