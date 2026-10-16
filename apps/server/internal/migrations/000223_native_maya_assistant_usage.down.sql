DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.agent_invocation_assistant_usage_events) THEN
        RAISE EXCEPTION 'Native assistant usage receipts remain; preserve them and repair forward';
    END IF;
END $$;

DROP TABLE public.agent_invocation_assistant_usage_events;
