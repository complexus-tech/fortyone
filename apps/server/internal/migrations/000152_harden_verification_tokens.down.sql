-- Digest-only codes cannot be reconstructed during rollback. Invalidate them
-- in place so an older application can start safely; affected users request a
-- fresh code after rollback.
UPDATE public.verification_tokens
SET
    token = 'rollback-invalidated:' || id::text,
    used_at = COALESCE(used_at, now()),
    updated_at = now()
WHERE token IS NULL;

DROP INDEX IF EXISTS public.idx_verification_tokens_digest_consume;
DROP INDEX IF EXISTS public.verification_tokens_digest_unique;

ALTER TABLE public.verification_tokens
    DROP CONSTRAINT IF EXISTS verification_tokens_storage_shape_check,
    DROP COLUMN IF EXISTS token_digest,
    DROP COLUMN IF EXISTS token_key_id,
    DROP COLUMN IF EXISTS token_version,
    ALTER COLUMN token SET NOT NULL;
