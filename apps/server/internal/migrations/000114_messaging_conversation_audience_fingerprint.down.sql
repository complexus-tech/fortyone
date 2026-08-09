DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.messaging_conversations
        WHERE audience_scope = 'channel'
        GROUP BY provider,
                 workspace_id,
                 external_workspace_id,
                 external_channel_id,
                 external_thread_id
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION
            'cannot roll back audience-fingerprinted conversations while multiple audience histories exist for one provider thread';
    END IF;
END
$$;

DROP INDEX public.messaging_conversations_channel_key;

CREATE UNIQUE INDEX messaging_conversations_channel_key
    ON public.messaging_conversations USING btree (
        provider,
        workspace_id,
        external_workspace_id,
        external_channel_id,
        external_thread_id
    )
    WHERE audience_scope = 'channel';

ALTER TABLE public.messaging_conversations
    DROP CONSTRAINT messaging_conversations_audience_fingerprint_check,
    DROP COLUMN audience_fingerprint;
