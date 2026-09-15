DO $$
BEGIN
    RAISE EXCEPTION 'account subscriber deletion obligations are forward-only; repair the dispatcher without discarding pending cleanup';
END $$;
