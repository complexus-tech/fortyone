-- 000023_attachments.up.sql
CREATE TABLE public.attachments (
    attachment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    filename character varying(255) NOT NULL,
    size bigint NOT NULL,
    mime_type character varying(255) NOT NULL,
    uploaded_by uuid,
    workspace_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    blob_name character varying(255) DEFAULT ''::character varying NOT NULL
);

ALTER TABLE ONLY public.attachments ADD CONSTRAINT attachments_pkey PRIMARY KEY (attachment_id);

ALTER TABLE ONLY public.attachments 
    ADD CONSTRAINT attachments_uploaded_by_fkey FOREIGN KEY (uploaded_by) REFERENCES public.users(user_id) ON UPDATE SET NULL;

ALTER TABLE ONLY public.attachments 
    ADD CONSTRAINT attachments_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_attachments_uploaded_by ON public.attachments USING btree (uploaded_by);
