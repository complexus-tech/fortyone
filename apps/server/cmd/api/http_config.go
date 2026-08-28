package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func validateAPIHTTPConfig(cfg Config) error {
	var validationErrors []error
	address := strings.TrimSpace(cfg.Web.APIHost)
	if address == "" {
		validationErrors = append(validationErrors, errors.New("APP_API_HOST is required"))
	} else if _, _, err := net.SplitHostPort(address); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("APP_API_HOST must be a host:port address: %w", err))
	}
	if cfg.Web.ReadHeaderTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_API_READ_HEADER_TIMEOUT must be positive"))
	}
	if cfg.Web.ReadTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_API_READ_TIMEOUT must be positive"))
	}
	if cfg.Web.WriteTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_API_WRITE_TIMEOUT must be positive"))
	}
	if cfg.Web.IdleTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_API_IDLE_TIMEOUT must be positive"))
	}
	if cfg.Web.ShutdownTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_API_SHUTDOWN_TIMEOUT must be positive"))
	}
	if cfg.Web.ReadinessCheckTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_API_READINESS_CHECK_TIMEOUT must be positive"))
	}
	if cfg.Web.TelemetryShutdownTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_API_TELEMETRY_SHUTDOWN_TIMEOUT must be positive"))
	}
	return errors.Join(validationErrors...)
}
