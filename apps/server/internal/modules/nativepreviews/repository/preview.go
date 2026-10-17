package nativepreviewrepository

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/nativepreviews/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Both revision and payload reads use the same live permission predicate. A
// cache hit never substitutes for workspace, team, or document access checks.
func (r *Repository) Read(ctx context.Context, actor auth.Actor, workspaceID uuid.UUID, target domain.Target, includeValue bool) (domain.Snapshot, error) {
	var query string
	args := []any{workspaceID, actor.PrincipalID}
	switch target.Kind {
	case "story":
		query = storyHeader
		if includeValue {
			query = storyValue
		}
		query += storyAccess
		args = append(args, nullableUUID(target.ID), target.TeamCode, target.Sequence, actor.TeamAccess.IsUnrestricted(), actor.TeamAccess.RestrictedTeamIDs())
	case "objective":
		query = objectiveHeader
		if includeValue {
			query = objectiveValue
		}
		query += objectiveAccess
		args = append(args, target.ID, target.TeamID, actor.TeamAccess.IsUnrestricted(), actor.TeamAccess.RestrictedTeamIDs())
	case "document":
		query = documentHeader
		if includeValue {
			query = documentValue
		}
		query += documentAccess
		args = append(args, target.ID)
	default:
		return domain.Snapshot{}, domain.ErrInvalidTarget
	}
	var snapshot domain.Snapshot
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&snapshot.ID, &snapshot.Version, &snapshot.Value); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Snapshot{}, domain.ErrNotFound
		}
		return domain.Snapshot{}, fmt.Errorf("read authorized native preview: %w", err)
	}
	return snapshot, nil
}

func nullableUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}

const storyRevision = `SELECT story.id, CAST(story.updated_at AS text) || ':' || team.code, `
const storyHeader = storyRevision + `NULL `
const storyValue = storyRevision + `jsonb_build_object(
    'id', story.id, 'title', story.title, 'sequenceId', story.sequence_id,
    'teamId', story.team_id, 'teamCode', team.code, 'workspaceId', story.workspace_id,
    'statusId', story.status_id, 'assigneeId', story.assignee_id,
    'priority', COALESCE(story.priority, 'No Priority'),
    'updatedAt', story.updated_at, 'archivedAt', story.archived_at, 'deletedAt', story.deleted_at
) `
const storyAccess = `FROM public.stories AS story
INNER JOIN public.teams AS team ON team.team_id = story.team_id AND team.workspace_id = story.workspace_id
INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
INNER JOIN public.workspace_members AS membership ON membership.workspace_id = story.workspace_id AND membership.user_id = $2
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.team_members AS team_membership ON team_membership.team_id = story.team_id AND team_membership.user_id = actor.user_id
WHERE story.workspace_id = $1 AND story.deleted_at IS NULL
AND ((CAST($3 AS uuid) IS NOT NULL AND story.id = CAST($3 AS uuid))
    OR (CAST($3 AS uuid) IS NULL AND UPPER(team.code) = UPPER(CAST($4 AS text)) AND story.sequence_id = $5))
AND (CAST($6 AS boolean) OR story.team_id = ANY(CAST($7 AS uuid[])))
ORDER BY story.id LIMIT 1`

const objectiveRevision = `SELECT objective.objective_id, CAST(objective.updated_at AS text), `
const objectiveHeader = objectiveRevision + `NULL `
const objectiveValue = objectiveRevision + `jsonb_build_object(
    'id', objective.objective_id, 'name', objective.name, 'sequenceId', objective.sequence_id,
    'shortSummary', LEFT(objective.short_summary, 400), 'teamId', objective.team_id,
    'workspaceId', objective.workspace_id, 'leadUser', objective.lead_user_id,
    'statusId', objective.status_id, 'priority', COALESCE(objective.priority, 'No Priority'),
    'updatedAt', objective.updated_at
) `
const objectiveAccess = `FROM public.objectives AS objective
INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = objective.workspace_id AND workspace.deleted_at IS NULL
INNER JOIN public.teams AS team ON team.team_id = objective.team_id AND team.workspace_id = objective.workspace_id
INNER JOIN public.workspace_members AS membership ON membership.workspace_id = objective.workspace_id AND membership.user_id = $2 AND membership.role IN ('member', 'admin')
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.team_members AS team_membership ON team_membership.team_id = objective.team_id AND team_membership.user_id = actor.user_id
WHERE objective.workspace_id = $1 AND objective.objective_id = $3 AND objective.team_id = $4
AND (CAST($5 AS boolean) OR objective.team_id = ANY(CAST($6 AS uuid[])))`

const documentRevision = `SELECT document.document_id, CAST(document.revision AS text) || ':' || CAST(document.updated_at AS text), `
const documentHeader = documentRevision + `NULL `
const documentValue = documentRevision + `jsonb_build_object(
    'id', document.document_id, 'title', document.title, 'contentText', LEFT(document.content_text, 400),
    'workspaceId', document.workspace_id, 'revision', document.revision, 'updatedAt', document.updated_at
) `
const documentAccess = `FROM public.documents AS document
INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = document.workspace_id AND workspace.deleted_at IS NULL
INNER JOIN public.workspace_members AS membership ON membership.workspace_id = document.workspace_id AND membership.user_id = $2
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
WHERE document.workspace_id = $1 AND document.document_id = $3 AND document.archived_at IS NULL
AND (document.visibility = 'workspace' OR document.created_by = $2 OR EXISTS (
    SELECT 1 FROM public.document_members AS reader WHERE document.visibility = 'restricted'
    AND reader.document_id = document.document_id AND reader.user_id = $2
))`
