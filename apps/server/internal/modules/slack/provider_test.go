package slack

import (
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

func TestProviderDescriptorMatchesImplementedSlackLifecycle(t *testing.T) {
	t.Parallel()
	descriptor := ProviderDescriptor()
	if descriptor.Disconnect.DeleteWebhook {
		t.Fatal("Slack app event subscriptions are application-owned and must not be advertised as per-installation deletion")
	}
	for _, capability := range descriptor.Capabilities {
		if capability.Key == integrations.CapabilityTokenRefresh {
			t.Fatal("Slack token rotation must not be advertised until refresh-token persistence and atomic rotation are implemented")
		}
	}
	foundPayloadSecret := false
	for _, requirement := range descriptor.Configuration {
		if requirement.Key == "APP_SLACK_WEBHOOK_PAYLOAD_SECRET" {
			foundPayloadSecret = requirement.Required && requirement.Sensitive
		}
	}
	if !foundPayloadSecret {
		t.Fatal("Slack durable payload encryption key is not declared as required sensitive configuration")
	}
}
