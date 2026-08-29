package calendar

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

const (
	GoogleProviderKey    integrations.ProviderKey = "google-calendar"
	MicrosoftProviderKey integrations.ProviderKey = "microsoft-calendar"
)

func GoogleProviderDescriptor() integrations.Descriptor {
	return calendarProviderDescriptor(
		GoogleProviderKey,
		"Google Calendar",
		[]integrations.ConfigurationRequirement{
			{Key: "APP_AUTH_GOOGLE_CLIENT_IDS", Required: true, Purpose: "OAuth client identities"},
			{Key: "APP_AUTH_GOOGLE_CLIENT_SECRET", Required: true, Sensitive: true, Purpose: "OAuth client authentication"},
			{Key: "APP_AUTH_GOOGLE_CALENDAR_REDIRECT_URL", Required: true, Purpose: "Exact OAuth callback"},
			{Key: "APP_AUTH_GOOGLE_CALENDAR_WEBHOOK_URL", Required: true, Purpose: "Change-notification target"},
		},
		"docs/integrations/providers.md#google-calendar",
	)
}

func MicrosoftProviderDescriptor() integrations.Descriptor {
	return calendarProviderDescriptor(
		MicrosoftProviderKey,
		"Microsoft Calendar",
		[]integrations.ConfigurationRequirement{
			{Key: "APP_AUTH_MICROSOFT_CLIENT_ID", Required: true, Purpose: "OAuth client identity"},
			{Key: "APP_AUTH_MICROSOFT_CLIENT_SECRET", Required: true, Sensitive: true, Purpose: "OAuth client authentication"},
			{Key: "APP_AUTH_MICROSOFT_TENANT", Required: true, Purpose: "Microsoft identity tenant"},
			{Key: "APP_AUTH_MICROSOFT_CALENDAR_REDIRECT_URL", Required: true, Purpose: "Exact OAuth callback"},
			{Key: "APP_AUTH_MICROSOFT_CALENDAR_WEBHOOK_URL", Required: true, Purpose: "Change-notification target"},
		},
		"docs/integrations/providers.md#microsoft-calendar",
	)
}

func calendarProviderDescriptor(
	key integrations.ProviderKey,
	displayName string,
	configuration []integrations.ConfigurationRequirement,
	runbook string,
) integrations.Descriptor {
	return integrations.Descriptor{
		Key:         key,
		DisplayName: displayName,
		Family:      integrations.FamilyCalendar,
		Capabilities: []integrations.Capability{
			{Key: integrations.CapabilityWebhookVerification, MajorVersion: 1},
			{Key: integrations.CapabilityTokenRefresh, MajorVersion: 1},
			{Key: integrations.CapabilityCalendarEvents, MajorVersion: 1},
			{Key: integrations.CapabilityCalendarAvailability, MajorVersion: 1},
			{Key: integrations.CapabilityCalendarWatch, MajorVersion: 1},
		},
		AuthStrategies: []integrations.AuthStrategy{integrations.AuthStrategyOAuthInstall},
		Configuration:  configuration,
		Disconnect: integrations.DisconnectPolicy{
			RevokeRemoteGrant:      true,
			DeleteWebhook:          true,
			DeleteCredentials:      true,
			RetainMappingMetadata:  true,
			MappingRetentionPeriod: 30 * 24 * time.Hour,
		},
		OperatorRunbook: runbook,
	}
}
