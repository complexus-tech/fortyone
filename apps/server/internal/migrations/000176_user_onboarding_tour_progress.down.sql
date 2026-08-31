-- Migration 000176 is forward-only once a user has stored onboarding state.
-- An empty table may be removed only before the new API is adopted.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.user_onboarding_tour_progress) THEN
        RAISE EXCEPTION 'migration 000176 is forward-only after onboarding progress exists'
            USING ERRCODE = '55000';
    END IF;
END $$;

DROP TABLE IF EXISTS public.user_onboarding_tour_progress;
