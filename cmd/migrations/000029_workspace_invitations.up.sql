-- 000029_workspace_invitations.up.sql
CREATE TABLE public.workspace_invitations (
    invitation_id uuid DEFAULT gen_random_uuid() NOT NULL,
    workspace_id uuid NOT NULL,
    inviter_id uuid NOT NULL,
    email character varying(255) NOT NULL,
    role public.user_role DEFAULT 'member'::public.user_role NOT NULL,
    token character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE ONLY public.workspace_invitations ADD CONSTRAINT workspace_invitations_pkey PRIMARY KEY (invitation_id);

ALTER TABLE ONLY public.workspace_invitations 
    ADD CONSTRAINT workspace_invitations_inviter_id_fkey FOREIGN KEY (inviter_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.workspace_invitations 
    ADD CONSTRAINT workspace_invitations_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_workspace_invitations_email ON public.workspace_invitations USING btree (email);
CREATE INDEX idx_workspace_invitations_workspace ON public.workspace_invitations USING btree (workspace_id);
CREATE UNIQUE INDEX workspace_invitations_token_key ON public.workspace_invitations USING btree (token);
