DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.users
        WHERE github_access_token_envelope_version > 0
           OR github_access_token_generation IS NOT NULL
    ) THEN
        RAISE EXCEPTION
            'cannot roll back provider credential vault metadata while encrypted GitHub credentials exist';
    END IF;
END
$$;

ALTER TABLE public.users
    DROP CONSTRAINT IF EXISTS users_github_access_token_metadata_check,
    DROP CONSTRAINT IF EXISTS users_github_access_token_envelope_version_check,
    DROP COLUMN IF EXISTS github_access_token_generation,
    DROP COLUMN IF EXISTS github_access_token_envelope_version;

COMMENT ON COLUMN public.users.github_access_token IS NULL;
