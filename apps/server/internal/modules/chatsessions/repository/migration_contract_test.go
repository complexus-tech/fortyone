package chatsessionsrepository

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMutationApprovalMigrationEnforcesDurableExecutionContract(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	migrationPath := filepath.Join(
		filepath.Dir(currentFile),
		"..",
		"..",
		"..",
		"migrations",
		"000146_chat_mutation_approval_executions.up.sql",
	)
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read mutation approval migration: %v", err)
	}
	migration := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")

	for _, contract := range []string{
		"primary key (session_id, user_id, workspace_id, tool_call_id)",
		"foreign key (session_id) references public.chat_sessions(id) on delete cascade",
		"fingerprint ~ '^[0-9a-f]{64}$'",
		"status in ('in_progress', 'completed')",
		"status = 'in_progress' and output is null and completed_at is null",
		"status = 'completed' and output is not null and completed_at is not null",
		"on public.chat_mutation_approval_executions (session_id, tool_call_id)",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("mutation approval migration is missing contract %q", contract)
		}
	}
}

func TestMutationApprovalLeaseMigrationFailsClosedAfterExecutionStarts(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	migrationPath := filepath.Join(
		filepath.Dir(currentFile),
		"..",
		"..",
		"..",
		"migrations",
		"000147_chat_mutation_approval_execution_leases.up.sql",
	)
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read mutation approval lease migration: %v", err)
	}
	migration := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")

	for _, contract := range []string{
		"add column lease_token uuid",
		"add column lease_expires_at timestamptz",
		"status in ('ready', 'executing', 'completed', 'failed_uncertain')",
		"where status = 'in_progress'",
		"failure_code = 'legacy_in_progress'",
		"status = 'ready' and output is null",
		"status = 'executing' and output is null",
		"status = 'failed_uncertain' and output is null",
		"on public.chat_mutation_approval_executions ( session_id, user_id, workspace_id, fingerprint ) where status in ('ready', 'executing', 'failed_uncertain')",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("mutation approval lease migration is missing contract %q", contract)
		}
	}
}

func TestMutationApprovalReconciliationMigrationPreservesAuditableEvidence(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	migrationPath := filepath.Join(
		filepath.Dir(currentFile),
		"..",
		"..",
		"..",
		"migrations",
		"000148_chat_mutation_approval_reconciliation.up.sql",
	)
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read mutation approval reconciliation migration: %v", err)
	}
	migration := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")

	for _, contract := range []string{
		"add column last_reconciliation_resolution text",
		"add column last_reconciliation_evidence jsonb",
		"add column last_reconciled_at timestamptz",
		"add column reconciliation_count integer not null default 0",
		"'verified_completed'",
		"'verified_not_applied'",
		"jsonb_typeof(last_reconciliation_evidence) = 'object'",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("mutation approval reconciliation migration is missing contract %q", contract)
		}
	}
}

func TestChatMessageWriteReservationMigrationAddsOpaqueCASState(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	migrationPath := filepath.Join(
		filepath.Dir(currentFile),
		"..",
		"..",
		"..",
		"migrations",
		"000149_chat_message_write_reservations.up.sql",
	)
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read chat message write migration: %v", err)
	}
	migration := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")
	for _, contract := range []string{
		"add column write_generation bigint not null default 0",
		"add column write_token uuid",
		"add column write_operation text",
		"add column write_finalized_at timestamptz",
		"write_operation in ('append', 'regenerate', 'approval')",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("chat message write migration is missing contract %q", contract)
		}
	}
}

func TestMutationApprovalFingerprintQuarantineSpansUserChats(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	migrationPath := filepath.Join(
		filepath.Dir(currentFile),
		"..",
		"..",
		"..",
		"migrations",
		"000150_chat_mutation_approval_global_fingerprints.up.sql",
	)
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read cross-session mutation quarantine migration: %v", err)
	}
	migration := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")

	for _, contract := range []string{
		"group by user_id, workspace_id, fingerprint",
		"having count(*) > 1",
		"raise exception",
		"on public.chat_mutation_approval_executions ( user_id, workspace_id, fingerprint )",
		"where status in ('ready', 'executing', 'failed_uncertain')",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("cross-session mutation quarantine migration is missing contract %q", contract)
		}
	}
}

func TestMutationApprovalSafeRetryMigrationIsDistinctBoundedAndGlobal(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	migrationDirectory := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "migrations")
	upData, err := os.ReadFile(filepath.Join(migrationDirectory, "000151_chat_mutation_approval_safe_retry.up.sql"))
	if err != nil {
		t.Fatalf("read safe retry migration: %v", err)
	}
	up := strings.Join(strings.Fields(strings.ToLower(string(upData))), " ")
	for _, contract := range []string{
		"'retry_ready'",
		"'safe_retry_prepared'",
		"status = 'retry_ready' and output is null",
		"reconciliation_count = 1",
		"last_reconciliation_resolution = 'safe_retry_prepared'",
		"on public.chat_mutation_approval_executions ( user_id, workspace_id, fingerprint )",
		"where status in ('ready', 'retry_ready', 'executing', 'failed_uncertain')",
		"where status in ('ready', 'retry_ready', 'executing')",
	} {
		if !strings.Contains(up, contract) {
			t.Errorf("safe retry migration is missing contract %q", contract)
		}
	}

	downData, err := os.ReadFile(filepath.Join(migrationDirectory, "000151_chat_mutation_approval_safe_retry.down.sql"))
	if err != nil {
		t.Fatalf("read safe retry down migration: %v", err)
	}
	down := strings.Join(strings.Fields(strings.ToLower(string(downData))), " ")
	for _, contract := range []string{
		"where status = 'retry_ready' or last_reconciliation_resolution = 'safe_retry_prepared'",
		"raise exception",
		"status in ('ready', 'executing', 'completed', 'failed_uncertain')",
		"where status in ('ready', 'executing', 'failed_uncertain')",
	} {
		if !strings.Contains(down, contract) {
			t.Errorf("safe retry down migration is missing contract %q", contract)
		}
	}
}
