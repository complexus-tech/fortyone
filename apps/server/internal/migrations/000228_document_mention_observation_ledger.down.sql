DO $$ BEGIN
    RAISE EXCEPTION 'Document mention observations prevent duplicate notifications; repair forward instead of discarding the ledger.';
END $$;
