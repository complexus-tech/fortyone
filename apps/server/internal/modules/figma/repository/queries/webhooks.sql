-- name: SaveWebhook :execrows
INSERT INTO public.figma_webhooks (
    connection_id, file_key, event_type, figma_webhook_id, passcode_hash
)
SELECT
    connection.id,
    sqlc.arg(file_key),
    sqlc.arg(event_type),
    sqlc.arg(figma_webhook_id),
    sqlc.arg(passcode_hash)
FROM public.figma_connections AS connection
WHERE connection.id = sqlc.arg(connection_id)
  AND connection.is_active
ON CONFLICT (connection_id, file_key, event_type) DO UPDATE SET
    figma_webhook_id = EXCLUDED.figma_webhook_id,
    passcode_hash = EXCLUDED.passcode_hash,
    is_active = TRUE,
    updated_at = sqlc.arg(updated_at);

-- name: GetAuthorizedWebhook :one
SELECT
    webhook.id,
    webhook.connection_id,
    connection.workspace_id,
    connection.installation_generation,
    webhook.file_key,
    webhook.event_type,
    webhook.figma_webhook_id,
    webhook.passcode_hash,
    webhook.is_active
FROM public.figma_webhooks AS webhook
INNER JOIN public.figma_connections AS connection
    ON connection.id = webhook.connection_id
   AND connection.is_active
WHERE webhook.figma_webhook_id = sqlc.arg(figma_webhook_id)
  AND webhook.is_active;

-- name: GetCurrentWebhook :one
SELECT
    webhook.id,
    webhook.connection_id,
    connection.workspace_id,
    connection.installation_generation,
    webhook.file_key,
    webhook.event_type,
    webhook.figma_webhook_id,
    webhook.passcode_hash,
    webhook.is_active
FROM public.figma_webhooks AS webhook
INNER JOIN public.figma_connections AS connection
    ON connection.id = webhook.connection_id
   AND connection.is_active
WHERE webhook.connection_id = sqlc.arg(connection_id)
  AND connection.installation_generation = sqlc.arg(installation_generation)
  AND webhook.figma_webhook_id = sqlc.arg(figma_webhook_id)
  AND webhook.is_active;

-- name: FindWebhook :one
SELECT id, connection_id, file_key, event_type, figma_webhook_id, passcode_hash, is_active
FROM public.figma_webhooks
WHERE connection_id = sqlc.arg(connection_id)
  AND file_key = sqlc.arg(file_key)
  AND event_type = sqlc.arg(event_type)
  AND is_active;

-- name: ListWebhooks :many
SELECT id, connection_id, file_key, event_type, figma_webhook_id, passcode_hash, is_active
FROM public.figma_webhooks
WHERE connection_id = sqlc.arg(connection_id)
  AND is_active
ORDER BY created_at, id;

-- name: DeactivateWebhook :execrows
UPDATE public.figma_webhooks
SET is_active = FALSE,
    updated_at = sqlc.arg(updated_at)
WHERE figma_webhook_id = sqlc.arg(figma_webhook_id)
  AND is_active;
