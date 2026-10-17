DO $$ BEGIN
    RAISE EXCEPTION 'Document comment agents contain durable requests and notifications; repair forward rather than discarding user data.';
END $$;
