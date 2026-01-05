-- 000024_story_attachments.up.sql
CREATE TABLE public.story_attachments (
    story_id uuid NOT NULL,
    attachment_id uuid NOT NULL
);

ALTER TABLE ONLY public.story_attachments ADD CONSTRAINT story_attachments_pkey PRIMARY KEY (story_id, attachment_id);

ALTER TABLE ONLY public.story_attachments 
    ADD CONSTRAINT story_attachments_attachment_id_fkey FOREIGN KEY (attachment_id) REFERENCES public.attachments(attachment_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_attachments 
    ADD CONSTRAINT story_attachments_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE;
