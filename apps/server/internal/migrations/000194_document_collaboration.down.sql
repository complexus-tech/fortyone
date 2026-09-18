DO $$ BEGIN
    RAISE EXCEPTION 'Document collaboration retains user history. Recover with a forward migration; do not discard revisions or collaboration state.';
END $$;
