package github

import (
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

func TestProviderDescriptorMatchesImplementedGitHubLifecycle(t *testing.T) {
	t.Parallel()
	descriptor := ProviderDescriptor()
	if descriptor.Disconnect.DeleteWebhook {
		t.Fatal("GitHub App webhooks are application-owned and must not be advertised as per-installation deletion")
	}
	for _, capability := range descriptor.Capabilities {
		if capability.Key == integrations.CapabilityTokenRefresh {
			t.Fatal("GitHub OAuth token refresh is not implemented")
		}
	}
	for _, requirement := range descriptor.Configuration {
		if requirement.Key == "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET" {
			t.Fatal("GitHub payload encryption is derived internally and must not be operator configuration")
		}
	}
}
