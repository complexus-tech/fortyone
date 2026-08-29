//go:build integration

package mayarepository

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRealtimeVoiceReservationSerializesMonthlyQuota(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	workspaceID, userID := insertRealtimeOwner(t, ctx, postgres.Pool, true)

	const attempts = 3
	start := make(chan struct{})
	results := make(chan error, attempts)
	var waitGroup sync.WaitGroup
	for range attempts {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := repository.ReserveRealtimeVoiceSession(
				ctx,
				workspaceID,
				userID,
				maya.RealtimeMonthlyVoiceLimit,
				maya.RealtimeMaxSessionDuration,
			)
			results <- err
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	var reserved, limited int
	for err := range results {
		switch {
		case err == nil:
			reserved++
		case errors.Is(err, maya.ErrRealtimeMonthlyLimitExceeded):
			limited++
		default:
			t.Fatalf("reserve realtime voice session: %v", err)
		}
	}
	if reserved != 2 || limited != 1 {
		t.Fatalf("reservations = %d, limited = %d; want 2 and 1", reserved, limited)
	}

	var openSessions int
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT count(*)
		FROM maya_realtime_voice_sessions
		WHERE workspace_id = $1 AND ended_at IS NULL
	`, workspaceID).Scan(&openSessions); err != nil {
		t.Fatalf("count open realtime sessions: %v", err)
	}
	if openSessions != 2 {
		t.Fatalf("open realtime sessions = %d, want 2", openSessions)
	}
}

func TestRealtimeVoiceAndToolCallsAreTenantBoundAndIdempotent(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	workspaceID, userID := insertRealtimeOwner(t, ctx, postgres.Pool, true)
	otherWorkspaceID, otherUserID := insertRealtimeOwner(t, ctx, postgres.Pool, true)

	lease, err := repository.ReserveRealtimeVoiceSession(
		ctx,
		workspaceID,
		userID,
		maya.RealtimeMonthlyVoiceLimit,
		maya.RealtimeMaxSessionDuration,
	)
	if err != nil {
		t.Fatalf("reserve realtime voice session: %v", err)
	}
	active, err := repository.RealtimeVoiceSessionIsActive(
		ctx,
		workspaceID,
		userID,
		lease.SessionID,
		maya.RealtimeMaxSessionDuration,
	)
	if err != nil || !active {
		t.Fatalf("own realtime session active = %t, error = %v", active, err)
	}
	active, err = repository.RealtimeVoiceSessionIsActive(
		ctx,
		otherWorkspaceID,
		otherUserID,
		lease.SessionID,
		maya.RealtimeMaxSessionDuration,
	)
	if err != nil || active {
		t.Fatalf("cross-tenant realtime session active = %t, error = %v", active, err)
	}

	const requestHash = "exact-request-hash"
	claim, err := repository.ClaimRealtimeToolCall(
		ctx,
		lease.SessionID,
		"call-1",
		"get_context",
		requestHash,
	)
	if err != nil || !claim.Claimed {
		t.Fatalf("initial tool claim = %+v, error = %v", claim, err)
	}
	_, err = repository.ClaimRealtimeToolCall(
		ctx,
		lease.SessionID,
		"call-1",
		"get_context",
		requestHash,
	)
	if !errors.Is(err, maya.ErrRealtimeToolCallInProgress) {
		t.Fatalf("duplicate incomplete tool claim error = %v", err)
	}

	response := json.RawMessage(`{"success":true}`)
	if err := repository.CompleteRealtimeToolCall(ctx, lease.SessionID, "call-1", response); err != nil {
		t.Fatalf("complete realtime tool call: %v", err)
	}
	claim, err = repository.ClaimRealtimeToolCall(
		ctx,
		lease.SessionID,
		"call-1",
		"get_context",
		requestHash,
	)
	var replayedResponse map[string]bool
	decodeErr := json.Unmarshal(claim.Response, &replayedResponse)
	if err != nil || decodeErr != nil || claim.Claimed || !replayedResponse["success"] {
		t.Fatalf("replayed tool claim = %+v, error = %v", claim, err)
	}
	_, err = repository.ClaimRealtimeToolCall(
		ctx,
		lease.SessionID,
		"call-1",
		"get_context",
		"different-request-hash",
	)
	if !errors.Is(err, maya.ErrRealtimeToolCallConflict) {
		t.Fatalf("conflicting tool claim error = %v", err)
	}
	if err := repository.CompleteRealtimeToolCall(ctx, lease.SessionID, "call-1", response); !errors.Is(err, maya.ErrRealtimeToolCallCompletionConflict) {
		t.Fatalf("second tool completion error = %v", err)
	}

	if err := repository.EndRealtimeVoiceSession(
		ctx,
		workspaceID,
		userID,
		lease.SessionID,
		maya.RealtimeMaxSessionDuration,
	); err != nil {
		t.Fatalf("end realtime voice session: %v", err)
	}
	active, err = repository.RealtimeVoiceSessionIsActive(
		ctx,
		workspaceID,
		userID,
		lease.SessionID,
		maya.RealtimeMaxSessionDuration,
	)
	if err != nil || active {
		t.Fatalf("ended realtime session active = %t, error = %v", active, err)
	}
}

func TestRealtimeReservationRequiresCurrentMayaAccess(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	workspaceID, userID := insertRealtimeOwner(t, ctx, postgres.Pool, false)

	_, err := repository.ReserveRealtimeVoiceSession(
		ctx,
		workspaceID,
		userID,
		maya.RealtimeMonthlyVoiceLimit,
		maya.RealtimeMaxSessionDuration,
	)
	if !errors.Is(err, maya.ErrMayaAccessDenied) {
		t.Fatalf("reservation without Maya access error = %v", err)
	}
}

func TestRealtimeReservationRequiresCurrentWorkspaceMembership(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	workspaceID, userID := insertRealtimeOwner(t, ctx, postgres.Pool, true)
	if _, err := postgres.Pool.Exec(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID); err != nil {
		t.Fatalf("revoke realtime workspace membership: %v", err)
	}

	_, err := repository.ReserveRealtimeVoiceSession(
		ctx,
		workspaceID,
		userID,
		maya.RealtimeMonthlyVoiceLimit,
		maya.RealtimeMaxSessionDuration,
	)
	if !errors.Is(err, maya.ErrMayaAccessDenied) {
		t.Fatalf("reservation after membership revocation error = %v", err)
	}
}

func insertRealtimeOwner(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	hasAccess bool,
) (uuid.UUID, uuid.UUID) {
	t.Helper()
	suffix := uuid.NewString()
	userID := uuid.New()
	workspaceID := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Realtime owner')
	`, userID, "realtime-"+suffix, "realtime-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert realtime user: %v", err)
	}
	var trialEndsAt *time.Time
	if hasAccess {
		endsAt := time.Now().UTC().Add(24 * time.Hour)
		trialEndsAt = &endsAt
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, trial_ends_on, created_by)
		VALUES ($1, 'Realtime workspace', $2, $3, $4)
	`, workspaceID, "realtime-"+suffix, trialEndsAt, userID); err != nil {
		t.Fatalf("insert realtime workspace: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert realtime workspace membership: %v", err)
	}
	return workspaceID, userID
}
