-- This expand migration is forward-only after digest-only invitations exist.
-- The guard keeps an operator from silently deleting the only lookup material.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.workspace_invitations
        WHERE token IS NULL
          AND token_digest IS NOT NULL
    ) THEN
        RAISE EXCEPTION '000155 cannot be reversed after digest-only invitations have been issued; deploy a forward fix';
    END IF;
END
$$;

DROP TABLE IF EXISTS public.workspace_invitation_outbox;

DROP INDEX IF EXISTS public.workspace_invitations_token_digest_key;

ALTER TABLE public.workspace_invitations
    DROP CONSTRAINT IF EXISTS workspace_invitations_token_storage_shape_check,
    DROP COLUMN IF EXISTS token_version,
    DROP COLUMN IF EXISTS token_key_id,
    DROP COLUMN IF EXISTS token_nonce,
    DROP COLUMN IF EXISTS token_digest,
    ALTER COLUMN token SET NOT NULL;
