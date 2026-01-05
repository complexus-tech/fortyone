-- 000041_story_github_links.up.sql
CREATE TABLE public.story_github_links (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    story_id uuid NOT NULL,
    github_pull_request_id bigint,
    github_branch_name text,
    github_repository_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.story_github_links ADD CONSTRAINT story_github_links_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.story_github_links 
    ADD CONSTRAINT story_github_links_github_repository_id_fkey FOREIGN KEY (github_repository_id) REFERENCES public.github_repositories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_github_links 
    ADD CONSTRAINT story_github_links_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE;

CREATE INDEX idx_story_github_links_branch ON public.story_github_links USING btree (github_branch_name);
CREATE INDEX idx_story_github_links_pr ON public.story_github_links USING btree (github_pull_request_id);
CREATE INDEX idx_story_github_links_story ON public.story_github_links USING btree (story_id);
