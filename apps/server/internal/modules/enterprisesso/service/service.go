package enterprisesso

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	domain "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/oidcclient"
	"github.com/google/uuid"
)

type Repository interface {
	AuthorizeAdmin(context.Context, domain.Scope) error
	Get(context.Context, domain.Scope) (*domain.Connection, error)
	Public(context.Context, string) (domain.Connection, error)
	Create(context.Context, domain.Scope, domain.Connection) (domain.Connection, error)
	Update(context.Context, domain.Scope, domain.Connection, int64) (domain.Connection, error)
	Archive(context.Context, domain.Scope, string) error
	Authenticate(context.Context, domain.Attempt, domain.Identity) (uuid.UUID, error)
	CheckSession(context.Context, domain.Scope, domain.SessionProof) error
}
type Provider interface {
	Validate(context.Context, oidcclient.Credential, string) error
	AuthorizationURL(context.Context, oidcclient.Credential, oidcclient.Attempt, string, string) (string, error)
	Exchange(context.Context, oidcclient.Credential, oidcclient.Attempt, string, string) (oidcclient.Identity, error)
}
type Service struct {
	repository Repository
	vault      *credentialvault.Vault
	provider   Provider
	redirect   string
}

func New(repository Repository, vault *credentialvault.Vault, provider Provider, redirect string) *Service {
	return &Service{repository: repository, vault: vault, provider: provider, redirect: redirect}
}
func scopeValid(scope domain.Scope) error {
	if scope.WorkspaceID == uuid.Nil || scope.ActorID == uuid.Nil {
		return domain.ErrForbidden
	}
	return nil
}
func binding(c domain.Connection) credentialvault.Context {
	return credentialvault.Context{Provider: "enterprise_oidc", TenantID: c.WorkspaceID.String(), SubjectID: c.ID.String(), CredentialType: "client_secret", Generation: strconv.FormatInt(c.Generation, 10)}
}
func (s *Service) Get(ctx context.Context, scope domain.Scope) (*domain.Connection, error) {
	if err := scopeValid(scope); err != nil {
		return nil, err
	}
	return s.repository.Get(ctx, scope)
}
func (s *Service) Create(ctx context.Context, scope domain.Scope, input domain.Create) (domain.Connection, error) {
	if err := scopeValid(scope); err != nil {
		return domain.Connection{}, err
	}
	if err := s.repository.AuthorizeAdmin(ctx, scope); err != nil {
		return domain.Connection{}, err
	}
	issuer := strings.TrimSpace(input.Issuer)
	clientID := strings.TrimSpace(input.ClientID)
	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || len(issuer) > 2048 || clientID == "" || len(clientID) > 512 || !utf8.ValidString(clientID) || len(input.ClientSecret) < 1 || len(input.ClientSecret) > 8192 || s.vault == nil {
		return domain.Connection{}, domain.ErrInvalid
	}
	if err := s.provider.Validate(ctx, oidcclient.Credential{Issuer: issuer, ClientID: clientID, ClientSecret: input.ClientSecret}, s.redirect); err != nil {
		return domain.Connection{}, domain.ErrProvider
	}
	connection := domain.Connection{ID: uuid.New(), WorkspaceID: scope.WorkspaceID, Issuer: issuer, ClientID: clientID, Generation: 1, Enabled: true}
	connection.SecretEnvelope, err = s.vault.Seal(binding(connection), []byte(input.ClientSecret))
	if err != nil {
		return domain.Connection{}, err
	}
	return s.repository.Create(ctx, scope, connection)
}
func (s *Service) Update(ctx context.Context, scope domain.Scope, input domain.Update, proof domain.SessionProof) (domain.Connection, error) {
	current, err := s.Get(ctx, scope)
	if err != nil {
		return domain.Connection{}, err
	}
	if current == nil {
		return domain.Connection{}, domain.ErrNotFound
	}
	if input.ExpectedVersion != current.Version {
		return domain.Connection{}, domain.ErrConflict
	}
	if input.RequireSSO && (!input.Enabled || input.ClientSecret != nil || proof.ConnectionID != current.ID || proof.Generation != current.Generation || proof.AuthenticatedAt.IsZero()) {
		return domain.Connection{}, errors.Join(domain.ErrInvalid, errors.New("test this SSO connection before requiring it"))
	}
	current.Enabled = input.Enabled
	current.RequireSSO = input.RequireSSO
	if input.ClientSecret != nil {
		if len(*input.ClientSecret) < 1 || len(*input.ClientSecret) > 8192 || s.vault == nil {
			return domain.Connection{}, domain.ErrInvalid
		}
		current.Generation++
		current.SecretEnvelope, err = s.vault.Seal(binding(*current), []byte(*input.ClientSecret))
		if err != nil {
			return domain.Connection{}, err
		}
	}
	return s.repository.Update(ctx, scope, *current, input.ExpectedVersion)
}
func (s *Service) Archive(ctx context.Context, scope domain.Scope, reason string) error {
	if err := scopeValid(scope); err != nil {
		return err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" || !utf8.ValidString(reason) || utf8.RuneCountInString(reason) > 240 {
		return domain.ErrInvalid
	}
	return s.repository.Archive(ctx, scope, reason)
}
func (s *Service) Public(ctx context.Context, slug string) (domain.Connection, error) {
	if slug == "" || len(slug) > 120 || strings.ContainsAny(slug, "/\\?#") {
		return domain.Connection{}, domain.ErrInvalid
	}
	return s.repository.Public(ctx, slug)
}
func (s *Service) credential(connection domain.Connection) (oidcclient.Credential, func(), error) {
	if s.vault == nil {
		return oidcclient.Credential{}, nil, domain.ErrProvider
	}
	secret, err := s.vault.Open(binding(connection), connection.SecretEnvelope)
	if err != nil {
		return oidcclient.Credential{}, nil, domain.ErrProvider
	}
	plain := secret.Reveal()
	credential := oidcclient.Credential{Issuer: connection.Issuer, ClientID: connection.ClientID, ClientSecret: string(plain)}
	return credential, func() {
		for index := range plain {
			plain[index] = 0
		}
		secret.Destroy()
	}, nil
}
func (s *Service) AuthorizationURL(ctx context.Context, connection domain.Connection, attempt domain.Attempt, state string) (string, error) {
	credential, release, err := s.credential(connection)
	if err != nil {
		return "", err
	}
	defer release()
	return s.provider.AuthorizationURL(ctx, credential, oidcclient.Attempt{Nonce: attempt.Nonce, Verifier: attempt.Verifier, CreatedAt: attempt.CreatedAt}, state, s.redirect)
}
func (s *Service) Complete(ctx context.Context, attempt domain.Attempt, code string) (uuid.UUID, domain.Identity, error) {
	connection, err := s.Public(ctx, attempt.Slug)
	if err != nil {
		return uuid.Nil, domain.Identity{}, err
	}
	if connection.ID != attempt.ConnectionID || connection.WorkspaceID != attempt.WorkspaceID || connection.Generation != attempt.Generation {
		return uuid.Nil, domain.Identity{}, domain.ErrConflict
	}
	credential, release, err := s.credential(connection)
	if err != nil {
		return uuid.Nil, domain.Identity{}, err
	}
	defer release()
	verified, err := s.provider.Exchange(ctx, credential, oidcclient.Attempt{Nonce: attempt.Nonce, Verifier: attempt.Verifier, CreatedAt: attempt.CreatedAt}, code, s.redirect)
	if err != nil {
		return uuid.Nil, domain.Identity{}, domain.ErrProvider
	}
	identity := domain.Identity{Subject: verified.Subject, Email: verified.Email, AuthenticatedAt: verified.AuthenticatedAt}
	userID, err := s.repository.Authenticate(ctx, attempt, identity)
	return userID, identity, err
}
func (s *Service) CheckSession(ctx context.Context, scope domain.Scope, proof domain.SessionProof) error {
	if err := scopeValid(scope); err != nil {
		return err
	}
	return s.repository.CheckSession(ctx, scope, proof)
}
