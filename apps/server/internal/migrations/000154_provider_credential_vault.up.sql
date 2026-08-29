ALTER TABLE public.users
    ADD COLUMN github_access_token_envelope_version smallint NOT NULL DEFAULT 0,
    ADD COLUMN github_access_token_generation uuid;

ALTER TABLE public.users
    ADD CONSTRAINT users_github_access_token_envelope_version_check
        CHECK (github_access_token_envelope_version >= 0),
    ADD CONSTRAINT users_github_access_token_metadata_check CHECK (
        (github_access_token IS NULL
            AND github_access_token_envelope_version = 0
            AND github_access_token_generation IS NULL)
        OR
        (github_access_token IS NOT NULL
            AND (
                (github_access_token_envelope_version = 0
                    AND github_access_token_generation IS NULL)
                OR
                (github_access_token_envelope_version >= 2
                    AND github_access_token_generation IS NOT NULL)
            ))
    );

COMMENT ON COLUMN public.users.github_access_token IS
    'Legacy plaintext only when envelope version is zero; vault envelope after the credential backfill.';
COMMENT ON COLUMN public.users.github_access_token_envelope_version IS
    'Zero identifies a pre-vault credential. Positive values identify the credential vault envelope format.';
COMMENT ON COLUMN public.users.github_access_token_generation IS
    'Generation bound into the vault associated data for the current GitHub user credential.';
