-- name: ListAutomationsForActor :many
SELECT automation.id, automation.team_id, automation.owner_id, automation.kind, automation.name, automation.configuration, automation.paused, automation.next_run_at, automation.last_run_at, automation.last_error, automation.created_at, CAST(COALESCE(EXISTS (SELECT 1 FROM public.workspace_members AS editor WHERE editor.workspace_id = automation.workspace_id AND editor.user_id = sqlc.arg(actor_id) AND editor.role IN ('admin', 'member') AND (editor.role = 'admin' OR automation.owner_id = editor.user_id)), FALSE) AS boolean) AS can_edit
FROM public.team_automations AS automation
WHERE automation.workspace_id = sqlc.arg(workspace_id) AND automation.team_id = sqlc.arg(team_id)
 AND automation.archived_at IS NULL AND EXISTS (
 SELECT 1 FROM public.workspace_members AS member
 JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
 JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
 JOIN public.teams AS team ON team.team_id = automation.team_id AND team.workspace_id = member.workspace_id
 WHERE member.workspace_id = automation.workspace_id AND member.user_id = sqlc.arg(actor_id)
 AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM public.team_members AS tm WHERE tm.team_id = team.team_id AND tm.user_id = member.user_id))
)
ORDER BY automation.created_at DESC, automation.id DESC LIMIT 100;

-- name: GetAutomationForActor :one
SELECT automation.id, automation.team_id, automation.owner_id, automation.kind, automation.name, automation.configuration, automation.paused, automation.next_run_at, automation.last_run_at, automation.last_error, automation.created_at, CAST(COALESCE(EXISTS (SELECT 1 FROM public.workspace_members AS editor WHERE editor.workspace_id = automation.workspace_id AND editor.user_id = sqlc.arg(actor_id) AND editor.role IN ('admin', 'member') AND (editor.role = 'admin' OR automation.owner_id = editor.user_id)), FALSE) AS boolean) AS can_edit
FROM public.team_automations AS automation
WHERE automation.workspace_id = sqlc.arg(workspace_id) AND automation.id = sqlc.arg(id)
 AND automation.archived_at IS NULL AND EXISTS (
 SELECT 1 FROM public.workspace_members AS member
 JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
 JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
 JOIN public.teams AS team ON team.team_id = automation.team_id AND team.workspace_id = member.workspace_id
 WHERE member.workspace_id = automation.workspace_id AND member.user_id = sqlc.arg(actor_id)
 AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM public.team_members AS tm WHERE tm.team_id = team.team_id AND tm.user_id = member.user_id))
)
LIMIT 1;


-- name: CreateAutomationForActor :one
INSERT INTO public.team_automations AS automation (workspace_id, team_id, owner_id, kind, name, configuration, next_run_at)
SELECT sqlc.arg(workspace_id), team.team_id, member.user_id, sqlc.arg(kind), sqlc.arg(name), sqlc.arg(configuration), sqlc.narg(next_run_at)
FROM public.workspace_members AS member
JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
JOIN public.teams AS team ON team.team_id = sqlc.arg(team_id) AND team.workspace_id = member.workspace_id
WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id)
 AND member.role IN ('member', 'admin')
 AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM public.team_members AS tm WHERE tm.team_id = team.team_id AND tm.user_id = member.user_id))
 AND (SELECT count(*) FROM public.team_automations AS existing WHERE existing.team_id = team.team_id AND existing.archived_at IS NULL) < 100
RETURNING automation.id, automation.team_id, automation.owner_id, automation.kind, automation.name, automation.configuration, automation.paused, automation.next_run_at, automation.last_run_at, automation.last_error, automation.created_at, TRUE AS can_edit;

-- name: PauseAutomationForActor :one
UPDATE public.team_automations AS automation
SET paused = sqlc.arg(paused), next_run_at = CASE WHEN automation.kind = 'recurrence' AND sqlc.arg(paused) = FALSE THEN CAST(sqlc.narg(next_run_at) AS timestamptz) ELSE automation.next_run_at END, updated_at = now(), lease_token = NULL, lease_until = NULL,
 event_at = CASE WHEN automation.paused = TRUE AND sqlc.arg(paused) = FALSE THEN now() ELSE automation.event_at END,
 event_id = CASE WHEN automation.paused = TRUE AND sqlc.arg(paused) = FALSE THEN '00000000-0000-0000-0000-000000000000' ELSE automation.event_id END
WHERE automation.id = sqlc.arg(id) AND automation.workspace_id = sqlc.arg(workspace_id)
 AND automation.archived_at IS NULL AND (automation.owner_id = sqlc.arg(actor_id) OR EXISTS (SELECT 1 FROM public.workspace_members AS admin WHERE admin.workspace_id = automation.workspace_id AND admin.user_id = sqlc.arg(actor_id) AND admin.role = 'admin')) AND EXISTS (
 SELECT 1 FROM public.workspace_members AS member
 JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
 JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
 JOIN public.teams AS team ON team.team_id = automation.team_id AND team.workspace_id = member.workspace_id
 WHERE member.workspace_id = automation.workspace_id AND member.user_id = sqlc.arg(actor_id)
 AND member.role IN ('member', 'admin') AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM public.team_members AS tm WHERE tm.team_id = team.team_id AND tm.user_id = member.user_id))
)
RETURNING automation.id, automation.team_id, automation.owner_id, automation.kind, automation.name, automation.configuration, automation.paused, automation.next_run_at, automation.last_run_at, automation.last_error, automation.created_at, TRUE AS can_edit;

-- name: ArchiveAutomationForActor :execrows
UPDATE public.team_automations AS automation SET archived_at = now(), updated_at = now(), lease_token = NULL, lease_until = NULL
WHERE automation.id = sqlc.arg(id) AND automation.workspace_id = sqlc.arg(workspace_id)
 AND automation.archived_at IS NULL AND (automation.owner_id = sqlc.arg(actor_id) OR EXISTS (SELECT 1 FROM public.workspace_members AS admin WHERE admin.workspace_id = automation.workspace_id AND admin.user_id = sqlc.arg(actor_id) AND admin.role = 'admin')) AND EXISTS (
 SELECT 1 FROM public.workspace_members AS member
 JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
 JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
 JOIN public.teams AS team ON team.team_id = automation.team_id AND team.workspace_id = member.workspace_id
 WHERE member.workspace_id = automation.workspace_id AND member.user_id = sqlc.arg(actor_id)
 AND member.role IN ('member', 'admin') AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM public.team_members AS tm WHERE tm.team_id = team.team_id AND tm.user_id = member.user_id))
);

-- name: ListAutomationRunsForActor :many
SELECT run.id, run.automation_id, run.occurrence, run.status, run.story_id, run.error, run.started_at, run.finished_at
FROM public.team_automation_runs AS run
JOIN public.team_automations AS automation ON automation.id = run.automation_id
WHERE automation.id = sqlc.arg(id) AND automation.workspace_id = sqlc.arg(workspace_id) AND EXISTS (
 SELECT 1 FROM public.workspace_members AS member
 JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
 JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
 JOIN public.teams AS team ON team.team_id = automation.team_id AND team.workspace_id = member.workspace_id
 WHERE member.workspace_id = automation.workspace_id AND member.user_id = sqlc.arg(actor_id)
 AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM public.team_members AS tm WHERE tm.team_id = team.team_id AND tm.user_id = member.user_id))
)
ORDER BY run.started_at DESC, run.id DESC LIMIT 50;

-- name: ClaimAutomation :one
WITH candidate AS (
 SELECT automation.id FROM public.team_automations AS automation
 WHERE automation.archived_at IS NULL AND automation.paused = FALSE
 AND (automation.lease_until IS NULL OR automation.lease_until < now())
 AND (automation.kind = 'rule' OR automation.next_run_at <= now())
 ORDER BY automation.updated_at, automation.id LIMIT 1 FOR UPDATE SKIP LOCKED
)
UPDATE public.team_automations AS automation
SET lease_token = sqlc.arg(lease_token), lease_until = now() + interval '2 minutes'
FROM candidate WHERE automation.id = candidate.id
RETURNING automation.*;

-- name: CurrentAutomationOwnerCanWrite :one
SELECT CAST(COALESCE(EXISTS (
 SELECT 1 FROM public.workspace_members AS member
 JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
 JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
 JOIN public.teams AS team ON team.team_id = automation.team_id AND team.workspace_id = member.workspace_id
 WHERE member.workspace_id = automation.workspace_id AND member.user_id = automation.owner_id
 AND member.role IN ('member', 'admin') AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM public.team_members AS tm WHERE tm.team_id = team.team_id AND tm.user_id = member.user_id))
), FALSE) AS boolean) AS allowed
FROM public.team_automations AS automation
WHERE automation.id = sqlc.arg(id) AND automation.lease_token = sqlc.arg(lease_token)
 AND automation.lease_until > now() AND automation.paused = FALSE AND automation.archived_at IS NULL;

-- name: ListAutomationEvents :many
SELECT event.event_id, event.subject_id, event.event_type, event.created_at
FROM public.outbound_webhook_events AS event
JOIN public.team_automations AS automation ON automation.workspace_id = event.workspace_id
JOIN public.stories AS story ON story.id = event.subject_id AND story.workspace_id = automation.workspace_id AND story.team_id = automation.team_id
WHERE automation.id = sqlc.arg(id) AND automation.lease_token = sqlc.arg(lease_token)
 AND automation.lease_until > now() AND event.subject_type = 'story'
 AND event.event_type = automation.configuration ->> 'trigger' AND event.actor_kind <> 'system'
 AND event.created_at >= automation.event_at
 AND NOT EXISTS (SELECT 1 FROM public.team_automation_runs AS existing WHERE existing.automation_id = automation.id AND existing.occurrence = 'event:' || CAST(event.event_id AS text) AND existing.status <> 'running')
ORDER BY event.created_at, event.event_id LIMIT 50;

-- name: ClaimAutomationRun :one
INSERT INTO public.team_automation_runs AS run (automation_id, occurrence, lease_token, status)
SELECT automation.id, sqlc.arg(occurrence), automation.lease_token, 'running'
FROM public.team_automations AS automation WHERE automation.id = sqlc.arg(id)
 AND automation.lease_token = sqlc.arg(lease_token) AND automation.lease_until > now()
 AND automation.paused = FALSE AND automation.archived_at IS NULL
ON CONFLICT (automation_id, occurrence) DO UPDATE SET lease_token = EXCLUDED.lease_token, attempt_count = run.attempt_count + 1
WHERE run.status = 'running' AND run.lease_token <> EXCLUDED.lease_token AND run.attempt_count < 3
RETURNING run.id;

-- name: CompleteAutomationRun :execrows
UPDATE public.team_automation_runs AS run SET status = sqlc.arg(status), story_id = sqlc.narg(story_id), error = sqlc.arg(error), finished_at = now()
WHERE run.id = sqlc.arg(id) AND run.lease_token = sqlc.arg(lease_token) AND run.status = 'running';

-- name: AdvanceAutomationEvent :execrows
UPDATE public.team_automations SET last_run_at = now(), last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id) AND lease_token = sqlc.arg(lease_token) AND lease_until > now();

-- name: ReleaseAutomation :execrows
UPDATE public.team_automations SET lease_token = NULL, lease_until = NULL, updated_at = now(),
 next_run_at = CASE WHEN kind = 'recurrence' THEN CAST(sqlc.narg(next_run_at) AS timestamptz) ELSE next_run_at END,
 last_error = sqlc.arg(last_error), last_run_at = CASE WHEN kind = 'recurrence' THEN now() ELSE last_run_at END
WHERE id = sqlc.arg(id) AND lease_token = sqlc.arg(lease_token);

-- name: FinishExhaustedAutomationRuns :execrows
UPDATE public.team_automation_runs AS run SET status = 'failed', error = 'Task action exhausted its recovery attempts.', finished_at = now()
FROM public.team_automations AS automation
WHERE automation.id = sqlc.arg(id) AND automation.lease_token = sqlc.arg(lease_token)
 AND run.automation_id = automation.id AND run.status = 'running' AND run.attempt_count = 3 AND run.lease_token <> automation.lease_token;
