-- name: AuthorizeWorkspaceAdmin :one
SELECT account.email
FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id = member.user_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
  AND member.role = 'admin' AND account.is_active = TRUE AND workspace.deleted_at IS NULL;

-- name: LockWorkspaceAdmin :one
SELECT account.email
FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id = member.user_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
  AND member.role = 'admin' AND account.is_active = TRUE AND workspace.deleted_at IS NULL
FOR UPDATE OF workspace FOR SHARE OF member, account;

-- name: SessionAccessState :one
SELECT account.email, CAST(member.role AS text) AS role,
    COALESCE(policy.allowed_domains, CAST('{}' AS text[])) AS allowed_domains,
    COALESCE(policy.allow_guests, TRUE) AS allow_guests,
    COALESCE(policy.max_session_age_hours, 0) AS max_session_age_hours,
    epoch.revoked_before, session.revoked_at
FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id = member.user_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id
LEFT JOIN public.workspace_security_policies AS policy ON policy.workspace_id = member.workspace_id
LEFT JOIN public.workspace_member_session_epochs AS epoch ON epoch.workspace_id = member.workspace_id AND epoch.user_id = member.user_id
LEFT JOIN public.workspace_browser_sessions AS session ON session.workspace_id = member.workspace_id AND session.user_id = member.user_id AND session.session_id = sqlc.arg(session_id)
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
  AND account.is_active = TRUE AND workspace.deleted_at IS NULL
FOR SHARE OF workspace, member, account;

-- name: TrackBrowserSession :execrows
INSERT INTO public.workspace_browser_sessions (workspace_id, session_id, user_id, authenticated_at, last_seen_at, expires_at, browser_name)
VALUES (sqlc.arg(workspace_id), sqlc.arg(session_id), sqlc.arg(actor_id), sqlc.arg(authenticated_at), GREATEST(CURRENT_TIMESTAMP, CAST(sqlc.arg(authenticated_at) AS timestamptz)), sqlc.arg(expires_at), sqlc.narg(browser_name))
ON CONFLICT (workspace_id, session_id) DO UPDATE SET
    last_seen_at = CASE WHEN workspace_browser_sessions.last_seen_at < CURRENT_TIMESTAMP - INTERVAL '1 minute' THEN CURRENT_TIMESTAMP ELSE workspace_browser_sessions.last_seen_at END,
    expires_at = EXCLUDED.expires_at,
    browser_name = COALESCE(workspace_browser_sessions.browser_name, EXCLUDED.browser_name)
WHERE workspace_browser_sessions.user_id = EXCLUDED.user_id
  AND workspace_browser_sessions.authenticated_at = EXCLUDED.authenticated_at
  AND workspace_browser_sessions.revoked_at IS NULL;

-- name: GetPolicy :one
SELECT COALESCE(policy.allowed_domains, CAST('{}' AS text[])) AS allowed_domains,
    COALESCE(policy.allow_guests, TRUE) AS allow_guests,
    COALESCE(policy.max_session_age_hours, 0) AS max_session_age_hours,
    COALESCE(policy.version, CAST(0 AS bigint)) AS version, policy.updated_at
FROM public.workspaces AS workspace
LEFT JOIN public.workspace_security_policies AS policy ON policy.workspace_id = workspace.workspace_id
WHERE workspace.workspace_id = sqlc.arg(workspace_id) AND workspace.deleted_at IS NULL;

-- name: PutPolicy :one
INSERT INTO public.workspace_security_policies (workspace_id, allowed_domains, allow_guests, max_session_age_hours, version)
VALUES (sqlc.arg(workspace_id), sqlc.arg(allowed_domains), sqlc.arg(allow_guests), sqlc.arg(max_session_age_hours), 1)
ON CONFLICT (workspace_id) DO UPDATE SET
    allowed_domains = EXCLUDED.allowed_domains, allow_guests = EXCLUDED.allow_guests,
    max_session_age_hours = EXCLUDED.max_session_age_hours, version = workspace_security_policies.version + 1, updated_at = CURRENT_TIMESTAMP
RETURNING allowed_domains, allow_guests, max_session_age_hours, version, updated_at;

-- name: LockTargetMember :one
SELECT member.user_id FROM public.workspace_members AS member
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(user_id)
FOR SHARE OF member;

-- name: RevokeMemberSessions :one
INSERT INTO public.workspace_member_session_epochs (workspace_id, user_id, revoked_before)
VALUES (sqlc.arg(workspace_id), sqlc.arg(user_id), CURRENT_TIMESTAMP)
ON CONFLICT (workspace_id, user_id) DO UPDATE
SET revoked_before = GREATEST(workspace_member_session_epochs.revoked_before, EXCLUDED.revoked_before)
RETURNING revoked_before;

-- name: LockBrowserSession :one
SELECT user_id, revoked_at FROM public.workspace_browser_sessions
WHERE workspace_id = sqlc.arg(workspace_id) AND session_id = sqlc.arg(session_id)
FOR UPDATE;

-- name: RevokeBrowserSession :execrows
UPDATE public.workspace_browser_sessions SET revoked_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id) AND session_id = sqlc.arg(session_id) AND revoked_at IS NULL;

-- name: ListBrowserSessions :many
SELECT session.session_id, session.user_id, account.full_name, account.username, account.email, CAST(member.role AS text) AS role, session.browser_name,
    session.authenticated_at, session.last_seen_at, session.expires_at, session.revoked_at, epoch.revoked_before
FROM public.workspace_browser_sessions AS session
JOIN public.workspace_members AS member ON member.workspace_id = session.workspace_id AND member.user_id = session.user_id
JOIN public.users AS account ON account.user_id = member.user_id
LEFT JOIN public.workspace_member_session_epochs AS epoch ON epoch.workspace_id = session.workspace_id AND epoch.user_id = session.user_id
WHERE session.workspace_id = sqlc.arg(workspace_id)
  AND (CAST(sqlc.narg(user_id) AS uuid) IS NULL OR session.user_id = sqlc.narg(user_id))
  AND (CAST(sqlc.arg(include_revoked) AS boolean) OR (session.revoked_at IS NULL AND session.expires_at > CURRENT_TIMESTAMP AND (epoch.revoked_before IS NULL OR session.authenticated_at > epoch.revoked_before)))
ORDER BY session.last_seen_at DESC, session.session_id DESC LIMIT 501;

-- name: AppendSecurityAudit :exec
INSERT INTO public.workspace_security_audit_events (event_id, workspace_id, actor_id, resource_type, resource_id, operation, metadata)
VALUES (sqlc.arg(event_id), sqlc.arg(workspace_id), sqlc.arg(actor_id), sqlc.arg(resource_type), sqlc.arg(resource_id), sqlc.arg(operation), sqlc.arg(metadata));
