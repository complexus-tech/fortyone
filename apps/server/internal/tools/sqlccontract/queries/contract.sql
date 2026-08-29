-- name: GetTypeContract :one
SELECT
    id,
    nullable_id,
    occurred_at,
    nullable_occurred_at,
    local_at,
    nullable_local_at,
    due_date,
    nullable_due_date,
    status,
    nullable_status,
    amount,
    payload,
    related_ids
FROM sqlc_type_contracts
WHERE id = sqlc.arg(id);

-- name: ListTypeContracts :many
SELECT
    id,
    nullable_id,
    occurred_at,
    nullable_occurred_at,
    local_at,
    nullable_local_at,
    due_date,
    nullable_due_date,
    status,
    nullable_status,
    amount,
    payload,
    related_ids
FROM sqlc_type_contracts
WHERE id = ANY(CAST(sqlc.arg(ids) AS uuid[]))
ORDER BY id;
