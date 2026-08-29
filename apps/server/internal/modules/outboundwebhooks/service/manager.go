package outboundwebhooksservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

const (
	defaultSecretOverlap    = 24 * time.Hour
	defaultEndpointPageSize = 25
	maximumEndpointPageSize = 100
	maxDisableReasonRunes   = 240
)

type EndpointRepository interface {
	CreateEndpoint(context.Context, outboundwebhooksdomain.CreateEndpoint, string) (outboundwebhooksdomain.Endpoint, error)
	GetEndpoint(context.Context, uuid.UUID, uuid.UUID) (outboundwebhooksdomain.Endpoint, error)
	ListEndpoints(context.Context, uuid.UUID, *outboundwebhooksdomain.EndpointCursor, int) ([]outboundwebhooksdomain.Endpoint, error)
	ReplaceSubscriptions(context.Context, platformauth.Actor, uuid.UUID, uuid.UUID, uuid.UUID, []outboundwebhooksdomain.EventType, time.Time, string) error
	DisableEndpoint(context.Context, platformauth.Actor, uuid.UUID, uuid.UUID, uuid.UUID, string, string, time.Time) error
	RotateEndpointSecret(context.Context, platformauth.Actor, uuid.UUID, uuid.UUID, uuid.UUID, int, string, time.Time, time.Time, string) (int, error)
}

type PrincipalResolver interface {
	ResolveHumanPrincipal(
		context.Context,
		platformauth.Actor,
		uuid.UUID,
		authorization.WorkspaceRole,
		string,
	) (uuid.UUID, error)
}

type EndpointValidator interface {
	Validate(context.Context, string) (string, error)
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewUUID() (uuid.UUID, error)
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

type randomIDGenerator struct{}

func (randomIDGenerator) NewUUID() (uuid.UUID, error) { return uuid.NewRandom() }

type Manager struct {
	repository EndpointRepository
	principals PrincipalResolver
	secrets    *SecretManager
	validator  EndpointValidator
	clock      Clock
	ids        IDGenerator
}

func NewManager(
	repository EndpointRepository,
	principals PrincipalResolver,
	secrets *SecretManager,
	validator EndpointValidator,
) (*Manager, error) {
	return newManager(repository, principals, secrets, validator, systemClock{}, randomIDGenerator{})
}

func newManager(
	repository EndpointRepository,
	principals PrincipalResolver,
	secrets *SecretManager,
	validator EndpointValidator,
	clock Clock,
	ids IDGenerator,
) (*Manager, error) {
	if repository == nil || principals == nil || secrets == nil || validator == nil || clock == nil || ids == nil {
		return nil, errors.New("outbound webhook manager dependencies are required")
	}
	return &Manager{
		repository: repository, principals: principals, secrets: secrets,
		validator: validator, clock: clock, ids: ids,
	}, nil
}

type Access struct {
	Actor         platformauth.Actor
	WorkspaceID   uuid.UUID
	WorkspaceRole authorization.WorkspaceRole
}

type CreateEndpointInput struct {
	Name          string
	URL           string
	Subscriptions []outboundwebhooksdomain.EventType
	RequestID     string
}

func (manager *Manager) CreateEndpoint(ctx context.Context, access Access, input CreateEndpointInput) (outboundwebhooksdomain.CreatedEndpoint, error) {
	if err := authorizeManagement(access); err != nil {
		return outboundwebhooksdomain.CreatedEndpoint{}, err
	}
	if input.Name == "" || input.Name != strings.TrimSpace(input.Name) || len(input.Name) > 120 ||
		input.URL == "" || input.URL != strings.TrimSpace(input.URL) || len(input.URL) > 2048 {
		return outboundwebhooksdomain.CreatedEndpoint{}, outboundwebhooksdomain.ErrInvalidEndpoint
	}
	if err := outboundwebhooksdomain.ValidateSubscriptions(input.Subscriptions); err != nil {
		return outboundwebhooksdomain.CreatedEndpoint{}, err
	}
	canonicalURL, err := manager.validator.Validate(ctx, input.URL)
	if err != nil {
		return outboundwebhooksdomain.CreatedEndpoint{}, errors.Join(outboundwebhooksdomain.ErrInvalidEndpoint, err)
	}
	now := manager.clock.Now().UTC()
	endpointID, err := manager.ids.NewUUID()
	if err != nil {
		return outboundwebhooksdomain.CreatedEndpoint{}, fmt.Errorf("generate outbound webhook endpoint id: %w", err)
	}
	auditID, err := manager.ids.NewUUID()
	if err != nil {
		return outboundwebhooksdomain.CreatedEndpoint{}, fmt.Errorf("generate outbound webhook audit id: %w", err)
	}
	ownerPrincipalID, err := manager.principals.ResolveHumanPrincipal(
		ctx, access.Actor, access.WorkspaceID, access.WorkspaceRole, input.RequestID,
	)
	if err != nil {
		return outboundwebhooksdomain.CreatedEndpoint{}, fmt.Errorf("resolve outbound webhook owner: %w", err)
	}
	secret, envelope, err := manager.secrets.Generate(access.WorkspaceID, endpointID, 1)
	if err != nil {
		return outboundwebhooksdomain.CreatedEndpoint{}, err
	}
	command := outboundwebhooksdomain.CreateEndpoint{
		ID: endpointID, AuditID: auditID, WorkspaceID: access.WorkspaceID,
		OwnerPrincipalID: ownerPrincipalID, Actor: access.Actor, WorkspaceRole: access.WorkspaceRole,
		Name: input.Name, URL: canonicalURL,
		Subscriptions: append([]outboundwebhooksdomain.EventType(nil), input.Subscriptions...),
		CreatedAt:     now, RequestID: input.RequestID,
	}
	endpoint, err := manager.repository.CreateEndpoint(ctx, command, envelope)
	if err != nil {
		return outboundwebhooksdomain.CreatedEndpoint{}, err
	}
	return outboundwebhooksdomain.CreatedEndpoint{Endpoint: endpoint, Secret: secret}, nil
}

func (manager *Manager) GetEndpoint(ctx context.Context, access Access, endpointID uuid.UUID) (outboundwebhooksdomain.Endpoint, error) {
	if err := authorizeManagement(access); err != nil {
		return outboundwebhooksdomain.Endpoint{}, err
	}
	if endpointID == uuid.Nil {
		return outboundwebhooksdomain.Endpoint{}, outboundwebhooksdomain.ErrEndpointNotFound
	}
	return manager.repository.GetEndpoint(ctx, access.WorkspaceID, endpointID)
}

func (manager *Manager) ListEndpoints(
	ctx context.Context,
	access Access,
	cursor *outboundwebhooksdomain.EndpointCursor,
	requestedPageSize int,
) (outboundwebhooksdomain.EndpointPage, error) {
	if err := authorizeManagement(access); err != nil {
		return outboundwebhooksdomain.EndpointPage{}, err
	}
	if cursor != nil {
		if err := cursor.Validate(); err != nil {
			return outboundwebhooksdomain.EndpointPage{}, err
		}
	}
	pageSize := requestedPageSize
	if pageSize == 0 {
		pageSize = defaultEndpointPageSize
	}
	if pageSize < 1 || pageSize > maximumEndpointPageSize {
		return outboundwebhooksdomain.EndpointPage{}, outboundwebhooksdomain.ErrInvalidEndpoint
	}
	endpoints, err := manager.repository.ListEndpoints(ctx, access.WorkspaceID, cursor, pageSize+1)
	if err != nil {
		return outboundwebhooksdomain.EndpointPage{}, err
	}
	page := outboundwebhooksdomain.EndpointPage{Items: endpoints}
	if len(endpoints) > pageSize {
		page.Items = endpoints[:pageSize]
		last := page.Items[len(page.Items)-1]
		page.NextCursor = &outboundwebhooksdomain.EndpointCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	if page.Items == nil {
		page.Items = []outboundwebhooksdomain.Endpoint{}
	}
	return page, nil
}

func (manager *Manager) ReplaceSubscriptions(
	ctx context.Context,
	access Access,
	endpointID uuid.UUID,
	subscriptions []outboundwebhooksdomain.EventType,
	requestID string,
) error {
	if err := authorizeManagement(access); err != nil {
		return err
	}
	if endpointID == uuid.Nil {
		return outboundwebhooksdomain.ErrEndpointNotFound
	}
	if err := outboundwebhooksdomain.ValidateSubscriptions(subscriptions); err != nil {
		return err
	}
	auditID, err := manager.ids.NewUUID()
	if err != nil {
		return fmt.Errorf("generate outbound webhook audit id: %w", err)
	}
	return manager.repository.ReplaceSubscriptions(
		ctx, access.Actor, access.WorkspaceID, endpointID, auditID,
		append([]outboundwebhooksdomain.EventType(nil), subscriptions...), manager.clock.Now().UTC(), requestID,
	)
}

func (manager *Manager) DisableEndpoint(ctx context.Context, access Access, endpointID uuid.UUID, reason, requestID string) error {
	if err := authorizeManagement(access); err != nil {
		return err
	}
	if endpointID == uuid.Nil || reason == "" || reason != strings.TrimSpace(reason) || len([]rune(reason)) > maxDisableReasonRunes {
		return outboundwebhooksdomain.ErrInvalidEndpoint
	}
	auditID, err := manager.ids.NewUUID()
	if err != nil {
		return fmt.Errorf("generate outbound webhook audit id: %w", err)
	}
	return manager.repository.DisableEndpoint(
		ctx, access.Actor, access.WorkspaceID, endpointID, auditID, reason, requestID, manager.clock.Now().UTC(),
	)
}

func (manager *Manager) RotateEndpointSecret(
	ctx context.Context,
	access Access,
	endpointID uuid.UUID,
	requestID string,
) (outboundwebhooksdomain.SigningSecret, int, time.Time, error) {
	if err := authorizeManagement(access); err != nil {
		return outboundwebhooksdomain.SigningSecret{}, 0, time.Time{}, err
	}
	if endpointID == uuid.Nil {
		return outboundwebhooksdomain.SigningSecret{}, 0, time.Time{}, outboundwebhooksdomain.ErrEndpointNotFound
	}
	endpoint, err := manager.repository.GetEndpoint(ctx, access.WorkspaceID, endpointID)
	if err != nil {
		return outboundwebhooksdomain.SigningSecret{}, 0, time.Time{}, err
	}
	if endpoint.Status != outboundwebhooksdomain.EndpointActive {
		return outboundwebhooksdomain.SigningSecret{}, 0, time.Time{}, outboundwebhooksdomain.ErrEndpointDisabled
	}
	newGeneration := endpoint.SecretGeneration + 1
	secret, envelope, err := manager.secrets.Generate(access.WorkspaceID, endpointID, newGeneration)
	if err != nil {
		return outboundwebhooksdomain.SigningSecret{}, 0, time.Time{}, err
	}
	auditID, err := manager.ids.NewUUID()
	if err != nil {
		return outboundwebhooksdomain.SigningSecret{}, 0, time.Time{}, fmt.Errorf("generate outbound webhook audit id: %w", err)
	}
	now := manager.clock.Now().UTC()
	overlapExpiresAt := now.Add(defaultSecretOverlap)
	persistedGeneration, err := manager.repository.RotateEndpointSecret(
		ctx, access.Actor, access.WorkspaceID, endpointID, auditID, endpoint.SecretGeneration,
		envelope, overlapExpiresAt, now, requestID,
	)
	if err != nil {
		return outboundwebhooksdomain.SigningSecret{}, 0, time.Time{}, err
	}
	if persistedGeneration != newGeneration {
		return outboundwebhooksdomain.SigningSecret{}, 0, time.Time{}, outboundwebhooksdomain.ErrEndpointConflict
	}
	return secret, persistedGeneration, overlapExpiresAt, nil
}

func authorizeManagement(access Access) error {
	err := authorization.AuthorizeWorkspace(authorization.WorkspacePolicyInput{
		Actor: access.Actor, WorkspaceID: access.WorkspaceID,
		WorkspaceRole: access.WorkspaceRole, MinimumWorkspaceRole: authorization.WorkspaceRoleAdmin,
		RequiredScopes: []platformauth.Scope{platformauth.ScopeWebhooksManage},
		AllowedPrincipalKinds: []platformauth.PrincipalKind{
			platformauth.PrincipalHumanUser,
			platformauth.PrincipalPersonalToken,
			platformauth.PrincipalOAuthUser,
		},
	})
	if err != nil {
		return errors.Join(outboundwebhooksdomain.ErrEndpointOwnerInactive, err)
	}
	return nil
}
