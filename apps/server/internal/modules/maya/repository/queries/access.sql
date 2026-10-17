-- name: WorkspaceCanUseMaya :one
SELECT public.workspace_can_use_maya(sqlc.arg(workspace_id)) AS allowed;
