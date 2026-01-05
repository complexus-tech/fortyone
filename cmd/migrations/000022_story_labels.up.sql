-- 000022_story_labels.up.sql
CREATE TABLE public.story_labels (
    story_id uuid NOT NULL,
    label_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE ONLY public.story_labels ADD CONSTRAINT story_labels_pkey PRIMARY KEY (story_id, label_id);

ALTER TABLE ONLY public.story_labels 
    ADD CONSTRAINT story_labels_label_id_fkey FOREIGN KEY (label_id) REFERENCES public.labels(label_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.story_labels 
    ADD CONSTRAINT story_labels_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE;

CREATE INDEX idx_story_labels_label_id_story_id ON public.story_labels USING btree (label_id, story_id);
CREATE INDEX idx_story_labels_story_id ON public.story_labels USING btree (story_id);
