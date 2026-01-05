-- 000025_story_links.up.sql
CREATE TABLE public.story_links (
    link_id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(255),
    url character varying(255) NOT NULL,
    story_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE ONLY public.story_links ADD CONSTRAINT story_links_pkey PRIMARY KEY (link_id);

ALTER TABLE ONLY public.story_links 
    ADD CONSTRAINT story_links_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE;

CREATE INDEX idx_story_links_story_id ON public.story_links USING btree (story_id);
