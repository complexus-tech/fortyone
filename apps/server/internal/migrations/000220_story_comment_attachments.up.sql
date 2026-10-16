CREATE TABLE public.story_comment_attachments (
    comment_id uuid NOT NULL REFERENCES public.story_comments(comment_id) ON DELETE CASCADE,
    attachment_id uuid NOT NULL,
    -- Check after transaction cascades, so workspace deletion can remove both
    -- attachments and comments while independent attachment deletion is denied.
    CONSTRAINT story_comment_attachments_attachment_fk
        FOREIGN KEY (attachment_id) REFERENCES public.attachments(attachment_id)
            ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED,
    PRIMARY KEY (comment_id, attachment_id)
);

CREATE INDEX story_comment_attachments_attachment_idx
    ON public.story_comment_attachments (attachment_id);
