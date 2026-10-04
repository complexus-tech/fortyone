DO $$
BEGIN
    RAISE EXCEPTION 'Migration 000212 is forward-only: notification receipts preserve discarded inbox payloads and replay identities. Preserve this history and repair forward.';
END;
$$;
