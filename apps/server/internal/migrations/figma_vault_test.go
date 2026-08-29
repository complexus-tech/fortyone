package migrations

import (
	"strings"
	"testing"
)

func TestFigmaVaultMigrationEstablishesVersionAndGenerationFences(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000169_figma_vault_installation_generation.up.sql")
	if err != nil {
		t.Fatalf("read Figma vault migration: %v", err)
	}
	source := string(forward)
	for _, contract := range []string{
		"credential_key_version smallint NOT NULL DEFAULT 1",
		"installation_generation uuid NOT NULL DEFAULT gen_random_uuid()",
		"figma_connections_credential_key_version_check",
		"CHECK (credential_key_version > 0)",
		"figma_connections_active_generation",
		"(workspace_id, installation_generation)",
		"WHERE is_active",
		"shared context-bound credential vault",
		"durable webhook deliveries",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("Figma vault migration is missing %q", contract)
		}
	}
}

func TestFigmaVaultRollbackFailsClosedAfterCredentialUpgrade(t *testing.T) {
	t.Parallel()

	rollback, err := FS.ReadFile("000169_figma_vault_installation_generation.down.sql")
	if err != nil {
		t.Fatalf("read Figma vault rollback: %v", err)
	}
	source := string(rollback)
	for _, contract := range []string{
		"WHERE credential_key_version <> 1",
		"cannot be reversed after Figma credentials use the shared vault",
		"DROP INDEX IF EXISTS public.figma_connections_active_generation",
		"DROP COLUMN IF EXISTS installation_generation",
		"DROP COLUMN IF EXISTS credential_key_version",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("Figma vault rollback is missing %q", contract)
		}
	}
}
