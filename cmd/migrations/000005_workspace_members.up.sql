-- 000005_workspace_members.up.sql
DROP TYPE IF EXISTS "public"."user_role";
CREATE TYPE "public"."user_role" AS ENUM ('admin', 'member', 'guest', 'system');

-- Table Definition
CREATE TABLE "public"."workspace_members" (
    "workspace_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    "role" "public"."user_role" NOT NULL DEFAULT 'member'::user_role,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "workspace_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    CONSTRAINT "workspace_members_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("workspace_id","user_id")
);


-- Indices
CREATE INDEX idx_workspace_members_user_id ON public.workspace_members USING btree (user_id);
CREATE INDEX idx_workspace_members_workspace_user ON public.workspace_members USING btree (workspace_id, user_id);
