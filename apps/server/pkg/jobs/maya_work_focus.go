package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	mayaWorkFocusMemberBatchSize = 500
	mayaWorkFocusEvidenceLimit   = 30
	mayaWorkFocusLookback        = 180 * 24 * time.Hour
)

type mayaWorkFocusMember struct {
	WorkspaceID           uuid.UUID `db:"workspace_id"`
	TeamID                uuid.UUID `db:"team_id"`
	UserID                uuid.UUID `db:"user_id"`
	ManualRoleTitle       string    `db:"ai_role_title"`
	ManualRoleDescription string    `db:"ai_role_description"`
}

type mayaWorkFocusEvidenceRow struct {
	Title       string         `db:"title"`
	Description sql.NullString `db:"description"`
	Labels      pq.StringArray `db:"labels"`
}

func ProcessMayaWorkFocusInference(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessMayaWorkFocusInference")
	defer span.End()

	members, err := listMayaWorkFocusCandidates(ctx, db)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		if log != nil {
			log.Info(ctx, "Maya work focus inference skipped: no candidates")
		}
		return nil
	}

	inferredCount := 0
	for _, member := range members {
		evidence, err := listMayaWorkFocusEvidence(ctx, db, member)
		if err != nil {
			if log != nil {
				log.Error(ctx, "failed to load Maya work focus evidence", "workspace_id", member.WorkspaceID, "team_id", member.TeamID, "user_id", member.UserID, "error", err)
			}
			continue
		}
		result := maya.InferWorkFocus(maya.WorkFocusInferenceInput{
			ManualRoleTitle:       member.ManualRoleTitle,
			ManualRoleDescription: member.ManualRoleDescription,
			Evidence:              evidence,
		})
		if !result.ShouldInfer {
			continue
		}
		updated, err := saveMayaInferredWorkFocus(ctx, db, member, result)
		if err != nil {
			if log != nil {
				log.Error(ctx, "failed to save Maya inferred work focus", "workspace_id", member.WorkspaceID, "team_id", member.TeamID, "user_id", member.UserID, "error", err)
			}
			continue
		}
		if updated {
			inferredCount++
		}
	}

	if log != nil {
		log.Info(ctx, "Maya work focus inference completed", "candidate_count", len(members), "inferred_count", inferredCount)
	}
	return nil
}

func listMayaWorkFocusCandidates(ctx context.Context, db *sqlx.DB) ([]mayaWorkFocusMember, error) {
	query := `
		SELECT
			t.workspace_id,
			tm.team_id,
			tm.user_id,
			tm.ai_role_title,
			tm.ai_role_description
		FROM team_members tm
		INNER JOIN teams t ON t.team_id = tm.team_id
		INNER JOIN users u ON u.user_id = tm.user_id
		WHERE u.is_active = TRUE
			AND u.is_system = FALSE
			AND tm.ai_role_title = ''
			AND tm.ai_role_description = ''
			AND EXISTS (
				SELECT 1
				FROM workspace_subscriptions ws
				WHERE ws.workspace_id = t.workspace_id
					AND ws.subscription_tier <> 'free'
					AND ws.subscription_status IN ('active', 'trialing', 'past_due')
				LIMIT 1
			)
		ORDER BY tm.inferred_ai_role_generated_at ASC NULLS FIRST, tm.updated_at ASC
		LIMIT $1
	`

	var members []mayaWorkFocusMember
	if err := db.SelectContext(ctx, &members, query, mayaWorkFocusMemberBatchSize); err != nil {
		return nil, fmt.Errorf("list Maya work focus candidates: %w", err)
	}
	return members, nil
}

func listMayaWorkFocusEvidence(ctx context.Context, db *sqlx.DB, member mayaWorkFocusMember) ([]maya.WorkFocusEvidence, error) {
	query := `
		SELECT
			s.title,
			s.description,
			COALESCE(array_agg(l.name ORDER BY l.name) FILTER (WHERE l.name IS NOT NULL), '{}') AS labels
		FROM stories s
		LEFT JOIN story_labels sl ON sl.story_id = s.id
		LEFT JOIN labels l ON l.label_id = sl.label_id
		WHERE s.workspace_id = $1
			AND s.team_id = $2
			AND s.assignee_id = $3
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
			AND s.is_draft = FALSE
			AND s.updated_at >= $4
		GROUP BY s.id, s.title, s.description, s.updated_at
		ORDER BY s.updated_at DESC
		LIMIT $5
	`

	var rows []mayaWorkFocusEvidenceRow
	if err := db.SelectContext(ctx, &rows, query, member.WorkspaceID, member.TeamID, member.UserID, time.Now().UTC().Add(-mayaWorkFocusLookback), mayaWorkFocusEvidenceLimit); err != nil {
		return nil, fmt.Errorf("list Maya work focus evidence: %w", err)
	}

	evidence := make([]maya.WorkFocusEvidence, len(rows))
	for i, row := range rows {
		evidence[i] = maya.WorkFocusEvidence{
			Title:       row.Title,
			Description: row.Description.String,
			Labels:      []string(row.Labels),
		}
	}
	return evidence, nil
}

func saveMayaInferredWorkFocus(ctx context.Context, db *sqlx.DB, member mayaWorkFocusMember, result maya.WorkFocusInferenceResult) (bool, error) {
	query := `
		UPDATE team_members tm
		SET
			inferred_ai_role_title = $1,
			inferred_ai_role_description = $2,
			inferred_ai_role_story_count = $3,
			inferred_ai_role_confidence = $4,
			inferred_ai_role_generated_at = NOW(),
			updated_at = NOW()
		FROM teams t
		WHERE tm.team_id = t.team_id
			AND tm.team_id = $5
			AND tm.user_id = $6
			AND t.workspace_id = $7
			AND tm.ai_role_title = ''
			AND tm.ai_role_description = ''
	`

	res, err := db.ExecContext(
		ctx,
		query,
		result.RoleTitle,
		result.RoleDescription,
		result.StoryCount,
		result.Confidence,
		member.TeamID,
		member.UserID,
		member.WorkspaceID,
	)
	if err != nil {
		return false, fmt.Errorf("save Maya inferred work focus: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read Maya inferred work focus rows affected: %w", err)
	}
	return rows > 0, nil
}
