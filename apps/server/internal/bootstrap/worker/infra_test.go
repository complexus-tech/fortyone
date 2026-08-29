package workerbootstrap

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisOptionsAlwaysVerifyTLSPeer(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Redis.Host = "redis.internal.example"
	cfg.Redis.Port = "6379"

	options := redisOptions(cfg)
	require.NotNil(t, options.TLSConfig)
	require.False(t, options.TLSConfig.InsecureSkipVerify)
	require.Equal(t, uint16(tls.VersionTLS12), options.TLSConfig.MinVersion)
	require.Equal(t, "redis.internal.example", options.TLSConfig.ServerName)

	cfg.Redis.DisableTLS = true
	require.Nil(t, redisOptions(cfg).TLSConfig)
}
