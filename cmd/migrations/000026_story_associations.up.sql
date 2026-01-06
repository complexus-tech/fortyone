-- 000026_story_associations.up.sql
CREATE TABLE "public"."story_associations" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "from_story_id" uuid NOT NULL,
    "to_story_id" uuid NOT NULL,
    "association_type" varchar(50) NOT NULL CHECK ((association_type)::text = ANY ((ARRAY['blocking'::character varying, 'related'::character varying, 'duplicate'::character varying])::text[])),
    "workspace_id" uuid NOT NULL,
    "created_at" timestamptz DEFAULT now(),
    CONSTRAINT "story_associations_to_story_id_fkey" FOREIGN KEY ("to_story_id") REFERENCES "public"."stories"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_story_associations_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "story_associations_from_story_id_fkey" FOREIGN KEY ("from_story_id") REFERENCES "public"."stories"("id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);


-- Indices
CREATE UNIQUE INDEX unique_story_association ON public.story_associations USING btree (from_story_id, to_story_id, association_type);
CREATE INDEX idx_story_associations_from ON public.story_associations USING btree (from_story_id);
CREATE INDEX idx_story_associations_to ON public.story_associations USING btree (to_story_id);
CREATE INDEX idx_story_associations_workspace ON public.story_associations USING btree (workspace_id);
