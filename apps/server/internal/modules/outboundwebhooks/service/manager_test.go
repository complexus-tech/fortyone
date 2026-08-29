package outboundwebhooksservice

import (
	"context"
	"errors"
	"testing"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type endpointRepositoryStub struct {
	created     outboundwebhooksdomain.CreateEndpoint
	createCalls int
	endpoint    outboundwebhooksdomain.Endpoint
	listed      []outboundwebhooksdomain.Endpoint
	listCursor  *outboundwebhooksdomain.EndpointCursor
	listSize    int
	createErr   error
}

func (repository *endpointRepositoryStub) CreateEndpoint(
	_ context.Context,
	input outboundwebhooksdomain.CreateEndpoint,
	secretEnvelope string,
) (outboundwebhooksdomain.Endpoint, error) {
	repository.createCalls++
	repository.created = input
	if secretEnvelope == "" {
		return outboundwebhooksdomain.Endpoint{}, errors.New("secret envelope is empty")
	}
	if repository.createErr != nil {
		return outboundwebhooksdomain.Endpoint{}, repository.createErr
	}
	endpoint := repository.endpoint
	endpoint.ID = input.ID
	endpoint.WorkspaceID = input.WorkspaceID
	endpoint.OwnerPrincipalID = input.OwnerPrincipalID
	endpoint.Name = input.Name
	endpoint.URL = input.URL
	endpoint.Subscriptions = append([]outboundwebhooksdomain.EventType(nil), input.Subscriptions...)
	endpoint.Status = outboundwebhooksdomain.EndpointActive
	endpoint.SecretGeneration = 1
	return endpoint, nil
}

func (repository *endpointRepositoryStub) GetEndpoint(context.Context, uuid.UUID, uuid.UUID) (outboundwebhooksdomain.Endpoint, error) {
	return repository.endpoint, nil
}

func (repository *endpointRepositoryStub) ListEndpoints(_ context.Context, _ uuid.UUID, cursor *outboundwebhooksdomain.EndpointCursor, pageSize int) ([]outboundwebhooksdomain.Endpoint, error) {
	repository.listCursor = cursor
	repository.listSize = pageSize
	return append([]outboundwebhooksdomain.Endpoint(nil), repository.listed...), nil
}

func (*endpointRepositoryStub) ReplaceSubscriptions(context.Context, platformauth.Actor, uuid.UUID, uuid.UUID, uuid.UUID, []outboundwebhooksdomain.EventType, time.Time, string) error {
	return nil
}

func (*endpointRepositoryStub) DisableEndpoint(context.Context, platformauth.Actor, uuid.UUID, uuid.UUID, uuid.UUID, string, string, time.Time) error {
	return nil
}

func (*endpointRepositoryStub) RotateEndpointSecret(context.Context, platformauth.Actor, uuid.UUID, uuid.UUID, uuid.UUID, int, string, time.Time, time.Time, string) (int, error) {
	return 2, nil
}

type principalResolverStub struct {
	id    uuid.UUID
	calls int
}

func (resolver *principalResolverStub) ResolveHumanPrincipal(context.Context, platformauth.Actor, uuid.UUID, authorization.WorkspaceRole, string) (uuid.UUID, error) {
	resolver.calls++
	return resolver.id, nil
}

type endpointValidatorStub struct {
	canonical string
	err       error
	calls     int
}

func (validator *endpointValidatorStub) Validate(_ context.Context, _ string) (string, error) {
	validator.calls++
	return validator.canonical, validator.err
}

func TestAuthorizeManagementFailsClosed(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()

	personalToken := newTestActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeWebhooksManage)
	serviceAccount := newTestActor(t, platformauth.PrincipalServiceAccount, workspaceID, platformauth.ScopeWebhooksManage)
	missingScope := newTestActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeStoriesRead)
	human, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("create human actor: %v", err)
	}
	wrongWorkspace, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(uuid.New())
	if err != nil {
		t.Fatalf("create cross-workspace actor: %v", err)
	}

	tests := []struct {
		name    string
		access  Access
		wantErr error
	}{
		{name: "human admin", access: Access{Actor: human, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin}},
		{name: "personal token admin", access: Access{Actor: personalToken, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin}},
		{name: "member", access: Access{Actor: human, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleMember}, wantErr: authorization.ErrInsufficientWorkspaceRole},
		{name: "service account", access: Access{Actor: serviceAccount, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin}, wantErr: authorization.ErrPrincipalKindDenied},
		{name: "missing scope", access: Access{Actor: missingScope, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin}, wantErr: authorization.ErrCredentialScopeDenied},
		{name: "workspace mismatch", access: Access{Actor: wrongWorkspace, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin}, wantErr: authorization.ErrWorkspaceMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := authorizeManagement(test.access)
			if test.wantErr == nil && err != nil {
				t.Fatalf("authorizeManagement() error = %v", err)
			}
			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("authorizeManagement() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestManagerCreateEndpointValidatesBeforeSideEffects(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	access := Access{Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin}

	for _, input := range []CreateEndpointInput{
		{Name: " leading", URL: "https://hooks.example.com", Subscriptions: []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated}},
		{Name: "Example", URL: "https://hooks.example.com ", Subscriptions: []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated}},
		{Name: "Example", URL: "https://hooks.example.com", Subscriptions: nil},
	} {
		repository := &endpointRepositoryStub{}
		principals := &principalResolverStub{id: uuid.New()}
		validator := &endpointValidatorStub{canonical: "https://hooks.example.com/"}
		manager, err := newManager(repository, principals, newTestSecretManager(t), validator,
			&testClock{values: []time.Time{time.Unix(1_700_000_000, 0)}}, &testIDs{values: []uuid.UUID{uuid.New(), uuid.New()}})
		if err != nil {
			t.Fatalf("create manager: %v", err)
		}
		if _, err := manager.CreateEndpoint(t.Context(), access, input); err == nil {
			t.Fatalf("CreateEndpoint(%+v) succeeded", input)
		}
		if repository.createCalls != 0 || principals.calls != 0 || validator.calls != 0 {
			t.Fatalf("invalid input caused side effects: repository=%d principal=%d validator=%d", repository.createCalls, principals.calls, validator.calls)
		}
	}
}

func TestManagerCreateEndpointReturnsCanonicalEndpointAndShowOnceSecret(t *testing.T) {
	t.Parallel()
	workspaceID, ownerID := uuid.New(), uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	repository := &endpointRepositoryStub{}
	principals := &principalResolverStub{id: ownerID}
	validator := &endpointValidatorStub{canonical: "https://hooks.example.com/receive"}
	endpointID, auditID := uuid.New(), uuid.New()
	manager, err := newManager(repository, principals, newTestSecretManager(t), validator,
		&testClock{values: []time.Time{time.Unix(1_700_000_000, 0)}}, &testIDs{values: []uuid.UUID{endpointID, auditID}})
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	created, err := manager.CreateEndpoint(t.Context(), Access{
		Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}, CreateEndpointInput{
		Name: "Production events", URL: "https://hooks.example.com/receive",
		Subscriptions: []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated}, RequestID: "request-1",
	})
	if err != nil {
		t.Fatalf("CreateEndpoint() error = %v", err)
	}
	if created.Endpoint.ID != endpointID || created.Endpoint.OwnerPrincipalID != ownerID || created.Endpoint.URL != validator.canonical {
		t.Fatalf("created endpoint = %+v", created.Endpoint)
	}
	if created.Secret.Reveal() == "" || created.Secret.String() != "[REDACTED]" {
		t.Fatal("created endpoint did not return a redacted show-once secret")
	}
	if repository.created.AuditID != auditID || repository.created.RequestID != "request-1" || repository.createCalls != 1 {
		t.Fatalf("repository command = %+v", repository.created)
	}
}

func TestManagerListEndpointsUsesBoundedKeysetPage(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	createdAt := time.Unix(1_700_000_000, 0).UTC()
	repository := &endpointRepositoryStub{listed: []outboundwebhooksdomain.Endpoint{
		{ID: uuid.New(), CreatedAt: createdAt.Add(3 * time.Minute)},
		{ID: uuid.New(), CreatedAt: createdAt.Add(2 * time.Minute)},
		{ID: uuid.New(), CreatedAt: createdAt.Add(time.Minute)},
	}}
	manager, err := newManager(repository, &principalResolverStub{id: uuid.New()}, newTestSecretManager(t),
		&endpointValidatorStub{}, &testClock{values: []time.Time{createdAt}}, &testIDs{})
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	cursor := &outboundwebhooksdomain.EndpointCursor{CreatedAt: createdAt.Add(time.Hour), ID: uuid.New()}
	page, err := manager.ListEndpoints(t.Context(), Access{
		Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}, cursor, 2)
	if err != nil {
		t.Fatalf("ListEndpoints() error = %v", err)
	}
	if len(page.Items) != 2 || page.NextCursor == nil || page.NextCursor.ID != page.Items[1].ID {
		t.Fatalf("endpoint page = %+v", page)
	}
	if repository.listSize != 3 || repository.listCursor != cursor {
		t.Fatalf("repository list size=%d cursor=%+v", repository.listSize, repository.listCursor)
	}
	if _, err := manager.ListEndpoints(t.Context(), Access{
		Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}, nil, maximumEndpointPageSize+1); !errors.Is(err, outboundwebhooksdomain.ErrInvalidEndpoint) {
		t.Fatalf("oversized ListEndpoints() error = %v", err)
	}
}

func newTestActor(t *testing.T, kind platformauth.PrincipalKind, workspaceID uuid.UUID, scopes ...platformauth.Scope) platformauth.Actor {
	t.Helper()
	actor, err := platformauth.NewActor(uuid.New(), kind, uuid.New(), platformauth.MustScopeSet(scopes...), platformauth.UnrestrictedTeamAccess())
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("select actor workspace: %v", err)
	}
	return actor
}
