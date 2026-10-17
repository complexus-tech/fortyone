DO $$ BEGIN
    RAISE EXCEPTION 'Chat notification identities may be in use; preserve notifications and repair forward.';
END $$;
