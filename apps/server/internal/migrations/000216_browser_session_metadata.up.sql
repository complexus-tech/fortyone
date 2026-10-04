ALTER TABLE public.workspace_browser_sessions
    ADD COLUMN browser_name varchar(64),
    ADD CONSTRAINT workspace_browser_sessions_browser_name_valid
        CHECK (browser_name IS NULL OR (browser_name = btrim(browser_name) AND length(browser_name) > 0));

COMMENT ON COLUMN public.workspace_browser_sessions.browser_name IS
    'Reported browser brand captured at session issuance; null when unavailable. Display metadata only, never authentication evidence.';
