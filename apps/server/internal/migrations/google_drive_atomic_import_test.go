package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestGoogleDriveAtomicImportMigrationPreservesIdempotentIdentity(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000180_google_drive_atomic_document_imports.up.sql")
	if err != nil {
		t.Fatalf("read atomic Google Drive import migration: %v", err)
	}
	source := string(forward)
	for _, contract := range []string{
		"google_drive_document_import_operations",
		"source_reference_id uuid NOT NULL",
		"document_id uuid NOT NULL UNIQUE",
		"UNIQUE (workspace_id, user_id, idempotency_key)",
		"attempt_generation uuid NOT NULL",
		"status IN ('pending', 'completed', 'failed')",
		"status = 'completed' AND completed_at IS NOT NULL AND error_code IS NULL",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("atomic Google Drive import migration is missing %q", contract)
		}
	}
	if strings.Contains(source, "document_id uuid NOT NULL UNIQUE REFERENCES") ||
		strings.Contains(source, "source_reference_id uuid NOT NULL REFERENCES") {
		t.Error("durable import identities must survive source and document deletion")
	}
}

func TestGoogleDriveAtomicImportQueriesFenceAndReauthorizeFinalization(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../modules/googledrive/repository/queries/imports.sql")
	if err != nil {
		t.Fatalf("read atomic Google Drive import queries: %v", err)
	}
	source := string(data)
	for _, contract := range []string{
		"ON CONFLICT (workspace_id, user_id, idempotency_key) DO NOTHING",
		"status = 'pending' AND updated_at <= sqlc.arg(stale_before)",
		"FOR UPDATE",
		"grant_record.verification_generation = sqlc.arg(grant_generation)",
		"GoogleDriveReferenceImportable",
		"CreateGoogleDriveImportedDocument",
		"CompleteGoogleDriveImportOperation",
		"attempt_generation = sqlc.arg(attempt_generation)",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("atomic Google Drive import query contract is missing %q", contract)
		}
	}
}

func TestGoogleDriveAtomicImportRollbackRefusesToDiscardOperations(t *testing.T) {
	t.Parallel()

	reverse, err := FS.ReadFile("000180_google_drive_atomic_document_imports.down.sql")
	if err != nil {
		t.Fatalf("read atomic Google Drive import down migration: %v", err)
	}
	source := string(reverse)
	if !strings.Contains(source, "EXISTS (SELECT 1 FROM public.google_drive_document_import_operations)") {
		t.Error("atomic Google Drive import rollback does not protect durable operations")
	}
	if !strings.Contains(source, "ERRCODE = '55000'") {
		t.Error("atomic Google Drive import rollback must fail with object_not_in_prerequisite_state")
	}
}
