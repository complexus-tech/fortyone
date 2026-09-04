package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestGoogleDriveMigrationEstablishesTenantAndCredentialBoundaries(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000178_google_drive_integration.up.sql")
	if err != nil {
		t.Fatalf("read Google Drive migration: %v", err)
	}
	source := string(forward)
	for _, contract := range []string{
		"google_drive_accounts_active_identity",
		"UNIQUE (account_id, user_id)",
		"REFERENCES public.workspace_members(workspace_id, user_id)",
		"FOREIGN KEY (account_id, user_id)",
		"FOREIGN KEY (file_id, workspace_id)",
		"target_type IN ('story', 'objective', 'document', 'comment')",
		"google_drive_create_operations",
		"google_drive_document_imports",
		"reference_id uuid REFERENCES public.google_drive_file_references(reference_id) ON DELETE SET NULL",
		"CREATE FUNCTION public.lock_google_drive_user_lifecycle_on_delete()",
		"BEFORE DELETE ON public.workspace_members",
		"google_drive_workspace_connections_revoke_orphaned_account",
		"credential_payload = ''",
		"google_subject = ''",
		"email = ''",
		"scopes = '{}'::text[]",
		"CREATE FUNCTION public.lock_google_drive_file_on_reference_delete()",
		"google_drive_file_references_delete_orphaned_file",
		"DELETE FROM public.google_drive_files AS file",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("Google Drive migration is missing %q", contract)
		}
	}

	for _, serializedLifecycle := range []string{
		"pg_advisory_xact_lock",
		"'google-drive:' || CAST(OLD.user_id AS text)",
		"FROM public.google_drive_accounts AS account\n        WHERE account.account_id = OLD.account_id\n        FOR UPDATE",
		"FROM public.google_drive_files AS file\n    WHERE file.file_id = OLD.file_id\n    FOR UPDATE",
	} {
		if !strings.Contains(source, serializedLifecycle) {
			t.Errorf("Google Drive migration is missing serialized lifecycle contract %q", serializedLifecycle)
		}
	}
}

func TestGoogleDriveQueriesSeparateReadAndMutationAccess(t *testing.T) {
	t.Parallel()

	queries, err := readGoogleDriveQueries()
	if err != nil {
		t.Fatalf("read Google Drive queries: %v", err)
	}
	for _, contract := range []string{
		"GoogleDriveTargetAccessible",
		"GoogleDriveTargetMutable",
		"workspace_member.role IN ('member', 'admin', 'guest')",
		"WHEN 'objective' THEN EXISTS",
		"workspace_member.role IN ('member', 'admin')",
		"document_member.role = 'editor'",
		"RevalidateGoogleDriveFileReference",
		"DeleteGoogleDriveFileGrantForActor",
		"MarkGoogleDriveFileUnavailable",
		"unavailable_at = NULL",
		"sqlc.arg(grant_generation)",
		"verification_generation = EXCLUDED.verification_generation",
		"grant_record.verification_generation = sqlc.arg(grant_generation)",
		"DeleteOrphanGoogleDriveFile",
		"ClaimGoogleDriveOperation",
		"updated_at = sqlc.arg(previous_updated_at)",
		"NOT EXISTS (\n      SELECT 1\n      FROM public.google_drive_file_references",
	} {
		if !strings.Contains(queries, contract) {
			t.Errorf("Google Drive query contract is missing %q", contract)
		}
	}
	objectivePolicyStart := strings.Index(queries, "WHEN 'objective' THEN EXISTS")
	if objectivePolicyStart == -1 {
		t.Fatal("Google Drive query contract is missing objective authorization")
	}
	remainingPolicies := queries[objectivePolicyStart:]
	objectivePolicyEnd := strings.Index(remainingPolicies, "WHEN 'document' THEN EXISTS")
	if objectivePolicyEnd == -1 {
		t.Fatal("Google Drive query contract has an incomplete objective authorization branch")
	}
	objectivePolicy := remainingPolicies[:objectivePolicyEnd]
	if strings.Contains(objectivePolicy, "'guest'") || strings.Contains(objectivePolicy, "objective.is_private") {
		t.Error("Google Drive objective authorization must mirror canonical objective visibility")
	}

	connectionQueries, err := os.ReadFile("../modules/googledrive/repository/queries/connections.sql")
	if err != nil {
		t.Fatalf("read Google Drive connection queries: %v", err)
	}
	for _, contract := range []string{
		"LockGoogleDriveUserLifecycle",
		"pg_advisory_xact_lock",
		"'google-drive:'",
		"WHERE google_drive_workspace_connections.account_id = EXCLUDED.account_id",
		"DeleteExpiredGoogleDriveOAuthStates",
		"DELETE FROM public.google_drive_oauth_states",
		"ConsumeGoogleDriveOAuthState :one\nDELETE FROM public.google_drive_oauth_states",
	} {
		if !strings.Contains(string(connectionQueries), contract) {
			t.Errorf("Google Drive connection query contract is missing %q", contract)
		}
	}
}

func TestGoogleDriveRollbackRefusesToDestroyDurableState(t *testing.T) {
	t.Parallel()

	reverse, err := FS.ReadFile("000178_google_drive_integration.down.sql")
	if err != nil {
		t.Fatalf("read Google Drive down migration: %v", err)
	}
	source := string(reverse)
	for _, durableTable := range []string{
		"google_drive_accounts",
		"google_drive_files",
		"google_drive_file_references",
		"google_drive_create_operations",
		"google_drive_document_imports",
	} {
		if !strings.Contains(source, "EXISTS (SELECT 1 FROM public."+durableTable+")") {
			t.Errorf("Google Drive rollback guard does not protect %s", durableTable)
		}
	}
	if !strings.Contains(source, "ERRCODE = '55000'") {
		t.Error("Google Drive rollback guard must use object_not_in_prerequisite_state")
	}
	for _, cleanup := range []string{
		"DROP TRIGGER IF EXISTS workspace_members_lock_google_drive_lifecycle",
		"DROP TRIGGER IF EXISTS google_drive_workspace_connections_revoke_orphaned_account",
		"DROP FUNCTION IF EXISTS public.lock_google_drive_user_lifecycle_on_delete()",
		"DROP TRIGGER IF EXISTS google_drive_file_references_delete_orphaned_file",
		"DROP FUNCTION IF EXISTS public.delete_orphaned_google_drive_file()",
	} {
		if !strings.Contains(source, cleanup) {
			t.Errorf("Google Drive rollback is missing lifecycle cleanup %q", cleanup)
		}
	}
}

func readGoogleDriveQueries() (string, error) {
	// Migration tests run with the server package working directory. Keeping
	// this as a narrow relative read avoids duplicating SQL authorization logic
	// in test fixtures while still guarding accidental policy collapse.
	data, err := os.ReadFile("../modules/googledrive/repository/queries/files.sql")
	return string(data), err
}
