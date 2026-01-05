-- 000018_story_activities.up.sql
CREATE TABLE public.story_activities (
    activity_id uuid DEFAULT gen_random_uuid() NOT NULL,
    story_id uuid NOT NULL,
    activity_type character varying(50) NOT NULL,
    field_changed character varying(50) NOT NULL,
    current_value text NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    workspace_id uuid,
    old_value jsonb,
    new_value jsonb
);

ALTER TABLE ONLY public.story_activities ADD CONSTRAINT story_activities_pkey PRIMARY KEY (activity_id);

ALTER TABLE ONLY public.story_activities 
    ADD CONSTRAINT story_activities_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_activities 
    ADD CONSTRAINT story_activities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_activities 
    ADD CONSTRAINT story_activities_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_story_activities_burndown_events ON public.story_activities USING btree (workspace_id, created_at) WHERE ((field_changed)::text = ANY ((ARRAY['status_id'::character varying, 'sprint_id'::character varying])::text[]));
CREATE INDEX idx_story_activities_story_history ON public.story_activities USING btree (story_id, created_at DESC);
CREATE INDEX idx_story_id ON public.story_activities USING btree (story_id);
CREATE INDEX idx_user_id ON public.story_activities USING btree (user_id);
