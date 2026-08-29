package testkit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// TestRedisURLEnv is the single Redis connection contract used by
	// integration tests. It must identify a disposable, non-production Redis
	// database that tests are permitted to mutate within their owned namespace.
	TestRedisURLEnv = "TEST_REDIS_URL"

	testRedisNamespacePrefix = "fortyone:test:"
	testRedisNamespaceBytes  = 16
	testRedisPoolSize        = 8
	testRedisScanBatch       = 128
	testRedisSetupTimeout    = 10 * time.Second
	testRedisCleanupTimeout  = 10 * time.Second
)

// Redis is a test-owned namespace on a shared Redis database. Callers must
// derive every key through Key so parallel tests cannot share mutable state.
// The client is intentionally bounded and is closed automatically by the test.
type Redis struct {
	Client    *redis.Client
	Namespace string
}

// Key qualifies a logical key with this test's cryptographically random
// namespace. The namespace is an opaque ownership boundary, not domain data.
func (r *Redis) Key(logicalKey string) string {
	return r.Namespace + logicalKey
}

// NewRedis connects to the required Redis integration service, verifies it is
// reachable, and registers namespace-only cleanup with the test. Integration
// tests deliberately fail when TEST_REDIS_URL or Redis is unavailable; they
// must never silently degrade into skipped coverage.
func NewRedis(t testing.TB) *Redis {
	t.Helper()

	options, err := parseTestRedisURL(os.Getenv(TestRedisURLEnv))
	if err != nil {
		t.Fatal(err)
	}
	namespace, err := newTestRedisNamespace()
	if err != nil {
		t.Fatalf("generate isolated Redis namespace: %v", err)
	}

	client := redis.NewClient(options)
	connected := false
	t.Cleanup(func() {
		var cleanupErr error
		if connected {
			cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), testRedisCleanupTimeout)
			cleanupErr = deleteRedisNamespace(cleanupCtx, client, namespace)
			cancelCleanup()
		}

		closeErr := client.Close()
		if cleanupErr != nil {
			t.Errorf("delete isolated Redis namespace: %v", cleanupErr)
		}
		if closeErr != nil {
			t.Errorf("close Redis integration client: %v", closeErr)
		}
	})

	setupCtx, cancelSetup := context.WithTimeout(t.Context(), testRedisSetupTimeout)
	defer cancelSetup()
	if err := client.Ping(setupCtx).Err(); err != nil {
		// Redis errors are intentionally omitted here. This keeps credentials out
		// of test output even if a future client error starts echoing its URL.
		t.Fatal("ping Redis integration service: connection failed; verify TEST_REDIS_URL and service health")
	}
	connected = true

	return &Redis{
		Client:    client,
		Namespace: namespace,
	}
}

func parseTestRedisURL(raw string) (*redis.Options, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf(
			"%s is required when integration tests are enabled; configure a disposable, non-production Redis database",
			TestRedisURLEnv,
		)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		// Neither the parser error nor the raw value is returned because both can
		// contain the complete URL, including its password.
		return nil, fmt.Errorf("%s must be a valid Redis URL", TestRedisURLEnv)
	}
	if parsed.Scheme != "redis" && parsed.Scheme != "rediss" {
		return nil, fmt.Errorf("%s must use the redis or rediss scheme", TestRedisURLEnv)
	}
	if parsed.Hostname() == "" || parsed.Port() == "" {
		return nil, fmt.Errorf("%s must include an explicit host and port", TestRedisURLEnv)
	}
	if parsed.Fragment != "" {
		return nil, fmt.Errorf("%s must not include a URL fragment", TestRedisURLEnv)
	}
	if err := validateRedisDatabasePath(parsed.Path); err != nil {
		return nil, err
	}

	options, err := redis.ParseURL(raw)
	if err != nil {
		// redis.ParseURL errors may include URL-derived values. Keep the public
		// error deliberately generic and credential-free.
		return nil, fmt.Errorf("%s contains unsupported Redis connection settings", TestRedisURLEnv)
	}
	options.MaxRetries = 1
	options.DialTimeout = 3 * time.Second
	options.ReadTimeout = 2 * time.Second
	options.WriteTimeout = 2 * time.Second
	options.ContextTimeoutEnabled = true
	options.PoolFIFO = true
	options.PoolSize = testRedisPoolSize
	options.PoolTimeout = 2 * time.Second
	options.MinIdleConns = 0
	options.MaxIdleConns = 2
	options.MaxActiveConns = testRedisPoolSize
	options.ConnMaxIdleTime = 30 * time.Second
	options.ConnMaxLifetime = 2 * time.Minute

	return options, nil
}

func validateRedisDatabasePath(path string) error {
	if path == "" {
		return nil
	}
	if !strings.HasPrefix(path, "/") || path == "/" || strings.Contains(strings.TrimPrefix(path, "/"), "/") {
		return fmt.Errorf("%s must include at most one numeric Redis database", TestRedisURLEnv)
	}

	if _, err := strconv.ParseUint(strings.TrimPrefix(path, "/"), 10, 31); err != nil {
		return fmt.Errorf("%s Redis database must be a non-negative integer", TestRedisURLEnv)
	}
	return nil
}

func newTestRedisNamespace() (string, error) {
	random := make([]byte, testRedisNamespaceBytes)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return testRedisNamespacePrefix + hex.EncodeToString(random) + ":", nil
}

func deleteRedisNamespace(ctx context.Context, client *redis.Client, namespace string) error {
	if !isTestRedisNamespace(namespace) {
		return fmt.Errorf("refusing to delete keys outside an owned Redis test namespace")
	}

	var cursor uint64
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, namespace+"*", testRedisScanBatch).Result()
		if err != nil {
			return fmt.Errorf("scan owned Redis namespace: %w", err)
		}
		if len(keys) > 0 {
			if err := client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("delete owned Redis namespace keys: %w", err)
			}
		}
		if nextCursor == 0 {
			return nil
		}
		cursor = nextCursor
	}
}

func isTestRedisNamespace(namespace string) bool {
	if !strings.HasPrefix(namespace, testRedisNamespacePrefix) || !strings.HasSuffix(namespace, ":") {
		return false
	}
	hexToken := strings.TrimSuffix(strings.TrimPrefix(namespace, testRedisNamespacePrefix), ":")
	if len(hexToken) != testRedisNamespaceBytes*2 {
		return false
	}
	_, err := hex.DecodeString(hexToken)
	return err == nil
}
