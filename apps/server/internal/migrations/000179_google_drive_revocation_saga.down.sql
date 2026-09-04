DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.google_drive_revocation_outbox) THEN
        RAISE EXCEPTION
            'cannot roll back Google Drive revocation saga while durable revocation records exist'
            USING ERRCODE = '55000';
    END IF;
END;
$$;

DROP TRIGGER IF EXISTS google_drive_accounts_stage_revocation_before_delete
    ON public.google_drive_accounts;
DROP FUNCTION IF EXISTS public.stage_google_drive_revocation_before_account_delete();

CREATE OR REPLACE FUNCTION public.revoke_orphaned_google_drive_account()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE public.google_drive_accounts AS account
    SET credential_payload = '',
        google_subject = '',
        email = '',
        display_name = NULL,
        scopes = '{}'::text[],
        requires_reauthorization = TRUE,
        last_error_code = NULL,
        revoked_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE account.account_id = OLD.account_id
      AND account.revoked_at IS NULL
      AND NOT EXISTS (
          SELECT 1
          FROM public.google_drive_workspace_connections AS connection
          WHERE connection.account_id = account.account_id
      );

    RETURN OLD;
END;
$$;

DROP FUNCTION IF EXISTS public.stage_google_drive_account_revocation(
    uuid,
    uuid,
    text,
    uuid,
    text,
    smallint
);
DROP INDEX IF EXISTS public.google_drive_accounts_active_google_subject;
DROP TABLE IF EXISTS public.google_drive_revocation_outbox;
