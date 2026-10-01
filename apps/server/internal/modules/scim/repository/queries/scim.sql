-- name: AuthorizeAdmin :one
SELECT member.user_id FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id = member.user_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
  AND member.role = 'admin' AND account.is_active AND NOT account.is_system AND workspace.deleted_at IS NULL;

-- name: LockAdmin :one
SELECT member.user_id FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id = member.user_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
  AND member.role = 'admin' AND account.is_active AND NOT account.is_system AND workspace.deleted_at IS NULL
FOR UPDATE OF workspace FOR SHARE OF member, account;

-- name: Authenticate :one
SELECT credential.workspace_id, credential.id, credential.issuer_id
FROM public.workspace_scim_credentials AS credential
JOIN public.workspaces AS workspace ON workspace.workspace_id = credential.workspace_id
JOIN public.workspace_members AS member ON member.workspace_id = credential.workspace_id AND member.user_id = credential.issuer_id
JOIN public.users AS account ON account.user_id = member.user_id
WHERE workspace.slug = sqlc.arg(slug) AND credential.token_digest = sqlc.arg(digest)
  AND credential.revoked_at IS NULL AND credential.expires_at > CURRENT_TIMESTAMP
  AND workspace.deleted_at IS NULL AND member.role = 'admin' AND account.is_active AND NOT account.is_system;

-- name: CheckScope :one
SELECT credential.id FROM public.workspace_scim_credentials AS credential
JOIN public.workspaces AS workspace ON workspace.workspace_id = credential.workspace_id
JOIN public.workspace_members AS member ON member.workspace_id = credential.workspace_id AND member.user_id = credential.issuer_id
JOIN public.users AS account ON account.user_id = member.user_id
WHERE credential.id = sqlc.arg(credential_id) AND credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.issuer_id = sqlc.arg(issuer_id) AND credential.revoked_at IS NULL AND credential.expires_at > CURRENT_TIMESTAMP
  AND workspace.deleted_at IS NULL AND member.role = 'admin' AND account.is_active AND NOT account.is_system;

-- name: LockCredential :one
SELECT credential.id FROM public.workspace_scim_credentials AS credential
WHERE credential.id = sqlc.arg(credential_id) AND credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.issuer_id = sqlc.arg(issuer_id) AND credential.revoked_at IS NULL AND credential.expires_at > CURRENT_TIMESTAMP
FOR UPDATE;

-- name: ListCredentials :many
SELECT id,name,token_prefix,created_at,expires_at,revoked_at FROM public.workspace_scim_credentials
WHERE workspace_id = sqlc.arg(workspace_id) ORDER BY revoked_at NULLS FIRST,expires_at DESC,created_at DESC,id DESC LIMIT 100;

-- name: CountLiveCredentials :one
SELECT count(*) FROM public.workspace_scim_credentials WHERE workspace_id = sqlc.arg(workspace_id)
 AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP;

-- name: CreateCredential :one
INSERT INTO public.workspace_scim_credentials(id,workspace_id,issuer_id,name,token_digest,token_prefix,expires_at)
VALUES(sqlc.arg(id),sqlc.arg(workspace_id),sqlc.arg(issuer_id),sqlc.arg(name),sqlc.arg(digest),sqlc.arg(prefix),sqlc.arg(expires_at))
RETURNING id,name,token_prefix,created_at,expires_at,revoked_at;

-- name: RevokeCredential :execrows
UPDATE public.workspace_scim_credentials SET revoked_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(id) AND revoked_at IS NULL;

-- name: GetUser :one
SELECT id,workspace_id,user_id,user_name,external_id,active,profile,previous_role,version,created_at,updated_at
FROM public.workspace_scim_users WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: LockUser :one
SELECT id,workspace_id,user_id,user_name,external_id,active,profile,previous_role,version,created_at,updated_at
FROM public.workspace_scim_users WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(id) AND deleted_at IS NULL FOR UPDATE;

-- name: CountUsers :one
SELECT count(*) FROM public.workspace_scim_users
WHERE workspace_id = sqlc.arg(workspace_id) AND deleted_at IS NULL
AND (sqlc.arg(field) = '' OR (sqlc.arg(field) = 'userName' AND lower(user_name) = lower(sqlc.arg(value)))
 OR (sqlc.arg(field) = 'externalId' AND external_id = sqlc.arg(value))
 OR (sqlc.arg(field) = 'id' AND CAST(id AS text) = sqlc.arg(value))
 OR (sqlc.arg(field) = 'active' AND active = sqlc.narg(active)));

-- name: ListUsers :many
SELECT id,workspace_id,user_id,user_name,external_id,active,profile,previous_role,version,created_at,updated_at
FROM public.workspace_scim_users
WHERE workspace_id = sqlc.arg(workspace_id) AND deleted_at IS NULL
AND (sqlc.arg(field) = '' OR (sqlc.arg(field) = 'userName' AND lower(user_name) = lower(sqlc.arg(value)))
 OR (sqlc.arg(field) = 'externalId' AND external_id = sqlc.arg(value))
 OR (sqlc.arg(field) = 'id' AND CAST(id AS text) = sqlc.arg(value))
 OR (sqlc.arg(field) = 'active' AND active = sqlc.narg(active)))
ORDER BY created_at,id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: FindAccount :one
SELECT user_id,is_active,is_system,is_internal FROM public.users WHERE lower(email) = lower(sqlc.arg(email)) FOR SHARE;

-- name: CreateAccount :one
WITH account AS (
 INSERT INTO public.users(user_id,username,email,full_name,timezone)
 VALUES(sqlc.arg(id),sqlc.arg(user_name),sqlc.arg(email),sqlc.arg(full_name),'UTC')
 RETURNING user_id,is_active,is_system,is_internal,full_name,email,created_at
), alert AS (
 INSERT INTO public.internal_slack_alerts(dedupe_key,kind,user_id,payload)
 SELECT 'account_created:' || CAST(user_id AS text),'account_created',user_id,
 jsonb_build_object('name',full_name,'email',email,'occurred_at',created_at) FROM account
 ON CONFLICT(dedupe_key) DO NOTHING
)
SELECT user_id,is_active,is_system,is_internal FROM account;

-- name: LockAccount :one
SELECT is_active,is_system,is_internal FROM public.users WHERE user_id = sqlc.arg(user_id) FOR SHARE;

-- name: CurrentRole :one
SELECT CAST(role AS text) FROM public.workspace_members WHERE workspace_id = sqlc.arg(workspace_id) AND user_id = sqlc.arg(user_id) FOR UPDATE;

-- name: CountAdmins :one
SELECT count(*) FROM public.workspace_members AS member JOIN public.users AS account ON account.user_id = member.user_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.role = 'admin' AND account.is_active AND NOT account.is_system;

-- name: ActivateMembership :exec
INSERT INTO public.workspace_members(workspace_id,user_id,role)
VALUES(sqlc.arg(workspace_id),sqlc.arg(user_id),CAST(sqlc.arg(role) AS user_role)) ON CONFLICT(workspace_id,user_id) DO NOTHING;

-- name: RemoveTeamMemberships :exec
DELETE FROM public.team_members AS member USING public.teams AS team
WHERE member.team_id = team.team_id AND team.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(user_id);

-- name: RemoveMembership :exec
DELETE FROM public.workspace_members WHERE workspace_id = sqlc.arg(workspace_id) AND user_id = sqlc.arg(user_id);

-- name: CreateUser :one
INSERT INTO public.workspace_scim_users(id,workspace_id,user_id,user_name,external_id,active,profile,previous_role)
VALUES(sqlc.arg(id),sqlc.arg(workspace_id),sqlc.arg(user_id),sqlc.arg(user_name),sqlc.narg(external_id),sqlc.arg(active),sqlc.arg(profile),sqlc.arg(previous_role))
RETURNING id,workspace_id,user_id,user_name,external_id,active,profile,previous_role,version,created_at,updated_at;

-- name: UpdateUser :one
UPDATE public.workspace_scim_users SET user_name=sqlc.arg(user_name),external_id=sqlc.narg(external_id),active=sqlc.arg(active),
 profile=sqlc.arg(profile),previous_role=sqlc.arg(previous_role),version=version+1,updated_at=CURRENT_TIMESTAMP,
 deleted_at=CASE WHEN CAST(sqlc.arg(deleted) AS boolean) THEN CURRENT_TIMESTAMP ELSE NULL END
WHERE workspace_id=sqlc.arg(workspace_id) AND id=sqlc.arg(id) AND version=sqlc.arg(expected_version) AND deleted_at IS NULL
RETURNING id,workspace_id,user_id,user_name,external_id,active,profile,previous_role,version,created_at,updated_at;

-- name: AppendAudit :exec
INSERT INTO public.workspace_scim_audit_events(event_id,workspace_id,actor_id,credential_id,resource_id,operation,metadata)
VALUES(sqlc.arg(id),sqlc.arg(workspace_id),sqlc.arg(actor_id),sqlc.narg(credential_id),sqlc.arg(resource_id),sqlc.arg(operation),sqlc.arg(metadata));

-- name: MarkSeatSync :exec
INSERT INTO public.workspace_scim_seat_reconciliation(workspace_id) VALUES(sqlc.arg(workspace_id))
ON CONFLICT(workspace_id) DO UPDATE SET generation=workspace_scim_seat_reconciliation.generation+1,last_error='',updated_at=CURRENT_TIMESTAMP;

-- name: SeatSync :one
SELECT generation,last_error FROM public.workspace_scim_seat_reconciliation WHERE workspace_id=sqlc.arg(workspace_id);

-- name: FinishSeatSync :exec
DELETE FROM public.workspace_scim_seat_reconciliation WHERE workspace_id=sqlc.arg(workspace_id) AND generation=sqlc.arg(generation);

-- name: FailSeatSync :exec
UPDATE public.workspace_scim_seat_reconciliation SET last_error='Seat sync failed. Retry from workspace security.',updated_at=CURRENT_TIMESTAMP
WHERE workspace_id=sqlc.arg(workspace_id) AND generation=sqlc.arg(generation);
