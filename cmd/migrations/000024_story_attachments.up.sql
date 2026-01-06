-- 000024_story_attachments.up.sql
CREATE TABLE "public"."story_attachments" (
    "story_id" uuid NOT NULL,
    "attachment_id" uuid NOT NULL,
    CONSTRAINT "story_attachments_attachment_id_fkey" FOREIGN KEY ("attachment_id") REFERENCES "public"."attachments"("attachment_id") ON DELETE CASCADE,
    CONSTRAINT "story_attachments_story_id_fkey" FOREIGN KEY ("story_id") REFERENCES "public"."stories"("id") ON DELETE CASCADE,
    PRIMARY KEY ("story_id","attachment_id")
);
