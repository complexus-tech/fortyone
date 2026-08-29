DROP TABLE IF EXISTS public.slack_agent_settings;
DROP TABLE IF EXISTS public.slack_channel_team_access;

ALTER TABLE public.messaging_outbound_deliveries
    DROP COLUMN IF EXISTS provider_payload;

DELETE FROM public.messaging_conversations actor_conversation
USING public.messaging_conversations channel_conversation
WHERE actor_conversation.audience_scope = 'actor'
  AND channel_conversation.audience_scope = 'channel'
  AND actor_conversation.provider = channel_conversation.provider
  AND actor_conversation.workspace_id = channel_conversation.workspace_id
  AND actor_conversation.external_workspace_id = channel_conversation.external_workspace_id
  AND actor_conversation.external_channel_id = channel_conversation.external_channel_id
  AND actor_conversation.external_thread_id = channel_conversation.external_thread_id
  AND actor_conversation.user_id = channel_conversation.user_id;

DROP INDEX IF EXISTS public.messaging_conversations_channel_key;
DROP INDEX IF EXISTS public.messaging_conversations_actor_key;

ALTER TABLE public.messaging_conversations
    DROP CONSTRAINT IF EXISTS messaging_conversations_audience_scope_check;

ALTER TABLE public.messaging_conversations
    DROP COLUMN audience_scope;

CREATE UNIQUE INDEX messaging_conversations_external_key
    ON public.messaging_conversations USING btree (
        provider,
        workspace_id,
        external_workspace_id,
        external_channel_id,
        external_thread_id,
        user_id
    );
