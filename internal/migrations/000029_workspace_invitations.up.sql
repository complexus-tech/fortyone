-- 000029_workspace_invitations.up.sql

-- Table Definition
CREATE TABLE "public"."workspace_invitations" (
    "invitation_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "workspace_id" uuid NOT NULL,
    "inviter_id" uuid NOT NULL,
    "email" varchar(255) NOT NULL,
    "role" "public"."user_role" NOT NULL DEFAULT 'member'::user_role,
    "token" varchar(255) NOT NULL,
    "expires_at" timestamptz NOT NULL,
    "used_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "workspace_invitations_inviter_id_fkey" FOREIGN KEY ("inviter_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    CONSTRAINT "workspace_invitations_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("invitation_id")
);


-- Indices
CREATE INDEX idx_workspace_invitations_email ON public.workspace_invitations USING btree (email);
CREATE INDEX idx_workspace_invitations_workspace ON public.workspace_invitations USING btree (workspace_id);
CREATE UNIQUE INDEX workspace_invitations_token_key ON public.workspace_invitations USING btree (token);
