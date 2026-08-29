package developeroauth

import (
	"bytes"
	"context"
	"testing"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testApplicationResource = "https://api.fortyone.app/api/v1"

type applicationManagementRepositoryStub struct {
	createCommand          *developeroauthdomain.CreateManagedApplication
	rotateCommand          *developeroauthdomain.RotateClientSecret
	installCommand         *developeroauthdomain.InstallApplication
	updateCommand          *developeroauthdomain.UpdateApplicationInstallation
	revokeSecret           *developeroauthdomain.RevokeClientSecret
	revokeInstall          *developeroauthdomain.RevokeApplicationInstallation
	createPrefixCollisions int
	rotatePrefixCollisions int
	createCalls            int
	rotateCalls            int
	listCalls              int
}

func (repository *applicationManagementRepositoryStub) CreateManagedApplication(
	_ context.Context,
	command developeroauthdomain.CreateManagedApplication,
) (developeroauthdomain.ManagedApplication, developeroauthdomain.ClientSecret, error) {
	repository.createCalls++
	repository.createCommand = &command
	if repository.createPrefixCollisions > 0 {
		repository.createPrefixCollisions--
		return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, developeroauthdomain.ErrSecretPrefixCollision
	}
	application := developeroauthdomain.ManagedApplication{
		Application: developeroauthdomain.Application{
			ID: command.ApplicationID, ClientID: command.ClientID, Name: command.Name,
			RegistrationKind: "confidential", RedirectURIs: append([]string(nil), command.RedirectURIs...),
			ExpiresAt: command.ExpiresAt, CreatedAt: command.CreatedAt,
		},
		OwnerWorkspaceID: command.OwnerWorkspaceID, OwnerUserID: command.OwnerUserID,
		Status: "active", UpdatedAt: command.CreatedAt,
	}
	secret := developeroauthdomain.ClientSecret{
		ID: command.Secret.ID, ApplicationID: command.ApplicationID,
		LookupPrefix: command.Secret.LookupPrefix, ExpiresAt: command.SecretExpiresAt,
		CreatedAt: command.CreatedAt,
	}
	return application, secret, nil
}

func (repository *applicationManagementRepositoryStub) ListManagedApplications(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
) ([]developeroauthdomain.ManagedApplication, error) {
	repository.listCalls++
	return []developeroauthdomain.ManagedApplication{}, nil
}

func (repository *applicationManagementRepositoryStub) ListClientSecrets(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
) ([]developeroauthdomain.ClientSecret, error) {
	return []developeroauthdomain.ClientSecret{}, nil
}

func (repository *applicationManagementRepositoryStub) RotateClientSecret(
	_ context.Context,
	command developeroauthdomain.RotateClientSecret,
) (developeroauthdomain.ClientSecret, error) {
	repository.rotateCalls++
	repository.rotateCommand = &command
	if repository.rotatePrefixCollisions > 0 {
		repository.rotatePrefixCollisions--
		return developeroauthdomain.ClientSecret{}, developeroauthdomain.ErrSecretPrefixCollision
	}
	previousID := uuid.New()
	return developeroauthdomain.ClientSecret{
		ID: command.Secret.ID, ApplicationID: command.ApplicationID,
		LookupPrefix: command.Secret.LookupPrefix, ExpiresAt: command.ExpiresAt,
		RotatedFromID: &previousID, CreatedAt: command.RotatedAt,
	}, nil
}

func (repository *applicationManagementRepositoryStub) RevokeClientSecret(
	_ context.Context,
	command developeroauthdomain.RevokeClientSecret,
) error {
	repository.revokeSecret = &command
	return nil
}

func (repository *applicationManagementRepositoryStub) InstallApplication(
	_ context.Context,
	command developeroauthdomain.InstallApplication,
) (developeroauthdomain.ApplicationInstallation, error) {
	repository.installCommand = &command
	return developeroauthdomain.ApplicationInstallation{
		ID: command.InstallationID, ApplicationID: uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		ClientID: command.ClientID, WorkspaceID: command.WorkspaceID, PrincipalID: command.PrincipalID,
		Resource: command.Resource, Scopes: append([]string(nil), command.Scopes...), Status: "active",
		InstalledBy: command.InstalledBy, CreatedAt: command.InstalledAt, UpdatedAt: command.InstalledAt,
	}, nil
}

func (repository *applicationManagementRepositoryStub) ListApplicationInstallations(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
) ([]developeroauthdomain.ApplicationInstallation, error) {
	return []developeroauthdomain.ApplicationInstallation{}, nil
}

func (repository *applicationManagementRepositoryStub) UpdateApplicationInstallation(
	_ context.Context,
	command developeroauthdomain.UpdateApplicationInstallation,
) (developeroauthdomain.ApplicationInstallation, error) {
	repository.updateCommand = &command
	return developeroauthdomain.ApplicationInstallation{
		ID: command.InstallationID, WorkspaceID: command.WorkspaceID,
		Resource: command.Resource, Scopes: append([]string(nil), command.Scopes...), Status: "active",
		UpdatedAt: command.UpdatedAt,
	}, nil
}

func (repository *applicationManagementRepositoryStub) RevokeApplicationInstallation(
	_ context.Context,
	command developeroauthdomain.RevokeApplicationInstallation,
) error {
	repository.revokeInstall = &command
	return nil
}

func TestApplicationManagerCreatesConfidentialApplicationAndRevealsSecretOnce(t *testing.T) {
	t.Parallel()
	manager, repository, now := newApplicationManagerFixture(t)
	workspaceID := uuid.New()
	userID := uuid.New()
	access := applicationManagementAccess(t, userID, workspaceID)

	issued, err := manager.CreateManagedApplication(context.Background(), access, CreateManagedApplicationInput{
		Name: "  Release automation  ", RedirectURIs: nil,
		ExpiresAt: now.Add(30 * 24 * time.Hour), SecretExpiresAt: now.Add(24 * time.Hour),
		RequestID: "request-create",
	})
	require.NoError(t, err)
	require.NotNil(t, repository.createCommand)
	require.Equal(t, "Release automation", repository.createCommand.Name)
	require.Empty(t, repository.createCommand.RedirectURIs)
	require.Equal(t, workspaceID, repository.createCommand.OwnerWorkspaceID)
	require.Equal(t, userID, repository.createCommand.OwnerUserID)
	require.Equal(t, developeroauthdomain.SecretClientSecret, repository.createCommand.Secret.Kind)
	require.NotEmpty(t, repository.createCommand.Secret.Digest)
	require.NotEmpty(t, repository.createCommand.Secret.LookupPrefix)
	require.Equal(t, repository.createCommand.Secret.ID, issued.Secret.Secret.ID)
	require.Equal(t, repository.createCommand.ClientID, issued.Application.ClientID)
	require.Equal(t, "confidential", issued.Application.RegistrationKind)
	require.Contains(t, issued.Secret.Plaintext.Reveal(), clientSecretHeader)
	require.Equal(t, "[REDACTED]", issued.Secret.Plaintext.String())
	require.NotContains(t, string(repository.createCommand.Secret.Digest), issued.Secret.Plaintext.Reveal())
	require.Equal(t, platformauth.PrincipalHumanUser, repository.createCommand.Audit.ActorKind)
	require.Equal(t, userID, *repository.createCommand.Audit.ActorID)
	require.Equal(t, userID, *repository.createCommand.Audit.PrincipalID)
	require.Equal(t, userID, *repository.createCommand.Audit.UserID)
	require.Equal(t, workspaceID, *repository.createCommand.Audit.WorkspaceID)
	require.Equal(t, repository.createCommand.ApplicationID, *repository.createCommand.Audit.ApplicationID)
	require.Equal(t, repository.createCommand.Secret.ID, *repository.createCommand.Audit.SecretID)
	require.Equal(t, repository.createCommand.ApplicationID, *repository.createCommand.Audit.SubjectID)
	require.Equal(t, "request-create", repository.createCommand.Audit.RequestID)
}

func TestApplicationManagerRequiresCurrentHumanWorkspaceAdmin(t *testing.T) {
	t.Parallel()
	manager, repository, _ := newApplicationManagerFixture(t)
	workspaceID := uuid.New()
	otherWorkspaceID := uuid.New()
	userID := uuid.New()

	limitedHuman, err := platformauth.NewActor(
		userID,
		platformauth.PrincipalHumanUser,
		uuid.Nil,
		platformauth.MustScopeSet(platformauth.ScopeWorkspacesRead),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	limitedHuman, err = limitedHuman.WithWorkspace(workspaceID)
	require.NoError(t, err)

	serviceActor, err := platformauth.NewActor(
		uuid.New(),
		platformauth.PrincipalServiceAccount,
		uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeIntegrationsManage),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	serviceActor, err = serviceActor.WithWorkspace(workspaceID)
	require.NoError(t, err)

	tests := []struct {
		name   string
		access developeroauthdomain.ManagementAccess
	}{
		{
			name: "member",
			access: developeroauthdomain.ManagementAccess{
				Actor:       applicationManagementAccess(t, userID, workspaceID).Actor,
				WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleMember,
			},
		},
		{
			name: "wrong workspace",
			access: developeroauthdomain.ManagementAccess{
				Actor:       applicationManagementAccess(t, userID, otherWorkspaceID).Actor,
				WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
		},
		{
			name: "human credential lacks scope",
			access: developeroauthdomain.ManagementAccess{
				Actor: limitedHuman, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
		},
		{
			name: "machine principal",
			access: developeroauthdomain.ManagementAccess{
				Actor: serviceActor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, err := manager.ListManagedApplications(context.Background(), test.access)
			require.ErrorIs(t, err, developeroauthdomain.ErrAccessDenied)
		})
	}
	require.Zero(t, repository.listCalls)
}

func TestApplicationManagerRotationRequiresExplicitBoundedOverlap(t *testing.T) {
	t.Parallel()
	manager, repository, now := newApplicationManagerFixture(t)
	access := applicationManagementAccess(t, uuid.New(), uuid.New())
	applicationID := uuid.New()

	for _, overlap := range []time.Duration{0, time.Minute - time.Nanosecond, 24*time.Hour + time.Nanosecond} {
		_, err := manager.RotateClientSecret(context.Background(), access, applicationID, RotateClientSecretInput{
			ExpiresAt: now.Add(24 * time.Hour), Overlap: overlap,
		})
		require.ErrorIs(t, err, developeroauthdomain.ErrInvalidRotationOverlap)
	}
	require.Nil(t, repository.rotateCommand)

	issued, err := manager.RotateClientSecret(context.Background(), access, applicationID, RotateClientSecretInput{
		ExpiresAt: now.Add(48 * time.Hour), Overlap: 15 * time.Minute, RequestID: "request-rotate",
	})
	require.NoError(t, err)
	require.NotNil(t, repository.rotateCommand)
	require.Equal(t, now.Add(15*time.Minute), repository.rotateCommand.OverlapExpiresAt)
	require.Equal(t, now.Add(48*time.Hour), repository.rotateCommand.ExpiresAt)
	require.Equal(t, repository.rotateCommand.Secret.ID, issued.Secret.ID)
	require.Nil(t, issued.Secret.OverlapExpiresAt, "the cutoff belongs to the previous secret generation")
	require.NotNil(t, issued.PreviousSecretOverlapExpiresAt)
	require.Equal(t, now.Add(15*time.Minute), *issued.PreviousSecretOverlapExpiresAt)
	require.NotEmpty(t, issued.Plaintext.Reveal())
	require.Equal(t, "client_secret.rotated", repository.rotateCommand.Audit.Operation)
	require.Equal(t, repository.rotateCommand.Secret.ID, *repository.rotateCommand.Audit.SubjectID)
}

func TestApplicationManagerInstallationUsesDedicatedPrincipalAndNarrowScope(t *testing.T) {
	t.Parallel()
	manager, repository, _ := newApplicationManagerFixture(t)
	workspaceID := uuid.New()
	installerID := uuid.New()
	access := applicationManagementAccess(t, installerID, workspaceID)

	installation, err := manager.InstallApplication(context.Background(), access, InstallApplicationInput{
		ClientID: "f41_oauth_managed_client", Resource: testApplicationResource,
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, RequestID: "request-install",
	})
	require.NoError(t, err)
	require.NotNil(t, repository.installCommand)
	require.Equal(t, workspaceID, repository.installCommand.WorkspaceID)
	require.Equal(t, installerID, repository.installCommand.InstalledBy)
	require.NotEqual(t, installerID, repository.installCommand.PrincipalID)
	require.Equal(t, repository.installCommand.PrincipalID, installation.PrincipalID)
	require.Equal(t, installerID, installation.InstalledBy)
	require.Equal(t, []string{string(platformauth.ScopeStoriesWrite)}, repository.installCommand.Scopes)
	require.Equal(t, platformauth.PrincipalHumanUser, repository.installCommand.Audit.ActorKind)
	require.Equal(t, installerID, *repository.installCommand.Audit.ActorID)
	require.Equal(t, repository.installCommand.PrincipalID, *repository.installCommand.Audit.PrincipalID)
	require.Equal(t, repository.installCommand.InstallationID, *repository.installCommand.Audit.InstallationID)
	require.Equal(t, repository.installCommand.InstallationID, *repository.installCommand.Audit.SubjectID)
	require.NotEqual(t, repository.installCommand.PrincipalID, *repository.installCommand.Audit.ActorID)

	for _, scopes := range [][]string{
		{string(platformauth.ScopeStoriesRead)},
		{string(platformauth.ScopeWebhooksManage)},
		{developeroauthdomain.ScopeOfflineAccess},
		{developeroauthdomain.ScopeMCPAccess},
		{},
	} {
		_, err := manager.InstallApplication(context.Background(), access, InstallApplicationInput{
			ClientID: "f41_oauth_managed_client", Resource: testApplicationResource, Scopes: scopes,
		})
		require.ErrorIs(t, err, developeroauthdomain.ErrInvalidScope)
	}

	_, err = manager.InstallApplication(context.Background(), access, InstallApplicationInput{
		ClientID: "f41_oauth_managed_client", Resource: testApplicationResource + "/",
		Scopes: []string{string(platformauth.ScopeStoriesWrite)},
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidResource)
}

func TestApplicationManagerUpdatesAndRevokesWithinWorkspace(t *testing.T) {
	t.Parallel()
	manager, repository, now := newApplicationManagerFixture(t)
	workspaceID := uuid.New()
	actorID := uuid.New()
	access := applicationManagementAccess(t, actorID, workspaceID)
	installationID := uuid.New()

	updated, err := manager.UpdateApplicationInstallation(
		context.Background(),
		access,
		installationID,
		UpdateApplicationInstallationInput{
			Resource: testApplicationResource, Scopes: []string{string(platformauth.ScopeStoriesWrite)},
			RequestID: "request-update",
		},
	)
	require.NoError(t, err)
	require.Equal(t, installationID, updated.ID)
	require.Equal(t, workspaceID, repository.updateCommand.WorkspaceID)
	require.Equal(t, actorID, repository.updateCommand.ActorUserID)
	require.Equal(t, now, repository.updateCommand.UpdatedAt)

	err = manager.RevokeApplicationInstallation(context.Background(), access, installationID, RevokeApplicationInput{
		Reason: "  no longer approved  ", RequestID: "request-revoke-installation",
	})
	require.NoError(t, err)
	require.Equal(t, "no longer approved", repository.revokeInstall.Reason)
	require.Equal(t, workspaceID, repository.revokeInstall.WorkspaceID)
	require.Equal(t, actorID, repository.revokeInstall.ActorUserID)

	applicationID := uuid.New()
	secretID := uuid.New()
	err = manager.RevokeClientSecret(context.Background(), access, applicationID, secretID, RevokeApplicationInput{
		Reason: " compromised ", RequestID: "request-revoke-secret",
	})
	require.NoError(t, err)
	require.Equal(t, applicationID, repository.revokeSecret.ApplicationID)
	require.Equal(t, secretID, repository.revokeSecret.SecretID)
	require.Equal(t, "compromised", repository.revokeSecret.Reason)

	err = manager.RevokeApplicationInstallation(context.Background(), access, installationID, RevokeApplicationInput{})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidReason)
	err = manager.RevokeClientSecret(context.Background(), access, applicationID, secretID, RevokeApplicationInput{})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidReason)
}

func TestApplicationManagerRetriesOnlyLookupPrefixCollisions(t *testing.T) {
	t.Parallel()

	manager, repository, now := newApplicationManagerFixture(t)
	repository.createPrefixCollisions = 1
	access := applicationManagementAccess(t, uuid.New(), uuid.New())
	created, err := manager.CreateManagedApplication(context.Background(), access, CreateManagedApplicationInput{
		Name: "Retrying app", ExpiresAt: now.Add(48 * time.Hour), SecretExpiresAt: now.Add(24 * time.Hour),
	})
	require.NoError(t, err)
	require.Equal(t, 2, repository.createCalls)
	require.NotEmpty(t, created.Secret.Plaintext.Reveal())

	repository.rotatePrefixCollisions = 1
	rotated, err := manager.RotateClientSecret(context.Background(), access, created.Application.ID, RotateClientSecretInput{
		ExpiresAt: now.Add(24 * time.Hour), Overlap: time.Minute,
	})
	require.NoError(t, err)
	require.Equal(t, 2, repository.rotateCalls)
	require.NotEmpty(t, rotated.Plaintext.Reveal())
}

func newApplicationManagerFixture(
	t *testing.T,
) (*ApplicationManager, *applicationManagementRepositoryStub, time.Time) {
	t.Helper()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	repository := &applicationManagementRepositoryStub{}
	tokens, err := newTokenManager(TokenKeyringConfig{
		Active: developeroauthdomain.DigestKeyRef{ID: "management-test"},
		Keys: []DigestKey{{
			Ref:      developeroauthdomain.DigestKeyRef{ID: "management-test"},
			Material: bytes.Repeat([]byte{0x6b}, digestKeyBytes),
		}},
	}, bytes.NewReader(testRandomBytes(4096)))
	require.NoError(t, err)
	manager, err := newApplicationManager(
		repository,
		tokens,
		fixedClock{now: now},
		&sequentialIDs{},
		bytes.NewReader(testRandomBytes(1024)),
		testApplicationResource,
	)
	require.NoError(t, err)
	return manager, repository, now
}

func applicationManagementAccess(
	t *testing.T,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) developeroauthdomain.ManagementAccess {
	t.Helper()
	actor, err := platformauth.NewHumanActor(userID).WithWorkspace(workspaceID)
	require.NoError(t, err)
	return developeroauthdomain.ManagementAccess{
		Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}
}
