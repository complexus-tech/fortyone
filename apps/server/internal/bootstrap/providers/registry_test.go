package providers

import (
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

func TestBuiltInRegistryDeclaresProviderCapabilities(t *testing.T) {
	t.Parallel()

	registry, err := BuiltInRegistry()
	if err != nil {
		t.Fatalf("build provider registry: %v", err)
	}

	if providers := registry.List(); len(providers) != 7 {
		t.Fatalf("provider count = %d, want 7", len(providers))
	}
	tests := []struct {
		provider   integrations.ProviderKey
		capability integrations.CapabilityKey
	}{
		{provider: "github", capability: integrations.CapabilityCodeHostRepositoryCatalog},
		{provider: "gitlab", capability: integrations.CapabilityCodeHostCommentWriter},
		{provider: "slack", capability: integrations.CapabilityMessagingDelivery},
		{provider: "figma", capability: integrations.CapabilityDesignFileContext},
		{provider: "google-drive", capability: integrations.CapabilityCloudFileContext},
		{provider: "google-calendar", capability: integrations.CapabilityCalendarWatch},
		{provider: "microsoft-calendar", capability: integrations.CapabilityCalendarWatch},
	}
	for _, test := range tests {
		if err := registry.RequireCapability(test.provider, integrations.Capability{
			Key:          test.capability,
			MajorVersion: 1,
		}); err != nil {
			t.Errorf("provider %s capability %s: %v", test.provider, test.capability, err)
		}
	}
}
