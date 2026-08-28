package developercredentialshttp

import (
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type createCredentialRequest struct {
	Name      string               `json:"name"`
	Scopes    []platformauth.Scope `json:"scopes"`
	TeamIDs   []uuid.UUID          `json:"teamIds,omitempty"`
	ExpiresAt time.Time            `json:"expiresAt"`
}

type rotateCredentialRequest struct {
	ExpiresAt      time.Time `json:"expiresAt"`
	OverlapSeconds int64     `json:"overlapSeconds,omitempty"`
}

type createServiceAccountRequest struct {
	Name          string                      `json:"name"`
	WorkspaceRole authorization.WorkspaceRole `json:"workspaceRole"`
}

type credentialResponse struct {
	ID            uuid.UUID                                 `json:"id"`
	PrincipalID   uuid.UUID                                 `json:"principalId"`
	Kind          developercredentialsdomain.CredentialKind `json:"kind"`
	Name          string                                    `json:"name"`
	Prefix        string                                    `json:"prefix"`
	Scopes        []platformauth.Scope                      `json:"scopes"`
	TeamIDs       []uuid.UUID                               `json:"teamIds"`
	ExpiresAt     time.Time                                 `json:"expiresAt"`
	LastUsedAt    *time.Time                                `json:"lastUsedAt,omitempty"`
	RotatedFromID *uuid.UUID                                `json:"rotatedFromId,omitempty"`
	RotatedAt     *time.Time                                `json:"rotatedAt,omitempty"`
	RevokedAt     *time.Time                                `json:"revokedAt,omitempty"`
	CreatedAt     time.Time                                 `json:"createdAt"`
}

type issuedCredentialResponse struct {
	Credential credentialResponse `json:"credential"`
	Token      string             `json:"token"`
}

type serviceAccountResponse struct {
	ID             uuid.UUID                                  `json:"id"`
	Name           string                                     `json:"name"`
	WorkspaceRole  authorization.WorkspaceRole                `json:"workspaceRole"`
	Status         developercredentialsdomain.PrincipalStatus `json:"status"`
	CreatedAt      time.Time                                  `json:"createdAt"`
	UpdatedAt      time.Time                                  `json:"updatedAt"`
	DisabledAt     *time.Time                                 `json:"disabledAt,omitempty"`
	DisabledReason *string                                    `json:"disabledReason,omitempty"`
}

func credentialModel(value developercredentialsdomain.Credential) credentialResponse {
	teamIDs := append([]uuid.UUID(nil), value.TeamIDs...)
	if teamIDs == nil {
		teamIDs = []uuid.UUID{}
	}
	return credentialResponse{
		ID: value.ID, PrincipalID: value.PrincipalID, Kind: value.Kind, Name: value.Name,
		Prefix: value.LookupPrefix, Scopes: append([]platformauth.Scope(nil), value.Scopes...),
		TeamIDs: teamIDs, ExpiresAt: value.ExpiresAt, LastUsedAt: value.LastUsedAt,
		RotatedFromID: value.RotatedFromID, RotatedAt: value.RotatedAt, RevokedAt: value.RevokedAt,
		CreatedAt: value.CreatedAt,
	}
}

func issuedCredentialModel(value developercredentialsdomain.IssuedCredential) issuedCredentialResponse {
	return issuedCredentialResponse{Credential: credentialModel(value.Credential), Token: value.Token.Reveal()}
}

func credentialModels(values []developercredentialsdomain.Credential) []credentialResponse {
	models := make([]credentialResponse, len(values))
	for index, value := range values {
		models[index] = credentialModel(value)
	}
	return models
}

func serviceAccountModel(value developercredentialsdomain.ServiceAccount) serviceAccountResponse {
	return serviceAccountResponse{
		ID: value.ID, Name: value.Name, WorkspaceRole: value.WorkspaceRole, Status: value.Status,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, DisabledAt: value.DisabledAt,
		DisabledReason: value.DisabledReason,
	}
}

func serviceAccountModels(values []developercredentialsdomain.ServiceAccount) []serviceAccountResponse {
	models := make([]serviceAccountResponse, len(values))
	for index, value := range values {
		models[index] = serviceAccountModel(value)
	}
	return models
}

func personalTokenInput(request createCredentialRequest, requestID string) developercredentials.CreatePersonalTokenInput {
	return developercredentials.CreatePersonalTokenInput{
		Name: request.Name, Scopes: request.Scopes, TeamIDs: request.TeamIDs,
		ExpiresAt: request.ExpiresAt, RequestID: requestID,
	}
}

func serviceAccountKeyInput(request createCredentialRequest, requestID string) developercredentials.CreateServiceAccountKeyInput {
	return developercredentials.CreateServiceAccountKeyInput{
		Name: request.Name, Scopes: request.Scopes, TeamIDs: request.TeamIDs,
		ExpiresAt: request.ExpiresAt, RequestID: requestID,
	}
}
