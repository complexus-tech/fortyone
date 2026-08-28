package testkit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestParseTestRedisURLRequiresExplicitSafeContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		redisURL  string
		wantError string
	}{
		{name: "missing", wantError: "is required"},
		{name: "malformed", redisURL: "redis://user:fake_password@%zz:6379/0", wantError: "valid Redis URL"},
		{name: "wrong scheme", redisURL: "http://user:fake_password@localhost:6379/0", wantError: "redis or rediss"},
		{name: "missing port", redisURL: "redis://user:fake_password@localhost/0", wantError: "host and port"},
		{name: "empty database", redisURL: "redis://user:fake_password@localhost:6379/", wantError: "numeric Redis database"},
		{name: "multiple path segments", redisURL: "redis://user:fake_password@localhost:6379/0/extra", wantError: "numeric Redis database"},
		{name: "negative database", redisURL: "redis://user:fake_password@localhost:6379/-1", wantError: "non-negative integer"},
		{name: "fragment", redisURL: "redis://user:fake_password@localhost:6379/0#fragment", wantError: "fragment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseTestRedisURL(tt.redisURL)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("parse error = %v, want error containing %q", err, tt.wantError)
			}
			if strings.Contains(err.Error(), "fake_password") {
				t.Fatal("parse error exposed Redis URL credentials")
			}
		})
	}
}

func TestParseTestRedisURLAppliesBoundedClientSettings(t *testing.T) {
	t.Parallel()

	options, err := parseTestRedisURL("rediss://user:fake_password@127.0.0.1:6379/3?protocol=2")
	if err != nil {
		t.Fatalf("parse valid Redis URL: %v", err)
	}
	if options.Addr != "127.0.0.1:6379" || options.DB != 3 {
		t.Fatalf("parsed Redis endpoint or database incorrectly")
	}
	if options.Password != "fake_password" || options.TLSConfig == nil {
		t.Fatal("Redis credentials or TLS scheme were not parsed")
	}
	if options.MaxRetries != 1 || options.DialTimeout != 3*time.Second {
		t.Fatalf("retry or dial bounds were not applied")
	}
	if options.ReadTimeout != 2*time.Second || options.WriteTimeout != 2*time.Second || options.PoolTimeout != 2*time.Second {
		t.Fatal("command or pool timeout bounds were not applied")
	}
	if !options.ContextTimeoutEnabled || options.PoolSize != testRedisPoolSize || options.MaxActiveConns != testRedisPoolSize {
		t.Fatal("context or connection pool bounds were not applied")
	}
}

func TestNewTestRedisNamespaceIsCryptographicallyShapedAndUnique(t *testing.T) {
	t.Parallel()

	first, err := newTestRedisNamespace()
	if err != nil {
		t.Fatalf("generate first namespace: %v", err)
	}
	second, err := newTestRedisNamespace()
	if err != nil {
		t.Fatalf("generate second namespace: %v", err)
	}
	if !isTestRedisNamespace(first) || !isTestRedisNamespace(second) {
		t.Fatalf("generated namespace did not satisfy the owned namespace contract")
	}
	if first == second {
		t.Fatal("independent Redis test namespaces matched")
	}
}

func TestRedisKeyAlwaysQualifiesLogicalKey(t *testing.T) {
	t.Parallel()

	namespace := testRedisNamespacePrefix + strings.Repeat("a", testRedisNamespaceBytes*2) + ":"
	testRedis := &Redis{Namespace: namespace}
	if got := testRedis.Key("sessions:shared"); got != namespace+"sessions:shared" {
		t.Fatalf("qualified key = %q", got)
	}
}

func TestDeleteRedisNamespaceDeletesOnlyOwnedKeys(t *testing.T) {
	t.Parallel()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis test client: %v", err)
		}
	})

	ownedNamespace := testRedisNamespacePrefix + strings.Repeat("a", testRedisNamespaceBytes*2) + ":"
	siblingNamespace := testRedisNamespacePrefix + strings.Repeat("b", testRedisNamespaceBytes*2) + ":"
	ctx := t.Context()
	if err := client.MSet(ctx,
		ownedNamespace+"one", "owned",
		ownedNamespace+"two", "owned",
		siblingNamespace+"one", "sibling",
		"unrelated", "outside",
	).Err(); err != nil {
		t.Fatalf("seed Redis keys: %v", err)
	}

	if err := deleteRedisNamespace(ctx, client, ownedNamespace); err != nil {
		t.Fatalf("delete owned namespace: %v", err)
	}
	assertRedisKeyExists(t, client, ownedNamespace+"one", false)
	assertRedisKeyExists(t, client, ownedNamespace+"two", false)
	assertRedisKeyExists(t, client, siblingNamespace+"one", true)
	assertRedisKeyExists(t, client, "unrelated", true)
}

func TestDeleteRedisNamespaceRejectsBroadPrefix(t *testing.T) {
	t.Parallel()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis test client: %v", err)
		}
	})

	ctx := t.Context()
	if err := client.Set(ctx, "unrelated", "outside", 0).Err(); err != nil {
		t.Fatalf("seed unrelated Redis key: %v", err)
	}
	if err := deleteRedisNamespace(ctx, client, testRedisNamespacePrefix); err == nil {
		t.Fatal("broad Redis namespace deletion was accepted")
	}
	assertRedisKeyExists(t, client, "unrelated", true)
}

func assertRedisKeyExists(t testing.TB, client *redis.Client, key string, want bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		t.Fatalf("inspect Redis key existence: %v", err)
	}
	if got := exists == 1; got != want {
		t.Fatalf("Redis key existence = %t, want %t", got, want)
	}
}
