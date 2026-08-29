package developeroauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

const (
	maximumManagedApplicationLifetime = 365 * 24 * time.Hour
	minimumManagedSecretOverlap       = time.Minute
	maximumManagedSecretOverlap       = 24 * time.Hour
	managedSecretIssueAttempts        = 3
)

type ApplicationManagementRepository interface {
	CreateManagedApplication(context.Context, developeroauthdomain.CreateManagedApplication) (developeroauthdomain.ManagedApplication, developeroauthdomain.ClientSecret, error)
	ListManagedApplications(context.Context, uuid.UUID, uuid.UUID) ([]developeroauthdomain.ManagedApplication, error)
	ListClientSecrets(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]developeroauthdomain.ClientSecret, error)
	RotateClientSecret(context.Context, developeroauthdomain.RotateClientSecret) (developeroauthdomain.ClientSecret, error)
	RevokeClientSecret(context.Context, developeroauthdomain.RevokeClientSecret) error
	InstallApplication(context.Context, developeroauthdomain.InstallApplication) (developeroauthdomain.ApplicationInstallation, error)
	ListApplicationInstallations(context.Context, uuid.UUID, uuid.UUID) ([]developeroauthdomain.ApplicationInstallation, error)
	UpdateApplicationInstallation(context.Context, developeroauthdomain.UpdateApplicationInstallation) (developeroauthdomain.ApplicationInstallation, error)
	RevokeApplicationInstallation(context.Context, developeroauthdomain.RevokeApplicationInstallation) error
}

type ApplicationManager struct {
	repository ApplicationManagementRepository
	tokens     *TokenManager
	clock      Clock
	ids        IDGenerator
	random     io.Reader
	resource   string
	scopes     scopePolicy
}

func NewApplicationManager(
	repository ApplicationManagementRepository,
	tokens *TokenManager,
	clock Clock,
	ids IDGenerator,
	resource string,
) (*ApplicationManager, error) {
	return newApplicationManager(repository, tokens, clock, ids, rand.Reader, resource)
}

func newApplicationManager(
	repository ApplicationManagementRepository,
	tokens *TokenManager,
	clock Clock,
	ids IDGenerator,
	random io.Reader,
	resource string,
) (*ApplicationManager, error) {
	if repository == nil || tokens == nil || clock == nil || ids == nil || random == nil {
		return nil, errors.New("OAuth application manager dependencies are required")
	}
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return nil, errors.New("OAuth application manager resource is required")
	}
	scopes, err := newScopePolicy(PublicAPIApplicationActorScopePolicy())
	if err != nil {
		return nil, err
	}
	return &ApplicationManager{
		repository: repository, tokens: tokens, clock: clock, ids: ids,
		random: random, resource: resource, scopes: scopes,
	}, nil
}

type CreateManagedApplicationInput struct {
	Name            string
	RedirectURIs    []string
	ExpiresAt       time.Time
	SecretExpiresAt time.Time
	RequestID       string
}

func (manager *ApplicationManager) CreateManagedApplication(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
	input CreateManagedApplicationInput,
) (developeroauthdomain.IssuedManagedApplication, error) {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return developeroauthdomain.IssuedManagedApplication{}, err
	}
	name, err := validateManagedApplicationName(input.Name)
	if err != nil {
		return developeroauthdomain.IssuedManagedApplication{}, err
	}
	redirects, err := normalizeManagedRedirectURIs(input.RedirectURIs)
	if err != nil {
		return developeroauthdomain.IssuedManagedApplication{}, err
	}
	now := manager.clock.Now().UTC()
	if err := validateManagedExpiry(now, input.ExpiresAt); err != nil {
		return developeroauthdomain.IssuedManagedApplication{}, err
	}
	if err := validateManagedExpiry(now, input.SecretExpiresAt); err != nil || input.SecretExpiresAt.After(input.ExpiresAt) {
		return developeroauthdomain.IssuedManagedApplication{}, developeroauthdomain.ErrInvalidExpiry
	}
	applicationID, err := manager.nextID()
	if err != nil {
		return developeroauthdomain.IssuedManagedApplication{}, err
	}
	auditID, err := manager.nextID()
	if err != nil {
		return developeroauthdomain.IssuedManagedApplication{}, err
	}
	clientID, err := manager.newClientID()
	if err != nil {
		return developeroauthdomain.IssuedManagedApplication{}, err
	}
	for attempt := 0; attempt < managedSecretIssueAttempts; attempt++ {
		secretID, err := manager.nextID()
		if err != nil {
			return developeroauthdomain.IssuedManagedApplication{}, err
		}
		secret, err := manager.tokens.Issue(developeroauthdomain.SecretClientSecret, secretID)
		if err != nil {
			return developeroauthdomain.IssuedManagedApplication{}, err
		}
		audit := managementAudit(
			auditID,
			access,
			"application.registered",
			"application",
			applicationID,
			input.RequestID,
			now,
		)
		audit.ApplicationID = &applicationID
		audit.SecretID = &secretID
		application, metadata, err := manager.repository.CreateManagedApplication(ctx, developeroauthdomain.CreateManagedApplication{
			ApplicationID: applicationID, ClientID: clientID, OwnerWorkspaceID: access.WorkspaceID,
			OwnerUserID: actorID, Name: name, RedirectURIs: redirects, ExpiresAt: input.ExpiresAt.UTC(),
			Secret: secret.Material, SecretExpiresAt: input.SecretExpiresAt.UTC(), CreatedAt: now, Audit: audit,
		})
		if errors.Is(err, developeroauthdomain.ErrSecretPrefixCollision) {
			continue
		}
		if err != nil {
			return developeroauthdomain.IssuedManagedApplication{}, err
		}
		return developeroauthdomain.IssuedManagedApplication{
			Application: application,
			Secret:      developeroauthdomain.IssuedClientSecret{Secret: metadata, Plaintext: secret.Plaintext},
		}, nil
	}
	return developeroauthdomain.IssuedManagedApplication{}, developeroauthdomain.ErrSecretPrefixCollision
}

func (manager *ApplicationManager) ListManagedApplications(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
) ([]developeroauthdomain.ManagedApplication, error) {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return nil, err
	}
	return manager.repository.ListManagedApplications(ctx, access.WorkspaceID, actorID)
}

func (manager *ApplicationManager) ListClientSecrets(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
	applicationID uuid.UUID,
) ([]developeroauthdomain.ClientSecret, error) {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return nil, err
	}
	if applicationID == uuid.Nil {
		return nil, developeroauthdomain.ErrApplicationNotFound
	}
	return manager.repository.ListClientSecrets(ctx, access.WorkspaceID, applicationID, actorID)
}

type RotateClientSecretInput struct {
	ExpiresAt time.Time
	Overlap   time.Duration
	RequestID string
}

func (manager *ApplicationManager) RotateClientSecret(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
	applicationID uuid.UUID,
	input RotateClientSecretInput,
) (developeroauthdomain.IssuedClientSecret, error) {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return developeroauthdomain.IssuedClientSecret{}, err
	}
	if applicationID == uuid.Nil {
		return developeroauthdomain.IssuedClientSecret{}, developeroauthdomain.ErrApplicationNotFound
	}
	now := manager.clock.Now().UTC()
	if err := validateManagedExpiry(now, input.ExpiresAt); err != nil {
		return developeroauthdomain.IssuedClientSecret{}, err
	}
	if input.Overlap < minimumManagedSecretOverlap || input.Overlap > maximumManagedSecretOverlap {
		return developeroauthdomain.IssuedClientSecret{}, developeroauthdomain.ErrInvalidRotationOverlap
	}
	auditID, err := manager.nextID()
	if err != nil {
		return developeroauthdomain.IssuedClientSecret{}, err
	}
	overlapExpiresAt := now.Add(input.Overlap)
	for attempt := 0; attempt < managedSecretIssueAttempts; attempt++ {
		secretID, err := manager.nextID()
		if err != nil {
			return developeroauthdomain.IssuedClientSecret{}, err
		}
		secret, err := manager.tokens.Issue(developeroauthdomain.SecretClientSecret, secretID)
		if err != nil {
			return developeroauthdomain.IssuedClientSecret{}, err
		}
		audit := managementAudit(auditID, access, "client_secret.rotated", "client_secret", secretID, input.RequestID, now)
		audit.ApplicationID = &applicationID
		audit.SecretID = &secretID
		metadata, err := manager.repository.RotateClientSecret(ctx, developeroauthdomain.RotateClientSecret{
			ApplicationID: applicationID, OwnerWorkspaceID: access.WorkspaceID, ActorUserID: actorID,
			Secret: secret.Material, ExpiresAt: input.ExpiresAt.UTC(), OverlapExpiresAt: overlapExpiresAt,
			RotatedAt: now,
			Audit:     audit,
		})
		if errors.Is(err, developeroauthdomain.ErrSecretPrefixCollision) {
			continue
		}
		if err != nil {
			return developeroauthdomain.IssuedClientSecret{}, err
		}
		var previousOverlapExpiresAt *time.Time
		if metadata.RotatedFromID != nil {
			cutoff := overlapExpiresAt
			previousOverlapExpiresAt = &cutoff
		}
		return developeroauthdomain.IssuedClientSecret{
			Secret: metadata, Plaintext: secret.Plaintext,
			PreviousSecretOverlapExpiresAt: previousOverlapExpiresAt,
		}, nil
	}
	return developeroauthdomain.IssuedClientSecret{}, developeroauthdomain.ErrSecretPrefixCollision
}

type RevokeApplicationInput struct {
	Reason    string
	RequestID string
}

func (manager *ApplicationManager) RevokeClientSecret(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
	applicationID uuid.UUID,
	secretID uuid.UUID,
	input RevokeApplicationInput,
) error {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return err
	}
	reason, err := validateManagementReason(input.Reason)
	if err != nil {
		return err
	}
	if applicationID == uuid.Nil || secretID == uuid.Nil {
		return developeroauthdomain.ErrSecretNotFound
	}
	now := manager.clock.Now().UTC()
	auditID, err := manager.nextID()
	if err != nil {
		return err
	}
	audit := managementAudit(auditID, access, "client_secret.revoked", "client_secret", secretID, input.RequestID, now)
	audit.ApplicationID = &applicationID
	audit.SecretID = &secretID
	return manager.repository.RevokeClientSecret(ctx, developeroauthdomain.RevokeClientSecret{
		ApplicationID: applicationID, SecretID: secretID, OwnerWorkspaceID: access.WorkspaceID,
		ActorUserID: actorID, Reason: reason, RevokedAt: now,
		Audit: audit,
	})
}

type InstallApplicationInput struct {
	ClientID  string
	Resource  string
	Scopes    []string
	RequestID string
}

func (manager *ApplicationManager) InstallApplication(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
	input InstallApplicationInput,
) (developeroauthdomain.ApplicationInstallation, error) {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	clientID := strings.TrimSpace(input.ClientID)
	if clientID == "" || clientID != input.ClientID {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInvalidClient
	}
	if input.Resource != manager.resource {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInvalidResource
	}
	scopes, err := manager.scopes.normalize(input.Scopes)
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	installationID, err := manager.nextID()
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	principalID, err := manager.nextID()
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	if principalID == actorID || principalID == installationID {
		return developeroauthdomain.ApplicationInstallation{}, errors.New("OAuth installation principal must be distinct from the installer and installation")
	}
	auditID, err := manager.nextID()
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	now := manager.clock.Now().UTC()
	audit := managementAudit(auditID, access, "installation.created", "installation", installationID, input.RequestID, now)
	audit.InstallationID = &installationID
	audit.PrincipalID = &principalID
	return manager.repository.InstallApplication(ctx, developeroauthdomain.InstallApplication{
		InstallationID: installationID, PrincipalID: principalID, ClientID: clientID,
		WorkspaceID: access.WorkspaceID, InstalledBy: actorID, Resource: manager.resource,
		Scopes: scopes, InstalledAt: now,
		Audit: audit,
	})
}

func (manager *ApplicationManager) ListApplicationInstallations(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
) ([]developeroauthdomain.ApplicationInstallation, error) {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return nil, err
	}
	return manager.repository.ListApplicationInstallations(ctx, access.WorkspaceID, actorID)
}

type UpdateApplicationInstallationInput struct {
	Resource  string
	Scopes    []string
	RequestID string
}

func (manager *ApplicationManager) UpdateApplicationInstallation(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
	installationID uuid.UUID,
	input UpdateApplicationInstallationInput,
) (developeroauthdomain.ApplicationInstallation, error) {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	if installationID == uuid.Nil {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInstallationNotFound
	}
	if input.Resource != manager.resource {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInvalidResource
	}
	scopes, err := manager.scopes.normalize(input.Scopes)
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	auditID, err := manager.nextID()
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	now := manager.clock.Now().UTC()
	audit := managementAudit(auditID, access, "installation.scopes_updated", "installation", installationID, input.RequestID, now)
	audit.InstallationID = &installationID
	audit.PrincipalID = nil // The repository fills the authoritative installation principal while holding the installation lock.
	return manager.repository.UpdateApplicationInstallation(ctx, developeroauthdomain.UpdateApplicationInstallation{
		InstallationID: installationID, WorkspaceID: access.WorkspaceID, ActorUserID: actorID,
		Resource: manager.resource, Scopes: scopes, UpdatedAt: now,
		Audit: audit,
	})
}

func (manager *ApplicationManager) RevokeApplicationInstallation(
	ctx context.Context,
	access developeroauthdomain.ManagementAccess,
	installationID uuid.UUID,
	input RevokeApplicationInput,
) error {
	actorID, err := authorizeApplicationManagement(access)
	if err != nil {
		return err
	}
	if installationID == uuid.Nil {
		return developeroauthdomain.ErrInstallationNotFound
	}
	reason, err := validateManagementReason(input.Reason)
	if err != nil {
		return err
	}
	auditID, err := manager.nextID()
	if err != nil {
		return err
	}
	now := manager.clock.Now().UTC()
	audit := managementAudit(auditID, access, "installation.revoked", "installation", installationID, input.RequestID, now)
	audit.InstallationID = &installationID
	audit.PrincipalID = nil // The repository fills the authoritative installation principal while holding the installation lock.
	return manager.repository.RevokeApplicationInstallation(ctx, developeroauthdomain.RevokeApplicationInstallation{
		InstallationID: installationID, WorkspaceID: access.WorkspaceID, ActorUserID: actorID,
		Reason: reason, RevokedAt: now,
		Audit: audit,
	})
}

func authorizeApplicationManagement(access developeroauthdomain.ManagementAccess) (uuid.UUID, error) {
	err := authorization.AuthorizeWorkspace(authorization.WorkspacePolicyInput{
		Actor: access.Actor, WorkspaceID: access.WorkspaceID, WorkspaceRole: access.WorkspaceRole,
		MinimumWorkspaceRole:  authorization.WorkspaceRoleAdmin,
		RequiredScopes:        []platformauth.Scope{platformauth.ScopeIntegrationsManage},
		AllowedPrincipalKinds: []platformauth.PrincipalKind{platformauth.PrincipalHumanUser},
	})
	if err != nil {
		return uuid.Nil, errors.Join(developeroauthdomain.ErrAccessDenied, err)
	}
	return access.Actor.UserID()
}

func validateManagedApplicationName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 120 {
		return "", developeroauthdomain.ErrInvalidName
	}
	return value, nil
}

func normalizeManagedRedirectURIs(values []string) ([]string, error) {
	if len(values) == 0 {
		return []string{}, nil
	}
	return normalizeRedirectURIs(values)
}

func validateManagedExpiry(now, expiry time.Time) error {
	expiry = expiry.UTC()
	if expiry.Before(now.Add(time.Minute)) || expiry.After(now.Add(maximumManagedApplicationLifetime)) {
		return developeroauthdomain.ErrInvalidExpiry
	}
	return nil
}

func validateManagementReason(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 240 {
		return "", developeroauthdomain.ErrInvalidReason
	}
	return value, nil
}

func managementAudit(
	id uuid.UUID,
	access developeroauthdomain.ManagementAccess,
	operation string,
	subjectType string,
	subjectID uuid.UUID,
	requestID string,
	createdAt time.Time,
) developeroauthdomain.AuditEvent {
	actorID := access.Actor.PrincipalID
	workspaceID := access.WorkspaceID
	return developeroauthdomain.AuditEvent{
		ID: id, WorkspaceID: &workspaceID, UserID: &actorID,
		PrincipalID: &actorID, ActorKind: access.Actor.Kind, ActorID: &actorID,
		Operation: operation, Result: "succeeded", RequestID: requestID,
		SubjectType: subjectType, SubjectID: &subjectID, CreatedAt: createdAt,
	}
}

func (manager *ApplicationManager) nextID() (uuid.UUID, error) {
	id, err := manager.ids.NewID()
	if err != nil {
		return uuid.Nil, err
	}
	if id == uuid.Nil {
		return uuid.Nil, errors.New("OAuth application manager ID generator returned a zero UUID")
	}
	return id, nil
}

func (manager *ApplicationManager) newClientID() (string, error) {
	material := make([]byte, 24)
	if _, err := io.ReadFull(manager.random, material); err != nil {
		return "", fmt.Errorf("generate OAuth client ID: %w", err)
	}
	defer zeroByteSlices([][]byte{material})
	return "f41_oauth_" + base64.RawURLEncoding.EncodeToString(material), nil
}
