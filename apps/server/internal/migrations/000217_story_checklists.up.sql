CREATE TABLE public.story_checklist_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id uuid NOT NULL REFERENCES public.stories(id) ON DELETE CASCADE,
    title varchar(255) NOT NULL CHECK (length(btrim(title)) > 0),
    completed boolean NOT NULL DEFAULT false,
    position integer NOT NULL CHECK (position >= 0),
    converted_story_id uuid REFERENCES public.stories(id) ON DELETE SET NULL,
    converted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT story_checklist_conversion_valid CHECK (
        (converted_story_id IS NULL OR converted_at IS NOT NULL)
        AND (converted_at IS NULL OR completed)
    ),
    CONSTRAINT story_checklist_order_unique UNIQUE (story_id, position) DEFERRABLE INITIALLY DEFERRED
);
CREATE INDEX story_checklist_converted_story ON public.story_checklist_items(converted_story_id)
    WHERE converted_story_id IS NOT NULL;
