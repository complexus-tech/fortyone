CREATE TABLE email_avatar_handles (
    handle uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL UNIQUE REFERENCES users(user_id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE email_avatar_handles IS
    'Non-expiring opaque email image handles. Resolve only the current profile image; never arbitrary storage keys.';
