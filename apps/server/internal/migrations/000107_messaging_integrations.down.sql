-- SQL cannot decrypt credentials written by the application. Fail closed
-- instead of silently producing unusable tokens or disconnecting customers.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.slack_workspaces
        WHERE credential_key_version > 0
          AND NULLIF(credential_payload, '') IS NOT NULL
    ) OR EXISTS (
        SELECT 1
        FROM public.slack_uninstall_outbox
        WHERE status <> 'completed'
    ) THEN
        RAISE EXCEPTION 'migration 000107 cannot be rolled back while active encrypted Slack credentials or unfinished uninstall records exist';
    END IF;
END
$$;

DROP TABLE IF EXISTS public.slack_uninstall_outbox;
DROP TABLE IF EXISTS public.messaging_outbound_deliveries;
DROP TABLE IF EXISTS public.messaging_messages;
DROP TABLE IF EXISTS public.messaging_conversations;
DROP TABLE IF EXISTS public.messaging_inbound_events;
DROP TABLE IF EXISTS public.messaging_nonces;

DROP INDEX IF EXISTS public.story_links_external_source_key_unique;

ALTER TABLE public.story_links
    DROP COLUMN IF EXISTS external_source_key;

ALTER TABLE public.stories
    DROP COLUMN IF EXISTS external_creation_key;

UPDATE public.slack_workspaces
SET bot_access_token = COALESCE(NULLIF(bot_access_token, ''), credential_payload, '')
WHERE credential_payload IS NOT NULL
  AND credential_key_version = 0;

ALTER TABLE public.slack_workspaces
    DROP COLUMN IF EXISTS revoked_at,
    DROP COLUMN IF EXISTS authed_user_id,
    DROP COLUMN IF EXISTS enterprise_id,
    DROP COLUMN IF EXISTS slack_app_id,
    DROP COLUMN IF EXISTS installation_authorized_at,
    DROP COLUMN IF EXISTS installation_generation,
    DROP COLUMN IF EXISTS credential_key_version,
    DROP COLUMN IF EXISTS credential_payload;

DROP INDEX IF EXISTS public.slack_workspaces_slack_team_id_key;

CREATE UNIQUE INDEX slack_workspaces_slack_team_id_key
    ON public.slack_workspaces USING btree (slack_team_id);
