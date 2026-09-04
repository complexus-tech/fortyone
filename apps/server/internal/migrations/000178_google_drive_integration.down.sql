-- Migration 000178 becomes forward-only as soon as Google Drive stores any
-- durable credential, linked-file, creation, or import state. An empty schema
-- may be removed before adoption; established integrations require a forward fix.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.google_drive_accounts)
        OR EXISTS (SELECT 1 FROM public.google_drive_files)
        OR EXISTS (SELECT 1 FROM public.google_drive_file_references)
        OR EXISTS (SELECT 1 FROM public.google_drive_create_operations)
        OR EXISTS (SELECT 1 FROM public.google_drive_document_imports) THEN
        RAISE EXCEPTION 'migration 000178 cannot be rolled back after Google Drive integration data exists; deploy a forward fix'
            USING ERRCODE = '55000';
    END IF;
END $$;

DROP TRIGGER IF EXISTS workspace_members_lock_google_drive_lifecycle ON public.workspace_members;
DROP TRIGGER IF EXISTS google_drive_workspace_connections_lock_lifecycle ON public.google_drive_workspace_connections;
DROP TRIGGER IF EXISTS google_drive_workspace_connections_revoke_orphaned_account ON public.google_drive_workspace_connections;
DROP FUNCTION IF EXISTS public.lock_google_drive_user_lifecycle_on_delete();
DROP FUNCTION IF EXISTS public.revoke_orphaned_google_drive_account();

DROP TRIGGER IF EXISTS google_drive_file_references_lock_file_lifecycle ON public.google_drive_file_references;
DROP TRIGGER IF EXISTS google_drive_file_references_delete_orphaned_file ON public.google_drive_file_references;
DROP FUNCTION IF EXISTS public.lock_google_drive_file_on_reference_delete();
DROP FUNCTION IF EXISTS public.delete_orphaned_google_drive_file();

DROP TABLE IF EXISTS public.google_drive_document_imports;
DROP TABLE IF EXISTS public.google_drive_create_operations;
DROP TABLE IF EXISTS public.google_drive_file_references;
DROP TABLE IF EXISTS public.google_drive_file_grants;
DROP TABLE IF EXISTS public.google_drive_files;
DROP TABLE IF EXISTS public.google_drive_oauth_states;
DROP TABLE IF EXISTS public.google_drive_workspace_connections;
DROP TABLE IF EXISTS public.google_drive_accounts;
