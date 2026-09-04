DROP TRIGGER IF EXISTS users_cleanup_google_drive_on_deactivation
    ON public.users;
DROP FUNCTION IF EXISTS public.cleanup_google_drive_on_user_deactivation();
