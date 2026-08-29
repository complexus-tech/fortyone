ALTER TABLE public.messaging_conversations
    ADD COLUMN audience_fingerprint text NOT NULL DEFAULT '';

UPDATE public.messaging_conversations
SET audience_fingerprint = 'legacy:' || id::text
WHERE audience_scope = 'channel';

ALTER TABLE public.messaging_conversations
    ADD CONSTRAINT messaging_conversations_audience_fingerprint_check
    CHECK (
        (audience_scope = 'actor' AND audience_fingerprint = '')
        OR (
            audience_scope = 'channel'
            AND btrim(audience_fingerprint) <> ''
            AND char_length(audience_fingerprint) <= 128
        )
    );

DROP INDEX public.messaging_conversations_channel_key;

CREATE UNIQUE INDEX messaging_conversations_channel_key
    ON public.messaging_conversations USING btree (
        provider,
        workspace_id,
        external_workspace_id,
        external_channel_id,
        external_thread_id,
        audience_fingerprint
    )
    WHERE audience_scope = 'channel';
