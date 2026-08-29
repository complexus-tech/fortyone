package workspaces

import (
	"context"
	"errors"
	"mime/multipart"
	"time"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrNotFound               = workspacedomain.ErrNotFound
	ErrMemberNotFound         = workspacedomain.ErrMemberNotFound
	ErrSlugTaken              = workspacedomain.ErrSlugTaken
	ErrRestrictedSlug         = workspacedomain.ErrRestrictedSlug
	ErrTx                     = errors.New("failed to create a workspace")
	ErrAlreadyWorkspaceMember = workspacedomain.ErrAlreadyWorkspaceMember
)

var restrictedSlugs = map[string]struct{}{
	"admin": {}, "internal": {}, "qa": {}, "staging": {}, "ops": {},
	"team": {}, "complexus": {}, "dev": {}, "test": {}, "prod": {},
	"development": {}, "testing": {}, "production": {}, "staff": {},
	"hr": {}, "finance": {}, "legal": {}, "marketing": {}, "sales": {},
	"support": {}, "it": {}, "security": {}, "engineering": {},
	"design": {}, "product": {}, "auth": {}, "fortyone": {}, "forty-one": {},
}

// AttachmentsService provides workspace logo operations.
type AttachmentsService interface {
	UploadWorkspaceLogo(context.Context, multipart.File, *multipart.FileHeader, uuid.UUID) (string, error)
	DeleteWorkspaceLogo(context.Context, string) error
	ResolveWorkspaceLogoURL(context.Context, string, time.Duration) (string, error)
}

// Repository owns workspace persistence. Every operation is either explicitly
// tenant-scoped or only addresses a workspace-global invariant such as slug
// uniqueness.
type Repository interface {
	List(context.Context, uuid.UUID) ([]CoreWorkspace, error)
	Update(context.Context, uuid.UUID, CoreWorkspace) (CoreWorkspace, error)
	Delete(context.Context, uuid.UUID, uuid.UUID) error
	Restore(context.Context, uuid.UUID, uuid.UUID) error
	GetWorkspaceAdminEmails(context.Context, uuid.UUID, uuid.UUID) ([]string, error)
	AddMember(context.Context, uuid.UUID, uuid.UUID, string) error
	Get(context.Context, uuid.UUID, uuid.UUID) (CoreWorkspace, error)
	GetByID(context.Context, uuid.UUID) (CoreWorkspace, error)
	GetBySlug(context.Context, string, uuid.UUID) (CoreWorkspace, error)
	GetPublicBySlug(context.Context, string) (CoreWorkspace, error)
	RemoveMember(context.Context, uuid.UUID, uuid.UUID) error
	CheckSlugAvailability(context.Context, string) (bool, error)
	UpdateMemberRole(context.Context, uuid.UUID, uuid.UUID, string) error
	GetWorkspaceSettings(context.Context, uuid.UUID) (CoreWorkspaceSettings, error)
	GetOrCreateWorkspaceSettings(context.Context, uuid.UUID) (CoreWorkspaceSettings, error)
	UpdateWorkspaceSettings(context.Context, uuid.UUID, CoreWorkspaceSettings) (CoreWorkspaceSettings, error)
	ResolveCurrentMembership(context.Context, string, uuid.UUID) (CurrentMembership, error)
	RecordAccess(context.Context, uuid.UUID, uuid.UUID) error
}

type WorkspaceUser struct {
	Email    string
	FullName string
	Username string
}

type UserDirectory interface {
	GetWorkspaceUser(context.Context, uuid.UUID) (WorkspaceUser, error)
	UpdateLastUsedWorkspace(context.Context, uuid.UUID, uuid.UUID) error
}

type SubscriptionManager interface {
	UpdateWorkspaceSeats(context.Context, uuid.UUID) error
	CancelWorkspaceSubscription(context.Context, uuid.UUID) error
}

type SeedContentCreator interface {
	CreateWorkspaceSeedContent(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}

type TrialStart struct {
	UserID        uuid.UUID
	Email         string
	FullName      string
	WorkspaceSlug string
	WorkspaceName string
}

type TrialScheduler interface {
	ScheduleWorkspaceTrialStart(TrialStart) error
}

type EventPublisher interface {
	Publish(context.Context, events.Event) error
}

type Service struct {
	repo          Repository
	log           *logger.Logger
	unitOfWork    UnitOfWork
	seedContent   SeedContentCreator
	users         UserDirectory
	subscriptions SubscriptionManager
	publisher     EventPublisher
	trials        TrialScheduler
}

// Dependencies groups cross-module collaborators. Persistence transaction
// details deliberately stay out of this surface.
type Dependencies struct {
	SeedContent   SeedContentCreator
	Users         UserDirectory
	Subscriptions SubscriptionManager
	Publisher     EventPublisher
	Trials        TrialScheduler
}

func New(log *logger.Logger, repo Repository, unitOfWork UnitOfWork, deps Dependencies) *Service {
	if unitOfWork == nil {
		unitOfWork = unavailableUnitOfWork{}
	}
	return &Service{
		repo: repo, log: log, unitOfWork: unitOfWork,
		seedContent: deps.SeedContent, users: deps.Users,
		subscriptions: deps.Subscriptions, publisher: deps.Publisher,
		trials: deps.Trials,
	}
}

var workspaceTracer = otel.Tracer("github.com/complexus-tech/projects-api/internal/modules/workspaces/service")

func startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return workspaceTracer.Start(ctx, name)
}
