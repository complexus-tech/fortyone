package domain

import "github.com/google/uuid"

// LegacyCredential is the bounded migration projection for the former
// provider-local Figma envelope. It never crosses an HTTP boundary.
type LegacyCredential struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	Payload                string
	InstallationGeneration uuid.UUID
}

type Credential struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	Payload                string
	CredentialVersion      int16
	InstallationGeneration uuid.UUID
}
