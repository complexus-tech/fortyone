DO $$
BEGIN
    RAISE EXCEPTION 'Internal Slack receipts prevent duplicate alerts; disable dispatch and repair forward.';
END $$;
