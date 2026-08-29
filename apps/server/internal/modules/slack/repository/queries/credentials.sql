-- name: ListSlackCredentialsForRewrap :many
SELECT
    id,
    workspace_id,
    slack_team_id,
    installation_generation,
    credential_payload,
    credential_key_version
FROM public.slack_workspaces
WHERE is_active = TRUE
  AND credential_key_version = CAST(sqlc.arg(credential_key_version) AS smallint)
  AND NULLIF(credential_payload, '') IS NOT NULL
  AND (
      CAST(sqlc.narg(after_id) AS uuid) IS NULL
      OR id > CAST(sqlc.narg(after_id) AS uuid)
  )
ORDER BY id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: RewrapSlackCredential :execrows
UPDATE public.slack_workspaces
SET credential_payload = CAST(sqlc.arg(replacement_credential) AS text),
    updated_at = NOW()
WHERE id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
  AND workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)
  AND credential_key_version = CAST(sqlc.arg(credential_key_version) AS smallint)
  AND credential_payload = CAST(sqlc.arg(previous_credential) AS text)
  AND is_active = TRUE;

-- name: ListSlackUninstallCredentialsForRewrap :many
SELECT
    id,
    workspace_id,
    slack_team_id,
    installation_generation,
    credential_payload,
    credential_key_version
FROM public.slack_uninstall_outbox
WHERE status <> 'completed'
  AND credential_key_version = CAST(sqlc.arg(credential_key_version) AS smallint)
  AND NULLIF(credential_payload, '') IS NOT NULL
  AND (
      CAST(sqlc.narg(after_id) AS uuid) IS NULL
      OR id > CAST(sqlc.narg(after_id) AS uuid)
  )
ORDER BY id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: RewrapSlackUninstallCredential :execrows
UPDATE public.slack_uninstall_outbox
SET credential_payload = CAST(sqlc.arg(replacement_credential) AS text),
    updated_at = NOW()
WHERE id = CAST(sqlc.arg(uninstall_id) AS uuid)
  AND workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)
  AND credential_key_version = CAST(sqlc.arg(credential_key_version) AS smallint)
  AND credential_payload = CAST(sqlc.arg(previous_credential) AS text)
  AND status <> 'completed';

-- name: UpgradeSlackCredential :execrows
UPDATE public.slack_workspaces
SET credential_payload = CAST(sqlc.arg(replacement_credential) AS text),
    credential_key_version = CAST(sqlc.arg(replacement_version) AS smallint),
    bot_access_token = '',
    updated_at = NOW()
WHERE id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
  AND is_active = TRUE
  AND credential_key_version = CAST(sqlc.arg(previous_version) AS smallint)
  AND COALESCE(NULLIF(credential_payload, ''), bot_access_token) = CAST(sqlc.arg(previous_credential) AS text);

-- name: ScrubVersionedLegacySlackCredentials :execrows
WITH candidates AS (
    SELECT id
    FROM public.slack_workspaces
    WHERE credential_key_version > 0
      AND NULLIF(credential_payload, '') IS NOT NULL
      AND NULLIF(bot_access_token, '') IS NOT NULL
    ORDER BY created_at, id
    LIMIT CAST(sqlc.arg(result_limit) AS integer)
)
UPDATE public.slack_workspaces AS installation
SET bot_access_token = '',
    updated_at = NOW()
FROM candidates
WHERE installation.id = candidates.id;

-- name: ListLegacySlackCredentials :many
SELECT
    id,
    workspace_id,
    slack_team_id,
    installation_generation,
    COALESCE(NULLIF(credential_payload, ''), bot_access_token) AS credential,
    credential_key_version
FROM public.slack_workspaces
WHERE is_active = TRUE
  AND credential_key_version < CAST(sqlc.arg(current_version) AS smallint)
  AND COALESCE(NULLIF(credential_payload, ''), bot_access_token) <> ''
ORDER BY created_at, id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: ListLegacySlackUninstallCredentials :many
SELECT
    id,
    workspace_id,
    slack_team_id,
    installation_generation,
    credential_payload,
    credential_key_version
FROM public.slack_uninstall_outbox
WHERE status <> 'completed'
  AND credential_key_version < CAST(sqlc.arg(current_version) AS smallint)
  AND NULLIF(credential_payload, '') IS NOT NULL
ORDER BY created_at, id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: UpgradeSlackUninstallCredential :execrows
UPDATE public.slack_uninstall_outbox
SET credential_payload = CAST(sqlc.arg(replacement_credential) AS text),
    credential_key_version = CAST(sqlc.arg(replacement_version) AS smallint),
    updated_at = NOW()
WHERE id = CAST(sqlc.arg(uninstall_id) AS uuid)
  AND status <> 'completed'
  AND credential_key_version = CAST(sqlc.arg(previous_version) AS smallint)
  AND credential_payload = CAST(sqlc.arg(previous_credential) AS text);

-- name: ListLegacySlackWebhookPayloads :many
SELECT
    event.id,
    event.provider,
    event.external_event_id,
    event.event_type,
    event.external_workspace_id,
    event.workspace_id,
    event.installation_id,
    event.installation_generation,
    COALESCE(event.trace_id, '') AS trace_id,
    event.received_at,
    event.status,
    event.attempt_count,
    event.recovery_generation,
    event.recovery_enqueued_at,
    event.processed_at,
    event.updated_at,
    event.payload_encrypted,
    event.payload_expires_at
FROM public.messaging_inbound_events AS event
WHERE event.provider = 'slack'
  AND event.workspace_id IS NOT NULL
  AND event.installation_id IS NOT NULL
  AND event.installation_generation IS NOT NULL
  AND NULLIF(event.payload_encrypted, '') IS NOT NULL
  AND event.payload_encrypted NOT LIKE CAST(sqlc.arg(current_prefix) AS text) || '%'
  AND event.id > CAST(sqlc.arg(after_id) AS uuid)
ORDER BY event.id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: UpgradeLegacySlackWebhookPayload :execrows
UPDATE public.messaging_inbound_events
SET payload_encrypted = CAST(sqlc.arg(replacement_payload) AS text),
    updated_at = NOW()
WHERE id = CAST(sqlc.arg(event_id) AS uuid)
  AND provider = 'slack'
  AND external_event_id = CAST(sqlc.arg(delivery_id) AS text)
  AND external_workspace_id = CAST(sqlc.arg(external_account_id) AS text)
  AND workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND installation_id = CAST(sqlc.arg(installation_id) AS uuid)
  AND installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)
  AND payload_encrypted = CAST(sqlc.arg(previous_payload) AS text);
