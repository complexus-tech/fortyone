ALTER TABLE public.users
    ADD COLUMN auth_session_version bigint NOT NULL DEFAULT 1;

ALTER TABLE public.users
    ADD CONSTRAINT users_auth_session_version_positive
    CHECK (auth_session_version > 0);

COMMENT ON COLUMN public.users.auth_session_version IS
    'Monotonic epoch for first-party browser sessions. Deactivation and explicit revocation increment this value; reactivation never reduces it.';
