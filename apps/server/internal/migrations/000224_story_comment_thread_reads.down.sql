DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.story_comment_thread_reads) THEN
        RAISE EXCEPTION 'Comment thread read markers remain; preserve them and repair forward';
    END IF;
END $$;

DROP TABLE public.story_comment_thread_reads;
