DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.maya_skills) THEN
        RAISE EXCEPTION 'Maya skills contain saved instructions; preserve them and repair forward';
    END IF;
END $$;

DROP TABLE public.maya_skills;
