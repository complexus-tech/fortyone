-- Removing the generation from a shared-vault credential would make its AAD
-- unrecoverable. Rollback is allowed only before any connection is upgraded or
-- created by the vault-backed runtime.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.figma_connections
        WHERE credential_key_version <> 1
    ) THEN
        RAISE EXCEPTION '000169 cannot be reversed after Figma credentials use the shared vault; deploy a forward fix';
    END IF;
END
$$;

DROP INDEX IF EXISTS public.figma_connections_active_generation;

ALTER TABLE public.figma_connections
    DROP CONSTRAINT IF EXISTS figma_connections_credential_key_version_check,
    DROP COLUMN IF EXISTS installation_generation,
    DROP COLUMN IF EXISTS credential_key_version;
