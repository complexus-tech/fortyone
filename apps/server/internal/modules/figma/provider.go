package figma

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

const ProviderKey integrations.ProviderKey = "figma"

func ProviderDescriptor() integrations.Descriptor {
	return integrations.Descriptor{
		Key:         ProviderKey,
		DisplayName: "Figma",
		Family:      integrations.FamilyDesignContext,
		Capabilities: []integrations.Capability{
			{Key: integrations.CapabilityWebhookVerification, MajorVersion: 1},
			{Key: integrations.CapabilityTokenRefresh, MajorVersion: 1},
			{Key: integrations.CapabilityDesignFileContext, MajorVersion: 1},
		},
		AuthStrategies: []integrations.AuthStrategy{integrations.AuthStrategyOAuthInstall},
		Configuration: []integrations.ConfigurationRequirement{
			{Key: "FIGMA_CLIENT_ID", Required: true, Purpose: "OAuth client identity"},
			{Key: "FIGMA_CLIENT_SECRET", Required: true, Sensitive: true, Purpose: "OAuth client authentication"},
			{Key: "FIGMA_REDIRECT_URL", Required: true, Purpose: "Exact OAuth callback"},
			{Key: "FIGMA_WEBHOOK_URL", Required: true, Purpose: "Webhook delivery target"},
		},
		Disconnect: integrations.DisconnectPolicy{
			RevokeRemoteGrant:      true,
			DeleteWebhook:          true,
			DeleteCredentials:      true,
			RetainMappingMetadata:  true,
			MappingRetentionPeriod: 30 * 24 * time.Hour,
		},
		OperatorRunbook: "docs/integrations/providers.md#figma",
	}
}
