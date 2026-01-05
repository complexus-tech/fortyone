-- 000002_users.up.sql
CREATE TABLE public.users (
    user_id uuid DEFAULT gen_random_uuid() NOT NULL,
    username character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    full_name character varying(255),
    avatar_url character varying(255),
    is_active boolean DEFAULT true NOT NULL,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_used_workspace_id uuid,
    has_seen_walkthrough boolean DEFAULT false NOT NULL,
    timezone character varying(255) DEFAULT 'Antarctica/Troll'::character varying NOT NULL,
    inactivity_warning_sent_at timestamp without time zone,
    is_system boolean DEFAULT false NOT NULL
);

ALTER TABLE ONLY public.users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);

CREATE INDEX idx_users_active ON public.users USING btree (is_active) WHERE (is_active = true);
CREATE INDEX idx_users_active_fullname ON public.users USING btree (is_active, full_name);
CREATE INDEX idx_users_last_used_workspace_id ON public.users USING btree (last_used_workspace_id);
