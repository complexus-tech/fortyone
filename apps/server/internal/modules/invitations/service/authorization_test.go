package invitations

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

var errInvitationCreateBoundaryReached = errors.New("invitation create boundary reached")

func TestInvitationAdministrationRoleMatrix(t *testing.T) {
	t.Parallel()

	operations := []struct {
		name       string
		run        func(context.Context, *Service, uuid.UUID, uuid.UUID) error
		adminError error
		wasCalled  func(*invitationAuthorizationRepository, *invitationUserStub) bool
	}{
		{
			name: "create",
			run: func(ctx context.Context, service *Service, workspaceID, actorID uuid.UUID) error {
				_, err := service.CreateBulkInvitations(ctx, workspaceID, actorID, []InvitationRequest{{
					Email: "person@example.com", Role: InvitationRoleMember,
				}})
				return err
			},
			adminError: errInvitationCreateBoundaryReached,
			wasCalled: func(_ *invitationAuthorizationRepository, users *invitationUserStub) bool {
				return users.getCalls > 0
			},
		},
		{
			name: "list",
			run: func(ctx context.Context, service *Service, workspaceID, actorID uuid.UUID) error {
				_, err := service.ListInvitations(ctx, workspaceID, actorID)
				return err
			},
			wasCalled: func(repo *invitationAuthorizationRepository, _ *invitationUserStub) bool { return repo.listCalls > 0 },
		},
		{
			name: "revoke",
			run: func(ctx context.Context, service *Service, workspaceID, actorID uuid.UUID) error {
				return service.RevokeInvitation(ctx, workspaceID, actorID, uuid.New())
			},
			wasCalled: func(repo *invitationAuthorizationRepository, _ *invitationUserStub) bool { return repo.revokeCalls > 0 },
		},
	}
	roles := []struct {
		name  string
		role  string
		admin bool
	}{
		{name: "guest", role: InvitationRoleGuest},
		{name: "member", role: InvitationRoleMember},
		{name: "admin", role: InvitationRoleAdmin, admin: true},
	}

	for _, operation := range operations {
		operation := operation
		for _, role := range roles {
			role := role
			t.Run(operation.name+"/"+role.name, func(t *testing.T) {
				t.Parallel()

				repo := &invitationAuthorizationRepository{}
				users := &invitationUserStub{getErr: errInvitationCreateBoundaryReached}
				workspaceAccess := &invitationWorkspaceStub{role: role.role}
				service := New(
					repo,
					newTestInvitationTokenManager(t),
					users,
					workspaceAccess,
				)
				workspaceID, actorID := uuid.New(), uuid.New()

				err := operation.run(context.Background(), service, workspaceID, actorID)
				if role.admin {
					if operation.adminError == nil && err != nil {
						t.Fatalf("admin operation error = %v, want nil", err)
					}
					if operation.adminError != nil && !errors.Is(err, operation.adminError) {
						t.Fatalf("admin operation error = %v, want %v", err, operation.adminError)
					}
					if !operation.wasCalled(repo, users) {
						t.Fatal("admin operation did not reach its protected dependency boundary")
					}
				} else {
					if !errors.Is(err, authorization.ErrWorkspaceAdminRequired) {
						t.Fatalf("%s operation error = %v, want ErrWorkspaceAdminRequired", role.name, err)
					}
					if operation.wasCalled(repo, users) {
						t.Fatal("unauthorized operation reached its repository boundary")
					}
				}
				if workspaceAccess.workspaceID != workspaceID || workspaceAccess.actorID != actorID {
					t.Fatalf("authorization lookup = %s/%s, want %s/%s", workspaceAccess.workspaceID, workspaceAccess.actorID, workspaceID, actorID)
				}
			})
		}
	}
}

func TestCreateBulkInvitationsRejectsUnassignableRolesBeforeMutation(t *testing.T) {
	t.Parallel()

	for _, role := range []string{"", "system", "owner", "Member"} {
		role := role
		t.Run(role, func(t *testing.T) {
			t.Parallel()

			repo := &invitationAuthorizationRepository{}
			users := &invitationUserStub{getErr: errInvitationCreateBoundaryReached}
			service := New(
				repo,
				newTestInvitationTokenManager(t),
				users,
				&invitationWorkspaceStub{role: InvitationRoleAdmin},
			)
			_, err := service.CreateBulkInvitations(context.Background(), uuid.New(), uuid.New(), []InvitationRequest{{
				Email: "person@example.com", Role: role,
			}})
			if !errors.Is(err, ErrInvalidInvitationRole) {
				t.Fatalf("CreateBulkInvitations() error = %v, want ErrInvalidInvitationRole", err)
			}
			if users.getCalls != 0 {
				t.Fatal("invalid invitation role reached the user/dependency boundary")
			}
		})
	}
}

func TestRevokeInvitationForwardsWorkspaceScope(t *testing.T) {
	t.Parallel()

	repo := &invitationAuthorizationRepository{}
	service := New(
		repo,
		newTestInvitationTokenManager(t),
		&invitationUserStub{},
		&invitationWorkspaceStub{role: InvitationRoleAdmin},
	)
	workspaceID, actorID, invitationID := uuid.New(), uuid.New(), uuid.New()

	if err := service.RevokeInvitation(context.Background(), workspaceID, actorID, invitationID); err != nil {
		t.Fatalf("RevokeInvitation() error = %v", err)
	}
	if repo.revokedWorkspaceID != workspaceID || repo.revokedActorID != actorID || repo.revokedInvitationID != invitationID {
		t.Fatalf("repository revoke scope = %s/%s/%s, want %s/%s/%s", repo.revokedWorkspaceID, repo.revokedActorID, repo.revokedInvitationID, workspaceID, actorID, invitationID)
	}
}

func TestListInvitationsForwardsActorAndWorkspaceScope(t *testing.T) {
	t.Parallel()

	repo := &invitationAuthorizationRepository{}
	service := New(
		repo,
		newTestInvitationTokenManager(t),
		&invitationUserStub{},
		&invitationWorkspaceStub{role: InvitationRoleAdmin},
	)
	workspaceID, actorID := uuid.New(), uuid.New()

	if _, err := service.ListInvitations(context.Background(), workspaceID, actorID); err != nil {
		t.Fatalf("ListInvitations() error = %v", err)
	}
	if repo.listedWorkspaceID != workspaceID || repo.listedActorID != actorID {
		t.Fatalf("repository list scope = %s/%s, want %s/%s", repo.listedWorkspaceID, repo.listedActorID, workspaceID, actorID)
	}
}

func TestCreateBulkInvitationsNormalizesAndKeepsBearerOutOfRepositoryCommand(t *testing.T) {
	t.Parallel()

	repo := &invitationAuthorizationRepository{}
	tokens := newTestInvitationTokenManager(t)
	service := New(
		repo,
		tokens,
		&invitationUserStub{},
		&invitationWorkspaceStub{role: InvitationRoleAdmin},
	)
	now := time.Date(2026, time.August, 28, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	teamID := uuid.New()
	workspaceID, actorID := uuid.New(), uuid.New()

	created, err := service.CreateBulkInvitations(context.Background(), workspaceID, actorID, []InvitationRequest{{
		Email: "  Person@Example.COM ", Role: InvitationRoleMember, TeamIDs: []uuid.UUID{teamID, teamID},
	}})
	if err != nil {
		t.Fatalf("CreateBulkInvitations() error = %v", err)
	}
	if len(created) != 1 || len(repo.createdCommands) != 1 {
		t.Fatalf("created/results = %d/%d, want 1/1", len(created), len(repo.createdCommands))
	}
	if repo.createdActorID != actorID || repo.createdCommands[0].Invitation.WorkspaceID != workspaceID {
		t.Fatalf("repository create scope = %s/%s, want %s/%s", repo.createdCommands[0].Invitation.WorkspaceID, repo.createdActorID, workspaceID, actorID)
	}
	command := repo.createdCommands[0]
	if command.Invitation.Email != "person@example.com" {
		t.Fatalf("normalized email = %q", command.Invitation.Email)
	}
	if len(command.Invitation.TeamIDs) != 1 || command.Invitation.TeamIDs[0] != teamID {
		t.Fatalf("deduplicated teams = %v", command.Invitation.TeamIDs)
	}
	if command.Invitation.ExpiresAt != now.Add(invitationLifetime) {
		t.Fatalf("expiration = %s, want %s", command.Invitation.ExpiresAt, now.Add(invitationLifetime))
	}
	rawToken, err := tokens.Restore(command.Token)
	if err != nil {
		t.Fatalf("restore generated token at test boundary: %v", err)
	}
	payload, err := json.Marshal(command.EmailOutbox)
	if err != nil {
		t.Fatalf("marshal email outbox payload: %v", err)
	}
	if strings.Contains(string(payload), rawToken) || strings.Contains(string(payload), `"token"`) {
		t.Fatal("repository email outbox command contains a raw invitation token")
	}
}

func TestCreateBulkInvitationsRejectsNormalizedDuplicatesAndBoundedBatch(t *testing.T) {
	t.Parallel()

	service := New(
		&invitationAuthorizationRepository{},
		newTestInvitationTokenManager(t),
		&invitationUserStub{},
		&invitationWorkspaceStub{role: InvitationRoleAdmin},
	)
	_, err := service.CreateBulkInvitations(context.Background(), uuid.New(), uuid.New(), []InvitationRequest{
		{Email: "person@example.com", Role: InvitationRoleMember},
		{Email: " PERSON@example.com ", Role: InvitationRoleMember},
	})
	if !errors.Is(err, ErrDuplicateInvitation) {
		t.Fatalf("duplicate normalized email error = %v, want ErrDuplicateInvitation", err)
	}

	requests := make([]InvitationRequest, maximumBulkInvitations+1)
	_, err = service.CreateBulkInvitations(context.Background(), uuid.New(), uuid.New(), requests)
	if !errors.Is(err, ErrTooManyInvitations) {
		t.Fatalf("oversized batch error = %v, want ErrTooManyInvitations", err)
	}
}

func newTestInvitationTokenManager(t *testing.T) *InvitationTokenManager {
	t.Helper()
	manager, err := NewInvitationTokenManager(InvitationTokenConfig{Current: InvitationTokenKey{
		ID:     "test-v1",
		Secret: "test-invitation-hmac-key-with-at-least-32-bytes",
	}})
	if err != nil {
		t.Fatalf("create test invitation token manager: %v", err)
	}
	return manager
}

type invitationAuthorizationRepository struct {
	Repository
	createdCommands     []NewWorkspaceInvitation
	createdActorID      uuid.UUID
	listCalls           int
	listedWorkspaceID   uuid.UUID
	listedActorID       uuid.UUID
	revokeCalls         int
	revokedWorkspaceID  uuid.UUID
	revokedActorID      uuid.UUID
	revokedInvitationID uuid.UUID
}

func (r *invitationAuthorizationRepository) CreateBulkInvitations(_ context.Context, actorID uuid.UUID, commands []NewWorkspaceInvitation) ([]CoreWorkspaceInvitation, error) {
	r.createdActorID = actorID
	r.createdCommands = append(r.createdCommands, commands...)
	created := make([]CoreWorkspaceInvitation, 0, len(commands))
	for _, command := range commands {
		created = append(created, command.Invitation)
	}
	return created, nil
}

func (r *invitationAuthorizationRepository) ListInvitations(_ context.Context, workspaceID, actorID uuid.UUID, _ time.Time) ([]CoreWorkspaceInvitation, error) {
	r.listCalls++
	r.listedWorkspaceID = workspaceID
	r.listedActorID = actorID
	return []CoreWorkspaceInvitation{}, nil
}

func (r *invitationAuthorizationRepository) RevokeInvitation(_ context.Context, workspaceID, actorID, invitationID uuid.UUID, _ time.Time) error {
	r.revokeCalls++
	r.revokedWorkspaceID = workspaceID
	r.revokedActorID = actorID
	r.revokedInvitationID = invitationID
	return nil
}

type invitationUserStub struct {
	UserService
	getCalls int
	getErr   error
}

func (s *invitationUserStub) GetUser(_ context.Context, userID uuid.UUID) (users.CoreUser, error) {
	s.getCalls++
	return users.CoreUser{ID: userID, FullName: "Test Admin"}, s.getErr
}

type invitationWorkspaceStub struct {
	WorkspaceService
	role        string
	err         error
	workspaceID uuid.UUID
	actorID     uuid.UUID
}

func (s *invitationWorkspaceStub) Get(_ context.Context, workspaceID, actorID uuid.UUID) (workspaces.CoreWorkspace, error) {
	s.workspaceID = workspaceID
	s.actorID = actorID
	return workspaces.CoreWorkspace{ID: workspaceID, Name: "Test Workspace", UserRole: s.role}, s.err
}
