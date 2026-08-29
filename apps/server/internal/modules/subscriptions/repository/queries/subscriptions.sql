-- name: UpdateWorkspaceSubscriptionSnapshot :execrows
UPDATE workspace_subscriptions
SET stripe_subscription_item_id = sqlc.arg(stripe_subscription_item_id),
    stripe_customer_id = sqlc.arg(stripe_customer_id),
    subscription_status = sqlc.arg(subscription_status),
    subscription_tier = sqlc.arg(subscription_tier),
    seat_count = sqlc.arg(seat_count),
    trial_end_date = sqlc.narg(trial_end_date),
    billing_interval = sqlc.narg(billing_interval),
    billing_ends_at = sqlc.narg(billing_ends_at),
    updated_at = NOW()
FROM (
    SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(stripe_customer_id), 0))
) AS customer_lock
WHERE workspace_subscriptions.workspace_id = sqlc.arg(workspace_id)
  AND workspace_subscriptions.stripe_subscription_id = sqlc.arg(stripe_subscription_id)
  AND NOT EXISTS (
      SELECT 1
      FROM workspace_subscriptions AS customer_binding
      WHERE customer_binding.stripe_customer_id = sqlc.arg(stripe_customer_id)
        AND customer_binding.workspace_id <> sqlc.arg(workspace_id)
  );

-- name: ApplyStripeSubscriptionSnapshot :one
WITH customer_lock AS MATERIALIZED (
    SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(stripe_customer_id), 0))
),
applied AS (
    UPDATE workspace_subscriptions
    SET stripe_subscription_item_id = sqlc.arg(stripe_subscription_item_id),
        stripe_customer_id = sqlc.arg(stripe_customer_id),
        subscription_status = sqlc.arg(subscription_status),
        subscription_tier = sqlc.arg(subscription_tier),
        seat_count = sqlc.arg(seat_count),
        trial_end_date = sqlc.narg(trial_end_date),
        billing_interval = sqlc.narg(billing_interval),
        billing_ends_at = sqlc.narg(billing_ends_at),
        last_stripe_event_created_at = CAST(sqlc.arg(event_created_at) AS timestamptz),
        last_stripe_event_priority = CAST(sqlc.arg(event_priority) AS smallint),
        last_stripe_event_id = CAST(sqlc.arg(event_id) AS varchar(255)),
        updated_at = NOW()
    FROM customer_lock
    WHERE workspace_subscriptions.stripe_subscription_id = sqlc.arg(stripe_subscription_id)
      AND NOT EXISTS (
          SELECT 1
          FROM workspace_subscriptions AS customer_binding
          WHERE customer_binding.stripe_customer_id = sqlc.arg(stripe_customer_id)
            AND customer_binding.workspace_id <> workspace_subscriptions.workspace_id
      )
      AND (
          last_stripe_event_created_at IS NULL
          OR last_stripe_event_created_at < CAST(sqlc.arg(event_created_at) AS timestamptz)
          OR (
              last_stripe_event_created_at = CAST(sqlc.arg(event_created_at) AS timestamptz)
              AND last_stripe_event_priority < CAST(sqlc.arg(event_priority) AS smallint)
          )
          OR (
              last_stripe_event_created_at = CAST(sqlc.arg(event_created_at) AS timestamptz)
              AND last_stripe_event_priority = CAST(sqlc.arg(event_priority) AS smallint)
              AND last_stripe_event_id <= CAST(sqlc.arg(event_id) AS varchar(255))
          )
      )
    RETURNING workspace_id
)
SELECT workspace_id, TRUE AS applied, FALSE AS identity_conflict
FROM applied
UNION ALL
SELECT subscription.workspace_id,
       FALSE AS applied,
       EXISTS (
           SELECT 1
           FROM workspace_subscriptions AS customer_binding
           WHERE customer_binding.stripe_customer_id = sqlc.arg(stripe_customer_id)
             AND customer_binding.workspace_id <> subscription.workspace_id
       ) AS identity_conflict
FROM workspace_subscriptions AS subscription
WHERE subscription.stripe_subscription_id = sqlc.arg(stripe_subscription_id)
  AND NOT EXISTS (SELECT 1 FROM applied)
LIMIT 1;

-- name: ApplyStripeSubscriptionDeletion :one
WITH applied AS (
    UPDATE workspace_subscriptions
    SET subscription_status = 'canceled',
        last_stripe_event_created_at = CAST(sqlc.arg(event_created_at) AS timestamptz),
        last_stripe_event_priority = CAST(sqlc.arg(event_priority) AS smallint),
        last_stripe_event_id = CAST(sqlc.arg(event_id) AS varchar(255)),
        updated_at = NOW()
    WHERE workspace_subscriptions.stripe_subscription_id = sqlc.arg(stripe_subscription_id)
      AND (
          last_stripe_event_created_at IS NULL
          OR last_stripe_event_created_at < CAST(sqlc.arg(event_created_at) AS timestamptz)
          OR (
              last_stripe_event_created_at = CAST(sqlc.arg(event_created_at) AS timestamptz)
              AND last_stripe_event_priority < CAST(sqlc.arg(event_priority) AS smallint)
          )
          OR (
              last_stripe_event_created_at = CAST(sqlc.arg(event_created_at) AS timestamptz)
              AND last_stripe_event_priority = CAST(sqlc.arg(event_priority) AS smallint)
              AND last_stripe_event_id <= CAST(sqlc.arg(event_id) AS varchar(255))
          )
      )
    RETURNING workspace_id
)
SELECT workspace_id, TRUE AS applied
FROM applied
UNION ALL
SELECT workspace_id, FALSE AS applied
FROM workspace_subscriptions
WHERE stripe_subscription_id = sqlc.arg(stripe_subscription_id)
  AND NOT EXISTS (SELECT 1 FROM applied)
LIMIT 1;

-- name: UpsertStripeSubscriptionSnapshot :one
WITH customer_lock AS MATERIALIZED (
    SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(stripe_customer_id), 0))
),
applied AS (
    INSERT INTO workspace_subscriptions (
        workspace_id,
        stripe_customer_id,
        stripe_subscription_id,
        stripe_subscription_item_id,
        subscription_status,
        subscription_tier,
        seat_count,
        trial_end_date,
        created_at,
        updated_at,
        billing_interval,
        billing_ends_at,
        last_stripe_event_created_at,
        last_stripe_event_priority,
        last_stripe_event_id
    )
    SELECT sqlc.arg(workspace_id),
           sqlc.arg(stripe_customer_id),
           sqlc.arg(stripe_subscription_id),
           sqlc.arg(stripe_subscription_item_id),
           sqlc.arg(subscription_status),
           sqlc.arg(subscription_tier),
           sqlc.arg(seat_count),
           sqlc.narg(trial_end_date),
           NOW(),
           NOW(),
           sqlc.narg(billing_interval),
           sqlc.narg(billing_ends_at),
           CAST(sqlc.arg(event_created_at) AS timestamptz),
           CAST(sqlc.arg(event_priority) AS smallint),
           CAST(sqlc.arg(event_id) AS varchar(255))
    FROM customer_lock
    WHERE NOT EXISTS (
        SELECT 1
        FROM workspace_subscriptions
        WHERE stripe_customer_id = sqlc.arg(stripe_customer_id)
          AND workspace_id <> sqlc.arg(workspace_id)
    )
    ON CONFLICT (stripe_subscription_id) DO UPDATE
    SET stripe_customer_id = EXCLUDED.stripe_customer_id,
        stripe_subscription_item_id = EXCLUDED.stripe_subscription_item_id,
        subscription_status = EXCLUDED.subscription_status,
        subscription_tier = EXCLUDED.subscription_tier,
        seat_count = EXCLUDED.seat_count,
        trial_end_date = EXCLUDED.trial_end_date,
        billing_interval = EXCLUDED.billing_interval,
        billing_ends_at = EXCLUDED.billing_ends_at,
        last_stripe_event_created_at = EXCLUDED.last_stripe_event_created_at,
        last_stripe_event_priority = EXCLUDED.last_stripe_event_priority,
        last_stripe_event_id = EXCLUDED.last_stripe_event_id,
        updated_at = NOW()
    WHERE workspace_subscriptions.workspace_id = EXCLUDED.workspace_id
      AND (
          workspace_subscriptions.last_stripe_event_created_at IS NULL
          OR workspace_subscriptions.last_stripe_event_created_at < EXCLUDED.last_stripe_event_created_at
          OR (
              workspace_subscriptions.last_stripe_event_created_at = EXCLUDED.last_stripe_event_created_at
              AND workspace_subscriptions.last_stripe_event_priority < EXCLUDED.last_stripe_event_priority
          )
          OR (
              workspace_subscriptions.last_stripe_event_created_at = EXCLUDED.last_stripe_event_created_at
              AND workspace_subscriptions.last_stripe_event_priority = EXCLUDED.last_stripe_event_priority
              AND workspace_subscriptions.last_stripe_event_id <= EXCLUDED.last_stripe_event_id
          )
      )
    RETURNING workspace_id
)
SELECT workspace_id, TRUE AS applied, FALSE AS identity_conflict
FROM applied
UNION ALL
SELECT subscription.workspace_id,
       FALSE AS applied,
       EXISTS (
           SELECT 1
           FROM workspace_subscriptions AS customer_binding
           WHERE customer_binding.stripe_customer_id = sqlc.arg(stripe_customer_id)
             AND customer_binding.workspace_id <> subscription.workspace_id
       ) AS identity_conflict
FROM workspace_subscriptions AS subscription
WHERE subscription.stripe_subscription_id = sqlc.arg(stripe_subscription_id)
  AND NOT EXISTS (SELECT 1 FROM applied)
LIMIT 1;
