package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAPIHTTPConfigurationDefaults(t *testing.T) {
	t.Parallel()

	configType := reflect.TypeOf(Config{}.Web)
	tests := map[string]struct {
		environment  string
		defaultValue string
	}{
		"ReadHeaderTimeout":        {environment: "APP_API_READ_HEADER_TIMEOUT", defaultValue: "10s"},
		"ReadinessCheckTimeout":    {environment: "APP_API_READINESS_CHECK_TIMEOUT", defaultValue: "2s"},
		"TelemetryShutdownTimeout": {environment: "APP_API_TELEMETRY_SHUTDOWN_TIMEOUT", defaultValue: "5s"},
	}
	for fieldName, expected := range tests {
		field, found := configType.FieldByName(fieldName)
		require.True(t, found, fieldName)
		require.Equal(t, expected.environment, field.Tag.Get("env"), fieldName)
		require.Equal(t, expected.defaultValue, field.Tag.Get("default"), fieldName)
	}
}

func TestValidateAPIHTTPConfig(t *testing.T) {
	t.Parallel()

	cfg := validAPIHTTPConfig()
	require.NoError(t, validateAPIHTTPConfig(cfg))

	cfg.Web.APIHost = "missing-port"
	cfg.Web.ReadHeaderTimeout = 0
	cfg.Web.ReadinessCheckTimeout = -time.Second
	err := validateAPIHTTPConfig(cfg)
	require.ErrorContains(t, err, "APP_API_HOST")
	require.ErrorContains(t, err, "APP_API_READ_HEADER_TIMEOUT")
	require.ErrorContains(t, err, "APP_API_READINESS_CHECK_TIMEOUT")
}

func validAPIHTTPConfig() Config {
	var cfg Config
	cfg.Web.APIHost = "127.0.0.1:8000"
	cfg.Web.ReadHeaderTimeout = 10 * time.Second
	cfg.Web.ReadTimeout = time.Minute
	cfg.Web.WriteTimeout = time.Minute
	cfg.Web.IdleTimeout = time.Minute
	cfg.Web.ShutdownTimeout = 30 * time.Second
	cfg.Web.ReadinessCheckTimeout = 2 * time.Second
	cfg.Web.TelemetryShutdownTimeout = 5 * time.Second
	return cfg
}
