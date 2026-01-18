-- 000046_user_memories.up.sql
CREATE TABLE "public"."user_memories" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "workspace_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    "content" text NOT NULL,
    "last_accessed_at" timestamptz DEFAULT now(),
    "created_at" timestamptz DEFAULT now(),
    "updated_at" timestamptz DEFAULT now(),
    CONSTRAINT "user_memories_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "user_memories_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);
