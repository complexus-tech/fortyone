package github

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

func TestGitHubWorkspaceAdministrationRoleMatrix(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	for _, test := range []struct {
		name string
		role authorization.WorkspaceRole
		err  error
	}{
		{name: "guest", role: authorization.WorkspaceRoleGuest, err: authorization.ErrWorkspaceAdminRequired},
		{name: "member", role: authorization.WorkspaceRoleMember, err: authorization.ErrWorkspaceAdminRequired},
		{name: "admin", role: authorization.WorkspaceRoleAdmin},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			roles := &githubWorkspaceRoleStub{role: test.role}
			service := &Service{workspaceRoles: roles}
			err := service.requireWorkspaceAdmin(context.Background(), workspaceID, actorID)
			if test.err == nil && err != nil {
				t.Fatalf("requireWorkspaceAdmin() error = %v, want nil", err)
			}
			if test.err != nil && !errors.Is(err, test.err) {
				t.Fatalf("requireWorkspaceAdmin() error = %v, want %v", err, test.err)
			}
			if roles.workspaceID != workspaceID || roles.actorID != actorID {
				t.Fatalf("role lookup = %s/%s, want %s/%s", roles.workspaceID, roles.actorID, workspaceID, actorID)
			}
		})
	}
}

func TestGitHubAdminMutationsReauthorizeBeforeValidationOrSideEffects(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	for _, role := range []authorization.WorkspaceRole{
		authorization.WorkspaceRoleGuest,
		authorization.WorkspaceRoleMember,
	} {
		role := role
		for _, operation := range githubAdminMutationOperations(workspaceID, actorID) {
			operation := operation
			t.Run(string(role)+"/"+operation.name, func(t *testing.T) {
				t.Parallel()

				roles := &githubWorkspaceRoleStub{role: role}
				service := &Service{
					workspaceRoles: roles,
					oauthStates:    newGitHubOAuthStateStoreStub(),
					cfg: Config{
						AppID: 1, AppSlug: "test-app", RedirectURL: "https://example.com/setup", WebhookPayloadSecret: "test-secret",
					},
					privateKey: &rsa.PrivateKey{},
				}
				err := operation.run(service)
				if !errors.Is(err, authorization.ErrWorkspaceAdminRequired) {
					t.Fatalf("%s error = %v, want ErrWorkspaceAdminRequired", operation.name, err)
				}
				if roles.calls != 1 {
					t.Fatalf("workspace role lookups = %d, want 1", roles.calls)
				}
			})
		}
	}
}

type githubAdminMutationOperation struct {
	name string
	run  func(*Service) error
}

func githubAdminMutationOperations(workspaceID, actorID uuid.UUID) []githubAdminMutationOperation {
	return []githubAdminMutationOperation{
		{
			name: "create install session",
			run: func(service *Service) error {
				_, err := service.CreateInstallSession(context.Background(), workspaceID, actorID, "test-workspace")
				return err
			},
		},
		{
			name: "complete install callback",
			run: func(service *Service) error {
				state, err := service.createInstallState(context.Background(), workspaceID, actorID, "test-workspace")
				if err != nil {
					return err
				}
				_, err = service.HandleSetup(context.Background(), 42, state)
				return err
			},
		},
		{
			name: "resync repositories",
			run: func(service *Service) error {
				return service.ResyncRepositories(context.Background(), workspaceID, actorID)
			},
		},
		{
			name: "update workspace settings",
			run: func(service *Service) error {
				_, err := service.UpdateWorkspaceSettings(context.Background(), workspaceID, actorID, CoreUpdateWorkspaceSettingsInput{})
				return err
			},
		},
		{
			name: "create issue sync link",
			run: func(service *Service) error {
				_, err := service.CreateIssueSyncLink(context.Background(), workspaceID, actorID, CoreIssueSyncLinkInput{})
				return err
			},
		},
		{
			name: "update issue sync link",
			run: func(service *Service) error {
				_, err := service.UpdateIssueSyncLink(context.Background(), workspaceID, actorID, uuid.New(), CoreUpdateIssueSyncLinkInput{})
				return err
			},
		},
		{
			name: "delete issue sync link",
			run: func(service *Service) error {
				return service.DeleteIssueSyncLink(context.Background(), workspaceID, actorID, uuid.New())
			},
		},
		{
			name: "update team settings",
			run: func(service *Service) error {
				_, err := service.UpdateTeamSettings(context.Background(), workspaceID, actorID, uuid.New(), CoreUpdateTeamGitHubSettings{})
				return err
			},
		},
	}
}

func TestGitHubWorkspaceAdministrationTreatsMissingMembershipAsForbidden(t *testing.T) {
	t.Parallel()

	service := &Service{workspaceRoles: &githubWorkspaceRoleStub{err: sql.ErrNoRows}}
	err := service.requireWorkspaceAdmin(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, authorization.ErrWorkspaceAdminRequired) {
		t.Fatalf("missing membership error = %v, want ErrWorkspaceAdminRequired", err)
	}
}

type githubWorkspaceRoleStub struct {
	role        authorization.WorkspaceRole
	err         error
	workspaceID uuid.UUID
	actorID     uuid.UUID
	calls       int
}

func (s *githubWorkspaceRoleStub) GetWorkspaceRole(_ context.Context, workspaceID, actorID uuid.UUID) (authorization.WorkspaceRole, error) {
	s.calls++
	s.workspaceID = workspaceID
	s.actorID = actorID
	return s.role, s.err
}
