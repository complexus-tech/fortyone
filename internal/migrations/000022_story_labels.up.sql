-- 000022_story_labels.up.sql
CREATE TABLE "public"."story_labels" (
    "story_id" uuid NOT NULL,
    "label_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "story_labels_label_id_fkey" FOREIGN KEY ("label_id") REFERENCES "public"."labels"("label_id") ON DELETE CASCADE,
    CONSTRAINT "story_labels_story_id_fkey" FOREIGN KEY ("story_id") REFERENCES "public"."stories"("id") ON DELETE CASCADE,
    PRIMARY KEY ("story_id","label_id")
);


-- Indices
CREATE INDEX idx_story_labels_label_id_story_id ON public.story_labels USING btree (label_id, story_id);
CREATE INDEX idx_story_labels_story_id ON public.story_labels USING btree (story_id);
