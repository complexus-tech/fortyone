DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.agent_invocations) THEN
        RAISE EXCEPTION 'agent invocations contain durable work; preserve receipts and repair forward';
    END IF;
END $$;
DROP TABLE public.agent_invocations;
