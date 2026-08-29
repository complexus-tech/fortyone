DO $$
BEGIN
    RAISE EXCEPTION 'migration 000123 is forward-only because publication event identities and delivery state are immutable';
END
$$;
