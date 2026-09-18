DO $$ BEGIN
    RAISE EXCEPTION 'Migration 195 is forward-only: consolidated and expired document revisions cannot be reconstructed';
END $$;
