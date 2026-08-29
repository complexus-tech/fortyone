DO $$
BEGIN
    RAISE EXCEPTION
        'migration 000130 is forward-only: legacy time estimates were moved out of estimate_unit and time-based team schemes were collapsed to tshirt';
END
$$;
