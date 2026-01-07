-- 000041_story_github_links.up.sql
CREATE TABLE "public"."story_github_links" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "story_id" uuid NOT NULL,
    "repository_id" uuid NOT NULL,
    "link_type" text NOT NULL CHECK (link_type = ANY (ARRAY['branch'::text, 'pull_request'::text, 'issue'::text])),
    "github_id" text NOT NULL,
    "github_url" text NOT NULL,
    "title" text,
    "state" text,
    "metadata" jsonb,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "story_github_links_story_id_fkey" FOREIGN KEY ("story_id") REFERENCES "public"."stories"("id") ON DELETE CASCADE,
    CONSTRAINT "story_github_links_repository_id_fkey" FOREIGN KEY ("repository_id") REFERENCES "public"."github_repositories"("id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);


-- Indices
CREATE INDEX idx_story_github_links_story ON public.story_github_links USING btree (story_id);
CREATE INDEX idx_story_github_links_repository ON public.story_github_links USING btree (repository_id);
CREATE UNIQUE INDEX story_github_links_story_id_repository_id_link_type_github__key ON public.story_github_links USING btree (story_id, repository_id, link_type, github_id);
