DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM email_avatar_handles) THEN
        RAISE EXCEPTION 'Email avatar handles may already be in delivered emails; preserve them and repair forward.';
    END IF;
END $$;
DROP TABLE email_avatar_handles;
