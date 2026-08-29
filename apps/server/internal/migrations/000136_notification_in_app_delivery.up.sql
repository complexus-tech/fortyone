-- Keep delivery-only notifications available to the email worker without
-- surfacing them in the workspace inbox.
ALTER TABLE public.notifications
    ADD COLUMN IF NOT EXISTS in_app_enabled boolean NOT NULL DEFAULT true;

-- Strategy guidance is an email-first, evidence-rich communication. Existing
-- generic strategy inbox rows should no longer compete with actionable work.
UPDATE public.notifications
SET in_app_enabled = false
WHERE CAST(type AS text) = 'strategy_update';

CREATE INDEX IF NOT EXISTS idx_notifications_in_app_recipient_workspace_created
    ON public.notifications (recipient_id, workspace_id, created_at DESC)
    WHERE in_app_enabled = true;
