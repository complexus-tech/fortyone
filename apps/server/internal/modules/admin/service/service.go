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
	UpdateWorkspaceDeleted(ctx context.Context, input UpdateWorkspaceDeletedInput) error
	ListUsers(ctx context.Context, input ListUsersInput) (ListResult[UserSummary], error)
	GetUserOverview(ctx context.Context, userID uuid.UUID) (UserOverview, error)
	UpdateUserState(ctx context.Context, input UpdateUserStateInput) error
	ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (ListResult[AuditLog], error)
	ListAdminNotes(ctx context.Context, input ListAdminNotesInput) (ListResult[AdminNote], error)
	CreateAdminNote(ctx context.Context, input CreateAdminNoteInput) (AdminNote, error)
	InsertAuditEntry(ctx context.Context, input AuditEntryInput) error
}

type AssetResolver interface {
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
	ResolveWorkspaceLogoURL(ctx context.Context, logo string, expiry time.Duration) (string, error)
}

type SubscriptionSyncer interface {
	SyncSubscription(ctx context.Context, workspaceID uuid.UUID) error
}

const adminAssetURLExpiry = 24 * time.Hour

type Service struct {
	repo               Repository
	assetResolver      AssetResolver
	subscriptionSyncer SubscriptionSyncer
	now                func() time.Time
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

func WithSubscriptionSyncer(syncer SubscriptionSyncer) Option {
	return func(s *Service) {
		s.subscriptionSyncer = syncer
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

	trimmedReason, err := requireReason(input.Reason)
	if err != nil {
		return WorkspaceOverview{}, err
	}
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

func (s *Service) UpdateWorkspaceDeleted(ctx context.Context, actorID, workspaceID uuid.UUID, input UpdateWorkspaceDeletedInput) (WorkspaceOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.UpdateWorkspaceDeleted")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return WorkspaceOverview{}, err
	}

	reason, err := requireReason(input.Reason)
	if err != nil {
		return WorkspaceOverview{}, err
	}

	overview, err := s.repo.GetWorkspaceOverview(ctx, workspaceID)
	if err != nil {
		return WorkspaceOverview{}, err
	}

	oldDeletedAt := overview.Workspace.DeletedAt
	if err := s.repo.UpdateWorkspaceDeleted(ctx, UpdateWorkspaceDeletedInput{
		WorkspaceID: workspaceID,
		ActorUserID: actorID,
		Deleted:     input.Deleted,
		Reason:      reason,
	}); err != nil {
		return WorkspaceOverview{}, err
	}

	action := "workspace.restored"
	var newValue any
	if input.Deleted {
		action = "workspace.deleted"
		newValue = s.now().UTC()
	}

	if err := s.repo.InsertAuditEntry(ctx, AuditEntryInput{
		ActorUserID: actorID,
		TargetType:  "workspace",
		TargetID:    &workspaceID,
		WorkspaceID: &workspaceID,
		Action:      action,
		FieldName:   "deleted_at",
		OldValue:    oldDeletedAt,
		NewValue:    newValue,
		Reason:      reason,
		Metadata: map[string]any{
			"workspace_name": overview.Workspace.Name,
			"workspace_slug": overview.Workspace.Slug,
		},
	}); err != nil {
		return WorkspaceOverview{}, err
	}

	overview, err = s.repo.GetWorkspaceOverview(ctx, workspaceID)
	if err != nil {
		return WorkspaceOverview{}, err
	}
	s.resolveWorkspaceLogo(ctx, &overview.Workspace)
	return overview, nil
}

func (s *Service) RequestWorkspaceSubscriptionSync(ctx context.Context, actorID, workspaceID uuid.UUID, input RequestWorkspaceSubscriptionSyncInput) (WorkspaceOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.RequestWorkspaceSubscriptionSync")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return WorkspaceOverview{}, err
	}

	reason, err := requireReason(input.Reason)
	if err != nil {
		return WorkspaceOverview{}, err
	}

	overview, err := s.repo.GetWorkspaceOverview(ctx, workspaceID)
	if err != nil {
		return WorkspaceOverview{}, err
	}
	oldStatus := stringPtrValue(overview.Workspace.SubscriptionStatus)

	if s.subscriptionSyncer != nil {
		if err := s.subscriptionSyncer.SyncSubscription(ctx, workspaceID); err != nil {
			return WorkspaceOverview{}, err
		}
		overview, err = s.repo.GetWorkspaceOverview(ctx, workspaceID)
		if err != nil {
			return WorkspaceOverview{}, err
		}
	}

	if err := s.repo.InsertAuditEntry(ctx, AuditEntryInput{
		ActorUserID: actorID,
		TargetType:  "subscription",
		TargetID:    &workspaceID,
		WorkspaceID: &workspaceID,
		Action:      "subscription.synced",
		FieldName:   "subscription_status",
		OldValue:    oldStatus,
		NewValue:    stringPtrValue(overview.Workspace.SubscriptionStatus),
		Reason:      reason,
		Metadata: map[string]any{
			"workspace_name":          overview.Workspace.Name,
			"workspace_slug":          overview.Workspace.Slug,
			"stripe_customer_id":      stringPtrValue(overview.Workspace.StripeCustomerID),
			"stripe_subscription_id":  stringPtrValue(overview.Workspace.StripeSubscriptionID),
			"subscription_tier":       stringPtrValue(overview.Workspace.SubscriptionTier),
			"subscription_seat_count": overview.Workspace.SubscriptionSeats,
		},
	}); err != nil {
		return WorkspaceOverview{}, err
	}

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

func (s *Service) UpdateUserState(ctx context.Context, actorID, userID uuid.UUID, input UpdateUserStateInput) (UserOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.UpdateUserState")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return UserOverview{}, err
	}
	if actorID == userID {
		return UserOverview{}, ErrSelfMutation
	}

	reason, err := requireReason(input.Reason)
	if err != nil {
		return UserOverview{}, err
	}
	if input.IsActive == nil && input.IsInternal == nil {
		return UserOverview{}, ErrInvalidAdminAction
	}

	overview, err := s.repo.GetUserOverview(ctx, userID)
	if err != nil {
		return UserOverview{}, err
	}

	updateInput := UpdateUserStateInput{
		UserID:     userID,
		IsActive:   input.IsActive,
		IsInternal: input.IsInternal,
		Reason:     reason,
	}
	if err := s.repo.UpdateUserState(ctx, updateInput); err != nil {
		return UserOverview{}, err
	}

	entries := userStateAuditEntries(actorID, overview.User, updateInput)
	for _, entry := range entries {
		if err := s.repo.InsertAuditEntry(ctx, entry); err != nil {
			return UserOverview{}, err
		}
	}

	overview, err = s.repo.GetUserOverview(ctx, userID)
	if err != nil {
		return UserOverview{}, err
	}
	s.resolveUserAvatar(ctx, &overview.User)
	return overview, nil
}

func (s *Service) RequestUserSessionRevocation(ctx context.Context, actorID, userID uuid.UUID, input RequestUserSessionRevocationInput) (UserOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.RequestUserSessionRevocation")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return UserOverview{}, err
	}

	reason, err := requireReason(input.Reason)
	if err != nil {
		return UserOverview{}, err
	}

	overview, err := s.repo.GetUserOverview(ctx, userID)
	if err != nil {
		return UserOverview{}, err
	}

	if err := s.repo.InsertAuditEntry(ctx, AuditEntryInput{
		ActorUserID: actorID,
		TargetType:  "user",
		TargetID:    &userID,
		Action:      "user.session_revocation_requested",
		FieldName:   "auth_session",
		OldValue:    "active",
		NewValue:    "revocation_requested",
		Reason:      reason,
		Metadata: map[string]any{
			"user_email": overview.User.Email,
			"user_name":  overview.User.FullName,
		},
	}); err != nil {
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

func (s *Service) ListAdminNotes(ctx context.Context, actorID uuid.UUID, input ListAdminNotesInput) (ListResult[AdminNote], error) {
	ctx, span := web.AddSpan(ctx, "business.admin.ListAdminNotes")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return ListResult[AdminNote]{}, err
	}
	return s.repo.ListAdminNotes(ctx, normalizeListAdminNotesInput(input))
}

func (s *Service) CreateAdminNote(ctx context.Context, actorID uuid.UUID, input CreateAdminNoteInput) (AdminNote, error) {
	ctx, span := web.AddSpan(ctx, "business.admin.CreateAdminNote")
	defer span.End()

	if _, err := s.ensureAdmin(ctx, actorID); err != nil {
		return AdminNote{}, err
	}

	input.CreatedByUserID = actorID
	normalized, err := s.normalizeAdminNoteInput(ctx, input)
	if err != nil {
		return AdminNote{}, err
	}

	note, err := s.repo.CreateAdminNote(ctx, normalized)
	if err != nil {
		return AdminNote{}, err
	}

	targetID := normalized.TargetID
	if err := s.repo.InsertAuditEntry(ctx, AuditEntryInput{
		ActorUserID: actorID,
		TargetType:  normalized.TargetType,
		TargetID:    &targetID,
		WorkspaceID: normalized.WorkspaceID,
		Action:      "admin_note.created",
		FieldName:   "note",
		NewValue:    normalized.Body,
		Reason:      "Admin note added",
		Metadata: map[string]any{
			"note_id": note.ID,
		},
	}); err != nil {
		return AdminNote{}, err
	}

	return note, nil
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
	input.Query = strings.ToLower(strings.TrimSpace(input.Query))
	input.Action = strings.ToLower(strings.TrimSpace(input.Action))
	input.ActorQuery = strings.ToLower(strings.TrimSpace(input.ActorQuery))
	if input.From != nil {
		from := input.From.UTC()
		input.From = &from
	}
	if input.To != nil {
		to := input.To.UTC()
		input.To = &to
	}
	return input
}

func normalizeListAdminNotesInput(input ListAdminNotesInput) ListAdminNotesInput {
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

func (s *Service) normalizeAdminNoteInput(ctx context.Context, input CreateAdminNoteInput) (CreateAdminNoteInput, error) {
	targetType := strings.ToLower(strings.TrimSpace(input.TargetType))
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return CreateAdminNoteInput{}, ErrInvalidAdminNote
	}
	if input.TargetID == uuid.Nil {
		return CreateAdminNoteInput{}, ErrInvalidAdminAction
	}

	switch targetType {
	case "workspace":
		if _, err := s.repo.GetWorkspaceOverview(ctx, input.TargetID); err != nil {
			return CreateAdminNoteInput{}, err
		}
		workspaceID := input.TargetID
		return CreateAdminNoteInput{
			TargetType:      targetType,
			TargetID:        input.TargetID,
			WorkspaceID:     &workspaceID,
			Body:            body,
			CreatedByUserID: input.CreatedByUserID,
		}, nil
	case "user":
		if _, err := s.repo.GetUserOverview(ctx, input.TargetID); err != nil {
			return CreateAdminNoteInput{}, err
		}
		return CreateAdminNoteInput{
			TargetType:      targetType,
			TargetID:        input.TargetID,
			WorkspaceID:     input.WorkspaceID,
			Body:            body,
			CreatedByUserID: input.CreatedByUserID,
		}, nil
	default:
		return CreateAdminNoteInput{}, ErrInvalidAdminAction
	}
}

func userStateAuditEntries(actorID uuid.UUID, user UserSummary, input UpdateUserStateInput) []AuditEntryInput {
	entries := make([]AuditEntryInput, 0, 2)
	targetID := user.ID

	if input.IsActive != nil && *input.IsActive != user.IsActive {
		action := "user.activated"
		if !*input.IsActive {
			action = "user.deactivated"
		}
		entries = append(entries, AuditEntryInput{
			ActorUserID: actorID,
			TargetType:  "user",
			TargetID:    &targetID,
			Action:      action,
			FieldName:   "is_active",
			OldValue:    user.IsActive,
			NewValue:    *input.IsActive,
			Reason:      input.Reason,
			Metadata: map[string]any{
				"user_email": user.Email,
				"user_name":  user.FullName,
			},
		})
	}

	if input.IsInternal != nil && *input.IsInternal != user.IsInternal {
		action := "user.internal_granted"
		if !*input.IsInternal {
			action = "user.internal_revoked"
		}
		entries = append(entries, AuditEntryInput{
			ActorUserID: actorID,
			TargetType:  "user",
			TargetID:    &targetID,
			Action:      action,
			FieldName:   "is_internal",
			OldValue:    user.IsInternal,
			NewValue:    *input.IsInternal,
			Reason:      input.Reason,
			Metadata: map[string]any{
				"user_email": user.Email,
				"user_name":  user.FullName,
			},
		})
	}

	if len(entries) == 0 {
		entries = append(entries, AuditEntryInput{
			ActorUserID: actorID,
			TargetType:  "user",
			TargetID:    &targetID,
			Action:      "user.state_reviewed",
			FieldName:   "state",
			Reason:      input.Reason,
			Metadata: map[string]any{
				"user_email": user.Email,
				"user_name":  user.FullName,
			},
		})
	}

	return entries
}

func requireReason(reason string) (string, error) {
	trimmedReason := strings.TrimSpace(reason)
	if trimmedReason == "" {
		return "", ErrReasonRequired
	}
	return trimmedReason, nil
}

func stringPtrValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
