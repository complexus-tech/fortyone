ALTER TABLE public.chat_messages
    DROP CONSTRAINT IF EXISTS chat_messages_write_reservation_coherent,
    DROP CONSTRAINT IF EXISTS chat_messages_write_generation_nonnegative,
    DROP COLUMN IF EXISTS write_finalized_at,
    DROP COLUMN IF EXISTS write_operation,
    DROP COLUMN IF EXISTS write_token,
    DROP COLUMN IF EXISTS write_generation;
