ALTER TABLE public.feedback_items
    ADD COLUMN description_html text NOT NULL DEFAULT '';

CREATE TABLE public.feedback_item_attachments (
    item_id uuid NOT NULL,
    attachment_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_item_attachments_item_id_fkey
        FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE CASCADE,
    CONSTRAINT feedback_item_attachments_attachment_id_fkey
        FOREIGN KEY (attachment_id) REFERENCES public.attachments(attachment_id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, attachment_id)
);

CREATE INDEX idx_feedback_item_attachments_attachment
    ON public.feedback_item_attachments (attachment_id);
