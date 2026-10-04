DO $$
BEGIN
    RAISE EXCEPTION 'Migration 000214 is forward-only: removing send-start fences can resend uncertain accepted emails. Preserve delivery facts and repair forward.';
END;
$$;
