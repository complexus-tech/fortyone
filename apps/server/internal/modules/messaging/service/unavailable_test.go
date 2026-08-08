package messaging

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestUnavailableAssistantReturnsPermanentConfigurationError(t *testing.T) {
	t.Parallel()

	assistant := NewUnavailableAssistant("OpenAI is disabled for this process")
	_, err := assistant.Respond(context.Background(), Request{})
	if !errors.Is(err, ErrAssistantNotConfigured) {
		t.Fatalf("expected configuration sentinel, got %v", err)
	}
	if !strings.Contains(err.Error(), "OpenAI is disabled for this process") {
		t.Fatalf("expected operator-safe reason, got %v", err)
	}
}

func TestUnavailableAssistantUsesSafeDefaultMessage(t *testing.T) {
	t.Parallel()

	assistant := NewUnavailableAssistant("  ")
	_, err := assistant.Respond(context.Background(), Request{})
	if !errors.Is(err, ErrAssistantNotConfigured) || !strings.Contains(err.Error(), defaultUnavailableMessage) {
		t.Fatalf("expected safe default configuration error, got %v", err)
	}
}
