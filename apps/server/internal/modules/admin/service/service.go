package admin

import (
	"context"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Repository interface {
	GetAdminUser(ctx context.Context, userID uuid.UUID) (UserSummary, error)
	GetDashboardSummary(ctx context.Context) (DashboardSummary, error)
	ListWorkspaces(ctx context.Context, input ListWorkspacesInput) (ListResult[WorkspaceSummary], error)
	GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID) (WorkspaceOverview, error)
	UpdateWorkspaceTrial(ctx context.Context, input UpdateWorkspaceTrialInput) error
	ListUsers(ctx context.Context, input ListUsersInput) (ListResult[UserSummary], error)
	GetUserOverview(ctx context.Context, userID uuid.UUID) (UserOverview, error)
	ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (ListResult[AuditLog], error)
	InsertAuditEntry(ctx context.Context, input AuditEntryInput) error
}

type AssetResolver interface {
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
	ResolveWorkspaceLogoURL(ctx context.Context, logo string, expiry time.Duration) (string, error)
}

const adminAssetURLExpiry = 24 * time.Hour

type Service struct {
	repo          Repository
	assetResolver AssetResolver
	now           func() time.Time
}

type Option func(*Service)

func WithNow(now func() time.Time) Option {
	return func(s *Service) {
		s.now = now
	}
}

func WithAssetResolver(resolver AssetResolver) Option {
	return func(s *Service) {
		s.assetResolver = resolver
	}
}

func New(repo Repository, opts ...Option) *Service {
	s := &Service{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) GetCurrentAdmin(ctx context.Context, actorID uuid.UUID) (UserSummary, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.GetCurrentAdmin")
	defer span.End()

	return s.ensureAdmin(ctx, actorID)
}

func (s *Service) GetDashboardSummary(ctx context.Context, actorID uuid.UUID) (DashboardSummary, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.GetDashboardSummary")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return DashboardSummary{}, err
	}
	return s.repo.GetDashboardSummary(ctx)
}

func (s *Service) ListWorkspaces(ctx context.Context, actorID uuid.UUID, input ListWorkspacesInput) (ListResult[WorkspaceSummary], error) {
	ctx, span := web.AddSpan(ctx, "business.admin.ListWorkspaces")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return ListResult[WorkspaceSummary]{}, err
	}
	result, err := s.repo.ListWorkspaces(ctx, normalizeListWorkspacesInput(input))
	if err != nil {
		return ListResult[WorkspaceSummary]{}, err
	}
	for i := range result.Items {
		s.resolveWorkspaceLogo(ctx, &result.Items[i])
	}
	return result, nil
}

func (s *Service) GetWorkspaceOverview(ctx context.Context, actorID, workspaceID uuid.UUID) (WorkspaceOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.GetWorkspaceOverview")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return WorkspaceOverview{}, err
	}
	overview, err := s.repo.GetWorkspaceOverview(ctx, workspaceID)
	if err != nil {
		return WorkspaceOverview{}, err
	}
	s.resolveWorkspaceLogo(ctx, &overview.Workspace)
	return overview, nil
}

func (s *Service) UpdateWorkspaceTrial(ctx context.Context, actorID, workspaceID uuid.UUID, input UpdateWorkspaceTrialInput) (WorkspaceOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.UpdateWorkspaceTrial")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return WorkspaceOverview{}, err
	}

	overview, err := s.repo.GetWorkspaceOverview(ctx, workspaceID)
	if err != nil {
		return WorkspaceOverview{}, err
	}

	trialEndsOn := input.TrialEndsOn.UTC()
	now := s.now().UTC()
	if !trialEndsOn.After(now) {
		return WorkspaceOverview{}, ErrInvalidTrialEndsOn
	}

	var oldTrialEndsOn any
	if current := overview.Workspace.TrialEndsOn; current != nil {
		oldTrialEndsOn = current.UTC()
		if current.After(now) && !trialEndsOn.After(current.UTC()) {
			return WorkspaceOverview{}, ErrInvalidTrialEndsOn
		}
	}

	trimmedReason := strings.TrimSpace(input.Reason)
	updateInput := UpdateWorkspaceTrialInput{
		WorkspaceID: workspaceID,
		TrialEndsOn: trialEndsOn,
		Reason:      trimmedReason,
	}
	if err := s.repo.UpdateWorkspaceTrial(ctx, updateInput); err != nil {
		return WorkspaceOverview{}, err
	}

	if err := s.repo.InsertAuditEntry(ctx, AuditEntryInput{
		ActorUserID: actorID,
		TargetType:  "workspace",
		TargetID:    &workspaceID,
		WorkspaceID: &workspaceID,
		Action:      "workspace.trial_updated",
		FieldName:   "trial_ends_on",
		OldValue:    oldTrialEndsOn,
		NewValue:    trialEndsOn,
		Reason:      trimmedReason,
		Metadata: map[string]any{
			"workspace_name": overview.Workspace.Name,
			"workspace_slug": overview.Workspace.Slug,
		},
	}); err != nil {
		return WorkspaceOverview{}, err
	}

	overview.Workspace.TrialEndsOn = &trialEndsOn
	s.resolveWorkspaceLogo(ctx, &overview.Workspace)
	return overview, nil
}

func (s *Service) ListUsers(ctx context.Context, actorID uuid.UUID, input ListUsersInput) (ListResult[UserSummary], error) {
	ctx, span := web.AddSpan(ctx, "business.admin.ListUsers")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return ListResult[UserSummary]{}, err
	}
	result, err := s.repo.ListUsers(ctx, normalizeListUsersInput(input))
	if err != nil {
		return ListResult[UserSummary]{}, err
	}
	for i := range result.Items {
		s.resolveUserAvatar(ctx, &result.Items[i])
	}
	return result, nil
}

func (s *Service) GetUserOverview(ctx context.Context, actorID, userID uuid.UUID) (UserOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.GetUserOverview")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return UserOverview{}, err
	}
	overview, err := s.repo.GetUserOverview(ctx, userID)
	if err != nil {
		return UserOverview{}, err
	}
	s.resolveUserAvatar(ctx, &overview.User)
	return overview, nil
}

func (s *Service) ListAuditLogs(ctx context.Context, actorID uuid.UUID, input ListAuditLogsInput) (ListResult[AuditLog], error) {
	ctx, span := web.AddSpan(ctx, "business.admin.ListAuditLogs")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return ListResult[AuditLog]{}, err
	}
	return s.repo.ListAuditLogs(ctx, normalizeListAuditLogsInput(input))
}

func (s *Service) ensureAdmin(ctx context.Context, actorID uuid.UUID) (UserSummary, error) {
	user, err := s.repo.GetAdminUser(ctx, actorID)
	if err != nil {
		return UserSummary{}, err
	}
	if !user.IsInternal || !user.IsActive {
		return UserSummary{}, ErrForbidden
	}
	s.resolveUserAvatar(ctx, &user)
	return user, nil
}

func (s *Service) resolveUserAvatar(ctx context.Context, user *UserSummary) {
	if s.assetResolver == nil || user == nil || strings.TrimSpace(user.AvatarURL) == "" {
		return
	}

	resolved, err := s.assetResolver.ResolveProfileImageURL(ctx, user.AvatarURL, adminAssetURLExpiry)
	if err != nil {
		user.AvatarURL = ""
		return
	}
	user.AvatarURL = resolved
}

func (s *Service) resolveWorkspaceLogo(ctx context.Context, workspace *WorkspaceSummary) {
	if s.assetResolver == nil || workspace == nil || workspace.AvatarURL == nil || strings.TrimSpace(*workspace.AvatarURL) == "" {
		return
	}

	resolved, err := s.assetResolver.ResolveWorkspaceLogoURL(ctx, *workspace.AvatarURL, adminAssetURLExpiry)
	if err != nil {
		workspace.AvatarURL = nil
		return
	}
	workspace.AvatarURL = &resolved
}

func normalizeListWorkspacesInput(input ListWorkspacesInput) ListWorkspacesInput {
	input.Pagination = normalizePagination(input.Pagination)
	input.Query = strings.ToLower(strings.TrimSpace(input.Query))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	return input
}

func normalizeListUsersInput(input ListUsersInput) ListUsersInput {
	input.Pagination = normalizePagination(input.Pagination)
	input.Query = strings.ToLower(strings.TrimSpace(input.Query))
	return input
}

func normalizeListAuditLogsInput(input ListAuditLogsInput) ListAuditLogsInput {
	input.Pagination = normalizePagination(input.Pagination)
	input.TargetType = strings.ToLower(strings.TrimSpace(input.TargetType))
	return input
}

func normalizePagination(input PaginationInput) PaginationInput {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 {
		input.Limit = defaultPageLimit
	}
	if input.Limit > maxPageLimit {
		input.Limit = maxPageLimit
	}
	return input
}
