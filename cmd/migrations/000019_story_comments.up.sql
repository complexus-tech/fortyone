-- 000019_story_comments.up.sql
CREATE TABLE public.story_comments (
    comment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    content text NOT NULL,
    story_id uuid NOT NULL,
    commenter_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    parent_id uuid
);

ALTER TABLE ONLY public.story_comments ADD CONSTRAINT story_comments_pkey PRIMARY KEY (comment_id);

ALTER TABLE ONLY public.story_comments 
    ADD CONSTRAINT story_comments_commenter_id_fkey FOREIGN KEY (commenter_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_comments 
    ADD CONSTRAINT story_comments_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.story_comments(comment_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_comments 
    ADD CONSTRAINT story_comments_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE;
