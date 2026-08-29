package figma

import (
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

// CredentialContext is the single AAD contract for Figma OAuth credentials.
// Runtime reads, migration, and key rotation must all use this exact binding.
func CredentialContext(
	workspaceID, connectionID, installationGeneration uuid.UUID,
) credentialvault.Context {
	// #nosec G101 -- these are public authenticated-context identifiers, not secrets.
	return credentialvault.Context{
		Provider:       string(ProviderKey),
		TenantID:       workspaceID.String(),
		SubjectID:      connectionID.String(),
		CredentialType: "oauth-token",
		Generation:     installationGeneration.String(),
	}
}
