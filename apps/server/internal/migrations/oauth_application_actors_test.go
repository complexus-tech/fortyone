package migrations

import (
	"strings"
	"testing"
)

func TestOAuthApplicationActorMigrationEstablishesCredentialAndActorFences(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile("000170_oauth_application_actors.up.sql")
	if err != nil {
		t.Fatalf("read OAuth application actor migration: %v", err)
	}
	migration := strings.ToLower(string(data))
	for _, contract := range []string{
		"kind in ('human_user', 'service_account', 'oauth_application')",
		"oauth_client_secrets_lookup_prefix_key unique (lookup_prefix)",
		"check (octet_length(secret_digest) = 32)",
		"overlap_expires_at is null or overlap_expires_at > created_at",
		"oauth_application_installations_active_identity_key",
		"check (scope = 'stories:write')",
		"oauth_application_installations_principal_kind_check",
		"oauth_audit_events_actor_shape_check",
		"oauth_audit_events_subject_created_idx",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("OAuth application actor migration is missing %q", contract)
		}
	}
	for _, forbidden := range []string{
		"client_secret varchar",
		"client_secret text",
		"plaintext_secret",
	} {
		if strings.Contains(migration, forbidden) {
			t.Errorf("OAuth application actor migration persists forbidden secret material %q", forbidden)
		}
	}
}

func TestOAuthApplicationActorRollbackRefusesAdoptedData(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile("000170_oauth_application_actors.down.sql")
	if err != nil {
		t.Fatalf("read OAuth application actor rollback: %v", err)
	}
	rollback := strings.ToLower(string(data))
	for _, contract := range []string{
		"exists (select 1 from public.oauth_client_secrets)",
		"exists (select 1 from public.oauth_application_installations)",
		"where kind = 'oauth_application'",
		"migration 000170 is forward-only after oauth application actor data exists",
	} {
		if !strings.Contains(rollback, contract) {
			t.Errorf("OAuth application actor rollback is missing %q", contract)
		}
	}
}
