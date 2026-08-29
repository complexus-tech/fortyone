// Package messagingbudget contains provider-neutral abuse and cost controls for
// conversational messaging integrations.
package messagingbudget

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	DefaultUserCallLimit       int64 = 12
	DefaultWorkspaceCallLimit  int64 = 120
	DefaultCallLimitWindow           = time.Minute
	DefaultEventIdempotencyTTL       = 24 * time.Hour
)

var admitCallScript = `
local existing = redis.call("GET", KEYS[1])
local user_count = tonumber(redis.call("GET", KEYS[2]) or "0")
local workspace_count = tonumber(redis.call("GET", KEYS[3]) or "0")

if existing == "allowed" then
  return {1, 1, user_count, workspace_count, 0, 0}
end
if existing == "user" then
  return {0, 1, user_count, workspace_count, 1, redis.call("PTTL", KEYS[2])}
end
if existing == "workspace" then
  return {0, 1, user_count, workspace_count, 2, redis.call("PTTL", KEYS[3])}
end

if user_count >= tonumber(ARGV[1]) then
  local retry_after = redis.call("PTTL", KEYS[2])
  redis.call("SET", KEYS[1], "user", "PX", ARGV[4])
  return {0, 0, user_count, workspace_count, 1, retry_after}
end

if workspace_count >= tonumber(ARGV[2]) then
  local retry_after = redis.call("PTTL", KEYS[3])
  redis.call("SET", KEYS[1], "workspace", "PX", ARGV[4])
  return {0, 0, user_count, workspace_count, 2, retry_after}
end

redis.call("SET", KEYS[1], "allowed", "PX", ARGV[4])
user_count = redis.call("INCR", KEYS[2])
workspace_count = redis.call("INCR", KEYS[3])

if redis.call("PTTL", KEYS[2]) < 0 then
  redis.call("PEXPIRE", KEYS[2], ARGV[3])
end
if redis.call("PTTL", KEYS[3]) < 0 then
  redis.call("PEXPIRE", KEYS[3], ARGV[3])
end

return {1, 0, user_count, workspace_count, 0, 0}
`

// RedisEvaler is the narrow Redis contract needed for an atomic multi-scope
// call admission. *redis.Client implements it directly.
type RedisEvaler interface {
	Eval(ctx context.Context, script string, keys []string, args ...any) *redis.Cmd
}

var _ RedisEvaler = (*redis.Client)(nil)

type CallLimiterConfig struct {
	UserLimit      int64
	WorkspaceLimit int64
	Window         time.Duration
	IdempotencyTTL time.Duration
}

type AdmissionInput struct {
	Provider            string
	WorkspaceID         uuid.UUID
	UserID              uuid.UUID
	ExternalWorkspaceID string
	ExternalEventID     string
}

type AdmissionDecision struct {
	Allowed        bool
	Duplicate      bool
	LimitedScope   string
	RetryAfter     time.Duration
	UserCount      int64
	WorkspaceCount int64
}

type RedisCallLimiter struct {
	redis          RedisEvaler
	userLimit      int64
	workspaceLimit int64
	window         time.Duration
	idempotencyTTL time.Duration
}

func NewRedisCallLimiter(client RedisEvaler, config CallLimiterConfig) (*RedisCallLimiter, error) {
	if client == nil {
		return nil, errors.New("messaging assistant call limiter Redis client is required")
	}
	if config.UserLimit == 0 {
		config.UserLimit = DefaultUserCallLimit
	}
	if config.WorkspaceLimit == 0 {
		config.WorkspaceLimit = DefaultWorkspaceCallLimit
	}
	if config.Window == 0 {
		config.Window = DefaultCallLimitWindow
	}
	if config.IdempotencyTTL == 0 {
		config.IdempotencyTTL = DefaultEventIdempotencyTTL
	}
	if config.UserLimit < 1 || config.WorkspaceLimit < 1 {
		return nil, errors.New("messaging assistant call limits must be positive")
	}
	if config.Window < time.Second {
		return nil, errors.New("messaging assistant call limit window must be at least one second")
	}
	if config.IdempotencyTTL < config.Window {
		return nil, errors.New("messaging assistant event idempotency TTL must not be shorter than the call limit window")
	}

	return &RedisCallLimiter{
		redis:          client,
		userLimit:      config.UserLimit,
		workspaceLimit: config.WorkspaceLimit,
		window:         config.Window,
		idempotencyTTL: config.IdempotencyTTL,
	}, nil
}

// Admit atomically applies the global user and workspace windows exactly once
// per provider event. All keys share one Redis Cluster hash tag so the Lua
// operation remains single-slot if Redis is clustered. Counters intentionally
// span providers so adding Teams or WhatsApp cannot bypass the Slack budget.
func (l *RedisCallLimiter) Admit(ctx context.Context, input AdmissionInput) (AdmissionDecision, error) {
	if l == nil || l.redis == nil {
		return AdmissionDecision{}, errors.New("messaging assistant call limiter is not configured")
	}
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	if !validProvider(provider) {
		return AdmissionDecision{}, errors.New("messaging assistant call limiter provider is invalid")
	}
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil {
		return AdmissionDecision{}, errors.New("messaging assistant call limiter workspace and user are required")
	}
	externalWorkspaceID := strings.TrimSpace(input.ExternalWorkspaceID)
	externalEventID := strings.TrimSpace(input.ExternalEventID)
	if externalWorkspaceID == "" || externalEventID == "" {
		return AdmissionDecision{}, errors.New("messaging assistant call limiter external workspace and event are required")
	}

	keys := callLimitKeys(provider, input.WorkspaceID, input.UserID, externalWorkspaceID, externalEventID)
	result, err := l.redis.Eval(
		ctx,
		admitCallScript,
		keys,
		l.userLimit,
		l.workspaceLimit,
		l.window.Milliseconds(),
		l.idempotencyTTL.Milliseconds(),
	).Result()
	if err != nil {
		return AdmissionDecision{}, fmt.Errorf("admit messaging assistant call: %w", err)
	}
	values, ok := result.([]any)
	if !ok || len(values) != 6 {
		return AdmissionDecision{}, fmt.Errorf("admit messaging assistant call: unexpected Redis response %T", result)
	}
	numbers := make([]int64, len(values))
	for index, value := range values {
		numbers[index], ok = redisInteger(value)
		if !ok {
			return AdmissionDecision{}, fmt.Errorf("admit messaging assistant call: Redis response item %d has type %T", index, value)
		}
	}

	decision := AdmissionDecision{
		Allowed:        numbers[0] == 1,
		Duplicate:      numbers[1] == 1,
		UserCount:      numbers[2],
		WorkspaceCount: numbers[3],
	}
	switch numbers[4] {
	case 0:
	case 1:
		decision.LimitedScope = "user"
	case 2:
		decision.LimitedScope = "workspace"
	default:
		return AdmissionDecision{}, fmt.Errorf("admit messaging assistant call: unknown limited scope %d", numbers[4])
	}
	if !decision.Allowed {
		retryMilliseconds := numbers[5]
		if retryMilliseconds < 1 {
			retryMilliseconds = l.window.Milliseconds()
		}
		decision.RetryAfter = time.Duration(retryMilliseconds) * time.Millisecond
	}
	return decision, nil
}

func callLimitKeys(provider string, workspaceID, userID uuid.UUID, externalWorkspaceID, externalEventID string) []string {
	digest := sha256.Sum256([]byte(provider + "\x00" + externalWorkspaceID + "\x00" + externalEventID))
	prefix := "messaging:assistant:{messaging-assistant}"
	return []string{
		prefix + ":event:" + hex.EncodeToString(digest[:]),
		prefix + ":user:" + userID.String(),
		prefix + ":workspace:" + workspaceID.String(),
	}
}

func validProvider(provider string) bool {
	if provider == "" {
		return false
	}
	for _, value := range provider {
		if !unicode.IsLower(value) && !unicode.IsDigit(value) && value != '-' && value != '_' {
			return false
		}
	}
	return true
}

func redisInteger(value any) (int64, bool) {
	switch number := value.(type) {
	case int64:
		return number, true
	case int:
		return int64(number), true
	case string:
		parsed, err := strconv.ParseInt(number, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}
