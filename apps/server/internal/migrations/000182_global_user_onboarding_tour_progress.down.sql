-- Migration 000182 is forward-only after global onboarding state exists.
-- Once the replacement API writes the global table, the workspace-scoped
-- source table is no longer a complete recovery source.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.user_onboarding_tour_progress_global) THEN
        RAISE EXCEPTION 'migration 000182 is forward-only after global onboarding progress exists'
            USING ERRCODE = '55000';
    END IF;
END $$;

DROP TRIGGER IF EXISTS mirror_user_onboarding_tour_progress_global
    ON public.user_onboarding_tour_progress;
DROP FUNCTION IF EXISTS public.mirror_user_onboarding_tour_progress_global();
DROP TABLE IF EXISTS public.user_onboarding_tour_progress_global;
