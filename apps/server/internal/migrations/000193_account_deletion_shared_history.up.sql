-- The attribution row is not a former person's account. Its guard remains in
-- force after this atomic, table-locked display-name correction.
BEGIN;
ALTER TABLE public.users DISABLE TRIGGER users_prevent_deleted_account_reactivation;
UPDATE public.users SET full_name = 'Former user', updated_at = CURRENT_TIMESTAMP
WHERE user_id = 'ffffffff-ffff-4fff-8fff-ffffffffffff' AND is_system AND NOT is_active;
ALTER TABLE public.users ENABLE TRIGGER users_prevent_deleted_account_reactivation;
COMMIT;
