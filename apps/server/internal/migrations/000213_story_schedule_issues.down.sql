DO $$
BEGIN
    RAISE EXCEPTION 'Migration 000213 is forward-only: durable scheduling issue identities prevent repeated and stale alerts. Preserve issue lifecycle facts and repair forward.';
END;
$$;
