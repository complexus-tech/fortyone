-- name: ClaimInternalSlackAlert :one
WITH candidate AS (
    SELECT id FROM public.internal_slack_alerts
    WHERE delivered_at IS NULL
      AND available_at <= now()
      AND (lease_until IS NULL OR lease_until <= now())
      AND (slack_team_id IS NULL OR (slack_team_id = CAST(sqlc.arg(team_id) AS text)
          AND slack_channel_id = CAST(sqlc.arg(channel_id) AS text)))
    ORDER BY available_at, created_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
UPDATE public.internal_slack_alerts AS alert
SET slack_team_id = CAST(sqlc.arg(team_id) AS text),
    slack_channel_id = CAST(sqlc.arg(channel_id) AS text),
    lease_token = sqlc.arg(lease_token),
    lease_until = now() + interval '1 minute',
    attempts = attempts + 1
FROM candidate WHERE alert.id = candidate.id
RETURNING alert.id, alert.dedupe_key, alert.kind, alert.payload, alert.attempts;

-- name: CompleteInternalSlackAlert :execrows
UPDATE public.internal_slack_alerts
SET delivered_at = now(), slack_message_ts = CAST(sqlc.arg(message_ts) AS text),
    payload = CAST('{}' AS jsonb), lease_token = NULL, lease_until = NULL
WHERE id = sqlc.arg(id) AND lease_token = sqlc.arg(lease_token)
  AND lease_until > now() AND delivered_at IS NULL;

-- name: RetryInternalSlackAlert :execrows
UPDATE public.internal_slack_alerts
SET available_at = sqlc.arg(available_at), lease_token = NULL, lease_until = NULL
WHERE id = sqlc.arg(id) AND lease_token = sqlc.arg(lease_token)
  AND lease_until > now() AND delivered_at IS NULL;
