-- 000002_users.up.sql
CREATE TABLE "public"."users" (
    "user_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "username" varchar(255) NOT NULL,
    "email" varchar(255) NOT NULL,
    "full_name" varchar(255),
    "avatar_url" varchar(255),
    "is_active" bool NOT NULL DEFAULT true,
    "last_login_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_used_workspace_id" uuid,
    "has_seen_walkthrough" bool NOT NULL DEFAULT false,
    "timezone" varchar(255) NOT NULL DEFAULT 'Antarctica/Troll'::character varying,
    "inactivity_warning_sent_at" timestamp,
    "is_system" bool NOT NULL DEFAULT false,
    PRIMARY KEY ("user_id")
);


-- Indices
CREATE UNIQUE INDEX users_email_key ON public.users USING btree (email);
CREATE INDEX idx_users_active_fullname ON public.users USING btree (is_active, full_name);
CREATE INDEX idx_users_last_used_workspace_id ON public.users USING btree (last_used_workspace_id);
CREATE INDEX idx_users_active ON public.users USING btree (is_active) WHERE (is_active = true);
