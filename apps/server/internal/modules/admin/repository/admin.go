package adminrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	admin "github.com/complexus-tech/projects-api/internal/modules/admin/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

type dbUserSummary struct {
	ID                  uuid.UUID  `db:"user_id"`
	Username            string     `db:"username"`
	Email               string     `db:"email"`
	FullName            *string    `db:"full_name"`
	AvatarURL           *string    `db:"avatar_url"`
	IsActive            bool       `db:"is_active"`
	IsSystem            bool       `db:"is_system"`
	IsInternal          bool       `db:"is_internal"`
	LastLoginAt         *time.Time `db:"last_login_at"`
	LastUsedWorkspaceID *uuid.UUID `db:"last_used_workspace_id"`
	LastUsedWorkspace   *string    `db:"last_used_workspace"`
	GitHubUsername      *string    `db:"github_username"`
	WorkspaceCount      int        `db:"workspace_count"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
	TotalCount          int        `db:"total_count"`
}

type dbUserMembership struct {
	WorkspaceID   uuid.UUID `db:"workspace_id"`
	WorkspaceName string    `db:"workspace_name"`
	WorkspaceSlug string    `db:"workspace_slug"`
	Role          string    `db:"role"`
	JoinedAt      time.Time `db:"joined_at"`
}

type dbWorkspaceSummary struct {
	ID                   uuid.UUID  `db:"workspace_id"`
	Name                 string     `db:"name"`
	Slug                 string     `db:"slug"`
	AvatarURL            *string    `db:"avatar_url"`
	Color                string     `db:"color"`
	TeamSize             string     `db:"team_size"`
	TrialEndsOn          *time.Time `db:"trial_ends_on"`
	DeletedAt            *time.Time `db:"deleted_at"`
	LastAccessedAt       *time.Time `db:"last_accessed_at"`
	CreatedByUserID      *uuid.UUID `db:"created_by"`
	CreatedByEmail       *string    `db:"created_by_email"`
	CreatedByName        *string    `db:"created_by_name"`
	MemberCount          int        `db:"member_count"`
	TeamCount            int        `db:"team_count"`
	StoryCount           int        `db:"story_count"`
	SubscriptionTier     *string    `db:"subscription_tier"`
	SubscriptionStatus   *string    `db:"subscription_status"`
	SubscriptionSeats    *int       `db:"subscription_seats"`
	StripeCustomerID     *string    `db:"stripe_customer_id"`
	StripeSubscriptionID *string    `db:"stripe_subscription_id"`
	SlackInstalled       bool       `db:"slack_installed"`
	GitHubInstalled      bool       `db:"github_installed"`
	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
	TotalCount           int        `db:"total_count"`
}

type dbWorkspaceMember struct {
	UserID     uuid.UUID `db:"user_id"`
	Email      string    `db:"email"`
	FullName   *string   `db:"full_name"`
	Role       string    `db:"role"`
	IsInternal bool      `db:"is_internal"`
	JoinedAt   time.Time `db:"joined_at"`
}

type dbAuditLog struct {
	ID            uuid.UUID       `db:"id"`
	ActorUserID   uuid.UUID       `db:"actor_user_id"`
	ActorEmail    string          `db:"actor_email"`
	ActorName     *string         `db:"actor_name"`
	TargetType    string          `db:"target_type"`
	TargetID      *uuid.UUID      `db:"target_id"`
	WorkspaceID   *uuid.UUID      `db:"workspace_id"`
	WorkspaceName *string         `db:"workspace_name"`
	WorkspaceSlug *string         `db:"workspace_slug"`
	Action        string          `db:"action"`
	FieldName     *string         `db:"field_name"`
	OldValue      json.RawMessage `db:"old_value"`
	NewValue      json.RawMessage `db:"new_value"`
	Reason        *string         `db:"reason"`
	Metadata      json.RawMessage `db:"metadata"`
	CreatedAt     time.Time       `db:"created_at"`
	TotalCount    int             `db:"total_count"`
}

type dbDashboardSummary struct {
	TotalWorkspaces      int `db:"total_workspaces"`
	ActiveTrials         int `db:"active_trials"`
	ExpiredTrials        int `db:"expired_trials"`
	PaidWorkspaces       int `db:"paid_workspaces"`
	DeletedWorkspaces    int `db:"deleted_workspaces"`
	TotalUsers           int `db:"total_users"`
	InternalUsers        int `db:"internal_users"`
	ActiveSubscriptions  int `db:"active_subscriptions"`
	SlackInstallations   int `db:"slack_installations"`
	GitHubInstallations  int `db:"github_installations"`
	RecentAdminAuditLogs int `db:"recent_admin_audit_logs"`
}

func (r *repo) GetAdminUser(ctx context.Context, userID uuid.UUID) (admin.UserSummary, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.GetAdminUser")
	defer span.End()

	const query = `
		SELECT
			u.user_id,
			u.username,
			u.email,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.is_system,
			u.is_internal,
			u.last_login_at,
			u.last_used_workspace_id,
			lw.name AS last_used_workspace,
			u.github_username,
			(
				SELECT COUNT(*)
				FROM public.workspace_members wm
				WHERE wm.user_id = u.user_id
			) AS workspace_count,
			u.created_at,
			u.updated_at,
			1 AS total_count
		FROM public.users u
		LEFT JOIN public.workspaces lw ON lw.workspace_id = u.last_used_workspace_id
		WHERE u.user_id = $1
	`

	var row dbUserSummary
	if err := r.db.GetContext(ctx, &row, query, userID); err != nil {
		if err == sql.ErrNoRows {
			return admin.UserSummary{}, admin.ErrNotFound
		}
		return admin.UserSummary{}, fmt.Errorf("get admin user: %w", err)
	}

	return toUserSummary(row), nil
}

func (r *repo) GetDashboardSummary(ctx context.Context) (admin.DashboardSummary, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.GetDashboardSummary")
	defer span.End()

	const query = `
		SELECT
			(SELECT COUNT(*) FROM public.workspaces WHERE deleted_at IS NULL) AS total_workspaces,
			(SELECT COUNT(*) FROM public.workspaces WHERE deleted_at IS NULL AND trial_ends_on > now()) AS active_trials,
			(SELECT COUNT(*) FROM public.workspaces WHERE deleted_at IS NULL AND trial_ends_on IS NOT NULL AND trial_ends_on <= now()) AS expired_trials,
			(
				SELECT COUNT(DISTINCT workspace_id)
				FROM public.workspace_subscriptions
				WHERE subscription_status IN ('active', 'trialing', 'past_due')
					AND COALESCE(subscription_tier::text, 'free') <> 'free'
			) AS paid_workspaces,
			(SELECT COUNT(*) FROM public.workspaces WHERE deleted_at IS NOT NULL) AS deleted_workspaces,
			(SELECT COUNT(*) FROM public.users) AS total_users,
			(SELECT COUNT(*) FROM public.users WHERE is_internal = true) AS internal_users,
			(
				SELECT COUNT(*)
				FROM public.workspace_subscriptions
				WHERE subscription_status IN ('active', 'trialing', 'past_due')
			) AS active_subscriptions,
			(SELECT COUNT(*) FROM public.slack_workspaces WHERE is_active = true) AS slack_installations,
			(SELECT COUNT(*) FROM public.github_installations WHERE is_active = true) AS github_installations,
			(SELECT COUNT(*) FROM public.admin_audit_logs WHERE created_at >= now() - interval '30 days') AS recent_admin_audit_logs
	`

	var row dbDashboardSummary
	if err := r.db.GetContext(ctx, &row, query); err != nil {
		return admin.DashboardSummary{}, fmt.Errorf("get admin dashboard summary: %w", err)
	}
	return admin.DashboardSummary{
		TotalWorkspaces:      row.TotalWorkspaces,
		ActiveTrials:         row.ActiveTrials,
		ExpiredTrials:        row.ExpiredTrials,
		PaidWorkspaces:       row.PaidWorkspaces,
		DeletedWorkspaces:    row.DeletedWorkspaces,
		TotalUsers:           row.TotalUsers,
		InternalUsers:        row.InternalUsers,
		ActiveSubscriptions:  row.ActiveSubscriptions,
		SlackInstallations:   row.SlackInstallations,
		GitHubInstallations:  row.GitHubInstallations,
		RecentAdminAuditLogs: row.RecentAdminAuditLogs,
	}, nil
}

func (r *repo) ListWorkspaces(ctx context.Context, input admin.ListWorkspacesInput) (admin.ListResult[admin.WorkspaceSummary], error) {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.ListWorkspaces")
	defer span.End()

	where, params := workspaceFilters(input)
	params["limit"] = input.Pagination.Limit
	params["offset"] = pageOffset(input.Pagination)

	query := fmt.Sprintf(`
		SELECT
			w.workspace_id,
			w.name,
			w.slug,
			w.avatar_url,
			w.color,
			w.team_size,
			w.trial_ends_on,
			w.deleted_at,
			w.last_accessed_at,
			w.created_by,
			creator.email AS created_by_email,
			creator.full_name AS created_by_name,
			(SELECT COUNT(*) FROM public.workspace_members wm WHERE wm.workspace_id = w.workspace_id) AS member_count,
			(SELECT COUNT(*) FROM public.teams t WHERE t.workspace_id = w.workspace_id) AS team_count,
			(SELECT COUNT(*) FROM public.stories s WHERE s.workspace_id = w.workspace_id AND s.deleted_at IS NULL) AS story_count,
			ws.subscription_tier::text AS subscription_tier,
			ws.subscription_status,
			ws.seat_count AS subscription_seats,
			ws.stripe_customer_id,
			ws.stripe_subscription_id,
			EXISTS (
				SELECT 1 FROM public.slack_workspaces sw
				WHERE sw.workspace_id = w.workspace_id AND sw.is_active = true
			) AS slack_installed,
			EXISTS (
				SELECT 1 FROM public.github_installations gi
				WHERE gi.workspace_id = w.workspace_id AND gi.is_active = true
			) AS github_installed,
			w.created_at,
			w.updated_at,
			COUNT(*) OVER() AS total_count
		FROM public.workspaces w
		LEFT JOIN public.users creator ON creator.user_id = w.created_by
		LEFT JOIN LATERAL (
			SELECT
				subscription_tier,
				subscription_status,
				seat_count,
				stripe_customer_id,
				stripe_subscription_id
			FROM public.workspace_subscriptions
			WHERE workspace_id = w.workspace_id
			ORDER BY updated_at DESC
			LIMIT 1
		) ws ON true
		%s
		ORDER BY w.created_at DESC
		LIMIT :limit OFFSET :offset
	`, where)

	rows, err := selectNamed[dbWorkspaceSummary](ctx, r.db, query, params)
	if err != nil {
		return admin.ListResult[admin.WorkspaceSummary]{}, fmt.Errorf("list admin workspaces: %w", err)
	}

	items := make([]admin.WorkspaceSummary, len(rows))
	for i, row := range rows {
		items[i] = toWorkspaceSummary(row)
	}
	return admin.ListResult[admin.WorkspaceSummary]{
		Items:      items,
		Pagination: paginationFromRows(input.Pagination, len(rows), workspaceTotal(rows)),
	}, nil
}

func (r *repo) GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID) (admin.WorkspaceOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.GetWorkspaceOverview")
	defer span.End()

	workspace, err := r.getWorkspaceSummary(ctx, workspaceID)
	if err != nil {
		return admin.WorkspaceOverview{}, err
	}

	const membersQuery = `
		SELECT
			u.user_id,
			u.email,
			u.full_name,
			wm.role::text AS role,
			u.is_internal,
			wm.created_at AS joined_at
		FROM public.workspace_members wm
		INNER JOIN public.users u ON u.user_id = wm.user_id
		WHERE wm.workspace_id = $1
		ORDER BY
			CASE wm.role::text WHEN 'admin' THEN 0 ELSE 1 END,
			u.full_name NULLS LAST,
			u.email
	`

	var memberRows []dbWorkspaceMember
	if err := r.db.SelectContext(ctx, &memberRows, membersQuery, workspaceID); err != nil {
		return admin.WorkspaceOverview{}, fmt.Errorf("get admin workspace members: %w", err)
	}

	members := make([]admin.WorkspaceMember, len(memberRows))
	for i, row := range memberRows {
		members[i] = admin.WorkspaceMember{
			UserID:     row.UserID,
			Email:      row.Email,
			FullName:   derefString(row.FullName),
			Role:       row.Role,
			IsInternal: row.IsInternal,
			JoinedAt:   row.JoinedAt,
		}
	}

	return admin.WorkspaceOverview{
		Workspace: workspace,
		Members:   members,
	}, nil
}

func (r *repo) UpdateWorkspaceTrial(ctx context.Context, input admin.UpdateWorkspaceTrialInput) error {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.UpdateWorkspaceTrial")
	defer span.End()

	const query = `
		UPDATE public.workspaces
		SET trial_ends_on = $2,
			updated_at = now()
		WHERE workspace_id = $1
	`
	result, err := r.db.ExecContext(ctx, query, input.WorkspaceID, input.TrialEndsOn)
	if err != nil {
		return fmt.Errorf("update workspace trial: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read workspace trial update result: %w", err)
	}
	if affected == 0 {
		return admin.ErrNotFound
	}
	return nil
}

func (r *repo) ListUsers(ctx context.Context, input admin.ListUsersInput) (admin.ListResult[admin.UserSummary], error) {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.ListUsers")
	defer span.End()

	where, params := userFilters(input)
	params["limit"] = input.Pagination.Limit
	params["offset"] = pageOffset(input.Pagination)

	query := fmt.Sprintf(`
		SELECT
			u.user_id,
			u.username,
			u.email,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.is_system,
			u.is_internal,
			u.last_login_at,
			u.last_used_workspace_id,
			lw.name AS last_used_workspace,
			u.github_username,
			(SELECT COUNT(*) FROM public.workspace_members wm WHERE wm.user_id = u.user_id) AS workspace_count,
			u.created_at,
			u.updated_at,
			COUNT(*) OVER() AS total_count
		FROM public.users u
		LEFT JOIN public.workspaces lw ON lw.workspace_id = u.last_used_workspace_id
		%s
		ORDER BY u.created_at DESC
		LIMIT :limit OFFSET :offset
	`, where)

	rows, err := selectNamed[dbUserSummary](ctx, r.db, query, params)
	if err != nil {
		return admin.ListResult[admin.UserSummary]{}, fmt.Errorf("list admin users: %w", err)
	}

	items := make([]admin.UserSummary, len(rows))
	for i, row := range rows {
		items[i] = toUserSummary(row)
	}
	return admin.ListResult[admin.UserSummary]{
		Items:      items,
		Pagination: paginationFromRows(input.Pagination, len(rows), userTotal(rows)),
	}, nil
}

func (r *repo) GetUserOverview(ctx context.Context, userID uuid.UUID) (admin.UserOverview, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.GetUserOverview")
	defer span.End()

	user, err := r.getUserSummary(ctx, userID)
	if err != nil {
		return admin.UserOverview{}, err
	}

	const membershipsQuery = `
		SELECT
			w.workspace_id,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			wm.role::text AS role,
			wm.created_at AS joined_at
		FROM public.workspace_members wm
		INNER JOIN public.workspaces w ON w.workspace_id = wm.workspace_id
		WHERE wm.user_id = $1
		ORDER BY wm.created_at DESC
	`

	var rows []dbUserMembership
	if err := r.db.SelectContext(ctx, &rows, membershipsQuery, userID); err != nil {
		return admin.UserOverview{}, fmt.Errorf("get admin user memberships: %w", err)
	}

	memberships := make([]admin.UserMembership, len(rows))
	for i, row := range rows {
		memberships[i] = admin.UserMembership{
			WorkspaceID:   row.WorkspaceID,
			WorkspaceName: row.WorkspaceName,
			WorkspaceSlug: row.WorkspaceSlug,
			Role:          row.Role,
			JoinedAt:      row.JoinedAt,
		}
	}

	return admin.UserOverview{
		User:        user,
		Memberships: memberships,
	}, nil
}

func (r *repo) ListAuditLogs(ctx context.Context, input admin.ListAuditLogsInput) (admin.ListResult[admin.AuditLog], error) {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.ListAuditLogs")
	defer span.End()

	where, params := auditLogFilters(input)
	params["limit"] = input.Pagination.Limit
	params["offset"] = pageOffset(input.Pagination)

	query := fmt.Sprintf(`
		SELECT
			a.id,
			a.actor_user_id,
			actor.email AS actor_email,
			actor.full_name AS actor_name,
			a.target_type,
			a.target_id,
			a.workspace_id,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			a.action,
			a.field_name,
			a.old_value,
			a.new_value,
			a.reason,
			a.metadata,
			a.created_at,
			COUNT(*) OVER() AS total_count
		FROM public.admin_audit_logs a
		INNER JOIN public.users actor ON actor.user_id = a.actor_user_id
		LEFT JOIN public.workspaces w ON w.workspace_id = a.workspace_id
		%s
		ORDER BY a.created_at DESC
		LIMIT :limit OFFSET :offset
	`, where)

	rows, err := selectNamed[dbAuditLog](ctx, r.db, query, params)
	if err != nil {
		return admin.ListResult[admin.AuditLog]{}, fmt.Errorf("list admin audit logs: %w", err)
	}

	items := make([]admin.AuditLog, len(rows))
	for i, row := range rows {
		items[i] = toAuditLog(row)
	}
	return admin.ListResult[admin.AuditLog]{
		Items:      items,
		Pagination: paginationFromRows(input.Pagination, len(rows), auditLogTotal(rows)),
	}, nil
}

func (r *repo) InsertAuditEntry(ctx context.Context, input admin.AuditEntryInput) error {
	ctx, span := web.AddSpan(ctx, "business.repository.admin.InsertAuditEntry")
	defer span.End()

	oldValue, err := marshalNullableJSON(input.OldValue)
	if err != nil {
		return fmt.Errorf("marshal old audit value: %w", err)
	}
	newValue, err := marshalNullableJSON(input.NewValue)
	if err != nil {
		return fmt.Errorf("marshal new audit value: %w", err)
	}
	metadata, err := marshalJSON(input.Metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	const query = `
		INSERT INTO public.admin_audit_logs (
			actor_user_id,
			target_type,
			target_id,
			workspace_id,
			action,
			field_name,
			old_value,
			new_value,
			reason,
			metadata
		)
		VALUES (
			:actor_user_id,
			:target_type,
			:target_id,
			:workspace_id,
			:action,
			:field_name,
			:old_value,
			:new_value,
			:reason,
			:metadata
		)
	`

	params := map[string]any{
		"actor_user_id": input.ActorUserID,
		"target_type":   input.TargetType,
		"target_id":     input.TargetID,
		"workspace_id":  input.WorkspaceID,
		"action":        input.Action,
		"field_name":    nullableString(input.FieldName),
		"old_value":     oldValue,
		"new_value":     newValue,
		"reason":        nullableString(input.Reason),
		"metadata":      metadata,
	}

	if _, err := r.db.NamedExecContext(ctx, query, params); err != nil {
		return fmt.Errorf("insert admin audit log: %w", err)
	}
	return nil
}

func (r *repo) getWorkspaceSummary(ctx context.Context, workspaceID uuid.UUID) (admin.WorkspaceSummary, error) {
	input := admin.ListWorkspacesInput{
		Pagination: admin.PaginationInput{Page: 1, Limit: 1},
	}
	where, params := workspaceFilters(input)
	if where == "" {
		where = "WHERE w.workspace_id = :workspace_id"
	} else {
		where += " AND w.workspace_id = :workspace_id"
	}
	params["workspace_id"] = workspaceID
	params["limit"] = 1
	params["offset"] = 0

	query := fmt.Sprintf(`
		SELECT
			w.workspace_id,
			w.name,
			w.slug,
			w.avatar_url,
			w.color,
			w.team_size,
			w.trial_ends_on,
			w.deleted_at,
			w.last_accessed_at,
			w.created_by,
			creator.email AS created_by_email,
			creator.full_name AS created_by_name,
			(SELECT COUNT(*) FROM public.workspace_members wm WHERE wm.workspace_id = w.workspace_id) AS member_count,
			(SELECT COUNT(*) FROM public.teams t WHERE t.workspace_id = w.workspace_id) AS team_count,
			(SELECT COUNT(*) FROM public.stories s WHERE s.workspace_id = w.workspace_id AND s.deleted_at IS NULL) AS story_count,
			ws.subscription_tier::text AS subscription_tier,
			ws.subscription_status,
			ws.seat_count AS subscription_seats,
			ws.stripe_customer_id,
			ws.stripe_subscription_id,
			EXISTS (
				SELECT 1 FROM public.slack_workspaces sw
				WHERE sw.workspace_id = w.workspace_id AND sw.is_active = true
			) AS slack_installed,
			EXISTS (
				SELECT 1 FROM public.github_installations gi
				WHERE gi.workspace_id = w.workspace_id AND gi.is_active = true
			) AS github_installed,
			w.created_at,
			w.updated_at,
			1 AS total_count
		FROM public.workspaces w
		LEFT JOIN public.users creator ON creator.user_id = w.created_by
		LEFT JOIN LATERAL (
			SELECT
				subscription_tier,
				subscription_status,
				seat_count,
				stripe_customer_id,
				stripe_subscription_id
			FROM public.workspace_subscriptions
			WHERE workspace_id = w.workspace_id
			ORDER BY updated_at DESC
			LIMIT 1
		) ws ON true
		%s
		LIMIT :limit OFFSET :offset
	`, where)

	rows, err := selectNamed[dbWorkspaceSummary](ctx, r.db, query, params)
	if err != nil {
		return admin.WorkspaceSummary{}, fmt.Errorf("get admin workspace: %w", err)
	}
	if len(rows) == 0 {
		return admin.WorkspaceSummary{}, admin.ErrNotFound
	}
	return toWorkspaceSummary(rows[0]), nil
}

func (r *repo) getUserSummary(ctx context.Context, userID uuid.UUID) (admin.UserSummary, error) {
	const query = `
		SELECT
			u.user_id,
			u.username,
			u.email,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.is_system,
			u.is_internal,
			u.last_login_at,
			u.last_used_workspace_id,
			lw.name AS last_used_workspace,
			u.github_username,
			(SELECT COUNT(*) FROM public.workspace_members wm WHERE wm.user_id = u.user_id) AS workspace_count,
			u.created_at,
			u.updated_at,
			1 AS total_count
		FROM public.users u
		LEFT JOIN public.workspaces lw ON lw.workspace_id = u.last_used_workspace_id
		WHERE u.user_id = $1
	`

	var row dbUserSummary
	if err := r.db.GetContext(ctx, &row, query, userID); err != nil {
		if err == sql.ErrNoRows {
			return admin.UserSummary{}, admin.ErrNotFound
		}
		return admin.UserSummary{}, fmt.Errorf("get admin user: %w", err)
	}
	return toUserSummary(row), nil
}

func workspaceFilters(input admin.ListWorkspacesInput) (string, map[string]any) {
	clauses := make([]string, 0, 2)
	params := make(map[string]any)

	if input.Query != "" {
		clauses = append(clauses, "(w.name ILIKE :query OR w.slug ILIKE :query OR creator.email ILIKE :query)")
		params["query"] = "%" + input.Query + "%"
	}

	switch input.Status {
	case "active":
		clauses = append(clauses, "w.deleted_at IS NULL")
	case "trialing":
		clauses = append(clauses, "w.deleted_at IS NULL AND w.trial_ends_on > now()")
	case "expired":
		clauses = append(clauses, "w.deleted_at IS NULL AND w.trial_ends_on IS NOT NULL AND w.trial_ends_on <= now()")
	case "paid":
		clauses = append(clauses, "w.deleted_at IS NULL AND ws.subscription_status IN ('active', 'trialing', 'past_due') AND COALESCE(ws.subscription_tier::text, 'free') <> 'free'")
	case "deleted":
		clauses = append(clauses, "w.deleted_at IS NOT NULL")
	}

	if len(clauses) == 0 {
		return "", params
	}
	return "WHERE " + strings.Join(clauses, " AND "), params
}

func userFilters(input admin.ListUsersInput) (string, map[string]any) {
	if input.Query == "" {
		return "", map[string]any{}
	}
	return "WHERE (u.email ILIKE :query OR u.username ILIKE :query OR u.full_name ILIKE :query)", map[string]any{
		"query": "%" + input.Query + "%",
	}
}

func auditLogFilters(input admin.ListAuditLogsInput) (string, map[string]any) {
	clauses := make([]string, 0, 2)
	params := make(map[string]any)

	if input.WorkspaceID != nil {
		clauses = append(clauses, "a.workspace_id = :workspace_id")
		params["workspace_id"] = *input.WorkspaceID
	}
	if input.TargetType != "" {
		clauses = append(clauses, "a.target_type = :target_type")
		params["target_type"] = input.TargetType
	}

	if len(clauses) == 0 {
		return "", params
	}
	return "WHERE " + strings.Join(clauses, " AND "), params
}

func selectNamed[T any](ctx context.Context, db *sqlx.DB, query string, params map[string]any) ([]T, error) {
	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var rows []T
	if err := stmt.SelectContext(ctx, &rows, params); err != nil {
		return nil, err
	}
	return rows, nil
}

func pageOffset(input admin.PaginationInput) int {
	return (input.Page - 1) * input.Limit
}

func paginationFromRows(input admin.PaginationInput, rowCount, total int) admin.Pagination {
	if total == 0 && rowCount > 0 {
		total = rowCount
	}
	return admin.Pagination{
		Total:  total,
		Page:   input.Page,
		Limit:  input.Limit,
		Offset: pageOffset(input),
	}
}

func workspaceTotal(rows []dbWorkspaceSummary) int {
	if len(rows) == 0 {
		return 0
	}
	return rows[0].TotalCount
}

func userTotal(rows []dbUserSummary) int {
	if len(rows) == 0 {
		return 0
	}
	return rows[0].TotalCount
}

func auditLogTotal(rows []dbAuditLog) int {
	if len(rows) == 0 {
		return 0
	}
	return rows[0].TotalCount
}

func toUserSummary(row dbUserSummary) admin.UserSummary {
	return admin.UserSummary{
		ID:                  row.ID,
		Username:            row.Username,
		Email:               row.Email,
		FullName:            derefString(row.FullName),
		AvatarURL:           derefString(row.AvatarURL),
		IsActive:            row.IsActive,
		IsSystem:            row.IsSystem,
		IsInternal:          row.IsInternal,
		LastLoginAt:         row.LastLoginAt,
		LastUsedWorkspaceID: row.LastUsedWorkspaceID,
		LastUsedWorkspace:   row.LastUsedWorkspace,
		GitHubUsername:      row.GitHubUsername,
		WorkspaceCount:      row.WorkspaceCount,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

func toWorkspaceSummary(row dbWorkspaceSummary) admin.WorkspaceSummary {
	return admin.WorkspaceSummary{
		ID:                   row.ID,
		Name:                 row.Name,
		Slug:                 row.Slug,
		AvatarURL:            row.AvatarURL,
		Color:                row.Color,
		TeamSize:             row.TeamSize,
		TrialEndsOn:          row.TrialEndsOn,
		DeletedAt:            row.DeletedAt,
		LastAccessedAt:       row.LastAccessedAt,
		CreatedByUserID:      row.CreatedByUserID,
		CreatedByEmail:       row.CreatedByEmail,
		CreatedByName:        row.CreatedByName,
		MemberCount:          row.MemberCount,
		TeamCount:            row.TeamCount,
		StoryCount:           row.StoryCount,
		SubscriptionTier:     row.SubscriptionTier,
		SubscriptionStatus:   row.SubscriptionStatus,
		SubscriptionSeats:    row.SubscriptionSeats,
		StripeCustomerID:     row.StripeCustomerID,
		StripeSubscriptionID: row.StripeSubscriptionID,
		SlackInstalled:       row.SlackInstalled,
		GitHubInstalled:      row.GitHubInstalled,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func toAuditLog(row dbAuditLog) admin.AuditLog {
	return admin.AuditLog{
		ID:            row.ID,
		ActorUserID:   row.ActorUserID,
		ActorEmail:    row.ActorEmail,
		ActorName:     derefString(row.ActorName),
		TargetType:    row.TargetType,
		TargetID:      row.TargetID,
		WorkspaceID:   row.WorkspaceID,
		WorkspaceName: row.WorkspaceName,
		WorkspaceSlug: row.WorkspaceSlug,
		Action:        row.Action,
		FieldName:     derefString(row.FieldName),
		OldValue:      decodeJSON(row.OldValue),
		NewValue:      decodeJSON(row.NewValue),
		Reason:        derefString(row.Reason),
		Metadata:      decodeJSON(row.Metadata),
		CreatedAt:     row.CreatedAt,
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func marshalNullableJSON(value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	return marshalJSON(value)
}

func marshalJSON(value any) ([]byte, error) {
	if value == nil {
		value = map[string]any{}
	}
	return json.Marshal(value)
}

func decodeJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	return value
}
