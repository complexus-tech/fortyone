-- 000038_github_commits.up.sql
CREATE TABLE public.github_commits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    repository_id uuid NOT NULL,
    story_id uuid,
    sha text NOT NULL,
    message text NOT NULL,
    author_username text NOT NULL,
    author_email text,
    committed_at timestamp without time zone NOT NULL,
    url text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.github_commits ADD CONSTRAINT github_commits_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.github_commits ADD CONSTRAINT github_commits_repository_id_sha_key UNIQUE (repository_id, sha);

ALTER TABLE ONLY public.github_commits 
    ADD CONSTRAINT github_commits_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.github_repositories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.github_commits 
    ADD CONSTRAINT github_commits_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE SET NULL;

CREATE INDEX idx_github_commits_repository_committed ON public.github_commits USING btree (repository_id, committed_at);
CREATE INDEX idx_github_commits_story ON public.github_commits USING btree (story_id);
