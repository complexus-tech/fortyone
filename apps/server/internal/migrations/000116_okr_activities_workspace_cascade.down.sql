ALTER TABLE public.okr_activities
    DROP CONSTRAINT okr_activities_workspace_id_fkey,
    ADD CONSTRAINT okr_activities_workspace_id_fkey
        FOREIGN KEY (workspace_id)
        REFERENCES public.workspaces(workspace_id);
