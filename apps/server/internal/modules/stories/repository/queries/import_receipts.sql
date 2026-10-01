-- name: RecordImportReceipt :exec
INSERT INTO public.story_import_receipts (
 workspace_id, team_id, creation_key, provider, source_digest, source_namespace,
 source_key, story_id, created, error_code, error_message, source_metadata
)
SELECT sqlc.arg(workspace_id), sqlc.arg(team_id), sqlc.arg(creation_key),
 sqlc.arg(provider), sqlc.arg(source_digest), sqlc.narg(source_namespace),
 sqlc.arg(source_key), sqlc.narg(story_id), sqlc.arg(created),
 sqlc.narg(error_code), sqlc.narg(error_message), CAST(sqlc.arg(source_metadata) AS jsonb)
WHERE EXISTS (SELECT 1 FROM public.teams AS team
 WHERE team.team_id = sqlc.arg(team_id) AND team.workspace_id = sqlc.arg(workspace_id))
ON CONFLICT (workspace_id, creation_key) DO UPDATE SET
 source_digest = EXCLUDED.source_digest,
 story_id = COALESCE(EXCLUDED.story_id, story_import_receipts.story_id),
 created = story_import_receipts.created OR EXCLUDED.created,
 error_code = EXCLUDED.error_code, error_message = EXCLUDED.error_message,
 source_metadata = EXCLUDED.source_metadata,
 attempts = story_import_receipts.attempts + 1, updated_at = clock_timestamp();

-- name: ListImportReceipts :many
SELECT receipt.* FROM public.story_import_receipts AS receipt
INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = receipt.workspace_id AND workspace.deleted_at IS NULL
INNER JOIN public.teams AS team ON team.team_id = receipt.team_id
WHERE receipt.workspace_id = sqlc.arg(workspace_id)
 AND receipt.provider = sqlc.arg(provider)
 AND ((CAST(sqlc.narg(source_namespace) AS text) IS NOT NULL AND receipt.source_namespace = sqlc.narg(source_namespace))
   OR (CAST(sqlc.narg(source_namespace) AS text) IS NULL AND receipt.source_digest = sqlc.arg(source_digest)))
 AND EXISTS (SELECT 1 FROM public.workspace_members AS member INNER JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
   WHERE member.workspace_id = receipt.workspace_id AND member.user_id = sqlc.arg(actor_id) AND member.role = 'admin')
 AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean) OR receipt.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
 AND (team.is_private = FALSE OR EXISTS (SELECT 1 FROM public.team_members WHERE team_id = receipt.team_id AND user_id = sqlc.arg(actor_id)))
ORDER BY receipt.creation_key
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);
