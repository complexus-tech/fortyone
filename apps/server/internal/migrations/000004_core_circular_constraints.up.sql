-- 000004_core_circular_constraints.up.sql
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_last_used_workspace_id_fkey FOREIGN KEY (last_used_workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE SET NULL;
