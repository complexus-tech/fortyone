-- 000049_chat_messages.up.sql
CREATE TABLE public.chat_messages (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    session_id text NOT NULL,
    messages jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chat_messages_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.chat_sessions(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

-- Indices
CREATE UNIQUE INDEX chat_messages_session_id_key ON public.chat_messages USING btree (session_id);
CREATE INDEX idx_chat_messages_session ON public.chat_messages USING btree (session_id);
