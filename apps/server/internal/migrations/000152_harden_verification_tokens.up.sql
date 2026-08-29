-- Expand verification token storage for digest-only writes while preserving a
-- short compatibility window for codes issued by the previous application
-- version. The application dual-reads legacy plaintext rows; all new rows use
-- the versioned HMAC fields and leave token NULL.
ALTER TABLE public.verification_tokens
    ALTER COLUMN token DROP NOT NULL,
    ADD COLUMN token_digest bytea,
    ADD COLUMN token_key_id varchar(64),
    ADD COLUMN token_version smallint;

ALTER TABLE public.verification_tokens
    ADD CONSTRAINT verification_tokens_storage_shape_check CHECK (
        (
            token IS NOT NULL
            AND token_digest IS NULL
            AND token_key_id IS NULL
            AND token_version IS NULL
        )
        OR
        (
            token IS NULL
            AND octet_length(token_digest) = 32
            AND token_key_id IS NOT NULL
            AND token_key_id <> ''
            AND token_version > 0
        )
    );

CREATE UNIQUE INDEX verification_tokens_digest_unique
    ON public.verification_tokens (token_digest)
    WHERE token_digest IS NOT NULL;

CREATE INDEX idx_verification_tokens_digest_consume
    ON public.verification_tokens (
        email,
        token_type,
        token_key_id,
        token_version,
        token_digest
    )
    WHERE used_at IS NULL AND token_digest IS NOT NULL;

