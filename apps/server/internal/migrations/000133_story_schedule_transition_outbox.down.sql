DO $$
BEGIN
    RAISE EXCEPTION 'migration 000133 is forward-only because schedule transition event snapshots and delivery state are immutable';
END
$$;
