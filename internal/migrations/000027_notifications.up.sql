-- 000027_notifications.up.sql
DROP TYPE IF EXISTS "public"."notification_type";
CREATE TYPE "public"."notification_type" AS ENUM ('story_update', 'story_comment', 'comment_reply', 'objective_update', 'key_result_update', 'mention');
DROP TYPE IF EXISTS "public"."entity_type";
CREATE TYPE "public"."entity_type" AS ENUM ('story', 'comment', 'objective', 'key_result');

-- Table Definition
CREATE TABLE "public"."notifications" (
    "notification_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "recipient_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "type" "public"."notification_type" NOT NULL,
    "entity_type" "public"."entity_type" NOT NULL,
    "entity_id" uuid NOT NULL,
    "actor_id" uuid NOT NULL,
    "title" text NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP,
    "read_at" timestamptz,
    "message" jsonb NOT NULL,
    CONSTRAINT "fk_notifications_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "fk_notifications_recipient" FOREIGN KEY ("recipient_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    CONSTRAINT "fk_notifications_actor" FOREIGN KEY ("actor_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    PRIMARY KEY ("notification_id")
);
