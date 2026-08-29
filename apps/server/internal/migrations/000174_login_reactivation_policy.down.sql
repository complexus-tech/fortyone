ALTER TABLE public.users
    DROP CONSTRAINT users_login_reactivation_policy_check,
    DROP COLUMN login_reactivation_policy;
