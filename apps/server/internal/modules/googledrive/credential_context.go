package googledrive

import (
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

// CredentialContext is the authenticated-data contract for Google Drive OAuth
// credentials. Google subjects are stable account identifiers and are never
// exposed by the API.
func CredentialContext(userID uuid.UUID, googleSubject string, generation uuid.UUID) credentialvault.Context {
	return credentialvault.Context{
		Provider:       string(ProviderKey),
		TenantID:       userID.String(),
		SubjectID:      googleSubject,
		CredentialType: "oauth-token",
		Generation:     generation.String(),
	}
}
