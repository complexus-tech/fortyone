DO $$ BEGIN RAISE EXCEPTION 'Migration 229 is forward-only; preserve chat history and repair forward.'; END $$;
