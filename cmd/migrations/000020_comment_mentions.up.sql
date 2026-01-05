-- 000020_comment_mentions.up.sql
CREATE TABLE public.comment_mentions (
    comment_id uuid NOT NULL,
    mentioned_user_id uuid NOT NULL
);

ALTER TABLE ONLY public.comment_mentions ADD CONSTRAINT comment_mentions_pkey PRIMARY KEY (comment_id, mentioned_user_id);

ALTER TABLE ONLY public.comment_mentions 
    ADD CONSTRAINT comment_mentions_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.story_comments(comment_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.comment_mentions 
    ADD CONSTRAINT comment_mentions_user_id_fkey FOREIGN KEY (mentioned_user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;
