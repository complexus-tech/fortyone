-- Only the reaction delivery intent is removed; uploaded files, messages,
-- reactions, existing notifications and realtime outbox events are retained.
DELETE FROM public.chat_outbox_recipients WHERE reason = 'reaction';
ALTER TABLE public.chat_outbox_recipients DROP CONSTRAINT chat_outbox_recipients_reason_check;
ALTER TABLE public.chat_outbox_recipients ADD CONSTRAINT chat_outbox_recipients_reason_check
    CHECK (reason IN ('dm', 'mention', 'reply', 'message'));

ALTER TABLE public.chat_outbox DROP COLUMN reaction_emoji;
