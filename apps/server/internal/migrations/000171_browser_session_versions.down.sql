DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.users
        WHERE auth_session_version <> 1
    ) THEN
        RAISE EXCEPTION
            'cannot remove browser session versions after a session epoch has been revoked'
            USING ERRCODE = '55000';
    END IF;
END;
$$;

ALTER TABLE public.users
    DROP CONSTRAINT users_auth_session_version_positive,
    DROP COLUMN auth_session_version;
