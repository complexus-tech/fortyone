ALTER TABLE public.sprints
    ADD COLUMN schedule_managed_by_automation boolean NOT NULL DEFAULT false;

UPDATE public.sprints s
SET schedule_managed_by_automation = true
WHERE s.start_date > CURRENT_DATE
  AND EXISTS (
      SELECT 1
      FROM public.audit_events ae
      WHERE ae.entity_type = 'sprint'
        AND ae.entity_id = s.sprint_id
        AND ae.event_type = 'sprint.auto_created'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.audit_events ae
      WHERE ae.entity_type = 'sprint'
        AND ae.entity_id = s.sprint_id
        AND ae.event_type = 'sprint.updated'
        AND ae.actor_type = 'user'
        AND (
            COALESCE((ae.metadata ->> 'start_date_changed')::boolean, false)
            OR COALESCE((ae.metadata ->> 'end_date_changed')::boolean, false)
        )
  );

CREATE INDEX idx_sprints_automation_managed_schedule
    ON public.sprints (team_id, workspace_id, start_date)
    WHERE schedule_managed_by_automation = true;
