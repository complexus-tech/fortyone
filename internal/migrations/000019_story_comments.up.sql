-- 000019_story_comments.up.sql
CREATE TABLE "public"."story_comments" (
    "comment_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "content" text NOT NULL,
    "story_id" uuid NOT NULL,
    "commenter_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "parent_id" uuid,
    CONSTRAINT "story_comments_story_id_fkey" FOREIGN KEY ("story_id") REFERENCES "public"."stories"("id") ON DELETE CASCADE,
    CONSTRAINT "story_comments_commenter_id_fkey" FOREIGN KEY ("commenter_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    CONSTRAINT "story_comments_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "public"."story_comments"("comment_id") ON DELETE CASCADE,
    PRIMARY KEY ("comment_id")
);
