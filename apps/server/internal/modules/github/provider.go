package github

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

const ProviderKey integrations.ProviderKey = "github"

// ProviderDescriptor declares GitHub control-plane and adapter capabilities.
// Concrete factories remain in bootstrap and the GitHub adapter packages.
func ProviderDescriptor() integrations.Descriptor {
	return integrations.Descriptor{
		Key:         ProviderKey,
		DisplayName: "GitHub",
		Family:      integrations.FamilyCodeHost,
		Capabilities: []integrations.Capability{
			{Key: integrations.CapabilityWebhookVerification, MajorVersion: 1},
			{Key: integrations.CapabilityExternalIdentity, MajorVersion: 1},
			{Key: integrations.CapabilityCodeHostRepositoryCatalog, MajorVersion: 1},
			{Key: integrations.CapabilityCodeHostWorkItemWriter, MajorVersion: 1},
			{Key: integrations.CapabilityCodeHostCommentWriter, MajorVersion: 1},
		},
		AuthStrategies: []integrations.AuthStrategy{
			integrations.AuthStrategyAppInstallation,
			integrations.AuthStrategyOAuthLink,
		},
		Configuration: []integrations.ConfigurationRequirement{
			{Key: "APP_GITHUB_APP_ID", Required: true, Purpose: "GitHub App identity"},
			{Key: "GITHUB_APP_SLUG", Required: true, Purpose: "GitHub App installation URL"},
			{Key: "GITHUB_CLIENT_ID", Required: true, Purpose: "OAuth client identity"},
			{Key: "GITHUB_CLIENT_SECRET", Required: true, Sensitive: true, Purpose: "OAuth client authentication"},
			{Key: "GITHUB_PRIVATE_KEY_BASE64", Required: true, Sensitive: true, Purpose: "GitHub App request signing"},
			{Key: "GITHUB_REDIRECT_URL", Required: true, Purpose: "Exact OAuth callback"},
			{Key: "GITHUB_WEBHOOK_SECRET", Required: true, Sensitive: true, Purpose: "Webhook verification"},
		},
		Disconnect: integrations.DisconnectPolicy{
			RevokeRemoteGrant:      true,
			DeleteWebhook:          false,
			DeleteCredentials:      true,
			RetainMappingMetadata:  true,
			MappingRetentionPeriod: 30 * 24 * time.Hour,
		},
		OperatorRunbook: "docs/integrations/providers.md#github",
	}
}
