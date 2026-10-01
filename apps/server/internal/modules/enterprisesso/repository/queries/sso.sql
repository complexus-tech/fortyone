-- name: AuthorizeAdmin :one
SELECT member.user_id FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id = member.user_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
  AND member.role = 'admin' AND account.is_active AND workspace.deleted_at IS NULL;

-- name: LockAdmin :one
SELECT member.user_id FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id = member.user_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
  AND member.role = 'admin' AND account.is_active AND workspace.deleted_at IS NULL
FOR UPDATE OF workspace FOR SHARE OF member, account;

-- name: GetConnection :one
SELECT id, workspace_id, issuer, client_id, client_secret_envelope, enabled, require_sso, generation, version, created_at, updated_at
FROM public.workspace_sso_connections WHERE workspace_id = sqlc.arg(workspace_id) AND archived_at IS NULL;

-- name: LockConnection :one
SELECT id, workspace_id, issuer, client_id, client_secret_envelope, enabled, require_sso, generation, version, created_at, updated_at
FROM public.workspace_sso_connections WHERE workspace_id = sqlc.arg(workspace_id) AND archived_at IS NULL FOR UPDATE;

-- name: LockLiveWorkspace :one
SELECT workspace_id FROM public.workspaces WHERE workspace_id = sqlc.arg(workspace_id) AND deleted_at IS NULL FOR SHARE;

-- name: PublicConnection :one
SELECT connection.id, connection.workspace_id, connection.issuer, connection.client_id, connection.client_secret_envelope,
    connection.enabled, connection.require_sso, connection.generation, connection.version, connection.created_at, connection.updated_at
FROM public.workspace_sso_connections AS connection
JOIN public.workspaces AS workspace ON workspace.workspace_id = connection.workspace_id
WHERE workspace.slug = sqlc.arg(slug) AND workspace.deleted_at IS NULL AND connection.archived_at IS NULL AND connection.enabled;

-- name: CreateConnection :one
INSERT INTO public.workspace_sso_connections(id, workspace_id, issuer, client_id, client_secret_envelope)
VALUES (sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(issuer), sqlc.arg(client_id), sqlc.arg(secret_envelope))
RETURNING id, workspace_id, issuer, client_id, client_secret_envelope, enabled, require_sso, generation, version, created_at, updated_at;

-- name: UpdateConnection :one
UPDATE public.workspace_sso_connections SET enabled = sqlc.arg(enabled), require_sso = sqlc.arg(require_sso),
    client_secret_envelope = sqlc.arg(secret_envelope), generation = sqlc.arg(generation), version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(id) AND archived_at IS NULL AND version = sqlc.arg(expected_version)
RETURNING id, workspace_id, issuer, client_id, client_secret_envelope, enabled, require_sso, generation, version, created_at, updated_at;

-- name: ArchiveConnection :execrows
UPDATE public.workspace_sso_connections SET enabled = FALSE, require_sso = FALSE, archived_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP, version = version + 1
WHERE workspace_id = sqlc.arg(workspace_id) AND archived_at IS NULL;

-- name: ResolveLinkedIdentity :one
SELECT identity.user_id
FROM public.workspace_sso_identities AS identity
JOIN public.workspace_members AS member ON member.workspace_id = identity.workspace_id AND member.user_id = identity.user_id
JOIN public.users AS account ON account.user_id = member.user_id
WHERE identity.connection_id = sqlc.arg(connection_id) AND identity.workspace_id = sqlc.arg(workspace_id)
  AND identity.subject = sqlc.arg(subject) AND account.is_active
FOR SHARE OF member, account;

-- name: LockLinkAccount :one
SELECT account.email FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id = member.user_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(user_id) AND account.is_active
FOR SHARE OF member, account;

-- name: LinkIdentity :exec
INSERT INTO public.workspace_sso_identities(connection_id, workspace_id, subject, user_id, last_authenticated_at)
VALUES (sqlc.arg(connection_id), sqlc.arg(workspace_id), sqlc.arg(subject), sqlc.arg(user_id), sqlc.arg(authenticated_at));

-- name: TouchIdentity :exec
UPDATE public.workspace_sso_identities SET last_authenticated_at = sqlc.arg(authenticated_at)
WHERE connection_id = sqlc.arg(connection_id) AND workspace_id = sqlc.arg(workspace_id) AND subject = sqlc.arg(subject);

-- name: AppendAudit :exec
INSERT INTO public.workspace_sso_audit_events(id,workspace_id,actor_id,connection_id,operation,metadata)
VALUES (sqlc.arg(id),sqlc.arg(workspace_id),sqlc.arg(actor_id),sqlc.arg(connection_id),sqlc.arg(operation),sqlc.arg(metadata));

-- name: RequiredConnection :one
SELECT connection.id, connection.generation
FROM public.workspace_sso_connections AS connection
JOIN public.workspaces AS workspace ON workspace.workspace_id = connection.workspace_id
JOIN public.workspace_members AS member ON member.workspace_id = workspace.workspace_id
JOIN public.users AS account ON account.user_id = member.user_id
WHERE connection.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
  AND account.is_active AND workspace.deleted_at IS NULL AND connection.archived_at IS NULL AND connection.enabled AND connection.require_sso;
