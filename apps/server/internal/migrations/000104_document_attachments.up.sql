CREATE TABLE public.document_attachments (
    document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
    attachment_id uuid NOT NULL REFERENCES public.attachments(attachment_id) ON DELETE CASCADE,
    created_by uuid NOT NULL REFERENCES public.users(user_id),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (document_id, attachment_id)
);

CREATE INDEX document_attachments_attachment
    ON public.document_attachments (attachment_id);
