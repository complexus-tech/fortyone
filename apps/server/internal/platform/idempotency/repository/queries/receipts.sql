-- name: LockReceiptScope :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        CAST(sqlc.arg(principal_kind) AS text)
        || chr(31)
        || CAST(CAST(sqlc.arg(principal_id) AS uuid) AS text)
        || chr(31)
        || COALESCE(CAST(CAST(sqlc.narg(workspace_id) AS uuid) AS text), '')
        || chr(31)
        || CAST(sqlc.arg(http_method) AS text)
        || chr(31)
        || CAST(sqlc.arg(route_operation) AS text)
        || chr(31)
        || encode(CAST(sqlc.arg(key_digest) AS bytea), 'hex'),
        0
    )
);

-- name: GetReceiptForUpdate :one
SELECT
    receipt.receipt_id,
    receipt.request_hash,
    receipt.state,
    receipt.lease_generation,
    receipt.lease_expires_at,
    receipt.response_status,
    receipt.response_body,
    receipt.response_content_type,
    receipt.expires_at
FROM public.api_idempotency_receipts AS receipt
WHERE receipt.principal_kind = sqlc.arg(principal_kind)
  AND receipt.principal_id = sqlc.arg(principal_id)
  AND receipt.workspace_id IS NOT DISTINCT FROM sqlc.narg(workspace_id)
  AND receipt.http_method = sqlc.arg(http_method)
  AND receipt.route_operation = sqlc.arg(route_operation)
  AND receipt.key_digest = sqlc.arg(key_digest)
LIMIT 1
FOR UPDATE;

-- name: CreateReceipt :one
INSERT INTO public.api_idempotency_receipts (
    receipt_id,
    principal_kind,
    principal_id,
    workspace_id,
    http_method,
    route_operation,
    key_digest,
    request_hash,
    state,
    lease_generation,
    lease_expires_at,
    expires_at,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(receipt_id),
    sqlc.arg(principal_kind),
    sqlc.arg(principal_id),
    sqlc.narg(workspace_id),
    sqlc.arg(http_method),
    sqlc.arg(route_operation),
    sqlc.arg(key_digest),
    sqlc.arg(request_hash),
    'in_progress',
    1,
    sqlc.arg(lease_expires_at),
    sqlc.arg(expires_at),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
)
RETURNING receipt_id, lease_generation, lease_expires_at;

-- name: RestartExpiredReceipt :one
UPDATE public.api_idempotency_receipts AS receipt
SET
    receipt_id = sqlc.arg(new_receipt_id),
    request_hash = sqlc.arg(request_hash),
    state = 'in_progress',
    lease_generation = 1,
    lease_expires_at = sqlc.arg(lease_expires_at),
    response_status = NULL,
    response_body = NULL,
    response_content_type = NULL,
    completed_at = NULL,
    expires_at = sqlc.arg(expires_at),
    created_at = sqlc.arg(restarted_at),
    updated_at = sqlc.arg(restarted_at)
WHERE receipt.receipt_id = sqlc.arg(receipt_id)
  AND receipt.expires_at <= sqlc.arg(restarted_at)
RETURNING receipt_id, lease_generation, lease_expires_at;

-- name: TakeOverStaleReceipt :one
UPDATE public.api_idempotency_receipts AS receipt
SET
    lease_generation = receipt.lease_generation + 1,
    lease_expires_at = sqlc.arg(lease_expires_at),
    expires_at = sqlc.arg(expires_at),
    updated_at = sqlc.arg(taken_over_at)
WHERE receipt.receipt_id = sqlc.arg(receipt_id)
  AND receipt.request_hash = sqlc.arg(request_hash)
  AND receipt.state = 'in_progress'
  AND receipt.lease_expires_at <= sqlc.arg(taken_over_at)
  AND receipt.expires_at > sqlc.arg(taken_over_at)
RETURNING receipt_id, lease_generation, lease_expires_at;

-- name: CompleteReceipt :execrows
UPDATE public.api_idempotency_receipts AS receipt
SET
    state = 'completed',
    lease_expires_at = NULL,
    response_status = sqlc.arg(response_status),
    response_body = sqlc.arg(response_body),
    response_content_type = sqlc.arg(response_content_type),
    completed_at = sqlc.arg(completed_at),
    expires_at = sqlc.arg(expires_at),
    updated_at = sqlc.arg(completed_at)
WHERE receipt.receipt_id = sqlc.arg(receipt_id)
  AND receipt.lease_generation = sqlc.arg(lease_generation)
  AND receipt.state = 'in_progress'
  AND receipt.lease_expires_at > sqlc.arg(completed_at)
  AND receipt.expires_at > sqlc.arg(completed_at);

-- name: DeleteExpiredReceipts :execrows
WITH expired AS (
    SELECT receipt.receipt_id
    FROM public.api_idempotency_receipts AS receipt
    WHERE receipt.expires_at <= sqlc.arg(expired_at)
    ORDER BY receipt.expires_at, receipt.receipt_id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(batch_size)
)
DELETE FROM public.api_idempotency_receipts AS receipt
USING expired
WHERE receipt.receipt_id = expired.receipt_id;
