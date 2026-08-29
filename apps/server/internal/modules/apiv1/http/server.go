package apiv1http

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

const (
	defaultPageLimit  = 50
	maximumPageLimit  = 100
	maximumPageOffset = 1_000_000
)

type WorkspaceReader interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (workspaces.CoreWorkspace, error)
}

type TeamReader interface {
	List(context.Context, uuid.UUID, uuid.UUID, ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error)
	GetByID(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (teams.CoreTeam, error)
}

type StoryReader interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error)
	List(context.Context, uuid.UUID, stories.CoreStoryFilters) ([]stories.CoreStoryList, error)
}

type StoryWriter interface {
	Create(context.Context, stories.CoreNewStory, uuid.UUID) (stories.CoreSingleStory, error)
}

type StoryService interface {
	StoryReader
	StoryWriter
}

type StoryCommentReader interface {
	GetComments(context.Context, uuid.UUID, uuid.UUID, int, int) ([]stories.CoreComment, bool, error)
	GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (stories.CoreComment, error)
}

type LabelReader interface {
	GetLabels(context.Context, uuid.UUID, uuid.UUID, labels.LabelFilters) ([]labels.CoreLabel, error)
}

type WorkflowStateReader interface {
	TeamListForMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]states.CoreState, error)
}

type SprintReader interface {
	ListQuery(context.Context, sprintdomain.ListQuery) ([]sprints.CoreSprint, error)
}

type ObjectiveReader interface {
	ListIntent(context.Context, objectivesdomain.ListQuery) ([]objectives.CoreObjective, error)
}

type KeyResultReader interface {
	ListPaginated(context.Context, keyresultsdomain.Filters) (keyresults.CoreKeyResultListResponse, error)
}

type IdempotencyManager interface {
	Begin(context.Context, idempotency.Scope, idempotency.Key, []byte) (idempotency.BeginResult, error)
	Complete(context.Context, idempotency.Lease, idempotency.Response) error
}

type WebhookManager interface {
	CreateEndpoint(context.Context, outboundwebhooksservice.Access, outboundwebhooksservice.CreateEndpointInput) (outboundwebhooksdomain.CreatedEndpoint, error)
	GetEndpoint(context.Context, outboundwebhooksservice.Access, uuid.UUID) (outboundwebhooksdomain.Endpoint, error)
	ListEndpoints(context.Context, outboundwebhooksservice.Access, *outboundwebhooksdomain.EndpointCursor, int) (outboundwebhooksdomain.EndpointPage, error)
	ReplaceSubscriptions(context.Context, outboundwebhooksservice.Access, uuid.UUID, []outboundwebhooksdomain.EventType, string) error
	DisableEndpoint(context.Context, outboundwebhooksservice.Access, uuid.UUID, string, string) error
	RotateEndpointSecret(context.Context, outboundwebhooksservice.Access, uuid.UUID, string) (outboundwebhooksdomain.SigningSecret, int, time.Time, error)
}

type cursorPage struct {
	Version     int       `json:"version"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	PrincipalID uuid.UUID `json:"principalId"`
	Offset      int       `json:"offset"`
	Limit       int       `json:"limit"`
	TeamID      uuid.UUID `json:"teamId,omitempty"`
	ResourceID  uuid.UUID `json:"resourceId,omitempty"`
}

type webhookCursor struct {
	Version     int       `json:"version"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	PrincipalID uuid.UUID `json:"principalId"`
	CreatedAt   time.Time `json:"createdAt"`
	EndpointID  uuid.UUID `json:"endpointId"`
	Limit       int       `json:"limit"`
}

type server struct {
	log                  *logger.Logger
	workspaces           WorkspaceReader
	teams                TeamReader
	stories              StoryService
	storyComments        StoryCommentReader
	labels               LabelReader
	workflowStates       WorkflowStateReader
	sprints              SprintReader
	objectives           ObjectiveReader
	keyResults           KeyResultReader
	idempotency          IdempotencyManager
	webhooks             WebhookManager
	teamCursors          pagination.CursorCodec[cursorPage]
	storyCursors         pagination.CursorCodec[cursorPage]
	labelCursors         pagination.CursorCodec[cursorPage]
	workflowStateCursors pagination.CursorCodec[cursorPage]
	sprintCursors        pagination.CursorCodec[cursorPage]
	objectiveCursors     pagination.CursorCodec[cursorPage]
	keyResultCursors     pagination.CursorCodec[cursorPage]
	commentCursors       pagination.CursorCodec[cursorPage]
	webhookCursors       pagination.CursorCodec[webhookCursor]
	createStoryOperation idempotency.Operation
}

type serverConfig struct {
	Log           *logger.Logger
	SecretKey     string
	Workspaces    WorkspaceReader
	Teams         TeamReader
	Stories       StoryService
	StoryComments StoryCommentReader
	Labels        LabelReader
	States        WorkflowStateReader
	Sprints       SprintReader
	Objectives    ObjectiveReader
	KeyResults    KeyResultReader
	Idempotency   IdempotencyManager
	Webhooks      WebhookManager
}

func newServer(config serverConfig) (*server, error) {
	if config.Log == nil || config.Workspaces == nil || config.Teams == nil || config.Stories == nil || config.StoryComments == nil || config.Labels == nil ||
		config.States == nil || config.Sprints == nil || config.Objectives == nil || config.KeyResults == nil ||
		config.Webhooks == nil || config.Idempotency == nil {
		return nil, errors.New("public API services are required")
	}
	if strings.TrimSpace(config.SecretKey) == "" {
		return nil, errors.New("public API cursor secret is required")
	}
	teamCursors, err := newCursorCodec[cursorPage](config.SecretKey, "teams")
	if err != nil {
		return nil, err
	}
	storyCursors, err := newCursorCodec[cursorPage](config.SecretKey, "stories")
	if err != nil {
		return nil, err
	}
	labelCursors, err := newCursorCodec[cursorPage](config.SecretKey, "labels")
	if err != nil {
		return nil, err
	}
	workflowStateCursors, err := newCursorCodec[cursorPage](config.SecretKey, "workflow-states")
	if err != nil {
		return nil, err
	}
	sprintCursors, err := newCursorCodec[cursorPage](config.SecretKey, "sprints")
	if err != nil {
		return nil, err
	}
	objectiveCursors, err := newCursorCodec[cursorPage](config.SecretKey, "objectives")
	if err != nil {
		return nil, err
	}
	keyResultCursors, err := newCursorCodec[cursorPage](config.SecretKey, "key-results")
	if err != nil {
		return nil, err
	}
	commentCursors, err := newCursorCodec[cursorPage](config.SecretKey, "story-comments")
	if err != nil {
		return nil, err
	}
	webhookCursors, err := newCursorCodec[webhookCursor](config.SecretKey, "webhook-endpoints")
	if err != nil {
		return nil, err
	}
	createStoryOperation, err := idempotency.ParseOperation("stories.create")
	if err != nil {
		return nil, fmt.Errorf("initialize story-create idempotency operation: %w", err)
	}
	return &server{
		log: config.Log, workspaces: config.Workspaces, teams: config.Teams,
		stories: config.Stories, storyComments: config.StoryComments, labels: config.Labels, workflowStates: config.States, sprints: config.Sprints,
		objectives: config.Objectives, keyResults: config.KeyResults, idempotency: config.Idempotency,
		webhooks: config.Webhooks, teamCursors: teamCursors, storyCursors: storyCursors, labelCursors: labelCursors,
		workflowStateCursors: workflowStateCursors, sprintCursors: sprintCursors, objectiveCursors: objectiveCursors,
		keyResultCursors: keyResultCursors, commentCursors: commentCursors, webhookCursors: webhookCursors,
		createStoryOperation: createStoryOperation,
	}, nil
}

func newCursorCodec[T any](rootSecret, family string) (pagination.CursorCodec[T], error) {
	key, err := pagination.DeriveSigningKey("v1", []byte(rootSecret), "public-api-v1:"+family)
	if err != nil {
		return pagination.CursorCodec[T]{}, fmt.Errorf("derive %s cursor key: %w", family, err)
	}
	codec, err := pagination.NewCursorCodec[T](key)
	if err != nil {
		return pagination.CursorCodec[T]{}, fmt.Errorf("initialize %s cursor codec: %w", family, err)
	}
	return codec, nil
}

func actorFor(ctx context.Context, workspaceID uuid.UUID, scope platformauth.Scope) (platformauth.Actor, *failure) {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		// The preview wire code predates OAuth and remains stable for existing
		// clients even though the bearer boundary now supports all documented
		// developer credential families.
		return platformauth.Actor{}, &failure{status: 401, code: "machine_authentication_required", message: "A valid machine bearer credential is required."}
	}
	if actor.WorkspaceID != workspaceID {
		return platformauth.Actor{}, &failure{status: 403, code: "workspace_access_denied", message: "The credential is bound to another workspace."}
	}
	if !actor.Scopes.Has(scope) {
		return platformauth.Actor{}, &failure{status: 403, code: "scope_required", message: fmt.Sprintf("The %s scope is required.", scope)}
	}
	return actor, nil
}

func requireUserCredential(actor platformauth.Actor) *failure {
	if actor.Kind != platformauth.PrincipalPersonalToken && actor.Kind != platformauth.PrincipalOAuthUser {
		return &failure{status: 403, code: "principal_not_supported", message: "This operation requires a personal access token or user-authorized OAuth token."}
	}
	return nil
}

func requireStoryWriter(actor platformauth.Actor) *failure {
	switch actor.Kind {
	case platformauth.PrincipalPersonalToken, platformauth.PrincipalServiceAccount,
		platformauth.PrincipalOAuthUser, platformauth.PrincipalOAuthApplication:
		return nil
	default:
		return &failure{status: 403, code: "principal_not_supported", message: "This operation requires a personal access token, service-account key, user-authorized OAuth token, or installed OAuth application."}
	}
}

var _ openapiv1.StrictServerInterface = (*server)(nil)
