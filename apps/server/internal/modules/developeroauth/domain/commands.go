package developeroauthdomain

import (
	"time"

	"github.com/google/uuid"
)

type RegisterApplication struct {
	ID               uuid.UUID
	ClientID         string
	Name             string
	RegistrationKind string
	RedirectURIs     []string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

type AuthorizeUser struct {
	GrantID       uuid.UUID
	Code          SecretMaterial
	Application   Application
	UserID        uuid.UUID
	Resource      string
	Scopes        []string
	RedirectURI   string
	CodeChallenge string
	AuthorizedAt  time.Time
	CodeExpiresAt time.Time
}

type ExchangeAuthorizationCode struct {
	LookupPrefix string
	UsedAt       time.Time
	FamilyID     uuid.UUID
	FamilyExpiry time.Time
	Refresh      SecretMaterial
}

type RotateRefreshToken struct {
	LookupPrefix string
	UsedAt       time.Time
	Replacement  SecretMaterial
}
