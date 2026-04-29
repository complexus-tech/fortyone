CREATE TABLE public.github_comment_links (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    story_id uuid NOT NULL,
    repository_id uuid NOT NULL,
    local_comment_id uuid,
    github_comment_id bigint NOT NULL,
    source text NOT NULL,
    created_by_user_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT github_comment_links_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT github_comment_links_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE,
    CONSTRAINT github_comment_links_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.github_repositories(id) ON DELETE CASCADE,
    CONSTRAINT github_comment_links_local_comment_id_fkey FOREIGN KEY (local_comment_id) REFERENCES public.story_comments(comment_id) ON DELETE SET NULL,
    CONSTRAINT github_comment_links_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT github_comment_links_source_check CHECK (source IN ('fortyone', 'github')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX github_comment_links_repository_comment_key
    ON public.github_comment_links USING btree (repository_id, github_comment_id);

CREATE INDEX idx_github_comment_links_story
    ON public.github_comment_links USING btree (workspace_id, story_id);
