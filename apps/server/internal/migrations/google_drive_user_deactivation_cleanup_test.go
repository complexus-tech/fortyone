package migrations

import (
	"strings"
	"testing"
)

func TestGoogleDriveUserDeactivationAtomicallyStagesExistingLifecycleCleanup(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000181_google_drive_user_deactivation_cleanup.up.sql")
	if err != nil {
		t.Fatalf("read Google Drive user-deactivation migration: %v", err)
	}
	source := string(forward)
	for _, contract := range []string{
		"CREATE FUNCTION public.cleanup_google_drive_on_user_deactivation()",
		"CREATE TRIGGER users_cleanup_google_drive_on_deactivation",
		"AFTER UPDATE OF is_active ON public.users",
		"WHEN (OLD.is_active = TRUE AND NEW.is_active = FALSE)",
		"pg_try_advisory_xact_lock",
		"'google-drive-provider-user:' || CAST(NEW.user_id AS text)",
		"'google-drive:' || CAST(NEW.user_id AS text)",
		"USING ERRCODE = '40001'",
		"DELETE FROM public.google_drive_workspace_connections AS connection",
		"WHERE connection.user_id = NEW.user_id",
		"account.is_active = FALSE",
		"ORDER BY connection.user_id",
		"pg_advisory_xact_lock",
		"WHERE account.user_id = deactivated_user_id",
		"FOR UPDATE",
		"IF NOT FOUND THEN",
		"CONTINUE",
		"WHERE connection.user_id = deactivated_user_id",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("Google Drive user-deactivation migration is missing %q", contract)
		}
	}

	providerLock := strings.Index(source, "'google-drive-provider-user:' || CAST(NEW.user_id AS text)")
	accountLock := strings.Index(source, "'google-drive:' || CAST(NEW.user_id AS text)")
	connectionDelete := strings.Index(source, "DELETE FROM public.google_drive_workspace_connections AS connection")
	if providerLock == -1 || accountLock == -1 || connectionDelete == -1 ||
		providerLock >= accountLock || accountLock >= connectionDelete {
		t.Error("user deactivation must fence provider and local lifecycle work before deleting Drive bindings")
	}
	if strings.Contains(source, "UPDATE public.users") {
		t.Error("the user-deactivation trigger must not update users and recurse")
	}

	backfillUserLock := strings.Index(source, "WHERE account.user_id = deactivated_user_id")
	backfillDelete := strings.LastIndex(source, "WHERE connection.user_id = deactivated_user_id")
	if backfillUserLock == -1 || backfillDelete == -1 || backfillUserLock >= backfillDelete {
		t.Error("the backfill must lock and recheck the inactive user before deleting stale Drive bindings")
	}
}

func TestGoogleDriveUserDeactivationRollbackDoesNotRestoreCredentials(t *testing.T) {
	t.Parallel()

	reverse, err := FS.ReadFile("000181_google_drive_user_deactivation_cleanup.down.sql")
	if err != nil {
		t.Fatalf("read Google Drive user-deactivation rollback: %v", err)
	}
	source := string(reverse)
	for _, contract := range []string{
		"DROP TRIGGER IF EXISTS users_cleanup_google_drive_on_deactivation",
		"DROP FUNCTION IF EXISTS public.cleanup_google_drive_on_user_deactivation()",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("Google Drive user-deactivation rollback is missing %q", contract)
		}
	}
	if strings.Contains(source, "INSERT INTO") || strings.Contains(source, "UPDATE public.google_drive_accounts") {
		t.Error("rollback must not recreate bindings or restore scrubbed Google credentials")
	}
}
