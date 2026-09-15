-- Shared attribution contains no identity of any deleted account. Keeping an
-- inert actor avoids destructive cascades or nullable-author reader regressions
-- in organization-owned documents and historical records.
INSERT INTO public.users (user_id, username, email, full_name, is_active, is_system,
    login_reactivation_policy, timezone)
VALUES ('ffffffff-ffff-4fff-8fff-ffffffffffff', 'deleted-user',
    'deleted-user@accounts.invalid', 'Deleted user', FALSE, TRUE, 'admin_only', 'UTC');

-- Only an anonymous account key remains while the existing calendar provider
-- dispatcher drains remote event deletion. No email, profile, content or token
-- is kept in this queue; calendar credentials stay in their sealed repository.
CREATE TABLE public.account_deletion_requests (
    user_id uuid PRIMARY KEY REFERENCES public.users(user_id) ON DELETE CASCADE,
    requested_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0)
);

CREATE INDEX account_deletion_requests_pending
    ON public.account_deletion_requests (updated_at, user_id);

COMMENT ON TABLE public.account_deletion_requests IS
    'Durable finalization queue for irreversibly anonymized accounts awaiting provider cleanup.';

CREATE FUNCTION public.prevent_deleted_account_reactivation()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.user_id = 'ffffffff-ffff-4fff-8fff-ffffffffffff'
       OR EXISTS (SELECT 1 FROM public.account_deletion_requests WHERE user_id = OLD.user_id) THEN
        RAISE EXCEPTION 'Deleted accounts cannot be changed or reactivated';
    END IF;
    RETURN NEW;
END $$;

CREATE TRIGGER users_prevent_deleted_account_reactivation
BEFORE UPDATE ON public.users FOR EACH ROW
EXECUTE FUNCTION public.prevent_deleted_account_reactivation();
