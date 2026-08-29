DROP TRIGGER IF EXISTS workspace_members_reanchor_account_calendar_connection ON public.workspace_members;
DROP FUNCTION IF EXISTS public.reanchor_account_calendar_connection();

DROP TRIGGER IF EXISTS team_members_cleanup_maya_schedule ON public.team_members;
DROP FUNCTION IF EXISTS public.cleanup_removed_team_member_maya_schedule();

DROP TRIGGER IF EXISTS stories_cleanup_retired_maya_schedule ON public.stories;
DROP FUNCTION IF EXISTS public.cleanup_retired_story_maya_schedule();

DROP TRIGGER IF EXISTS calendar_schedule_blocks_enqueue_provider_delete ON public.calendar_schedule_blocks;
DROP FUNCTION IF EXISTS public.enqueue_deleted_maya_schedule_event();

DROP TABLE IF EXISTS public.calendar_schedule_event_outbox;

DROP INDEX IF EXISTS public.idx_maya_agent_runs_schedule_recovery;

DROP TABLE IF EXISTS public.calendar_maya_schedule_ownerships;

DROP INDEX IF EXISTS public.uq_calendar_maya_story_segment;

ALTER TABLE public.calendar_schedule_blocks
    DROP CONSTRAINT IF EXISTS calendar_schedule_blocks_external_mapping_check,
    DROP CONSTRAINT IF EXISTS calendar_schedule_blocks_segment_index_check,
    DROP COLUMN IF EXISTS external_synced_at,
    DROP COLUMN IF EXISTS external_sync_hash,
    DROP COLUMN IF EXISTS external_event_id,
    DROP COLUMN IF EXISTS external_calendar_id,
    DROP COLUMN IF EXISTS external_provider,
    DROP COLUMN IF EXISTS segment_index;

-- Every connection maintained by the forward migration has a valid anchor.
-- Remove any manually-created invalid rows before restoring the legacy
-- workspace-owned constraints so rollback cannot fail partway through.
DELETE FROM public.calendar_connections
WHERE cleanup_pending_at IS NOT NULL;

DELETE FROM public.calendar_connections connection
WHERE NOT EXISTS (
    SELECT 1
    FROM public.workspace_members member
    WHERE member.workspace_id = connection.workspace_id
      AND member.user_id = connection.user_id
);

ALTER TABLE public.calendar_events
    DROP CONSTRAINT calendar_events_connection_scope_fkey,
    ADD CONSTRAINT calendar_events_connection_scope_fkey
        FOREIGN KEY (connection_id, workspace_id, user_id)
        REFERENCES public.calendar_connections (connection_id, workspace_id, user_id)
        ON DELETE CASCADE;

ALTER TABLE public.calendar_connections
    DROP COLUMN cleanup_pending_at,
    ADD CONSTRAINT calendar_connections_workspace_id_fkey
        FOREIGN KEY (workspace_id)
        REFERENCES public.workspaces(workspace_id)
        ON DELETE CASCADE,
    ADD CONSTRAINT calendar_connections_workspace_member_fkey
        FOREIGN KEY (workspace_id, user_id)
        REFERENCES public.workspace_members(workspace_id, user_id)
        ON DELETE CASCADE;
