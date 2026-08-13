package migrations

import (
	"strings"
	"testing"
)

const feedbackWidgetOptionalIdentityMigration = "000126_feedback_widget_optional_identity"

func TestFeedbackWidgetOptionalIdentityForwardMigration(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackWidgetOptionalIdentityMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	if !strings.Contains(migration, "CHECK (NOT enabled OR cardinality(allowed_origins) > 0)") {
		t.Fatal("enabled widgets must require an allowed origin")
	}
	if strings.Contains(migration, "signing_secret_encrypted IS NOT NULL") {
		t.Fatal("basic embeds must not require custom identity signing material")
	}
}

func TestFeedbackWidgetOptionalIdentityRollbackDisablesIncompatibleEmbeds(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackWidgetOptionalIdentityMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"SET enabled = false",
		"signing_secret_encrypted IS NULL",
		"signing_secret_version <= 0",
		"signing_secret_encrypted IS NOT NULL",
		"signing_secret_version > 0",
	} {
		if !strings.Contains(migration, contract) {
			t.Fatalf("rollback is missing contract %q", contract)
		}
	}
}
