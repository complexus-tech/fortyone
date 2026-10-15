DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM desktop_operation_receipts) THEN
        RAISE EXCEPTION 'desktop operation receipts cannot be removed while replayable operations exist';
    END IF;
END $$;
DROP TABLE desktop_operation_receipts;
