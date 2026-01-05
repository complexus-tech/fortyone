-- 000015_okr_activities.up.sql
CREATE TABLE public.okr_activities (
    activity_id uuid DEFAULT gen_random_uuid() NOT NULL,
    objective_id uuid NOT NULL,
    key_result_id uuid,
    user_id uuid NOT NULL,
    activity_type public.okr_activity_type NOT NULL,
    update_type public.okr_update_type NOT NULL,
    field_changed character varying(100),
    current_value text,
    comment text,
    created_at timestamp without time zone DEFAULT now(),
    workspace_id uuid NOT NULL
);

ALTER TABLE ONLY public.okr_activities ADD CONSTRAINT okr_activities_pkey PRIMARY KEY (activity_id);

ALTER TABLE ONLY public.okr_activities 
    ADD CONSTRAINT okr_activities_key_result_id_fkey FOREIGN KEY (key_result_id) REFERENCES public.key_results(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.okr_activities 
    ADD CONSTRAINT okr_activities_objective_id_fkey FOREIGN KEY (objective_id) REFERENCES public.objectives(objective_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.okr_activities 
    ADD CONSTRAINT okr_activities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);

ALTER TABLE ONLY public.okr_activities 
    ADD CONSTRAINT okr_activities_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id);
