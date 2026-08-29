package developeroauth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	"github.com/google/uuid"
)

const (
	defaultAccessTokenTTL       = 15 * time.Minute
	defaultAuthorizationCodeTTL = 5 * time.Minute
	defaultRefreshTokenTTL      = 30 * 24 * time.Hour
	defaultDynamicClientTTL     = 30 * 24 * time.Hour
	grantTouchInterval          = 15 * time.Minute
)

type Repository interface {
	CreateApplication(context.Context, developeroauthdomain.RegisterApplication) (developeroauthdomain.Application, error)
	GetActiveApplication(context.Context, string, time.Time) (developeroauthdomain.Application, error)
	AuthorizeUser(context.Context, developeroauthdomain.AuthorizeUser) (developeroauthdomain.Grant, error)
	ExchangeAuthorizationCode(
		context.Context,
		developeroauthdomain.ExchangeAuthorizationCode,
		func(developeroauthdomain.AuthorizationCode) error,
	) (developeroauthdomain.Grant, error)
	RotateRefreshToken(
		context.Context,
		developeroauthdomain.RotateRefreshToken,
		func(developeroauthdomain.RefreshToken) error,
	) (developeroauthdomain.Grant, error)
	RevokeRefreshToken(
		context.Context,
		string,
		time.Time,
		func(developeroauthdomain.RefreshToken) error,
	) error
	GetActiveGrant(context.Context, uuid.UUID, uuid.UUID, string, time.Time) (developeroauthdomain.Grant, error)
	TouchGrant(context.Context, uuid.UUID, time.Time, time.Time) error
}

// ApplicationActorRepository is the revocation-aware persistence capability
// needed by client_credentials issuance and access-token verification. It is
// configured only on resources that intentionally release application actors.
type ApplicationActorRepository interface {
	AuthenticateApplication(
		context.Context,
		developeroauthdomain.AuthenticateApplicationCredential,
		func(developeroauthdomain.ClientSecretRecord, developeroauthdomain.ApplicationInstallation) error,
	) (developeroauthdomain.ApplicationInstallation, error)
	GetActiveApplicationInstallation(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		string,
		time.Time,
	) (developeroauthdomain.ApplicationInstallation, error)
	TouchApplicationInstallation(context.Context, uuid.UUID, time.Time, time.Time) error
}

type Clock interface {
	Now() time.Time
}

type WallClock struct{}

func (WallClock) Now() time.Time {
	return time.Now().UTC()
}

type IDGenerator interface {
	NewID() (uuid.UUID, error)
}

type RandomIDGenerator struct{}

func (RandomIDGenerator) NewID() (uuid.UUID, error) {
	return uuid.NewRandom()
}

type Config struct {
	Issuer                 string
	Resource               string
	ScopePolicy            ScopePolicy
	AccessTokenSigningKey  string
	AccessTokenTTL         time.Duration
	AuthorizationCodeTTL   time.Duration
	RefreshTokenTTL        time.Duration
	DynamicClientTTL       time.Duration
	ApplicationActors      ApplicationActorRepository
	ApplicationActorScopes ScopePolicy
}

type Service struct {
	repository             Repository
	tokens                 *TokenManager
	clock                  Clock
	ids                    IDGenerator
	random                 io.Reader
	issuer                 string
	resource               string
	scopePolicy            scopePolicy
	accessTokenSigningKey  []byte
	accessTokenTTL         time.Duration
	authorizationCodeTTL   time.Duration
	refreshTokenTTL        time.Duration
	dynamicClientTTL       time.Duration
	applicationActors      ApplicationActorRepository
	applicationActorScopes scopePolicy
}

func New(
	repository Repository,
	tokens *TokenManager,
	clock Clock,
	ids IDGenerator,
	config Config,
) (*Service, error) {
	return newService(repository, tokens, clock, ids, rand.Reader, config)
}

func newService(
	repository Repository,
	tokens *TokenManager,
	clock Clock,
	ids IDGenerator,
	random io.Reader,
	config Config,
) (*Service, error) {
	if repository == nil {
		return nil, errors.New("developer OAuth repository is required")
	}
	if tokens == nil {
		return nil, errors.New("developer OAuth token manager is required")
	}
	if clock == nil || ids == nil || random == nil {
		return nil, errors.New("developer OAuth clock, ID generator, and random source are required")
	}
	issuer := strings.TrimRight(strings.TrimSpace(config.Issuer), "/")
	resource := strings.TrimSpace(config.Resource)
	if err := validateIssuerAndResource(issuer, resource); err != nil {
		return nil, err
	}
	scopes, err := newScopePolicy(config.ScopePolicy)
	if err != nil {
		return nil, err
	}
	if len(config.AccessTokenSigningKey) < 32 {
		return nil, errors.New("OAuth access-token signing key must contain at least 32 bytes")
	}
	var applicationScopes scopePolicy
	if config.ApplicationActors != nil {
		applicationScopes, err = newScopePolicy(config.ApplicationActorScopes)
		if err != nil {
			return nil, fmt.Errorf("OAuth application-actor scope policy: %w", err)
		}
		if !applicationScopes.requireExplicit {
			return nil, errors.New("OAuth application-actor scopes must be explicitly requested")
		}
	}
	return &Service{
		repository: repository, tokens: tokens, clock: clock, ids: ids, random: random,
		issuer: issuer, resource: resource, scopePolicy: scopes,
		accessTokenSigningKey:  []byte(config.AccessTokenSigningKey),
		accessTokenTTL:         durationOrDefault(config.AccessTokenTTL, defaultAccessTokenTTL),
		authorizationCodeTTL:   durationOrDefault(config.AuthorizationCodeTTL, defaultAuthorizationCodeTTL),
		refreshTokenTTL:        durationOrDefault(config.RefreshTokenTTL, defaultRefreshTokenTTL),
		dynamicClientTTL:       durationOrDefault(config.DynamicClientTTL, defaultDynamicClientTTL),
		applicationActors:      config.ApplicationActors,
		applicationActorScopes: applicationScopes,
	}, nil
}

func (service *Service) Resource() string {
	return service.resource
}

func (service *Service) SupportedScopes() []string {
	if service == nil {
		return nil
	}
	return service.scopePolicy.scopes()
}

func durationOrDefault(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func (service *Service) nextID() (uuid.UUID, error) {
	id, err := service.ids.NewID()
	if err != nil {
		return uuid.Nil, err
	}
	if id == uuid.Nil {
		return uuid.Nil, errors.New("developer OAuth ID generator returned a zero UUID")
	}
	return id, nil
}
