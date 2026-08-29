package messagingbudget

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type redisEvalStub struct {
	result any
	err    error
	keys   []string
	args   []any
}

func (s *redisEvalStub) Eval(_ context.Context, _ string, keys []string, args ...any) *redis.Cmd {
	s.keys = append([]string(nil), keys...)
	s.args = append([]any(nil), args...)
	return redis.NewCmdResult(s.result, s.err)
}

func TestRedisCallLimiterAdmitsAllScopesAtomically(t *testing.T) {
	t.Parallel()

	redisStub := &redisEvalStub{result: []any{int64(1), int64(0), int64(4), int64(18), int64(0), int64(0)}}
	limiter, err := NewRedisCallLimiter(redisStub, CallLimiterConfig{})
	require.NoError(t, err)

	decision, err := limiter.Admit(context.Background(), validAdmissionInput())

	require.NoError(t, err)
	require.True(t, decision.Allowed)
	require.False(t, decision.Duplicate)
	require.Equal(t, int64(4), decision.UserCount)
	require.Equal(t, int64(18), decision.WorkspaceCount)
	require.Len(t, redisStub.keys, 3)
	for _, key := range redisStub.keys {
		require.Contains(t, key, "{messaging-assistant}")
	}
	require.Contains(t, redisStub.keys[1], ":user:22222222-2222-2222-2222-222222222222")
	require.Contains(t, redisStub.keys[2], ":workspace:11111111-1111-1111-1111-111111111111")
	require.Equal(t, []any{
		DefaultUserCallLimit,
		DefaultWorkspaceCallLimit,
		DefaultCallLimitWindow.Milliseconds(),
		DefaultEventIdempotencyTTL.Milliseconds(),
	}, redisStub.args)
}

func TestRedisCallLimiterReturnsIdempotentDuplicate(t *testing.T) {
	t.Parallel()

	redisStub := &redisEvalStub{result: []any{int64(1), int64(1), int64(12), int64(120), int64(0), int64(0)}}
	limiter, err := NewRedisCallLimiter(redisStub, CallLimiterConfig{})
	require.NoError(t, err)

	decision, err := limiter.Admit(context.Background(), validAdmissionInput())

	require.NoError(t, err)
	require.True(t, decision.Allowed)
	require.True(t, decision.Duplicate)
}

func TestRedisCallLimiterReturnsIdempotentDeniedDuplicate(t *testing.T) {
	t.Parallel()

	redisStub := &redisEvalStub{result: []any{int64(0), int64(1), int64(12), int64(80), int64(1), int64(-2)}}
	limiter, err := NewRedisCallLimiter(redisStub, CallLimiterConfig{})
	require.NoError(t, err)

	decision, err := limiter.Admit(context.Background(), validAdmissionInput())

	require.NoError(t, err)
	require.False(t, decision.Allowed)
	require.True(t, decision.Duplicate)
	require.Equal(t, "user", decision.LimitedScope)
	require.Equal(t, DefaultCallLimitWindow, decision.RetryAfter)
}

func TestRedisCallLimiterReturnsScopedDenialAndTTL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		scope int64
		want  string
	}{
		{name: "user", scope: 1, want: "user"},
		{name: "workspace", scope: 2, want: "workspace"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			redisStub := &redisEvalStub{result: []any{int64(0), int64(0), int64(12), int64(120), test.scope, int64(17_500)}}
			limiter, err := NewRedisCallLimiter(redisStub, CallLimiterConfig{})
			require.NoError(t, err)

			decision, err := limiter.Admit(context.Background(), validAdmissionInput())

			require.NoError(t, err)
			require.False(t, decision.Allowed)
			require.Equal(t, test.want, decision.LimitedScope)
			require.Equal(t, 17_500*time.Millisecond, decision.RetryAfter)
		})
	}
}

func TestRedisCallLimiterFailsClosedOnRedisError(t *testing.T) {
	t.Parallel()

	limiter, err := NewRedisCallLimiter(&redisEvalStub{err: errors.New("redis unavailable")}, CallLimiterConfig{})
	require.NoError(t, err)

	decision, err := limiter.Admit(context.Background(), validAdmissionInput())

	require.ErrorContains(t, err, "redis unavailable")
	require.False(t, decision.Allowed)
}

func TestRedisCallLimiterHashesUntrustedProviderIdentifiers(t *testing.T) {
	t.Parallel()

	redisStub := &redisEvalStub{result: []any{int64(1), int64(0), int64(1), int64(1), int64(0), int64(0)}}
	limiter, err := NewRedisCallLimiter(redisStub, CallLimiterConfig{})
	require.NoError(t, err)
	input := validAdmissionInput()
	input.ExternalWorkspaceID = strings.Repeat("team:", 100)
	input.ExternalEventID = strings.Repeat("event:", 100)

	_, err = limiter.Admit(context.Background(), input)

	require.NoError(t, err)
	require.Less(t, len(redisStub.keys[0]), 200)
	require.NotContains(t, redisStub.keys[0], input.ExternalEventID)
}

func TestCallLimitCountersSpanMessagingProviders(t *testing.T) {
	t.Parallel()

	input := validAdmissionInput()
	slackKeys := callLimitKeys("slack", input.WorkspaceID, input.UserID, "T1", "Ev1")
	teamsKeys := callLimitKeys("teams", input.WorkspaceID, input.UserID, "tenant-1", "event-1")

	require.NotEqual(t, slackKeys[0], teamsKeys[0], "provider events need distinct idempotency keys")
	require.Equal(t, slackKeys[1], teamsKeys[1], "user calls must be limited across providers")
	require.Equal(t, slackKeys[2], teamsKeys[2], "workspace calls must be limited across providers")
}

func validAdmissionInput() AdmissionInput {
	return AdmissionInput{
		Provider:            "slack",
		WorkspaceID:         uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		UserID:              uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ExternalWorkspaceID: "T1",
		ExternalEventID:     "Ev1",
	}
}
