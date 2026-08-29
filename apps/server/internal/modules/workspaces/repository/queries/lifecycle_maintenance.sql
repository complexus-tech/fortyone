-- name: ListInactiveWorkspaceWarningCandidates :many
SELECT
    workspace.workspace_id,
    CAST(workspace.name AS text) AS name,
    CAST(workspace.slug AS text) AS slug,
    workspace.last_accessed_at,
    CAST(ARRAY(
        SELECT CAST(account.email AS text)
        FROM public.workspace_members AS membership
        INNER JOIN public.users AS account
            ON account.user_id = membership.user_id
        WHERE membership.workspace_id = workspace.workspace_id
          AND membership.role = 'admin'
          AND account.is_active = TRUE
        ORDER BY account.user_id
    ) AS text[]) AS admin_emails
FROM public.workspaces AS workspace
WHERE workspace.last_accessed_at < (
        CAST(sqlc.arg(inactive_before) AS timestamptz) AT TIME ZONE 'UTC'
    )
  AND workspace.deleted_at IS NULL
  AND workspace.inactivity_warning_sent_at IS NULL
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR workspace.last_accessed_at > (
          CAST(sqlc.arg(after_last_accessed_at) AS timestamptz) AT TIME ZONE 'UTC'
      )
      OR (
          workspace.last_accessed_at = (
              CAST(sqlc.arg(after_last_accessed_at) AS timestamptz) AT TIME ZONE 'UTC'
          )
          AND workspace.workspace_id > sqlc.arg(after_workspace_id)
      )
  )
ORDER BY workspace.last_accessed_at, workspace.workspace_id
LIMIT CAST(sqlc.arg(batch_size) AS integer);

-- name: MarkWorkspaceInactivityWarningSent :execrows
UPDATE public.workspaces AS workspace
SET inactivity_warning_sent_at = (
        CAST(sqlc.arg(warning_sent_at) AS timestamptz) AT TIME ZONE 'UTC'
    )
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace.last_accessed_at < (
      CAST(sqlc.arg(inactive_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND workspace.deleted_at IS NULL
  AND workspace.inactivity_warning_sent_at IS NULL;

-- name: LockWorkspaceIntegrationLifecycle :exec
SELECT pg_advisory_xact_lock(sqlc.arg(lock_key));

-- name: ListInactiveWorkspaceDeletionCandidates :many
SELECT
    workspace.workspace_id,
    workspace.last_accessed_at
FROM public.workspaces AS workspace
WHERE workspace.last_accessed_at < (
        CAST(sqlc.arg(inactive_before) AS timestamptz) AT TIME ZONE 'UTC'
    )
  AND workspace.deleted_at IS NULL
  AND workspace.inactivity_warning_sent_at IS NOT NULL
  AND workspace.inactivity_warning_sent_at <= (
      CAST(sqlc.arg(warning_sent_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR workspace.last_accessed_at > (
          CAST(sqlc.arg(after_last_accessed_at) AS timestamptz) AT TIME ZONE 'UTC'
      )
      OR (
          workspace.last_accessed_at = (
              CAST(sqlc.arg(after_last_accessed_at) AS timestamptz) AT TIME ZONE 'UTC'
          )
          AND workspace.workspace_id > sqlc.arg(after_workspace_id)
      )
  )
ORDER BY workspace.last_accessed_at, workspace.workspace_id
LIMIT CAST(sqlc.arg(batch_size) AS integer)
FOR UPDATE OF workspace SKIP LOCKED;

-- name: ListDeletedWorkspacePurgeCandidates :many
SELECT
    workspace.workspace_id,
    workspace.deleted_at
FROM public.workspaces AS workspace
WHERE workspace.deleted_at IS NOT NULL
  AND workspace.deleted_at < (
      CAST(sqlc.arg(deleted_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR workspace.deleted_at > (
          CAST(sqlc.arg(after_deleted_at) AS timestamptz) AT TIME ZONE 'UTC'
      )
      OR (
          workspace.deleted_at = (
              CAST(sqlc.arg(after_deleted_at) AS timestamptz) AT TIME ZONE 'UTC'
          )
          AND workspace.workspace_id > sqlc.arg(after_workspace_id)
      )
  )
ORDER BY workspace.deleted_at, workspace.workspace_id
LIMIT CAST(sqlc.arg(batch_size) AS integer)
FOR UPDATE OF workspace SKIP LOCKED;

-- name: CountWorkspaceDeletionCandidatesAwaitingSlackEncryption :one
SELECT COUNT(DISTINCT installation.workspace_id)
FROM public.slack_workspaces AS installation
WHERE installation.workspace_id = ANY(CAST(sqlc.arg(workspace_ids) AS uuid[]))
  AND installation.is_active = TRUE
  AND (
      installation.credential_key_version <> sqlc.arg(credential_key_version)
      OR NULLIF(installation.credential_payload, '') IS NULL
      OR installation.credential_payload NOT LIKE CAST(sqlc.arg(credential_envelope_pattern) AS text)
      OR NULLIF(installation.bot_access_token, '') IS NOT NULL
  );

-- name: SnapshotWorkspaceSlackUninstalls :execrows
INSERT INTO public.slack_uninstall_outbox (
    slack_workspace_id,
    workspace_id,
    installation_generation,
    slack_team_id,
    uninstall_kind,
    credential_payload,
    credential_key_version,
    status,
    next_attempt_at,
    created_at,
    updated_at
)
SELECT
    installation.id,
    installation.workspace_id,
    installation.installation_generation,
    installation.slack_team_id,
    'workspace_delete',
    installation.credential_payload,
    installation.credential_key_version,
    'pending',
    sqlc.arg(processed_at),
    sqlc.arg(processed_at),
    sqlc.arg(processed_at)
FROM public.slack_workspaces AS installation
WHERE installation.workspace_id = ANY(CAST(sqlc.arg(workspace_ids) AS uuid[]))
  AND installation.is_active = TRUE
  AND installation.credential_key_version = sqlc.arg(credential_key_version)
  AND installation.credential_payload LIKE CAST(sqlc.arg(credential_envelope_pattern) AS text)
  AND NULLIF(installation.bot_access_token, '') IS NULL
ON CONFLICT (slack_workspace_id, installation_generation, uninstall_kind) DO UPDATE
SET workspace_id = EXCLUDED.workspace_id,
    slack_team_id = EXCLUDED.slack_team_id,
    credential_payload = EXCLUDED.credential_payload,
    credential_key_version = EXCLUDED.credential_key_version,
    status = 'pending',
    attempt_count = 0,
    last_error = NULL,
    next_attempt_at = EXCLUDED.next_attempt_at,
    processing_started_at = NULL,
    completed_at = NULL,
    updated_at = EXCLUDED.updated_at
WHERE slack_uninstall_outbox.status = 'completed';

-- name: CancelWorkspaceSlackInboundEvents :execrows
UPDATE public.messaging_inbound_events AS event
SET status = 'cancelled',
    payload_encrypted = NULL,
    last_error = 'FortyOne workspace deleted',
    recovery_enqueued_at = NULL,
    processed_at = sqlc.arg(processed_at),
    updated_at = sqlc.arg(processed_at)
WHERE event.provider = 'slack'
  AND event.status IN ('pending', 'processing', 'failed')
  AND EXISTS (
      SELECT 1
      FROM public.slack_workspaces AS installation
      WHERE installation.workspace_id = ANY(CAST(sqlc.arg(workspace_ids) AS uuid[]))
        AND installation.slack_team_id = event.external_workspace_id
        AND installation.is_active = TRUE
        AND installation.credential_key_version = sqlc.arg(credential_key_version)
        AND installation.credential_payload LIKE CAST(sqlc.arg(credential_envelope_pattern) AS text)
        AND NULLIF(installation.bot_access_token, '') IS NULL
  );

-- name: CancelWorkspaceSlackOutboundDeliveries :execrows
UPDATE public.messaging_outbound_deliveries AS delivery
SET status = 'cancelled',
    content = NULL,
    last_error = 'FortyOne workspace deleted',
    updated_at = sqlc.arg(processed_at)
WHERE delivery.provider = 'slack'
  AND delivery.status IN ('pending', 'delivering', 'failed')
  AND EXISTS (
      SELECT 1
      FROM public.slack_workspaces AS installation
      WHERE installation.workspace_id = ANY(CAST(sqlc.arg(workspace_ids) AS uuid[]))
        AND installation.slack_team_id = delivery.external_workspace_id
        AND installation.is_active = TRUE
        AND installation.credential_key_version = sqlc.arg(credential_key_version)
        AND installation.credential_payload LIKE CAST(sqlc.arg(credential_envelope_pattern) AS text)
        AND NULLIF(installation.bot_access_token, '') IS NULL
  );

-- name: DeleteInactiveWorkspaceCandidates :execrows
DELETE FROM public.workspaces AS workspace
WHERE workspace.workspace_id = ANY(CAST(sqlc.arg(workspace_ids) AS uuid[]))
  AND workspace.last_accessed_at < (
      CAST(sqlc.arg(inactive_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND workspace.deleted_at IS NULL
  AND workspace.inactivity_warning_sent_at IS NOT NULL
  AND workspace.inactivity_warning_sent_at <= (
      CAST(sqlc.arg(warning_sent_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.slack_workspaces AS installation
      WHERE installation.workspace_id = workspace.workspace_id
        AND installation.is_active = TRUE
        AND (
            installation.credential_key_version <> sqlc.arg(credential_key_version)
            OR NULLIF(installation.credential_payload, '') IS NULL
            OR installation.credential_payload NOT LIKE CAST(sqlc.arg(credential_envelope_pattern) AS text)
            OR NULLIF(installation.bot_access_token, '') IS NOT NULL
        )
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.slack_workspaces AS installation
      WHERE installation.workspace_id = workspace.workspace_id
        AND installation.is_active = TRUE
        AND NOT EXISTS (
            SELECT 1
            FROM public.slack_uninstall_outbox AS uninstall
            WHERE uninstall.slack_workspace_id = installation.id
              AND uninstall.installation_generation = installation.installation_generation
              AND uninstall.uninstall_kind = 'workspace_delete'
              AND uninstall.status <> 'completed'
              AND NULLIF(uninstall.credential_payload, '') IS NOT NULL
        )
  );

-- name: DeleteSoftDeletedWorkspaceCandidates :execrows
DELETE FROM public.workspaces AS workspace
WHERE workspace.workspace_id = ANY(CAST(sqlc.arg(workspace_ids) AS uuid[]))
  AND workspace.deleted_at IS NOT NULL
  AND workspace.deleted_at < (
      CAST(sqlc.arg(deleted_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.slack_workspaces AS installation
      WHERE installation.workspace_id = workspace.workspace_id
        AND installation.is_active = TRUE
        AND (
            installation.credential_key_version <> sqlc.arg(credential_key_version)
            OR NULLIF(installation.credential_payload, '') IS NULL
            OR installation.credential_payload NOT LIKE CAST(sqlc.arg(credential_envelope_pattern) AS text)
            OR NULLIF(installation.bot_access_token, '') IS NOT NULL
        )
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.slack_workspaces AS installation
      WHERE installation.workspace_id = workspace.workspace_id
        AND installation.is_active = TRUE
        AND NOT EXISTS (
            SELECT 1
            FROM public.slack_uninstall_outbox AS uninstall
            WHERE uninstall.slack_workspace_id = installation.id
              AND uninstall.installation_generation = installation.installation_generation
              AND uninstall.uninstall_kind = 'workspace_delete'
              AND uninstall.status <> 'completed'
              AND NULLIF(uninstall.credential_payload, '') IS NOT NULL
        )
  );
