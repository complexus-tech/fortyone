package developeroauth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type ClientCredentialsExchange struct {
	ClientID       string
	ClientSecret   string
	InstallationID uuid.UUID
	Resource       string
	Scopes         []string
	RequestID      string
}

func (service *Service) ExchangeClientCredentials(
	ctx context.Context,
	request ClientCredentialsExchange,
) (developeroauthdomain.ApplicationAccessToken, error) {
	if service == nil || service.applicationActors == nil {
		return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrApplicationActorUnavailable
	}
	if request.Resource != service.resource {
		return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrInvalidResource
	}
	if strings.TrimSpace(request.ClientID) != request.ClientID || request.ClientID == "" ||
		strings.TrimSpace(request.ClientSecret) != request.ClientSecret || request.ClientSecret == "" ||
		request.InstallationID == uuid.Nil {
		return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrInvalidClient
	}
	scopes, err := service.applicationActorScopes.normalize(request.Scopes)
	if err != nil {
		return developeroauthdomain.ApplicationAccessToken{}, err
	}
	prefix, err := service.tokens.ParseLookupPrefix(request.ClientSecret, developeroauthdomain.SecretClientSecret)
	if err != nil {
		return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrInvalidClient
	}

	now := service.clock.Now().UTC()
	accessTokenID, err := service.nextID()
	if err != nil {
		return developeroauthdomain.ApplicationAccessToken{}, fmt.Errorf("generate OAuth application access token ID: %w", err)
	}
	auditID, err := service.nextID()
	if err != nil {
		return developeroauthdomain.ApplicationAccessToken{}, fmt.Errorf("generate OAuth application audit ID: %w", err)
	}
	var accessToken developeroauthdomain.PlaintextSecret
	installation, err := service.applicationActors.AuthenticateApplication(
		ctx,
		developeroauthdomain.AuthenticateApplicationCredential{
			LookupPrefix: prefix, InstallationID: request.InstallationID, AccessTokenID: accessTokenID,
			AuditID: auditID, RequestID: request.RequestID, AuthenticatedAt: now,
		},
		func(record developeroauthdomain.ClientSecretRecord, installation developeroauthdomain.ApplicationInstallation) error {
			if err := service.tokens.Verify(request.ClientSecret, record.Material); err != nil {
				return developeroauthdomain.ErrInvalidClient
			}
			if record.Application.RegistrationKind != "confidential" ||
				record.Application.ID != installation.ApplicationID ||
				record.Application.ClientID != request.ClientID || record.ClientID != request.ClientID ||
				installation.ID != request.InstallationID || installation.Resource != request.Resource ||
				installation.ClientID != request.ClientID || installation.PrincipalID == uuid.Nil ||
				installation.WorkspaceID == uuid.Nil || installation.Status != "active" {
				return developeroauthdomain.ErrInvalidClient
			}
			if !service.applicationActorScopes.accepts(installation.Scopes) || !scopesAreSubset(scopes, installation.Scopes) {
				return developeroauthdomain.ErrInvalidScope
			}
			tokenInstallation := installation
			tokenInstallation.Scopes = append([]string(nil), scopes...)
			signed, err := service.signApplicationAccessToken(tokenInstallation, accessTokenID, now)
			if err != nil {
				return err
			}
			accessToken = signed
			return nil
		},
	)
	if err != nil {
		if errors.Is(err, developeroauthdomain.ErrInvalidScope) {
			return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrInvalidScope
		}
		if errors.Is(err, developeroauthdomain.ErrInvalidClient) || errors.Is(err, developeroauthdomain.ErrClientSecret) ||
			errors.Is(err, developeroauthdomain.ErrInstallationNotFound) || errors.Is(err, developeroauthdomain.ErrInstallationRevoked) {
			return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrInvalidClient
		}
		return developeroauthdomain.ApplicationAccessToken{}, fmt.Errorf("authenticate OAuth application: %w", err)
	}
	if installation.ApplicationID == uuid.Nil || installation.PrincipalID == uuid.Nil ||
		installation.WorkspaceID == uuid.Nil || installation.ID != request.InstallationID || accessToken.Reveal() == "" {
		return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrInvalidClient
	}
	return developeroauthdomain.ApplicationAccessToken{
		AccessToken: accessToken,
		ExpiresIn:   service.accessTokenTTL,
		Scopes:      append([]string(nil), scopes...),
	}, nil
}

func validateApplicationInstallationIdentity(
	identity developeroauthdomain.ApplicationInstallation,
	claims accessTokenClaims,
) error {
	if identity.ID.String() != claims.InstallationID || identity.WorkspaceID.String() != claims.WorkspaceID ||
		identity.PrincipalID.String() != claims.Subject || identity.ApplicationID.String() != claims.ApplicationID ||
		identity.ClientID != claims.ClientID || identity.Resource == "" || identity.Status != "active" {
		return developeroauthdomain.ErrInvalidGrant
	}
	if identity.PrincipalID == uuid.Nil || identity.WorkspaceID == uuid.Nil || identity.ID == uuid.Nil ||
		identity.ApplicationID == uuid.Nil || strings.TrimSpace(identity.ClientID) == "" {
		return developeroauthdomain.ErrInvalidGrant
	}
	if platformauth.PrincipalKind(claims.ActorKind) != platformauth.PrincipalOAuthApplication {
		return developeroauthdomain.ErrInvalidGrant
	}
	return nil
}
