-- A begun send is uncertain until its delivery completion is durable. Keep
-- that day closed even if the worker exits after SMTP accepts the message.
ALTER TABLE public.routine_email_deliveries
    ADD COLUMN send_started_at timestamptz;

COMMENT ON COLUMN public.routine_email_deliveries.send_started_at IS
    'Durable send-start fence; processing attempts with this marker are never automatically reclaimed.';
