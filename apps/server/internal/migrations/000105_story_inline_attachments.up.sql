CREATE TABLE public.story_inline_attachments (
    story_id uuid NOT NULL REFERENCES public.stories(id) ON DELETE CASCADE,
    attachment_id uuid NOT NULL REFERENCES public.attachments(attachment_id) ON DELETE CASCADE,
    created_by uuid NOT NULL REFERENCES public.users(user_id),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (story_id, attachment_id)
);

CREATE INDEX story_inline_attachments_attachment
    ON public.story_inline_attachments (attachment_id);
