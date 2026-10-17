-- Snapshot additions separately from removals so delayed/replayed events cannot
-- turn a removed reaction or a pre-migration event into an author notification.
ALTER TABLE public.chat_outbox ADD COLUMN reaction_emoji text;
ALTER TABLE public.chat_outbox ADD CONSTRAINT chat_outbox_reaction_emoji_check
    CHECK (reaction_emoji IS NULL OR (event_type = 'reaction.updated' AND length(reaction_emoji) > 0));

ALTER TABLE public.chat_outbox_recipients DROP CONSTRAINT chat_outbox_recipients_reason_check;
ALTER TABLE public.chat_outbox_recipients ADD CONSTRAINT chat_outbox_recipients_reason_check
    CHECK (reason IN ('dm', 'mention', 'reply', 'message', 'reaction'));
