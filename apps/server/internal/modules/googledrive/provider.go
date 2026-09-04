// Package googledrive implements the personal Google Drive integration.
package googledrive

import "github.com/complexus-tech/projects-api/internal/platform/integrations"

const ProviderKey integrations.ProviderKey = "google-drive"

func ProviderDescriptor() integrations.Descriptor {
	return integrations.Descriptor{
		Key:         ProviderKey,
		DisplayName: "Google Drive",
		Family:      integrations.FamilyCloudContent,
		Capabilities: []integrations.Capability{
			{Key: integrations.CapabilityTokenRefresh, MajorVersion: 1},
			{Key: integrations.CapabilityCloudFileContext, MajorVersion: 1},
			{Key: integrations.CapabilityCloudFileCreate, MajorVersion: 1},
			{Key: integrations.CapabilityCloudFileImport, MajorVersion: 1},
		},
		AuthStrategies: []integrations.AuthStrategy{integrations.AuthStrategyOAuthLink},
		Configuration: []integrations.ConfigurationRequirement{
			{Key: "GOOGLE_DRIVE_CLIENT_ID", Required: true, Purpose: "OAuth client identity"},
			{Key: "GOOGLE_DRIVE_CLIENT_SECRET", Required: true, Sensitive: true, Purpose: "OAuth client authentication"},
			{Key: "GOOGLE_DRIVE_REDIRECT_URL", Required: true, Purpose: "Exact OAuth callback"},
			{Key: "GOOGLE_DRIVE_PICKER_API_KEY", Required: true, Purpose: "Browser-restricted Picker key"},
			{Key: "GOOGLE_DRIVE_APP_ID", Required: true, Purpose: "Numeric Google Cloud project number"},
		},
		Disconnect: integrations.DisconnectPolicy{
			RevokeRemoteGrant:      true,
			DeleteCredentials:      true,
			RetainMappingMetadata:  false,
			MappingRetentionPeriod: 0,
		},
		OperatorRunbook: "docs/integrations/providers.md#google-drive",
	}
}
