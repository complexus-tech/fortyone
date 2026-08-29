package gitlab

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

const ProviderKey integrations.ProviderKey = "gitlab"

func ProviderDescriptor() integrations.Descriptor {
	return integrations.Descriptor{
		Key:         ProviderKey,
		DisplayName: "GitLab",
		Family:      integrations.FamilyCodeHost,
		Capabilities: []integrations.Capability{
			{Key: integrations.CapabilityWebhookVerification, MajorVersion: 1},
			{Key: integrations.CapabilityCodeHostRepositoryCatalog, MajorVersion: 1},
			{Key: integrations.CapabilityCodeHostWorkItemWriter, MajorVersion: 1},
			{Key: integrations.CapabilityCodeHostCommentWriter, MajorVersion: 1},
		},
		AuthStrategies: []integrations.AuthStrategy{
			integrations.AuthStrategyOAuthInstall,
			integrations.AuthStrategyWebhookOnly,
		},
		Configuration: []integrations.ConfigurationRequirement{
			{Key: "GITLAB_BASE_URL", Required: true, Purpose: "GitLab instance API origin"},
			{Key: "GITLAB_OAUTH_CLIENT_ID", Required: true, Purpose: "OAuth client identity"},
			{Key: "GITLAB_OAUTH_CLIENT_SECRET", Required: true, Sensitive: true, Purpose: "OAuth client authentication"},
			{Key: "GITLAB_WEBHOOK_SIGNING_TOKEN", Required: true, Sensitive: true, Purpose: "Standard Webhooks HMAC verification"},
		},
		Disconnect: integrations.DisconnectPolicy{
			RevokeRemoteGrant:      true,
			DeleteWebhook:          true,
			DeleteCredentials:      true,
			RetainMappingMetadata:  true,
			MappingRetentionPeriod: 30 * 24 * time.Hour,
		},
		OperatorRunbook: "docs/integrations/code-hosts.md#gitlab-proof",
	}
}
