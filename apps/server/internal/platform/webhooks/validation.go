package webhooks

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	defaultMaxBodyBytes   = 1 << 20
	defaultMaxHeaderBytes = 64 << 10
	defaultMaxHeaderCount = 128
	defaultRetention      = 30 * 24 * time.Hour
	maxRetention          = 90 * 24 * time.Hour
)

var safeTraceID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

type Config struct {
	MaxBodyBytes     int
	MaxHeaderBytes   int
	MaxHeaderCount   int
	PayloadRetention time.Duration
	Now              func() time.Time
}

func normalizeConfig(config Config) (Config, error) {
	if config.MaxBodyBytes == 0 {
		config.MaxBodyBytes = defaultMaxBodyBytes
	}
	if config.MaxHeaderBytes == 0 {
		config.MaxHeaderBytes = defaultMaxHeaderBytes
	}
	if config.MaxHeaderCount == 0 {
		config.MaxHeaderCount = defaultMaxHeaderCount
	}
	if config.PayloadRetention == 0 {
		config.PayloadRetention = defaultRetention
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.MaxBodyBytes < 1 || config.MaxHeaderBytes < 1 || config.MaxHeaderCount < 1 {
		return Config{}, fmt.Errorf("%w: request limits must be positive", ErrNotConfigured)
	}
	if config.PayloadRetention < time.Hour || config.PayloadRetention > maxRetention {
		return Config{}, fmt.Errorf("%w: payload retention must be between one hour and 90 days", ErrNotConfigured)
	}
	return config, nil
}

func validateRequest(request SignedRequest, config Config) error {
	if len(request.Body) == 0 {
		return fmt.Errorf("%w: body is required", ErrInvalidRequest)
	}
	if len(request.Body) > config.MaxBodyBytes {
		return ErrPayloadTooLarge
	}
	if len(request.Headers) > config.MaxHeaderCount {
		return ErrHeadersTooLarge
	}
	headerBytes := 0
	for name, values := range request.Headers {
		headerBytes += len(name)
		for _, value := range values {
			headerBytes += len(value)
		}
		if headerBytes > config.MaxHeaderBytes {
			return ErrHeadersTooLarge
		}
	}
	return nil
}

func ValidateEnvelope(envelope Envelope) error {
	if envelope.Version != CurrentEnvelopeVersion || envelope.Provider == "" {
		return ErrInvalidDelivery
	}
	if envelope.WorkspaceID == uuid.Nil || envelope.InstallationID == uuid.Nil || envelope.InstallationGeneration == uuid.Nil {
		return ErrInvalidDelivery
	}
	if !validIdentifier(envelope.DeliveryID, 255) ||
		!validIdentifier(envelope.ExternalAccountID, 255) ||
		!validIdentifier(envelope.EventType, 200) {
		return ErrInvalidDelivery
	}
	if envelope.TraceID != "" && !safeTraceID.MatchString(envelope.TraceID) {
		return ErrInvalidDelivery
	}
	if envelope.ReceivedAt.IsZero() {
		return ErrInvalidDelivery
	}
	return nil
}

func ValidatePayloadRetention(receivedAt, expiresAt time.Time) error {
	retention := expiresAt.Sub(receivedAt)
	if retention < time.Hour || retention > maxRetention {
		return ErrInvalidDelivery
	}
	return nil
}

func validIdentifier(value string, max int) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > max || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func verificationError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	case errors.Is(err, ErrDeliveryIgnored):
		return ErrDeliveryIgnored
	case errors.Is(err, ErrReplay):
		return ErrReplay
	case errors.Is(err, ErrUnauthenticated):
		return ErrUnauthenticated
	case errors.Is(err, ErrVerificationUnavailable):
		return ErrVerificationUnavailable
	default:
		return ErrVerificationFailed
	}
}
