DO $$
BEGIN
    RAISE EXCEPTION 'shared account-deletion history retention is forward-only; erased identities cannot be reconstructed';
END $$;
