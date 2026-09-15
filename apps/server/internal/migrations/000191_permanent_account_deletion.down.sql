-- Account erasure is irreversible. Never restore deleted identities or remove
-- pending cleanup requests during rollback. Deploy a forward repair instead.
DO $$ BEGIN
    RAISE EXCEPTION 'Permanent account deletion cannot be rolled back; apply a forward repair';
END $$;
