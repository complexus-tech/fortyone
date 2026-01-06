-- 000025_story_links.up.sql
CREATE TABLE "public"."story_links" (
    "link_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "title" varchar(255),
    "url" varchar(255) NOT NULL,
    "story_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "story_links_story_id_fkey" FOREIGN KEY ("story_id") REFERENCES "public"."stories"("id") ON DELETE CASCADE,
    PRIMARY KEY ("link_id")
);


-- Indices
CREATE INDEX idx_story_links_story_id ON public.story_links USING btree (story_id);
