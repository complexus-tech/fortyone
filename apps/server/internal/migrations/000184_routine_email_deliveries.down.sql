DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM routine_email_deliveries) THEN
        RAISE EXCEPTION 'Email delivery claims may cover sent messages; preserve them and repair forward.';
    END IF;
END $$;
DROP TABLE routine_email_deliveries;
