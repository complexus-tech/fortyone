-- name: UpsertWorkspaceInvoice :one
WITH saved_invoice AS (
    INSERT INTO subscription_invoices (
        workspace_id,
        stripe_invoice_id,
        amount_paid,
        invoice_date,
        status,
        seats_count,
        hosted_url,
        customer_name,
        created_at
    )
    SELECT sqlc.arg(workspace_id),
           sqlc.arg(stripe_invoice_id),
           CAST(sqlc.arg(amount_paid) AS double precision),
           sqlc.arg(invoice_date),
           sqlc.arg(status),
           sqlc.arg(seats_count),
           sqlc.narg(hosted_url),
           sqlc.narg(customer_name),
           NOW()
    WHERE EXISTS (
        SELECT 1
        FROM workspace_subscriptions
        WHERE workspace_id = sqlc.arg(workspace_id)
          AND stripe_customer_id = sqlc.arg(stripe_customer_id)
    )
    ON CONFLICT (stripe_invoice_id) DO UPDATE
    SET amount_paid = EXCLUDED.amount_paid,
        invoice_date = EXCLUDED.invoice_date,
        status = EXCLUDED.status,
        seats_count = EXCLUDED.seats_count,
        hosted_url = EXCLUDED.hosted_url,
        customer_name = EXCLUDED.customer_name
    WHERE subscription_invoices.workspace_id = EXCLUDED.workspace_id
    RETURNING invoice_id, workspace_id, stripe_invoice_id, invoice_date, customer_name, hosted_url
), payment_alert AS (
    INSERT INTO public.internal_slack_alerts (dedupe_key, kind, workspace_id, payload)
    SELECT 'payment_received:' || saved_invoice.stripe_invoice_id, 'payment_received', saved_invoice.workspace_id,
           CAST(sqlc.narg(payment_payload) AS jsonb) || jsonb_build_object(
               'name', saved_invoice.customer_name, 'occurred_at', saved_invoice.invoice_date,
               'invoice_id', saved_invoice.stripe_invoice_id,
               'workspace_name', workspace.name, 'workspace_slug', workspace.slug)
    FROM saved_invoice
    JOIN public.workspaces AS workspace ON workspace.workspace_id = saved_invoice.workspace_id
    WHERE CAST(sqlc.narg(payment_payload) AS jsonb) IS NOT NULL
    ON CONFLICT (dedupe_key) DO NOTHING
)
SELECT invoice_id FROM saved_invoice;

-- name: GetInvoiceWorkspaceByProviderID :one
SELECT workspace_id
FROM subscription_invoices
WHERE stripe_invoice_id = sqlc.arg(stripe_invoice_id);
