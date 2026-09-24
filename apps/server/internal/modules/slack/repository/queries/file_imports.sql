-- name: RegisterSlackFileImport :one
INSERT INTO public.slack_file_imports (
    workspace_id, slack_workspace_id, installation_generation, slack_team_id,
    slack_user_id, slack_channel_id, slack_thread_ts, slack_message_ts,
    slack_file_id, idempotency_key, story_id, actor_id
)
SELECT
    installation.workspace_id, installation.id, installation.installation_generation,
    installation.slack_team_id, CAST(sqlc.arg(slack_user_id) AS text),
    CAST(sqlc.arg(slack_channel_id) AS text), CAST(sqlc.arg(slack_thread_ts) AS text),
    CAST(sqlc.arg(slack_message_ts) AS text), CAST(sqlc.arg(slack_file_id) AS text),
    CAST(sqlc.arg(idempotency_key) AS text), story.id, actor.user_id
FROM public.slack_workspaces AS installation
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = installation.workspace_id AND workspace.deleted_at IS NULL
JOIN public.slack_user_links AS link
  ON link.workspace_id = installation.workspace_id
 AND link.slack_workspace_id = installation.id
 AND link.slack_team_id = installation.slack_team_id
 AND link.slack_user_id = CAST(sqlc.arg(slack_user_id) AS text)
 AND link.user_id = CAST(sqlc.arg(actor_id) AS uuid)
JOIN public.workspace_members AS membership
  ON membership.workspace_id = installation.workspace_id
 AND membership.user_id = link.user_id
 AND membership.role IN ('admin', 'member')
JOIN public.users AS actor
  ON actor.user_id = membership.user_id AND actor.is_active = TRUE
JOIN public.stories AS story
  ON story.id = CAST(sqlc.arg(story_id) AS uuid)
 AND story.workspace_id = installation.workspace_id
 AND story.deleted_at IS NULL
JOIN public.team_members AS team_member
  ON team_member.team_id = story.team_id AND team_member.user_id = actor.user_id
WHERE installation.id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
  AND installation.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND installation.installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)
  AND installation.slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND installation.is_active = TRUE
  AND POSITION(',files:read,' IN ',' || REPLACE(COALESCE(installation.scope, ''), ' ', '') || ',') > 0
ON CONFLICT (workspace_id, story_id, slack_file_id)
DO UPDATE SET
    slack_workspace_id = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.slack_workspace_id
        ELSE EXCLUDED.slack_workspace_id
    END,
    installation_generation = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.installation_generation
        ELSE EXCLUDED.installation_generation
    END,
    slack_team_id = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.slack_team_id
        ELSE EXCLUDED.slack_team_id
    END,
    slack_user_id = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.slack_user_id
        ELSE EXCLUDED.slack_user_id
    END,
    slack_channel_id = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.slack_channel_id
        ELSE EXCLUDED.slack_channel_id
    END,
    slack_thread_ts = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.slack_thread_ts
        ELSE EXCLUDED.slack_thread_ts
    END,
    slack_message_ts = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.slack_message_ts
        ELSE EXCLUDED.slack_message_ts
    END,
    actor_id = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.actor_id
        ELSE EXCLUDED.actor_id
    END,
    status = CASE
        WHEN public.slack_file_imports.status = 'complete' THEN 'complete'
        WHEN public.slack_file_imports.status = 'processing'
         AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation THEN 'processing'
        ELSE 'pending'
    END,
    attempt_count = CASE
        WHEN public.slack_file_imports.status = 'complete'
          OR (public.slack_file_imports.status = 'processing'
              AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation)
          THEN public.slack_file_imports.attempt_count
        ELSE 0
    END,
    lease_until = CASE
        WHEN public.slack_file_imports.status = 'processing'
         AND public.slack_file_imports.installation_generation = EXCLUDED.installation_generation
          THEN public.slack_file_imports.lease_until
        ELSE NULL
    END,
    last_error = CASE
        WHEN public.slack_file_imports.status = 'complete' THEN public.slack_file_imports.last_error
        ELSE NULL
    END,
    updated_at = now()
RETURNING id;

-- name: ClaimSlackFileImport :one
UPDATE public.slack_file_imports
SET status = 'processing',
    attempt_count = attempt_count + 1,
    lease_until = now() + INTERVAL '90 minutes',
    last_error = NULL,
    updated_at = now()
WHERE id = CAST(sqlc.arg(import_id) AS uuid)
  AND attempt_count < 8
  AND (
    status IN ('pending', 'failed')
    OR (status = 'processing' AND lease_until < now())
  )
RETURNING
    id, workspace_id, slack_workspace_id, installation_generation, slack_team_id,
    slack_user_id, slack_channel_id, slack_thread_ts, slack_message_ts,
    slack_file_id, story_id, actor_id, attempt_count;

-- name: AuthorizeSlackFileImport :one
SELECT EXISTS (
    SELECT 1
    FROM public.slack_file_imports AS file_import
    JOIN public.slack_workspaces AS installation
      ON installation.id = file_import.slack_workspace_id
     AND installation.workspace_id = file_import.workspace_id
     AND installation.installation_generation = file_import.installation_generation
     AND installation.slack_team_id = file_import.slack_team_id
     AND installation.is_active = TRUE
    JOIN public.workspaces AS workspace
      ON workspace.workspace_id = file_import.workspace_id AND workspace.deleted_at IS NULL
    JOIN public.slack_user_links AS link
      ON link.workspace_id = file_import.workspace_id
     AND link.slack_workspace_id = installation.id
     AND link.slack_team_id = file_import.slack_team_id
     AND link.slack_user_id = file_import.slack_user_id
     AND link.user_id = file_import.actor_id
    JOIN public.workspace_members AS membership
      ON membership.workspace_id = file_import.workspace_id
     AND membership.user_id = file_import.actor_id
     AND membership.role IN ('admin', 'member')
    JOIN public.users AS actor
      ON actor.user_id = file_import.actor_id AND actor.is_active = TRUE
    JOIN public.stories AS story
      ON story.id = file_import.story_id
     AND story.workspace_id = file_import.workspace_id
     AND story.deleted_at IS NULL
    JOIN public.team_members AS team_member
      ON team_member.team_id = story.team_id AND team_member.user_id = file_import.actor_id
    WHERE file_import.id = CAST(sqlc.arg(import_id) AS uuid)
      AND file_import.status = 'processing'
      AND POSITION(',files:read,' IN ',' || REPLACE(COALESCE(installation.scope, ''), ' ', '') || ',') > 0
);

-- name: CompleteSlackFileImport :execrows
UPDATE public.slack_file_imports
SET status = 'complete', attachment_id = CAST(sqlc.arg(attachment_id) AS uuid),
    lease_until = NULL, last_error = NULL, updated_at = now()
WHERE id = CAST(sqlc.arg(import_id) AS uuid)
  AND status = 'processing'
  AND attempt_count = CAST(sqlc.arg(attempt_count) AS integer);

-- name: FailSlackFileImport :execrows
UPDATE public.slack_file_imports
SET status = 'failed', lease_until = NULL,
    last_error = CASE
        WHEN attempt_count >= 8 THEN 'Slack file import failed after retries'
        ELSE 'Slack file import failed; retry scheduled'
    END,
    updated_at = now()
WHERE id = CAST(sqlc.arg(import_id) AS uuid)
  AND status = 'processing'
  AND attempt_count = CAST(sqlc.arg(attempt_count) AS integer);

-- name: CancelSlackFileImport :execrows
UPDATE public.slack_file_imports
SET status = 'cancelled', lease_until = NULL,
    last_error = 'Slack access or story permission is no longer available', updated_at = now()
WHERE id = CAST(sqlc.arg(import_id) AS uuid)
  AND status = 'processing'
  AND attempt_count = CAST(sqlc.arg(attempt_count) AS integer);

-- name: ListRecoverableSlackFileImports :many
SELECT id
FROM public.slack_file_imports
WHERE attempt_count < 8
  AND (
    status IN ('pending', 'failed')
    OR (status = 'processing' AND lease_until < now())
  )
ORDER BY updated_at, id
LIMIT CAST(sqlc.arg(result_limit) AS integer);
