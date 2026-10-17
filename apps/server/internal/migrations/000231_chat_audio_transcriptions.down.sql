DO $$ BEGIN
    RAISE EXCEPTION 'Voice notes and requested transcripts may be in use; preserve recordings and transcripts and repair forward.';
END $$;
