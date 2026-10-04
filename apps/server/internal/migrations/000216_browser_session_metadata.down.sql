DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.workspace_browser_sessions WHERE browser_name IS NOT NULL) THEN
        RAISE EXCEPTION 'Browser session metadata has been recorded; preserve it and repair forward';
    END IF;
END $$;

ALTER TABLE public.workspace_browser_sessions
    DROP CONSTRAINT workspace_browser_sessions_browser_name_valid,
    DROP COLUMN browser_name;
