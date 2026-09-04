CREATE FUNCTION public.cleanup_google_drive_on_user_deactivation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    -- The application acquires this session-level key before OAuth exchange,
    -- disconnect, and revocation. A non-blocking transaction lock avoids the
    -- row-lock/advisory-lock inversion that a blocking AFTER trigger could
    -- otherwise create. SQLSTATE 40001 makes the whole deactivation retryable.
    IF NOT pg_try_advisory_xact_lock(
        hashtextextended(
            'google-drive-provider-user:' || CAST(NEW.user_id AS text),
            0
        )
    ) THEN
        RAISE EXCEPTION
            'Google Drive user lifecycle is busy; retry user deactivation'
            USING ERRCODE = '40001';
    END IF;

    IF NOT pg_try_advisory_xact_lock(
        hashtextextended(
            'google-drive:' || CAST(NEW.user_id AS text),
            0
        )
    ) THEN
        RAISE EXCEPTION
            'Google Drive account lifecycle is busy; retry user deactivation'
            USING ERRCODE = '40001';
    END IF;

    -- The existing per-connection trigger purges workspace-scoped Picker
    -- grants. The final binding also stages the sealed credential in the
    -- generation-fenced outbox before scrubbing the local account.
    DELETE FROM public.google_drive_workspace_connections AS connection
    WHERE connection.user_id = NEW.user_id;

    RETURN NEW;
END;
$$;

CREATE TRIGGER users_cleanup_google_drive_on_deactivation
AFTER UPDATE OF is_active ON public.users
FOR EACH ROW
WHEN (OLD.is_active = TRUE AND NEW.is_active = FALSE)
EXECUTE FUNCTION public.cleanup_google_drive_on_user_deactivation();

-- Repair users deactivated before the trigger existed. Locks are acquired in
-- stable user order and held only for this migration transaction; rollout must
-- still pause lifecycle traffic as required by migration 000179.
DO $$
DECLARE
    deactivated_user_id uuid;
BEGIN
    FOR deactivated_user_id IN
        SELECT DISTINCT connection.user_id
        FROM public.google_drive_workspace_connections AS connection
        INNER JOIN public.users AS account
            ON account.user_id = connection.user_id
           AND account.is_active = FALSE
        ORDER BY connection.user_id
    LOOP
        PERFORM pg_advisory_xact_lock(
            hashtextextended(
                'google-drive-provider-user:' || CAST(deactivated_user_id AS text),
                0
            )
        );
        PERFORM pg_advisory_xact_lock(
            hashtextextended(
                'google-drive:' || CAST(deactivated_user_id AS text),
                0
            )
        );

        -- Verified sign-in may reactivate an account while this migration is
        -- working through the initial candidate list. Lock and recheck the
        -- user only after the provider/local gates so an active account never
        -- loses a connection selected from a stale snapshot.
        PERFORM account.user_id
        FROM public.users AS account
        WHERE account.user_id = deactivated_user_id
          AND account.is_active = FALSE
        FOR UPDATE;

        IF NOT FOUND THEN
            CONTINUE;
        END IF;

        DELETE FROM public.google_drive_workspace_connections AS connection
        WHERE connection.user_id = deactivated_user_id;
    END LOOP;
END;
$$;
