-- Existing definitions and immutable audit records can contain expanded icon keys.
-- Keep the wider catalog during rollback; reverting it requires an explicit data plan.
DO $$
BEGIN
    RAISE EXCEPTION 'Migration 000210 is forward-only; preserve selected icons and repair forward';
END;
$$;
