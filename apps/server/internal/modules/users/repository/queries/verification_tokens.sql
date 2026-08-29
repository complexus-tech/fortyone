-- name: AcquireVerificationTokenIssueLock :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        CAST(sqlc.arg(email) AS text),
        hashtextextended(CAST(sqlc.arg(token_type) AS text), 0)
    )
);

-- name: CountRecentVerificationTokenIssues :one
SELECT COUNT(*)::integer
FROM public.verification_tokens
WHERE email = CAST(sqlc.arg(email) AS text)
  AND token_type = CAST(sqlc.arg(token_type) AS public.token_type)
  AND created_at >= sqlc.arg(rate_limit_since);

-- name: CreateVerificationToken :one
INSERT INTO public.verification_tokens (
    token,
    email,
    expires_at,
    token_type,
    token_digest,
    token_key_id,
    token_version,
    created_at,
    updated_at
)
VALUES (
    NULL,
    CAST(sqlc.arg(email) AS text),
    sqlc.arg(expires_at),
    CAST(sqlc.arg(token_type) AS public.token_type),
    sqlc.arg(token_digest),
    CAST(sqlc.arg(token_key_id) AS text),
    CAST(sqlc.arg(token_version) AS smallint),
    sqlc.arg(issued_at),
    sqlc.arg(issued_at)
)
RETURNING
    id,
    email,
    user_id,
    expires_at,
    used_at,
    CAST(token_type AS text) AS token_type,
    token_key_id,
    token_version,
    created_at,
    updated_at;

-- name: ConsumeVerificationToken :one
WITH digest_candidate AS (
    SELECT
        (CAST(sqlc.arg(token_digests) AS bytea[]))[candidate_index] AS token_digest,
        (CAST(sqlc.arg(token_key_ids) AS text[]))[candidate_index] AS token_key_id,
        (CAST(sqlc.arg(token_versions) AS smallint[]))[candidate_index] AS token_version
    FROM generate_subscripts(
        CAST(sqlc.arg(token_digests) AS bytea[]),
        1
    ) AS candidate_index
),
candidate AS (
    SELECT verification_token.id
    FROM public.verification_tokens AS verification_token
    WHERE verification_token.email = CAST(sqlc.arg(email) AS text)
      AND CAST(verification_token.token_type AS text) = ANY(CAST(sqlc.arg(token_types) AS text[]))
      AND verification_token.expires_at > CAST(sqlc.arg(consumed_at) AS timestamptz)
      AND verification_token.used_at IS NULL
      AND (
          (
              EXISTS (
                  SELECT 1
                  FROM digest_candidate
                  WHERE digest_candidate.token_digest = verification_token.token_digest
                    AND digest_candidate.token_key_id = verification_token.token_key_id
                    AND digest_candidate.token_version = verification_token.token_version
              )
          )
          OR (
              verification_token.token_digest IS NULL
              AND verification_token.token = CAST(sqlc.arg(legacy_token) AS text)
          )
      )
    ORDER BY verification_token.created_at DESC, verification_token.id DESC
    FOR UPDATE
    LIMIT 1
)
UPDATE public.verification_tokens AS verification_token
SET
    used_at = CAST(sqlc.arg(consumed_at) AS timestamptz),
    updated_at = CAST(sqlc.arg(consumed_at) AS timestamptz)
FROM candidate
WHERE verification_token.id = candidate.id
  AND verification_token.used_at IS NULL
RETURNING
    verification_token.id,
    verification_token.email,
    verification_token.user_id,
    verification_token.expires_at,
    verification_token.used_at,
    CAST(verification_token.token_type AS text) AS token_type,
    verification_token.token_key_id,
    verification_token.token_version,
    verification_token.created_at,
    verification_token.updated_at;

-- name: InvalidateVerificationTokens :execrows
UPDATE public.verification_tokens
SET
    used_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE email = CAST(sqlc.arg(email) AS text)
  AND used_at IS NULL;
