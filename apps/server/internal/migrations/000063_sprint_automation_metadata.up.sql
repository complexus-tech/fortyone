-- 000063_sprint_automation_metadata.up.sql
ALTER TABLE public.team_sprint_settings
    ADD COLUMN next_auto_sprint_number int4,
    ADD COLUMN auto_create_disabled_at timestamptz,
    ADD COLUMN auto_create_disabled_reason text;

UPDATE public.team_sprint_settings
SET next_auto_sprint_number = last_auto_sprint_number + 1
WHERE next_auto_sprint_number IS NULL;

ALTER TABLE public.team_sprint_settings
    ALTER COLUMN next_auto_sprint_number SET DEFAULT 1,
    ALTER COLUMN next_auto_sprint_number SET NOT NULL,
    ADD CONSTRAINT team_sprint_settings_next_auto_sprint_number_check
        CHECK (next_auto_sprint_number >= 1 AND next_auto_sprint_number <= 10000);

CREATE TABLE public.audit_events (
    event_id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    team_id uuid,
    actor_type varchar(32) NOT NULL,
    actor_id uuid,
    entity_type varchar(64) NOT NULL,
    entity_id uuid,
    event_type varchar(128) NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT audit_events_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT audit_events_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE,
    CONSTRAINT audit_events_actor_id_fkey
        FOREIGN KEY (actor_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    PRIMARY KEY (event_id)
);

CREATE INDEX idx_audit_events_workspace_created
    ON public.audit_events USING btree (workspace_id, created_at DESC);

CREATE INDEX idx_audit_events_team_created
    ON public.audit_events USING btree (team_id, created_at DESC)
    WHERE team_id IS NOT NULL;

CREATE INDEX idx_audit_events_entity
    ON public.audit_events USING btree (entity_type, entity_id, created_at DESC)
    WHERE entity_id IS NOT NULL;

CREATE INDEX idx_audit_events_event_type_created
    ON public.audit_events USING btree (event_type, created_at DESC);
