-- 000026_story_associations.up.sql
CREATE TABLE public.story_associations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    from_story_id uuid NOT NULL,
    to_story_id uuid NOT NULL,
    association_type character varying(50) NOT NULL,
    workspace_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT no_self_association CHECK ((from_story_id <> to_story_id)),
    CONSTRAINT story_associations_association_type_check CHECK (((association_type)::text = ANY ((ARRAY['blocking'::character varying, 'related'::character varying, 'duplicate'::character varying])::text[])))
);

ALTER TABLE ONLY public.story_associations ADD CONSTRAINT story_associations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.story_associations 
    ADD CONSTRAINT fk_story_associations_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_associations 
    ADD CONSTRAINT story_associations_from_story_id_fkey FOREIGN KEY (from_story_id) REFERENCES public.stories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_associations 
    ADD CONSTRAINT story_associations_to_story_id_fkey FOREIGN KEY (to_story_id) REFERENCES public.stories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_associations ADD CONSTRAINT unique_story_association UNIQUE (from_story_id, to_story_id, association_type);

CREATE INDEX idx_story_associations_from ON public.story_associations USING btree (from_story_id);
CREATE INDEX idx_story_associations_to ON public.story_associations USING btree (to_story_id);
CREATE INDEX idx_story_associations_workspace ON public.story_associations USING btree (workspace_id);
