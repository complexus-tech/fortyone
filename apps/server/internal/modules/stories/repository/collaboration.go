package storiesrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	errStoryNotFound       = errors.New("story not found")
	errInvalidCollaborator = errors.New("collaborators must be active, non-system members of the story team")
	errStoryAccessDenied   = errors.New("user cannot access this story")
)

type storyCollaborationScope struct {
	TeamID     uuid.UUID  `db:"team_id"`
	AssigneeID *uuid.UUID `db:"assignee_id"`
}

func (r *repo) GetCollaborators(ctx context.Context, storyID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT sc.user_id
		FROM story_collaborators sc
		INNER JOIN stories s ON s.id = sc.story_id
		WHERE sc.story_id = $1
			AND s.workspace_id = $2
		ORDER BY sc.created_at, sc.user_id;
	`

	collaboratorIDs := make([]uuid.UUID, 0)
	if err := r.db.SelectContext(ctx, &collaboratorIDs, query, storyID, workspaceID); err != nil {
		return nil, fmt.Errorf("get story collaborators: %w", err)
	}
	return collaboratorIDs, nil
}

func (r *repo) SetCollaborators(ctx context.Context, storyID, workspaceID uuid.UUID, collaboratorIDs []uuid.UUID) error {
	uniqueIDs := uniqueSortedUUIDs(collaboratorIDs)

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin collaborator update: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	scope, err := getStoryCollaborationScope(ctx, tx, storyID, workspaceID)
	if err != nil {
		return err
	}

	for _, collaboratorID := range uniqueIDs {
		if scope.AssigneeID != nil && collaboratorID == *scope.AssigneeID {
			return fmt.Errorf("%w: assignee cannot also be a collaborator", errInvalidCollaborator)
		}

		var valid bool
		if err := tx.GetContext(ctx, &valid, `
			SELECT EXISTS (
				SELECT 1
				FROM team_members tm
				INNER JOIN users u ON u.user_id = tm.user_id
				WHERE tm.team_id = $1
					AND tm.user_id = $2
					AND u.is_active = true
					AND u.is_system = false
			);
		`, scope.TeamID, collaboratorID); err != nil {
			return fmt.Errorf("validate collaborator: %w", err)
		}
		if !valid {
			return errInvalidCollaborator
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM story_collaborators WHERE story_id = $1`, storyID); err != nil {
		return fmt.Errorf("clear story collaborators: %w", err)
	}

	for _, collaboratorID := range uniqueIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO story_collaborators (story_id, team_id, user_id)
			VALUES ($1, $2, $3);
		`, storyID, scope.TeamID, collaboratorID); err != nil {
			return fmt.Errorf("add story collaborator: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit collaborator update: %w", err)
	}
	return nil
}

func (r *repo) SetWatching(ctx context.Context, storyID, workspaceID, userID uuid.UUID, watching bool) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin story watch update: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var teamID uuid.UUID
	if err := tx.GetContext(ctx, &teamID, `
		SELECT s.team_id
		FROM stories s
		INNER JOIN team_members tm
			ON tm.team_id = s.team_id
			AND tm.user_id = $3
		INNER JOIN users u
			ON u.user_id = tm.user_id
			AND u.is_active = true
			AND u.is_system = false
		WHERE s.id = $1 AND s.workspace_id = $2;
	`, storyID, workspaceID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errStoryAccessDenied
		}
		return fmt.Errorf("validate story watch access: %w", err)
	}

	var hasAutomaticAudienceRole bool
	if err := tx.GetContext(ctx, &hasAutomaticAudienceRole, `
		SELECT EXISTS (
			SELECT 1
			FROM stories s
			WHERE s.id = $1
				AND s.workspace_id = $2
				AND (
					s.assignee_id = $3
					OR EXISTS (
						SELECT 1
						FROM story_collaborators sc
						WHERE sc.story_id = s.id AND sc.user_id = $3
					)
				)
		);
	`, storyID, workspaceID, userID); err != nil {
		return fmt.Errorf("get automatic story audience role: %w", err)
	}

	if watching {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM story_notification_mutes
			WHERE story_id = $1 AND user_id = $2;
		`, storyID, userID); err != nil {
			return fmt.Errorf("unmute story notifications: %w", err)
		}
		if !hasAutomaticAudienceRole {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO story_watchers (story_id, team_id, user_id)
				VALUES ($1, $3, $2)
				ON CONFLICT (story_id, user_id) DO NOTHING;
			`, storyID, userID, teamID); err != nil {
				return fmt.Errorf("watch story: %w", err)
			}
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM story_watchers
			WHERE story_id = $1 AND user_id = $2;
		`, storyID, userID); err != nil {
			return fmt.Errorf("stop watching story: %w", err)
		}
		if hasAutomaticAudienceRole {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO story_notification_mutes (story_id, team_id, user_id)
				VALUES ($1, $3, $2)
				ON CONFLICT (story_id, user_id) DO NOTHING;
			`, storyID, userID, teamID); err != nil {
				return fmt.Errorf("mute story notifications: %w", err)
			}
		} else if _, err := tx.ExecContext(ctx, `
			DELETE FROM story_notification_mutes
			WHERE story_id = $1 AND user_id = $2;
		`, storyID, userID); err != nil {
			return fmt.Errorf("clear story notification mute: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit story watch update: %w", err)
	}
	return nil
}

func (r *repo) GetNotificationAudience(ctx context.Context, storyID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT audience.user_id
		FROM (
			SELECT s.assignee_id AS user_id
			FROM stories s
			WHERE s.id = $1 AND s.workspace_id = $2
			UNION
			SELECT sc.user_id
			FROM story_collaborators sc
			INNER JOIN stories s ON s.id = sc.story_id
			WHERE s.id = $1 AND s.workspace_id = $2
			UNION
			SELECT sw.user_id
			FROM story_watchers sw
			INNER JOIN stories s ON s.id = sw.story_id
			WHERE s.id = $1 AND s.workspace_id = $2
		) audience
		WHERE audience.user_id IS NOT NULL
			AND EXISTS (
				SELECT 1
				FROM users u
				WHERE u.user_id = audience.user_id
					AND u.is_active = true
					AND u.is_system = false
			)
			AND NOT EXISTS (
				SELECT 1
				FROM story_notification_mutes snm
				WHERE snm.story_id = $1 AND snm.user_id = audience.user_id
			)
		ORDER BY audience.user_id;
	`

	audienceIDs := make([]uuid.UUID, 0)
	if err := r.db.SelectContext(ctx, &audienceIDs, query, storyID, workspaceID); err != nil {
		return nil, fmt.Errorf("get story notification audience: %w", err)
	}
	return audienceIDs, nil
}

func getStoryCollaborationScope(ctx context.Context, tx *sqlx.Tx, storyID, workspaceID uuid.UUID) (storyCollaborationScope, error) {
	var scope storyCollaborationScope
	if err := tx.GetContext(ctx, &scope, `
		SELECT team_id, assignee_id
		FROM stories
		WHERE id = $1 AND workspace_id = $2
		FOR UPDATE;
	`, storyID, workspaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storyCollaborationScope{}, errStoryNotFound
		}
		return storyCollaborationScope{}, fmt.Errorf("get story collaboration scope: %w", err)
	}
	return scope, nil
}

func uniqueSortedUUIDs(values []uuid.UUID) []uuid.UUID {
	unique := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value != uuid.Nil {
			unique[value] = struct{}{}
		}
	}

	result := make([]uuid.UUID, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}
