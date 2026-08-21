CREATE TABLE public.user_external_identities (
    identity_id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    provider varchar(32) NOT NULL,
    issuer varchar(512) NOT NULL,
    subject varchar(255) NOT NULL,
    email_at_link varchar(255) NOT NULL,
    last_authenticated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_external_identities_pkey PRIMARY KEY (identity_id),
    CONSTRAINT user_external_identities_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT user_external_identities_provider_check
        CHECK (provider IN ('microsoft')),
    CONSTRAINT user_external_identities_provider_issuer_subject_key
        UNIQUE (provider, issuer, subject),
    CONSTRAINT user_external_identities_user_provider_key
        UNIQUE (user_id, provider)
);

CREATE INDEX user_external_identities_user_id_idx
    ON public.user_external_identities (user_id);
