package developeroauthhttp

import (
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	"github.com/google/uuid"
)

const (
	minimumRotationOverlapSeconds int64 = 60
	maximumRotationOverlapSeconds int64 = 24 * 60 * 60
)

type createManagedApplicationRequest struct {
	Name            string    `json:"name"`
	RedirectURIs    []string  `json:"redirectUris,omitempty"`
	ExpiresAt       time.Time `json:"expiresAt"`
	SecretExpiresAt time.Time `json:"secretExpiresAt"`
}

func (request createManagedApplicationRequest) Validate() error {
	if request.ExpiresAt.IsZero() || request.SecretExpiresAt.IsZero() {
		return developeroauthdomain.ErrInvalidExpiry
	}
	return nil
}

type rotateClientSecretRequest struct {
	ExpiresAt      time.Time `json:"expiresAt"`
	OverlapSeconds *int64    `json:"overlapSeconds"`
}

func (request rotateClientSecretRequest) Validate() error {
	if request.ExpiresAt.IsZero() {
		return developeroauthdomain.ErrInvalidExpiry
	}
	if request.OverlapSeconds == nil ||
		*request.OverlapSeconds < minimumRotationOverlapSeconds ||
		*request.OverlapSeconds > maximumRotationOverlapSeconds {
		return developeroauthdomain.ErrInvalidRotationOverlap
	}
	return nil
}

type installApplicationRequest struct {
	ClientID string   `json:"clientId"`
	Resource string   `json:"resource"`
	Scopes   []string `json:"scopes"`
}

func (request installApplicationRequest) Validate() error {
	if request.ClientID == "" {
		return developeroauthdomain.ErrInvalidClient
	}
	if request.Resource == "" {
		return developeroauthdomain.ErrInvalidResource
	}
	if len(request.Scopes) == 0 {
		return developeroauthdomain.ErrInvalidScope
	}
	return nil
}

type updateApplicationInstallationRequest struct {
	Resource string   `json:"resource"`
	Scopes   []string `json:"scopes"`
}

func (request updateApplicationInstallationRequest) Validate() error {
	if request.Resource == "" {
		return developeroauthdomain.ErrInvalidResource
	}
	if len(request.Scopes) == 0 {
		return developeroauthdomain.ErrInvalidScope
	}
	return nil
}

type managedApplicationResponse struct {
	ID               uuid.UUID  `json:"id"`
	ClientID         string     `json:"clientId"`
	Name             string     `json:"name"`
	RegistrationKind string     `json:"registrationKind"`
	RedirectURIs     []string   `json:"redirectUris"`
	ExpiresAt        time.Time  `json:"expiresAt"`
	OwnerWorkspaceID uuid.UUID  `json:"ownerWorkspaceId"`
	OwnerUserID      uuid.UUID  `json:"ownerUserId"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
}

type clientSecretResponse struct {
	ID               uuid.UUID  `json:"id"`
	ApplicationID    uuid.UUID  `json:"applicationId"`
	Prefix           string     `json:"prefix"`
	ExpiresAt        time.Time  `json:"expiresAt"`
	LastUsedAt       *time.Time `json:"lastUsedAt,omitempty"`
	RotatedFromID    *uuid.UUID `json:"rotatedFromId,omitempty"`
	OverlapExpiresAt *time.Time `json:"overlapExpiresAt,omitempty"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type issuedManagedApplicationResponse struct {
	Application  managedApplicationResponse `json:"application"`
	ClientSecret clientSecretResponse       `json:"clientSecret"`
	Secret       string                     `json:"secret"`
}

type issuedClientSecretResponse struct {
	ClientSecret                   clientSecretResponse `json:"clientSecret"`
	Secret                         string               `json:"secret"`
	PreviousSecretOverlapExpiresAt *time.Time           `json:"previousSecretOverlapExpiresAt,omitempty"`
}

type applicationInstallationResponse struct {
	ID            uuid.UUID  `json:"id"`
	ApplicationID uuid.UUID  `json:"applicationId"`
	ClientID      string     `json:"clientId"`
	WorkspaceID   uuid.UUID  `json:"workspaceId"`
	PrincipalID   uuid.UUID  `json:"principalId"`
	Resource      string     `json:"resource"`
	Scopes        []string   `json:"scopes"`
	Status        string     `json:"status"`
	InstalledBy   uuid.UUID  `json:"installedBy"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	LastUsedAt    *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt     *time.Time `json:"revokedAt,omitempty"`
	RevokedBy     *uuid.UUID `json:"revokedBy,omitempty"`
	RevokedReason *string    `json:"revokedReason,omitempty"`
}

func managedApplicationModel(value developeroauthdomain.ManagedApplication) managedApplicationResponse {
	redirectURIs := append([]string(nil), value.RedirectURIs...)
	if redirectURIs == nil {
		redirectURIs = []string{}
	}
	return managedApplicationResponse{
		ID: value.ID, ClientID: value.ClientID, Name: value.Name,
		RegistrationKind: value.RegistrationKind, RedirectURIs: redirectURIs,
		ExpiresAt: value.ExpiresAt, OwnerWorkspaceID: value.OwnerWorkspaceID,
		OwnerUserID: value.OwnerUserID, Status: value.Status, CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt, RevokedAt: value.RevokedAt,
	}
}

func managedApplicationModels(values []developeroauthdomain.ManagedApplication) []managedApplicationResponse {
	models := make([]managedApplicationResponse, len(values))
	for index, value := range values {
		models[index] = managedApplicationModel(value)
	}
	return models
}

func clientSecretModel(value developeroauthdomain.ClientSecret) clientSecretResponse {
	return clientSecretResponse{
		ID: value.ID, ApplicationID: value.ApplicationID, Prefix: value.LookupPrefix,
		ExpiresAt: value.ExpiresAt, LastUsedAt: value.LastUsedAt,
		RotatedFromID: value.RotatedFromID, OverlapExpiresAt: value.OverlapExpiresAt,
		RevokedAt: value.RevokedAt, CreatedAt: value.CreatedAt,
	}
}

func clientSecretModels(values []developeroauthdomain.ClientSecret) []clientSecretResponse {
	models := make([]clientSecretResponse, len(values))
	for index, value := range values {
		models[index] = clientSecretModel(value)
	}
	return models
}

func issuedManagedApplicationModel(value developeroauthdomain.IssuedManagedApplication) issuedManagedApplicationResponse {
	return issuedManagedApplicationResponse{
		Application:  managedApplicationModel(value.Application),
		ClientSecret: clientSecretModel(value.Secret.Secret),
		Secret:       value.Secret.Plaintext.Reveal(),
	}
}

func issuedClientSecretModel(value developeroauthdomain.IssuedClientSecret) issuedClientSecretResponse {
	return issuedClientSecretResponse{
		ClientSecret: clientSecretModel(value.Secret), Secret: value.Plaintext.Reveal(),
		PreviousSecretOverlapExpiresAt: value.PreviousSecretOverlapExpiresAt,
	}
}

func applicationInstallationModel(value developeroauthdomain.ApplicationInstallation) applicationInstallationResponse {
	scopes := append([]string(nil), value.Scopes...)
	if scopes == nil {
		scopes = []string{}
	}
	return applicationInstallationResponse{
		ID: value.ID, ApplicationID: value.ApplicationID, ClientID: value.ClientID,
		WorkspaceID: value.WorkspaceID, PrincipalID: value.PrincipalID,
		Resource: value.Resource, Scopes: scopes, Status: value.Status,
		InstalledBy: value.InstalledBy, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
		LastUsedAt: value.LastUsedAt, RevokedAt: value.RevokedAt,
		RevokedBy: value.RevokedBy, RevokedReason: value.RevokedReason,
	}
}

func applicationInstallationModels(values []developeroauthdomain.ApplicationInstallation) []applicationInstallationResponse {
	models := make([]applicationInstallationResponse, len(values))
	for index, value := range values {
		models[index] = applicationInstallationModel(value)
	}
	return models
}
