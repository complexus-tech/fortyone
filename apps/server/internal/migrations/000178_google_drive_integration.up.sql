CREATE TABLE public.google_drive_accounts (
    account_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    google_subject text NOT NULL,
    email text NOT NULL,
    display_name text,
    credential_payload text NOT NULL,
    credential_key_version smallint NOT NULL,
    installation_generation uuid NOT NULL,
    scopes text[] NOT NULL DEFAULT '{}'::text[],
    expires_at timestamptz NOT NULL,
    requires_reauthorization boolean NOT NULL DEFAULT false,
    last_error_code text,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (account_id, user_id)
);

CREATE UNIQUE INDEX google_drive_accounts_active_identity
    ON public.google_drive_accounts (user_id, google_subject)
    WHERE revoked_at IS NULL;

CREATE TABLE public.google_drive_workspace_connections (
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    account_id uuid NOT NULL,
    connected_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, user_id),
    FOREIGN KEY (workspace_id, user_id)
        REFERENCES public.workspace_members(workspace_id, user_id)
        ON DELETE CASCADE,
    FOREIGN KEY (account_id, user_id)
        REFERENCES public.google_drive_accounts(account_id, user_id)
        ON DELETE CASCADE
);

CREATE INDEX google_drive_workspace_connections_account
    ON public.google_drive_workspace_connections (account_id);

-- Workspace bindings and their shared personal credential are one lifecycle.
-- The user-scoped advisory lock is also acquired by application writes so a
-- reconnect cannot race the removal of the final binding. The account row lock
-- additionally coordinates with foreign-key checks from defensive direct SQL.
CREATE FUNCTION public.lock_google_drive_user_lifecycle_on_delete()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(
        hashtextextended('google-drive:' || CAST(OLD.user_id AS text), 0)
    );

    IF TG_TABLE_NAME = 'google_drive_workspace_connections' THEN
        PERFORM account.account_id
        FROM public.google_drive_accounts AS account
        WHERE account.account_id = OLD.account_id
        FOR UPDATE;
    END IF;

    RETURN OLD;
END;
$$;

CREATE TRIGGER workspace_members_lock_google_drive_lifecycle
BEFORE DELETE ON public.workspace_members
FOR EACH ROW
EXECUTE FUNCTION public.lock_google_drive_user_lifecycle_on_delete();

CREATE TRIGGER google_drive_workspace_connections_lock_lifecycle
BEFORE DELETE ON public.google_drive_workspace_connections
FOR EACH ROW
EXECUTE FUNCTION public.lock_google_drive_user_lifecycle_on_delete();

CREATE FUNCTION public.revoke_orphaned_google_drive_account()
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

CREATE TRIGGER google_drive_workspace_connections_revoke_orphaned_account
AFTER DELETE ON public.google_drive_workspace_connections
FOR EACH ROW
EXECUTE FUNCTION public.revoke_orphaned_google_drive_account();

CREATE TABLE public.google_drive_oauth_states (
    state_hash text PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    workspace_slug text NOT NULL,
    return_url text,
    code_verifier text NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX google_drive_oauth_states_expiry
    ON public.google_drive_oauth_states (expires_at);

CREATE TABLE public.google_drive_files (
    file_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    google_file_id text NOT NULL,
    resource_key text,
    name text NOT NULL,
    mime_type text NOT NULL,
    web_view_link text NOT NULL,
    drive_id text,
    version text,
    size_bytes bigint CHECK (size_bytes IS NULL OR size_bytes >= 0),
    modified_at timestamptz,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    last_synced_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unavailable_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (workspace_id, google_file_id),
    UNIQUE (file_id, workspace_id)
);

CREATE TABLE public.google_drive_file_grants (
    file_id uuid NOT NULL REFERENCES public.google_drive_files(file_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    account_id uuid NOT NULL,
    verification_generation uuid NOT NULL DEFAULT gen_random_uuid(),
    last_verified_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (file_id, user_id),
    FOREIGN KEY (account_id, user_id)
        REFERENCES public.google_drive_accounts(account_id, user_id)
        ON DELETE CASCADE
);

CREATE INDEX google_drive_file_grants_account
    ON public.google_drive_file_grants (account_id);

CREATE TABLE public.google_drive_file_references (
    reference_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    file_id uuid NOT NULL,
    target_type text NOT NULL CHECK (target_type IN ('story', 'objective', 'document', 'comment')),
    story_id uuid REFERENCES public.stories(id) ON DELETE CASCADE,
    objective_id uuid REFERENCES public.objectives(objective_id) ON DELETE CASCADE,
    document_id uuid REFERENCES public.documents(document_id) ON DELETE CASCADE,
    comment_id uuid REFERENCES public.story_comments(comment_id) ON DELETE CASCADE,
    created_by_user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id, workspace_id)
        REFERENCES public.google_drive_files(file_id, workspace_id)
        ON DELETE CASCADE,
    CHECK (
        (target_type = 'story' AND story_id IS NOT NULL AND objective_id IS NULL AND document_id IS NULL AND comment_id IS NULL)
        OR (target_type = 'objective' AND story_id IS NULL AND objective_id IS NOT NULL AND document_id IS NULL AND comment_id IS NULL)
        OR (target_type = 'document' AND story_id IS NULL AND objective_id IS NULL AND document_id IS NOT NULL AND comment_id IS NULL)
        OR (target_type = 'comment' AND story_id IS NULL AND objective_id IS NULL AND document_id IS NULL AND comment_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX google_drive_story_reference_unique
    ON public.google_drive_file_references (file_id, story_id)
    WHERE target_type = 'story';

CREATE UNIQUE INDEX google_drive_objective_reference_unique
    ON public.google_drive_file_references (file_id, objective_id)
    WHERE target_type = 'objective';

CREATE UNIQUE INDEX google_drive_document_reference_unique
    ON public.google_drive_file_references (file_id, document_id)
    WHERE target_type = 'document';

CREATE UNIQUE INDEX google_drive_comment_reference_unique
    ON public.google_drive_file_references (file_id, comment_id)
    WHERE target_type = 'comment';

CREATE INDEX google_drive_file_references_target
    ON public.google_drive_file_references (
        workspace_id,
        target_type,
        COALESCE(story_id, objective_id, document_id, comment_id),
        created_at
    );

-- Target foreign keys cascade references, so keep file cleanup inside the
-- database rather than relying only on the explicit unlink endpoint. Locking
-- the parent file serializes concurrent final-reference deletions. Removing
-- the file then cascades its actor grants.
CREATE FUNCTION public.lock_google_drive_file_on_reference_delete()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM file.file_id
    FROM public.google_drive_files AS file
    WHERE file.file_id = OLD.file_id
    FOR UPDATE;

    RETURN OLD;
END;
$$;

CREATE TRIGGER google_drive_file_references_lock_file_lifecycle
BEFORE DELETE ON public.google_drive_file_references
FOR EACH ROW
EXECUTE FUNCTION public.lock_google_drive_file_on_reference_delete();

CREATE FUNCTION public.delete_orphaned_google_drive_file()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM public.google_drive_files AS file
    WHERE file.file_id = OLD.file_id
      AND NOT EXISTS (
          SELECT 1
          FROM public.google_drive_file_references AS reference
          WHERE reference.file_id = file.file_id
      );

    RETURN OLD;
END;
$$;

CREATE TRIGGER google_drive_file_references_delete_orphaned_file
AFTER DELETE ON public.google_drive_file_references
FOR EACH ROW
EXECUTE FUNCTION public.delete_orphaned_google_drive_file();

CREATE TABLE public.google_drive_create_operations (
    operation_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 1 AND 200),
    request_hash text NOT NULL,
    target_type text NOT NULL CHECK (target_type IN ('story', 'objective', 'document', 'comment')),
    target_id uuid NOT NULL,
    file_type text NOT NULL CHECK (file_type IN ('document', 'spreadsheet')),
    title text NOT NULL,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
    google_file_id text,
    reference_id uuid REFERENCES public.google_drive_file_references(reference_id) ON DELETE SET NULL,
    error_code text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (workspace_id, user_id, idempotency_key)
);

CREATE TABLE public.google_drive_document_imports (
    import_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    document_id uuid NOT NULL UNIQUE REFERENCES public.documents(document_id) ON DELETE CASCADE,
    reference_id uuid REFERENCES public.google_drive_file_references(reference_id) ON DELETE SET NULL,
    google_file_id text NOT NULL,
    source_version text,
    imported_by_user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE RESTRICT,
    imported_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX google_drive_document_imports_source
    ON public.google_drive_document_imports (workspace_id, google_file_id, imported_at DESC);
