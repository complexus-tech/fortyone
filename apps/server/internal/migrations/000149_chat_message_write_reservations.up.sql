ALTER TABLE public.chat_messages
    ADD COLUMN write_generation BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN write_token UUID,
    ADD COLUMN write_operation TEXT,
    ADD COLUMN write_finalized_at TIMESTAMPTZ;

ALTER TABLE public.chat_messages
    ADD CONSTRAINT chat_messages_write_generation_nonnegative
        CHECK (write_generation >= 0),
    ADD CONSTRAINT chat_messages_write_reservation_coherent
        CHECK (
            (write_token IS NULL AND write_operation IS NULL AND write_finalized_at IS NULL)
            OR (
                write_token IS NOT NULL
                AND write_operation IN ('append', 'regenerate', 'approval')
            )
        );
