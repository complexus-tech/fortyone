-- 000038_github_commits.up.sql
CREATE TABLE "public"."github_commits" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "repository_id" uuid NOT NULL,
    "story_id" uuid,
    "sha" text NOT NULL,
    "message" text NOT NULL,
    "author_username" text NOT NULL,
    "author_email" text,
    "committed_at" timestamp NOT NULL,
    "url" text NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "github_commits_repository_id_fkey" FOREIGN KEY ("repository_id") REFERENCES "public"."github_repositories"("id") ON DELETE CASCADE,
    CONSTRAINT "github_commits_story_id_fkey" FOREIGN KEY ("story_id") REFERENCES "public"."stories"("id") ON DELETE SET NULL,
    PRIMARY KEY ("id")
);

-- Indices
CREATE UNIQUE INDEX github_commits_repository_id_sha_key ON public.github_commits USING btree (repository_id, sha);
CREATE INDEX idx_github_commits_story ON public.github_commits USING btree (story_id);
CREATE INDEX idx_github_commits_repository_committed ON public.github_commits USING btree (repository_id, committed_at);
