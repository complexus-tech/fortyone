package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestGoogleDriveRevocationSagaPersistsEncryptedGenerationFences(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000179_google_drive_revocation_saga.up.sql")
	if err != nil {
		t.Fatalf("read Google Drive revocation migration: %v", err)
	}
	source := string(forward)
	for _, contract := range []string{
		"CREATE TABLE public.google_drive_revocation_outbox",
		"source_account_id uuid",
		"UNIQUE (google_subject, installation_generation)",
		"credential_payload LIKE 'vault.v2.%'",
		"google_drive_accounts_active_google_subject",
		"CREATE FUNCTION public.stage_google_drive_account_revocation(",
		"CREATE OR REPLACE FUNCTION public.revoke_orphaned_google_drive_account()",
		"DELETE FROM public.google_drive_file_grants AS grant_record",
		"file.workspace_id = OLD.workspace_id",
		"grant_record.user_id = OLD.user_id",
		"grant_record.account_id = OLD.account_id",
		"google_drive_accounts_stage_revocation_before_delete",
		"BEFORE DELETE ON public.google_drive_accounts",
		"ON CONFLICT (google_subject, installation_generation) DO NOTHING",
		"credential_payload = ''",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("Google Drive revocation migration is missing %q", contract)
		}
	}

	tableEnd := strings.Index(source, "CREATE INDEX google_drive_revocation_outbox_ready")
	if tableEnd == -1 {
		t.Fatal("Google Drive revocation migration has no outbox table boundary")
	}
	if strings.Contains(source[:tableEnd], "REFERENCES") {
		t.Error("Google Drive revocation outbox must survive user and account deletion without foreign keys")
	}
	if strings.Contains(source[:tableEnd], "source_account_id uuid NOT NULL") {
		t.Error("failed OAuth cleanup must support a revocation without a persisted source account")
	}
}

func TestGoogleDriveRevocationQueriesFenceReconnectBeforeClaim(t *testing.T) {
	t.Parallel()

	connectionQueries, err := os.ReadFile("../modules/googledrive/repository/queries/connections.sql")
	if err != nil {
		t.Fatalf("read Google Drive connection queries: %v", err)
	}
	for _, contract := range []string{
		"AcquireGoogleDriveProviderLifecycleLock",
		"pg_advisory_lock",
		"ReleaseGoogleDriveProviderLifecycleLock",
		"LockGoogleDriveSubjectLifecycle",
		"GetActiveGoogleDriveAccountBySubject",
		"SupersedeGoogleDriveRevocationsForGeneration",
	} {
		combined, revocationErr := os.ReadFile("../modules/googledrive/repository/queries/revocations.sql")
		if revocationErr != nil {
			t.Fatalf("read Google Drive revocation queries: %v", revocationErr)
		}
		if !strings.Contains(string(connectionQueries)+string(combined), contract) {
			t.Errorf("Google Drive lifecycle queries are missing %q", contract)
		}
	}
	upsertStart := strings.Index(string(connectionQueries), "-- name: UpsertGoogleDriveWorkspaceConnection")
	if upsertStart == -1 {
		t.Fatal("Google Drive connection queries are missing the durable workspace upsert")
	}
	upsertSource := string(connectionQueries)[upsertStart:]
	if nextQuery := strings.Index(upsertSource[1:], "-- name:"); nextQuery >= 0 {
		upsertSource = upsertSource[:nextQuery+1]
	}
	if !strings.Contains(upsertSource, "membership.role IN ('member', 'admin')") {
		t.Error("Google Drive callback persistence must recheck a non-guest workspace role")
	}
}

func TestGoogleDriveRevocationSagaRollbackPreservesDurableWork(t *testing.T) {
	t.Parallel()

	reverse, err := FS.ReadFile("000179_google_drive_revocation_saga.down.sql")
	if err != nil {
		t.Fatalf("read Google Drive revocation down migration: %v", err)
	}
	source := string(reverse)
	for _, contract := range []string{
		"EXISTS (SELECT 1 FROM public.google_drive_revocation_outbox)",
		"ERRCODE = '55000'",
		"DROP TRIGGER IF EXISTS google_drive_accounts_stage_revocation_before_delete",
		"DROP FUNCTION IF EXISTS public.stage_google_drive_account_revocation(",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("Google Drive revocation rollback is missing %q", contract)
		}
	}
}
