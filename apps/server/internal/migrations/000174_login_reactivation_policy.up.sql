ALTER TABLE public.users
    ADD COLUMN login_reactivation_policy text NOT NULL DEFAULT 'verified_sign_in';

-- Historical inactive accounts do not consistently record whether the last
-- deactivation was self-service, scheduled inactivity, or an administrator.
-- A latest admin deactivation is the strongest available provenance and is
-- retained as admin-only. Every other inactive legacy row fails closed into a
-- distinct review state instead of guessing that verified sign-in is safe.
UPDATE public.users AS account
SET login_reactivation_policy = CASE
    WHEN (
        SELECT audit.action
        FROM public.admin_audit_logs AS audit
        WHERE audit.target_type = 'user'
          AND audit.target_id = account.user_id
          AND audit.action IN ('user.activated', 'user.deactivated')
        ORDER BY audit.created_at DESC, audit.id DESC
        LIMIT 1
    ) = 'user.deactivated'
        THEN 'admin_only'
    ELSE 'legacy_admin_review'
END
WHERE account.is_active = FALSE;

ALTER TABLE public.users
    ADD CONSTRAINT users_login_reactivation_policy_check
    CHECK (
        login_reactivation_policy IN (
            'verified_sign_in',
            'admin_only',
            'legacy_admin_review'
        )
    );

COMMENT ON COLUMN public.users.login_reactivation_policy IS
    'Durable sign-in reactivation gate. Only verified_sign_in may reactivate through verified authentication; admin_only and legacy_admin_review require an explicit administrator enable action.';
