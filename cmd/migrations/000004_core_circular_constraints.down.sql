-- 000004_core_circular_constraints.down.sql
ALTER TABLE ONLY public.users DROP CONSTRAINT IF EXISTS users_last_used_workspace_id_fkey;
ALTER TABLE ONLY public.workspaces DROP CONSTRAINT IF EXISTS workspaces_created_by_fkey;
ALTER TABLE ONLY public.workspaces DROP CONSTRAINT IF EXISTS workspaces_deleted_by_fkey;
