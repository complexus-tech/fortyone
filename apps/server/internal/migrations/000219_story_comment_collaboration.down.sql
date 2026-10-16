DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.story_comment_reactions)
       OR EXISTS (SELECT 1 FROM public.story_comment_resolutions)
       OR EXISTS (SELECT 1 FROM public.story_comment_reads) THEN
        RAISE EXCEPTION 'story comment collaboration data exists; use a forward fix';
    END IF;
END;
$$;

DROP INDEX public.idx_story_comments_unread;
DROP TABLE public.story_comment_reads;
DROP TABLE public.story_comment_resolutions;
DROP TABLE public.story_comment_reactions;
