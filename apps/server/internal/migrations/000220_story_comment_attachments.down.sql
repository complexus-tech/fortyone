DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.story_comment_attachments) THEN
        RAISE EXCEPTION 'Comment attachment references exist; repair forward instead of removing retained references';
    END IF;
END;
$$;

DROP TABLE public.story_comment_attachments;
