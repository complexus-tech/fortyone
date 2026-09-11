DO $$
BEGIN
    RAISE EXCEPTION 'Email receipts prevent repeated delivery; preserve them and repair forward.';
END $$;
