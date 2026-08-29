-- name: GetSubscriptionByWorkspaceID :one
SELECT workspace_id,
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
       billing_ends_at
FROM workspace_subscriptions
WHERE workspace_id = sqlc.arg(workspace_id)
ORDER BY created_at DESC NULLS LAST, stripe_subscription_id DESC
LIMIT 1;

-- name: ListWorkspaceInvoices :many
SELECT invoice_id,
       workspace_id,
       stripe_invoice_id,
       CAST(amount_paid AS double precision) AS amount_paid,
       invoice_date,
       status,
       seats_count,
       hosted_url,
       customer_name,
       created_at
FROM subscription_invoices
WHERE workspace_id = sqlc.arg(workspace_id)
ORDER BY invoice_date DESC, invoice_id DESC
LIMIT sqlc.arg(result_limit);

-- name: CountBillableWorkspaceUsers :one
SELECT count(*)
FROM workspace_members AS membership
JOIN users AS workspace_user ON workspace_user.user_id = membership.user_id
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('admin', 'member')
  AND workspace_user.is_active = TRUE;

-- name: HasActiveWorkspaceSubscription :one
SELECT EXISTS (
    SELECT 1
    FROM workspace_subscriptions
    WHERE workspace_id = sqlc.arg(workspace_id)
      AND subscription_status IN ('active', 'trialing', 'past_due')
);

-- name: GetWorkspaceCreatorEmail :one
SELECT workspace_user.email
FROM workspaces AS workspace
JOIN users AS workspace_user ON workspace_user.user_id = workspace.created_by
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace_user.is_active = TRUE;

-- name: GetSubscriptionWorkspaceByProviderID :one
SELECT workspace_id
FROM workspace_subscriptions
WHERE stripe_subscription_id = sqlc.arg(stripe_subscription_id);

-- name: HasCustomerBindingOutsideWorkspace :one
SELECT EXISTS (
    SELECT 1
    FROM workspace_subscriptions
    WHERE stripe_customer_id = sqlc.arg(stripe_customer_id)
      AND workspace_id <> sqlc.arg(workspace_id)
);
