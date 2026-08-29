package admin

import (
	"context"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	platformclock "github.com/complexus-tech/projects-api/internal/platform/clock"
	"github.com/google/uuid"
)

type Repository interface {
	GetAdminUser(context.Context, uuid.UUID) (admindomain.UserSummary, error)
	GetDashboardSummary(context.Context, admindomain.DashboardSummaryQuery) (admindomain.DashboardSummary, error)
	ListWorkspaces(context.Context, admindomain.ListWorkspacesQuery) (admindomain.ListResult[admindomain.WorkspaceSummary], error)
	GetWorkspaceOverview(context.Context, admindomain.GetWorkspaceQuery) (admindomain.WorkspaceOverview, error)
	UpdateWorkspaceTrial(context.Context, admindomain.UpdateWorkspaceTrialCommand) (admindomain.WorkspaceOverview, error)
	SetWorkspaceDeleted(context.Context, admindomain.SetWorkspaceDeletedCommand) (admindomain.WorkspaceOverview, error)
	ListUsers(context.Context, admindomain.ListUsersQuery) (admindomain.ListResult[admindomain.UserSummary], error)
	GetUserOverview(context.Context, admindomain.GetUserQuery) (admindomain.UserOverview, error)
	UpdateUserState(context.Context, admindomain.UpdateUserStateCommand) (admindomain.UserOverview, error)
	RequestSessionRevocation(context.Context, admindomain.RequestSessionRevocationCommand) (admindomain.UserOverview, error)
	ListAuditLogs(context.Context, admindomain.ListAuditLogsQuery) (admindomain.ListResult[admindomain.AuditLog], error)
	ListAdminNotes(context.Context, admindomain.ListAdminNotesQuery) (admindomain.ListResult[admindomain.AdminNote], error)
	CreateAdminNote(context.Context, admindomain.CreateAdminNoteCommand) (admindomain.AdminNote, error)
	BeginSubscriptionSync(context.Context, admindomain.BeginSubscriptionSyncCommand) (admindomain.SubscriptionSyncAttempt, admindomain.WorkspaceOverview, error)
	FinishSubscriptionSync(context.Context, admindomain.FinishSubscriptionSyncCommand) (admindomain.WorkspaceOverview, error)
}

type AssetResolver interface {
	ResolveProfileImageURL(context.Context, string, time.Duration) (string, error)
	ResolveWorkspaceLogoURL(context.Context, string, time.Duration) (string, error)
}

type SubscriptionSyncer interface {
	SyncSubscription(context.Context, uuid.UUID) error
}

type Service struct {
	repo               Repository
	assetResolver      AssetResolver
	subscriptionSyncer SubscriptionSyncer
	clock              platformclock.Clock
}

type Option func(*Service)

func WithClock(clock platformclock.Clock) Option {
	return func(service *Service) {
		if clock != nil {
			service.clock = clock
		}
	}
}

// WithNow remains as a narrow compatibility seam for existing tests and
// callers while all production composition uses the shared Clock contract.
func WithNow(now func() time.Time) Option {
	return WithClock(functionClock(now))
}

func WithAssetResolver(resolver AssetResolver) Option {
	return func(service *Service) { service.assetResolver = resolver }
}

func WithSubscriptionSyncer(syncer SubscriptionSyncer) Option {
	return func(service *Service) { service.subscriptionSyncer = syncer }
}

func New(repository Repository, options ...Option) *Service {
	service := &Service{repo: repository, clock: platformclock.System{}}
	for _, option := range options {
		option(service)
	}
	return service
}

type functionClock func() time.Time

func (clock functionClock) Now() time.Time {
	if clock == nil {
		return time.Now()
	}
	return clock()
}
