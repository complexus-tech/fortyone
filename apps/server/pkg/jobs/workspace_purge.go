package jobs

import (
	"context"
	"fmt"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/jmoiron/sqlx"
)

const workspacePurgeCandidatesTable = "workspace_purge_candidates"

// deleteWorkspacesWithSlackCleanup snapshots uninstall authority before the
// parent workspace and its Slack installation are removed. Workspaces with a
// legacy plaintext-only credential remain eligible for the next run after the
// worker credential backfill encrypts them.
func deleteWorkspacesWithSlackCleanup(ctx context.Context, db *sqlx.DB, candidateSelect string) (deleted, blocked int64, err error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("begin workspace purge: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, slackrepository.SlackInstallationLifecycleAdvisoryKey); err != nil {
		return 0, 0, fmt.Errorf("lock Slack lifecycle for workspace purge: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `
		CREATE TEMP TABLE `+workspacePurgeCandidatesTable+` ON COMMIT DROP AS
	`+candidateSelect); err != nil {
		return 0, 0, fmt.Errorf("select workspace purge candidates: %w", err)
	}
	var lockedWorkspaceIDs []string
	if err = tx.SelectContext(ctx, &lockedWorkspaceIDs, `
		SELECT CAST(w.workspace_id AS text)
		FROM workspaces w
		JOIN `+workspacePurgeCandidatesTable+` candidates ON candidates.workspace_id = w.workspace_id
		FOR UPDATE
	`); err != nil {
		return 0, 0, fmt.Errorf("lock workspace purge candidates: %w", err)
	}

	if err = tx.GetContext(ctx, &blocked, `
		SELECT COUNT(*)
		FROM `+workspacePurgeCandidatesTable+` candidates
		WHERE EXISTS (
			SELECT 1
			FROM slack_workspaces sw
			WHERE sw.workspace_id = candidates.workspace_id
			  AND sw.is_active = true
			  AND (
				sw.credential_key_version <= 0
				OR NULLIF(sw.credential_payload, '') IS NULL
			  )
		)
	`); err != nil {
		return 0, 0, fmt.Errorf("count workspaces awaiting Slack credential encryption: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO slack_uninstall_outbox (
			slack_workspace_id,
			workspace_id,
			installation_generation,
			slack_team_id,
			uninstall_kind,
			credential_payload,
			credential_key_version,
			status,
			next_attempt_at
		)
		SELECT sw.id,
		       sw.workspace_id,
		       sw.installation_generation,
		       sw.slack_team_id,
		       'workspace_delete',
		       sw.credential_payload,
		       sw.credential_key_version,
		       'pending',
		       NOW()
		FROM slack_workspaces sw
		JOIN `+workspacePurgeCandidatesTable+` candidates ON candidates.workspace_id = sw.workspace_id
		WHERE sw.is_active = true
		  AND sw.credential_key_version > 0
		  AND NULLIF(sw.credential_payload, '') IS NOT NULL
		ON CONFLICT (slack_workspace_id, installation_generation, uninstall_kind) DO UPDATE
		SET workspace_id = EXCLUDED.workspace_id,
		    slack_team_id = EXCLUDED.slack_team_id,
		    credential_payload = EXCLUDED.credential_payload,
		    credential_key_version = EXCLUDED.credential_key_version,
		    status = 'pending',
		    attempt_count = 0,
		    last_error = NULL,
		    next_attempt_at = NOW(),
		    processing_started_at = NULL,
		    completed_at = NULL,
		    updated_at = NOW()
		WHERE slack_uninstall_outbox.status = 'completed'
	`); err != nil {
		return 0, 0, fmt.Errorf("snapshot Slack uninstalls for workspace purge: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE messaging_inbound_events mie
		SET status = 'cancelled',
		    payload_encrypted = NULL,
		    last_error = 'FortyOne workspace deleted',
		    recovery_enqueued_at = NULL,
		    processed_at = NOW(),
		    updated_at = NOW()
		WHERE mie.provider = 'slack'
		  AND mie.status IN ('pending', 'processing', 'failed')
		  AND EXISTS (
			SELECT 1
			FROM slack_workspaces sw
			JOIN `+workspacePurgeCandidatesTable+` candidates ON candidates.workspace_id = sw.workspace_id
			WHERE sw.slack_team_id = mie.external_workspace_id
			  AND sw.is_active = true
			  AND sw.credential_key_version > 0
			  AND NULLIF(sw.credential_payload, '') IS NOT NULL
		  )
	`); err != nil {
		return 0, 0, fmt.Errorf("cancel Slack inbox work for workspace purge: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries mod
		SET status = 'cancelled',
		    content = NULL,
		    last_error = 'FortyOne workspace deleted',
		    updated_at = NOW()
		WHERE mod.provider = 'slack'
		  AND mod.status IN ('pending', 'delivering', 'failed')
		  AND EXISTS (
			SELECT 1
			FROM slack_workspaces sw
			JOIN `+workspacePurgeCandidatesTable+` candidates ON candidates.workspace_id = sw.workspace_id
			WHERE sw.slack_team_id = mod.external_workspace_id
			  AND sw.is_active = true
			  AND sw.credential_key_version > 0
			  AND NULLIF(sw.credential_payload, '') IS NOT NULL
		  )
	`); err != nil {
		return 0, 0, fmt.Errorf("cancel Slack outbound work for workspace purge: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		DELETE FROM workspaces w
		USING `+workspacePurgeCandidatesTable+` candidates
		WHERE w.workspace_id = candidates.workspace_id
		  AND NOT EXISTS (
			SELECT 1
			FROM slack_workspaces sw
			WHERE sw.workspace_id = w.workspace_id
			  AND sw.is_active = true
			  AND (
				sw.credential_key_version <= 0
				OR NULLIF(sw.credential_payload, '') IS NULL
			  )
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM slack_workspaces sw
			WHERE sw.workspace_id = w.workspace_id
			  AND sw.is_active = true
			  AND NOT EXISTS (
				SELECT 1
				FROM slack_uninstall_outbox suo
				WHERE suo.slack_workspace_id = sw.id
				  AND suo.installation_generation = sw.installation_generation
				  AND suo.uninstall_kind = 'workspace_delete'
				  AND suo.status <> 'completed'
				  AND NULLIF(suo.credential_payload, '') IS NOT NULL
			  )
		  )
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("delete workspace purge candidates: %w", err)
	}
	deleted, err = result.RowsAffected()
	if err != nil {
		return 0, 0, fmt.Errorf("count deleted workspace purge candidates: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit workspace purge: %w", err)
	}
	return deleted, blocked, nil
}
