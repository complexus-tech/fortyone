-- 000048_chat_sessions.up.sql
CREATE TABLE "public"."chat_sessions" (
    "id" text NOT NULL,
    "user_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "title" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    CONSTRAINT "chat_sessions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    CONSTRAINT "chat_sessions_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);

-- Indices
CREATE INDEX idx_chat_sessions_user_workspace ON public.chat_sessions USING btree (user_id, workspace_id);
CREATE INDEX idx_chat_sessions_workspace ON public.chat_sessions USING btree (workspace_id);
CREATE INDEX idx_sessions_user_date ON public.chat_sessions USING btree (user_id, created_at);
CREATE INDEX idx_chat_sessions_deleted_at ON public.chat_sessions USING btree (deleted_at) WHERE (deleted_at IS NULL);
