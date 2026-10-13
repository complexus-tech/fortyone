DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.story_checklist_items) THEN
        RAISE EXCEPTION 'Story checklists contain user data; preserve it and repair forward';
    END IF;
END $$;
DROP TABLE public.story_checklist_items;
