package slack

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

const ProviderKey integrations.ProviderKey = "slack"

func ProviderDescriptor() integrations.Descriptor {
	return integrations.Descriptor{
		Key:         ProviderKey,
		DisplayName: "Slack",
		Family:      integrations.FamilyMessaging,
		Capabilities: []integrations.Capability{
			{Key: integrations.CapabilityWebhookVerification, MajorVersion: 1},
			{Key: integrations.CapabilityExternalIdentity, MajorVersion: 1},
			{Key: integrations.CapabilityMessagingDelivery, MajorVersion: 1},
			{Key: integrations.CapabilityMessagingCommands, MajorVersion: 1},
			{Key: integrations.CapabilityMessagingThreads, MajorVersion: 1},
		},
		AuthStrategies: []integrations.AuthStrategy{
			integrations.AuthStrategyOAuthInstall,
			integrations.AuthStrategyOAuthLink,
		},
		Configuration: []integrations.ConfigurationRequirement{
			{Key: "SLACK_CLIENT_ID", Required: true, Purpose: "OAuth client identity"},
			{Key: "SLACK_CLIENT_SECRET", Required: true, Sensitive: true, Purpose: "OAuth client authentication"},
			{Key: "SLACK_SIGNING_SECRET", Required: true, Sensitive: true, Purpose: "Request signature verification"},
			{Key: "SLACK_REDIRECT_URL", Required: true, Purpose: "Exact OAuth callback"},
			{Key: "APP_SLACK_WEBHOOK_PAYLOAD_SECRET", Required: true, Sensitive: true, Purpose: "Durable webhook payload encryption"},
		},
		Disconnect: integrations.DisconnectPolicy{
			RevokeRemoteGrant:      true,
			DeleteWebhook:          false,
			DeleteCredentials:      true,
			RetainMappingMetadata:  true,
			MappingRetentionPeriod: 30 * 24 * time.Hour,
		},
		OperatorRunbook: "docs/integrations/providers.md#slack",
	}
}
