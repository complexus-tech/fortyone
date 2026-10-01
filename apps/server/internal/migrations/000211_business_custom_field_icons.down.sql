-- Expanded keys may already appear in definitions and immutable audit records.
DO $$
BEGIN
    RAISE EXCEPTION 'Migration 000211 is forward-only; preserve selected icons and repair forward';
END;
$$;
