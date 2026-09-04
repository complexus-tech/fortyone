CREATE TABLE public.google_drive_revocation_outbox (
    revocation_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_account_id uuid,
    user_id uuid NOT NULL,
    google_subject text NOT NULL CHECK (google_subject <> ''),
    installation_generation uuid NOT NULL,
    credential_payload text,
    credential_key_version smallint NOT NULL CHECK (credential_key_version > 0),
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'completed', 'superseded', 'failed')),
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    available_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    claim_token uuid,
    lease_expires_at timestamptz,
    last_error text CHECK (last_error IS NULL OR char_length(last_error) <= 2000),
    terminal_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (google_subject, installation_generation),
    CHECK (
        credential_payload IS NULL
        OR credential_payload LIKE 'vault.v2.%'
    ),
    CHECK (
        (status = 'processing' AND claim_token IS NOT NULL AND lease_expires_at IS NOT NULL)
        OR (status <> 'processing' AND claim_token IS NULL AND lease_expires_at IS NULL)
    ),
    CHECK (
        (status IN ('completed', 'superseded') AND credential_payload IS NULL AND terminal_at IS NOT NULL)
        OR (status = 'failed' AND credential_payload IS NOT NULL AND terminal_at IS NOT NULL)
        OR (status IN ('pending', 'processing') AND credential_payload IS NOT NULL AND terminal_at IS NULL)
    )
);

CREATE INDEX google_drive_revocation_outbox_ready
    ON public.google_drive_revocation_outbox (available_at, created_at, revocation_id)
    WHERE status = 'pending';

CREATE INDEX google_drive_revocation_outbox_stale_lease
    ON public.google_drive_revocation_outbox (lease_expires_at, revocation_id)
    WHERE status = 'processing';

CREATE INDEX google_drive_revocation_outbox_subject
    ON public.google_drive_revocation_outbox (google_subject, created_at, revocation_id)
    WHERE status IN ('pending', 'processing', 'failed');

-- Google revocation applies to the Google subject's grant across the Cloud
-- project, not only to one FortyOne user row. Refuse an ambiguous pre-existing
-- ownership state rather than silently selecting an owner during deployment.
DO $$
BEGIN
    IF EXISTS (
        SELECT account.google_subject
        FROM public.google_drive_accounts AS account
        WHERE account.revoked_at IS NULL
        GROUP BY account.google_subject
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION
            'cannot enforce exclusive Google Drive subject ownership while duplicate active subjects exist'
            USING ERRCODE = '55000';
    END IF;
END;
$$;

CREATE UNIQUE INDEX google_drive_accounts_active_google_subject
    ON public.google_drive_accounts (google_subject)
    WHERE revoked_at IS NULL;

-- This helper copies only an existing vault envelope. It has no provider I/O,
-- no foreign keys by design, and remains durable when a user/account is
-- deleted. The immutable generation is the remote-revocation fence.
CREATE FUNCTION public.stage_google_drive_account_revocation(
    staged_account_id uuid,
    staged_user_id uuid,
    staged_google_subject text,
    staged_installation_generation uuid,
    staged_credential_payload text,
    staged_credential_key_version smallint
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    IF staged_google_subject = ''
       OR staged_credential_payload = ''
       OR staged_credential_payload NOT LIKE 'vault.v2.%' THEN
        RETURN;
    END IF;

    INSERT INTO public.google_drive_revocation_outbox (
        source_account_id,
        user_id,
        google_subject,
        installation_generation,
        credential_payload,
        credential_key_version
    ) VALUES (
        staged_account_id,
        staged_user_id,
        staged_google_subject,
        staged_installation_generation,
        staged_credential_payload,
        staged_credential_key_version
    )
    ON CONFLICT (google_subject, installation_generation) DO NOTHING;
END;
$$;

-- Replace the original local-only orphan cleanup. Staging and local teardown
-- commit together; a provider outage can never roll back the local disconnect.
CREATE OR REPLACE FUNCTION public.revoke_orphaned_google_drive_account()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    orphaned_account public.google_drive_accounts%ROWTYPE;
BEGIN
    -- A Picker grant is proof for one actor, account, and workspace. Removing
    -- that workspace binding must destroy the proof even when another
    -- workspace still keeps the personal Google account connected.
    DELETE FROM public.google_drive_file_grants AS grant_record
    USING public.google_drive_files AS file
    WHERE grant_record.file_id = file.file_id
      AND file.workspace_id = OLD.workspace_id
      AND grant_record.user_id = OLD.user_id
      AND grant_record.account_id = OLD.account_id;

    SELECT account.*
    INTO orphaned_account
    FROM public.google_drive_accounts AS account
    WHERE account.account_id = OLD.account_id
      AND account.revoked_at IS NULL
      AND NOT EXISTS (
          SELECT 1
          FROM public.google_drive_workspace_connections AS connection
          WHERE connection.account_id = account.account_id
      )
    FOR UPDATE;

    IF NOT FOUND THEN
        RETURN OLD;
    END IF;

    PERFORM public.stage_google_drive_account_revocation(
        orphaned_account.account_id,
        orphaned_account.user_id,
        orphaned_account.google_subject,
        orphaned_account.installation_generation,
        orphaned_account.credential_payload,
        orphaned_account.credential_key_version
    );

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
    WHERE account.account_id = orphaned_account.account_id
      AND account.revoked_at IS NULL;

    RETURN OLD;
END;
$$;

-- PostgreSQL cascade ordering must not determine whether the credential is
-- retained for remote cleanup. Account deletion stages the same immutable job
-- before the row and its user foreign key disappear; the unique key makes the
-- connection-trigger and account-trigger paths idempotent.
CREATE FUNCTION public.stage_google_drive_revocation_before_account_delete()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.revoked_at IS NULL THEN
        PERFORM public.stage_google_drive_account_revocation(
            OLD.account_id,
            OLD.user_id,
            OLD.google_subject,
            OLD.installation_generation,
            OLD.credential_payload,
            OLD.credential_key_version
        );
    END IF;

    RETURN OLD;
END;
$$;

CREATE TRIGGER google_drive_accounts_stage_revocation_before_delete
BEFORE DELETE ON public.google_drive_accounts
FOR EACH ROW
EXECUTE FUNCTION public.stage_google_drive_revocation_before_account_delete();
