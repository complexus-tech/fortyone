package chatsessionsrepository

import (
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/google/uuid"
)

func TestMutationApprovalExecutionStateConversion(t *testing.T) {
	t.Parallel()

	completed, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		Status: string(chatsessions.MutationApprovalExecutionCompleted),
		Output: json.RawMessage(`{"success":true}`),
	})
	if err != nil {
		t.Fatalf("convert completed execution: %v", err)
	}
	if completed.State != chatsessions.MutationApprovalExecutionCompleted {
		t.Fatalf("state = %q, want completed", completed.State)
	}
	if string(completed.Output) != `{"success":true}` {
		t.Fatalf("output = %s", completed.Output)
	}

	executingLeaseExpiresAt := time.Now().Add(time.Minute)
	inProgress, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		LeaseExpiresAt: sql.NullTime{Time: executingLeaseExpiresAt, Valid: true},
		Status:         "executing",
	})
	if err != nil {
		t.Fatalf("convert in-progress execution: %v", err)
	}
	if inProgress.State != chatsessions.MutationApprovalExecutionExecuting || inProgress.Output != nil || inProgress.LeaseExpiresAt == nil {
		t.Fatalf("unexpected in-progress execution: %#v", inProgress)
	}

	failureCode := "execution_lease_expired"
	failed, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		FailureCode: sql.NullString{String: failureCode, Valid: true},
		Status:      "failed_uncertain",
	})
	if err != nil {
		t.Fatalf("convert failed execution: %v", err)
	}
	if failed.State != chatsessions.MutationApprovalExecutionFailed || failed.FailureCode != failureCode {
		t.Fatalf("unexpected failed execution: %#v", failed)
	}

	leaseToken := uuid.New()
	leaseExpiresAt := time.Now().Add(time.Minute)
	claimed, err := toClaimedMutationApprovalExecution(dbMutationApprovalExecution{
		LeaseExpiresAt: sql.NullTime{Time: leaseExpiresAt, Valid: true},
		LeaseToken:     uuid.NullUUID{UUID: leaseToken, Valid: true},
		Status:         "ready",
	})
	if err != nil {
		t.Fatalf("convert claimed execution: %v", err)
	}
	if claimed.State != chatsessions.MutationApprovalExecutionClaimed || claimed.LeaseToken == nil || *claimed.LeaseToken != leaseToken {
		t.Fatalf("unexpected claimed execution: %#v", claimed)
	}

	retryClaimed, err := toClaimedMutationApprovalExecution(dbMutationApprovalExecution{
		LeaseExpiresAt: sql.NullTime{Time: leaseExpiresAt, Valid: true},
		LeaseToken:     uuid.NullUUID{UUID: leaseToken, Valid: true},
		Status:         "retry_ready",
	})
	if err != nil {
		t.Fatalf("convert safe retry claim: %v", err)
	}
	if retryClaimed.State != chatsessions.MutationApprovalExecutionClaimed || retryClaimed.LeaseToken == nil || *retryClaimed.LeaseToken != leaseToken {
		t.Fatalf("unexpected safe retry claim: %#v", retryClaimed)
	}

	retryPending, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		LeaseExpiresAt: sql.NullTime{Time: leaseExpiresAt, Valid: true},
		Status:         "retry_ready",
	})
	if err != nil {
		t.Fatalf("convert pending safe retry: %v", err)
	}
	if retryPending.State != chatsessions.MutationApprovalExecutionReady || retryPending.LeaseExpiresAt == nil {
		t.Fatalf("unexpected pending safe retry: %#v", retryPending)
	}
}

func TestMutationApprovalExecutionRejectsInvalidDurableState(t *testing.T) {
	t.Parallel()

	if _, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		Status: string(chatsessions.MutationApprovalExecutionCompleted),
		Output: json.RawMessage(`not-json`),
	}); err == nil {
		t.Fatal("completed execution with invalid JSON must fail closed")
	}
	if _, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		Status: "unknown",
	}); err == nil {
		t.Fatal("unknown durable execution state must fail closed")
	}
	if _, err := toClaimedMutationApprovalExecution(dbMutationApprovalExecution{
		Status: "ready",
	}); err == nil {
		t.Fatal("ready execution without a complete lease must fail closed")
	}
	if _, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		Status: "failed_uncertain",
	}); err == nil {
		t.Fatal("uncertain execution without a failure code must fail closed")
	}
	if _, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		Status: "ready",
	}); err == nil {
		t.Fatal("ready execution without a lease expiry must fail closed")
	}
	if _, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		Status: "executing",
	}); err == nil {
		t.Fatal("executing execution without a lease expiry must fail closed")
	}
	if _, err := toCoreMutationApprovalExecution(dbMutationApprovalExecution{
		Status: "retry_ready",
	}); err == nil {
		t.Fatal("prepared retry without a lease expiry must fail closed at the API boundary")
	}
}

func TestMutationApprovalQueriesPreserveOwnershipAndAtomicClaimContracts(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("mutation_approvals.go")
	if err != nil {
		t.Fatalf("read mutation approval repository: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")

	for _, contract := range []string{
		"insert into chat_mutation_approval_executions",
		"select session.id, session.user_id, session.workspace_id, $4, $5, 'ready', $6",
		"session.user_id = $2",
		"session.workspace_id = $3",
		"session.deleted_at is null",
		"on conflict do nothing",
		"execution.fingerprint = $5",
		"execution.lease_token = $6",
		"execution.status in ('ready', 'retry_ready')",
		"execution.status = 'executing'",
		"execution.status = 'retry_ready'",
		"retry_requires_original_approval",
		"tool_call_id = $4",
		"execution.session_id <> $1 or execution.tool_call_id <> $4",
		"with terminalized as",
		"set status = 'completed'",
		"cross join terminalized",
		"destination_session.id = $1",
		"set status = 'completed'",
		"set status = 'failed_uncertain'",
		"last_reconciliation_resolution = $6",
		"last_reconciliation_evidence = cast($7 as jsonb)",
		"reconciliation_count = reconciliation_count + 1",
		"execution.status = 'failed_uncertain'",
		"execution.status in ('ready', 'retry_ready', 'executing', 'failed_uncertain')",
		"for update of execution",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("mutation approval repository is missing contract %q", contract)
		}
	}
}

func TestChatSessionTitleUpdateExcludesDeletedSessions(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read chat session commands: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")
	updateStart := strings.Index(source, "func (r *repo) updatesession")
	updateEnd := strings.Index(source[updateStart+1:], "func (r *repo)")
	if updateStart < 0 || updateEnd < 0 {
		t.Fatal("could not locate UpdateSession implementation")
	}
	updateSource := source[updateStart : updateStart+1+updateEnd]
	if !strings.Contains(updateSource, "and deleted_at is null") {
		t.Fatal("UpdateSession must not mutate a soft-deleted chat session")
	}
}

func TestChatSessionCreationIsOwnerScopedAndIdempotent(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read chat session commands: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")
	createStart := strings.Index(source, "func (r *repo) createsessionwithmessages")
	createEnd := strings.Index(source[createStart+1:], "func (r *repo)")
	if createStart < 0 || createEnd < 0 {
		t.Fatal("could not locate CreateSessionWithMessages implementation")
	}
	createSource := source[createStart : createStart+1+createEnd]

	for _, contract := range []string{
		"on conflict (id) do update",
		"chat_sessions.user_id = excluded.user_id",
		"chat_sessions.workspace_id = excluded.workspace_id",
		"chat_sessions.deleted_at is null",
		"on conflict (session_id) do nothing",
	} {
		if !strings.Contains(createSource, contract) {
			t.Errorf("idempotent chat persistence is missing contract %q", contract)
		}
	}
}

func TestLatestAssistantMessageQueryIsOwnerScopedAndReturnsOneMessage(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("queries.go")
	if err != nil {
		t.Fatalf("read chat session queries: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")
	queryStart := strings.Index(source, "func (r *repo) getlatestassistantmessage")
	queryEnd := strings.Index(source[queryStart+1:], "func (r *repo)")
	if queryStart < 0 || queryEnd < 0 {
		t.Fatal("could not locate GetLatestAssistantMessage implementation")
	}
	querySource := source[queryStart : queryStart+1+queryEnd]

	for _, contract := range []string{
		"jsonb_array_length(coalesce(messages.messages, cast('[]' as jsonb))) - 1",
		"generate_series(",
		"and sessions.user_id = :user_id",
		"and sessions.workspace_id = :workspace_id",
		"and sessions.deleted_at is null",
		"->> 'role' = 'assistant'",
		"limit 1",
	} {
		if !strings.Contains(querySource, contract) {
			t.Errorf("latest assistant query is missing contract %q", contract)
		}
	}
}

func TestChatSessionNamedQueriesUseBinderSafeCasts(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("queries.go")
	if err != nil {
		t.Fatalf("read chat session queries: %v", err)
	}
	if strings.Contains(string(data), "::") {
		t.Fatal("named chat-session queries must use CAST(value AS type), not PostgreSQL double-colon casts")
	}
}
