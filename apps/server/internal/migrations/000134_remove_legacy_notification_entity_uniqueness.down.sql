DO $$
BEGIN
    RAISE EXCEPTION 'migration 000134 is forward-only because restoring legacy notification uniqueness would break distinct strategy communications';
END
$$;

