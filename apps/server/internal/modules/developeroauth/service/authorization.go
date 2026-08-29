package developeroauth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type AuthorizationRequest struct {
	ClientID      string
	UserID        uuid.UUID
	RedirectURI   string
	Resource      string
	Scopes        []string
	CodeChallenge string
}

func (service *Service) AuthorizeUser(
	ctx context.Context,
	request AuthorizationRequest,
) (developeroauthdomain.PlaintextSecret, error) {
	if request.UserID == uuid.Nil {
		return developeroauthdomain.PlaintextSecret{}, developeroauthdomain.ErrAuthorizationDenied
	}
	application, scopes, err := service.PrepareAuthorization(ctx, request)
	if err != nil {
		return developeroauthdomain.PlaintextSecret{}, err
	}
	grantID, err := service.nextID()
	if err != nil {
		return developeroauthdomain.PlaintextSecret{}, fmt.Errorf("generate OAuth grant ID: %w", err)
	}
	codeID, err := service.nextID()
	if err != nil {
		return developeroauthdomain.PlaintextSecret{}, fmt.Errorf("generate OAuth authorization code ID: %w", err)
	}
	code, err := service.tokens.Issue(developeroauthdomain.SecretAuthorizationCode, codeID)
	if err != nil {
		return developeroauthdomain.PlaintextSecret{}, err
	}
	now := service.clock.Now().UTC()
	_, err = service.repository.AuthorizeUser(ctx, developeroauthdomain.AuthorizeUser{
		GrantID: grantID, Code: code.Material, Application: application, UserID: request.UserID,
		Resource: service.resource, Scopes: scopes, RedirectURI: request.RedirectURI,
		CodeChallenge: request.CodeChallenge, AuthorizedAt: now, CodeExpiresAt: now.Add(service.authorizationCodeTTL),
	})
	if err != nil {
		return developeroauthdomain.PlaintextSecret{}, err
	}
	return code.Plaintext, nil
}

func (service *Service) PrepareAuthorization(
	ctx context.Context,
	request AuthorizationRequest,
) (developeroauthdomain.Application, []string, error) {
	application, err := service.GetApplication(ctx, request.ClientID)
	if err != nil {
		return developeroauthdomain.Application{}, nil, err
	}
	if !containsExact(application.RedirectURIs, request.RedirectURI) {
		return developeroauthdomain.Application{}, nil, developeroauthdomain.ErrInvalidRedirectURI
	}
	// Client and exact redirect validation intentionally precede every error
	// that the HTTP adapter may return through the registered callback. This
	// ordering prevents an invalid resource from becoming an open redirect.
	if request.Resource != service.resource {
		return application, nil, developeroauthdomain.ErrInvalidResource
	}
	if err := validatePKCEChallenge(request.CodeChallenge); err != nil {
		return application, nil, err
	}
	scopes, err := service.scopePolicy.normalize(request.Scopes)
	if err != nil {
		return application, nil, err
	}
	return application, scopes, nil
}

type AuthorizationCodeExchange struct {
	Code         string
	ClientID     string
	RedirectURI  string
	Resource     string
	CodeVerifier string
}

func (service *Service) ExchangeAuthorizationCode(
	ctx context.Context,
	request AuthorizationCodeExchange,
) (developeroauthdomain.TokenPair, error) {
	if request.Resource != service.resource {
		return developeroauthdomain.TokenPair{}, developeroauthdomain.ErrInvalidResource
	}
	if err := validatePKCEVerifier(request.CodeVerifier); err != nil {
		return developeroauthdomain.TokenPair{}, errors.Join(developeroauthdomain.ErrAuthorizationCode, err)
	}
	prefix, err := service.tokens.ParseLookupPrefix(request.Code, developeroauthdomain.SecretAuthorizationCode)
	if err != nil {
		return developeroauthdomain.TokenPair{}, developeroauthdomain.ErrAuthorizationCode
	}
	familyID, err := service.nextID()
	if err != nil {
		return developeroauthdomain.TokenPair{}, fmt.Errorf("generate OAuth refresh family ID: %w", err)
	}
	refreshID, err := service.nextID()
	if err != nil {
		return developeroauthdomain.TokenPair{}, fmt.Errorf("generate OAuth refresh token ID: %w", err)
	}
	refresh, err := service.tokens.Issue(developeroauthdomain.SecretRefreshToken, refreshID)
	if err != nil {
		return developeroauthdomain.TokenPair{}, err
	}
	now := service.clock.Now().UTC()
	grant, err := service.repository.ExchangeAuthorizationCode(
		ctx,
		developeroauthdomain.ExchangeAuthorizationCode{
			LookupPrefix: prefix, UsedAt: now, FamilyID: familyID,
			FamilyExpiry: now.Add(service.refreshTokenTTL), Refresh: refresh.Material,
		},
		func(record developeroauthdomain.AuthorizationCode) error {
			if err := service.tokens.Verify(request.Code, developeroauthdomain.SecretMaterial{
				ID: record.ID, Kind: developeroauthdomain.SecretAuthorizationCode,
				LookupPrefix: record.LookupPrefix, Digest: record.Digest, DigestKey: record.DigestKey,
			}); err != nil {
				return developeroauthdomain.ErrAuthorizationCode
			}
			if record.ClientID != request.ClientID || record.RedirectURI != request.RedirectURI ||
				record.Resource != request.Resource || record.ActorKind != platformauth.PrincipalOAuthUser {
				return developeroauthdomain.ErrInvalidClient
			}
			digest := sha256.Sum256([]byte(request.CodeVerifier))
			expectedChallenge := base64.RawURLEncoding.EncodeToString(digest[:])
			if subtle.ConstantTimeCompare([]byte(expectedChallenge), []byte(record.CodeChallenge)) != 1 {
				return developeroauthdomain.ErrAuthorizationCode
			}
			return nil
		},
	)
	if err != nil {
		return developeroauthdomain.TokenPair{}, err
	}
	accessToken, err := service.signAccessToken(grant, now)
	if err != nil {
		return developeroauthdomain.TokenPair{}, err
	}
	return developeroauthdomain.TokenPair{
		AccessToken: accessToken, RefreshToken: refresh.Plaintext,
		ExpiresIn: service.accessTokenTTL, Scopes: append([]string(nil), grant.Scopes...),
	}, nil
}

type RefreshExchange struct {
	RefreshToken string
	ClientID     string
	Resource     string
}

func (service *Service) ExchangeRefreshToken(
	ctx context.Context,
	request RefreshExchange,
) (developeroauthdomain.TokenPair, error) {
	if request.Resource != service.resource {
		return developeroauthdomain.TokenPair{}, developeroauthdomain.ErrInvalidResource
	}
	prefix, err := service.tokens.ParseLookupPrefix(request.RefreshToken, developeroauthdomain.SecretRefreshToken)
	if err != nil {
		return developeroauthdomain.TokenPair{}, developeroauthdomain.ErrRefreshToken
	}
	replacementID, err := service.nextID()
	if err != nil {
		return developeroauthdomain.TokenPair{}, fmt.Errorf("generate OAuth refresh token ID: %w", err)
	}
	replacement, err := service.tokens.Issue(developeroauthdomain.SecretRefreshToken, replacementID)
	if err != nil {
		return developeroauthdomain.TokenPair{}, err
	}
	now := service.clock.Now().UTC()
	grant, err := service.repository.RotateRefreshToken(
		ctx,
		developeroauthdomain.RotateRefreshToken{
			LookupPrefix: prefix, UsedAt: now, Replacement: replacement.Material,
		},
		func(record developeroauthdomain.RefreshToken) error {
			if err := service.tokens.Verify(request.RefreshToken, developeroauthdomain.SecretMaterial{
				ID: record.ID, Kind: developeroauthdomain.SecretRefreshToken,
				LookupPrefix: record.LookupPrefix, Digest: record.Digest, DigestKey: record.DigestKey,
			}); err != nil {
				return developeroauthdomain.ErrRefreshToken
			}
			if record.Grant.ClientID != request.ClientID || record.Grant.Resource != request.Resource ||
				record.Grant.ActorKind != platformauth.PrincipalOAuthUser {
				return developeroauthdomain.ErrInvalidClient
			}
			return nil
		},
	)
	if err != nil {
		return developeroauthdomain.TokenPair{}, err
	}
	accessToken, err := service.signAccessToken(grant, now)
	if err != nil {
		return developeroauthdomain.TokenPair{}, err
	}
	return developeroauthdomain.TokenPair{
		AccessToken: accessToken, RefreshToken: replacement.Plaintext,
		ExpiresIn: service.accessTokenTTL, Scopes: append([]string(nil), grant.Scopes...),
	}, nil
}

func (service *Service) RevokeRefreshToken(ctx context.Context, raw string) error {
	prefix, err := service.tokens.ParseLookupPrefix(raw, developeroauthdomain.SecretRefreshToken)
	if err != nil {
		return developeroauthdomain.ErrRefreshToken
	}
	now := service.clock.Now().UTC()
	return service.repository.RevokeRefreshToken(ctx, prefix, now, func(record developeroauthdomain.RefreshToken) error {
		if err := service.tokens.Verify(raw, developeroauthdomain.SecretMaterial{
			ID: record.ID, Kind: developeroauthdomain.SecretRefreshToken,
			LookupPrefix: record.LookupPrefix, Digest: record.Digest, DigestKey: record.DigestKey,
		}); err != nil {
			return developeroauthdomain.ErrRefreshToken
		}
		return nil
	})
}

func ScopeString(scopes []string) string {
	return strings.Join(scopes, " ")
}
