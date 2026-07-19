package feedbackrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *Repo {
	return &Repo{log: log, db: db}
}

type portalRow struct {
	ID          uuid.UUID `db:"id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"`
	IsPublic    bool      `db:"is_public"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type boardRow struct {
	ID          uuid.UUID `db:"id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	PortalID    uuid.UUID `db:"portal_id"`
	TeamID      uuid.UUID `db:"team_id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"`
	Color       string    `db:"color"`
	OrderIndex  int       `db:"order_index"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type itemRow struct {
	ID               uuid.UUID  `db:"id"`
	WorkspaceID      uuid.UUID  `db:"workspace_id"`
	PortalID         uuid.UUID  `db:"portal_id"`
	BoardID          uuid.UUID  `db:"board_id"`
	AuthorID         *uuid.UUID `db:"author_id"`
	AuthorName       string     `db:"author_name"`
	AuthorEmail      string     `db:"author_email"`
	AuthorAvatar     *string    `db:"author_avatar"`
	Title            string     `db:"title"`
	Description      string     `db:"description"`
	Slug             string     `db:"slug"`
	Status           string     `db:"status"`
	VoteCount        int        `db:"vote_count"`
	CommentCount     int        `db:"comment_count"`
	RoadmapSummary   *string    `db:"roadmap_summary"`
	BoardTeamID      uuid.UUID  `db:"board_team_id"`
	BoardName        string     `db:"board_name"`
	BoardSlug        string     `db:"board_slug"`
	BoardColor       string     `db:"board_color"`
	BoardOrder       int        `db:"board_order_index"`
	BoardCreatedAt   time.Time  `db:"board_created_at"`
	BoardUpdatedAt   time.Time  `db:"board_updated_at"`
	PrimaryLinkID    *uuid.UUID `db:"primary_link_id"`
	PrimaryStoryID   *uuid.UUID `db:"primary_story_id"`
	PrimaryRelation  *string    `db:"primary_relationship"`
	PrimaryCreator   *uuid.UUID `db:"primary_created_by_user_id"`
	PrimaryCreatedAt *time.Time `db:"primary_created_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	StatusChanged    bool       `db:"status_changed"`
}

type commentRow struct {
	ID           uuid.UUID  `db:"id"`
	WorkspaceID  uuid.UUID  `db:"workspace_id"`
	ItemID       uuid.UUID  `db:"item_id"`
	AuthorID     *uuid.UUID `db:"author_id"`
	AuthorName   string     `db:"author_name"`
	AuthorAvatar *string    `db:"author_avatar"`
	Body         string     `db:"body"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

type storyLinkRow struct {
	ID              uuid.UUID  `db:"id"`
	WorkspaceID     uuid.UUID  `db:"workspace_id"`
	ItemID          uuid.UUID  `db:"item_id"`
	StoryID         uuid.UUID  `db:"story_id"`
	Relationship    string     `db:"relationship"`
	IsPrimary       bool       `db:"is_primary"`
	CreatedByUserID *uuid.UUID `db:"created_by_user_id"`
	CreatedAt       time.Time  `db:"created_at"`
}

const feedbackItemSearchVector = "to_tsvector('english', fi.title || ' ' || fi.description || ' ' || fi.slug)"

const projectedFeedbackStatus = `CASE
	WHEN primary_link.id IS NULL THEN fi.status
	WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
	WHEN projected_state.category = 'backlog' THEN 'reviewing'
	WHEN projected_state.category = 'unstarted' THEN 'planned'
	WHEN projected_state.category = 'started' THEN 'in_progress'
	WHEN projected_state.category = 'paused' THEN 'planned'
	WHEN projected_state.category = 'completed' THEN 'completed'
	WHEN projected_state.category = 'cancelled' THEN 'closed'
	ELSE fi.status
END`

func (r *Repo) GetPortalBySlug(ctx context.Context, slug string) (feedback.CorePortal, error) {
	var row portalRow
	err := r.db.GetContext(ctx, &row, `
		SELECT fp.id, fp.workspace_id, w.name, w.slug, fp.is_public, fp.created_at, fp.updated_at
		FROM feedback_portals fp
		INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
		WHERE w.slug = $1 AND fp.is_public = true
		LIMIT 1
	`, slug)
	if err != nil {
		return feedback.CorePortal{}, err
	}
	return toCorePortal(row), nil
}

func (r *Repo) GetPortalByWorkspaceSlugAndSlug(ctx context.Context, workspaceSlug, slug string) (feedback.CorePortal, error) {
	var row portalRow
	err := r.db.GetContext(ctx, &row, `
		SELECT fp.id, fp.workspace_id, w.name, w.slug, fp.is_public, fp.created_at, fp.updated_at
		FROM feedback_portals fp
		INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
		WHERE w.slug = $1 AND w.slug = $2 AND fp.is_public = true
	`, workspaceSlug, slug)
	if err != nil {
		return feedback.CorePortal{}, err
	}
	return toCorePortal(row), nil
}

func (r *Repo) GetPortal(ctx context.Context, workspaceID, portalID uuid.UUID) (feedback.CorePortal, error) {
	var row portalRow
	err := r.db.GetContext(ctx, &row, `
		SELECT fp.id, fp.workspace_id, w.name, w.slug, fp.is_public, fp.created_at, fp.updated_at
		FROM feedback_portals fp
		INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
		WHERE fp.workspace_id = $1 AND fp.id = $2
	`, workspaceID, portalID)
	if err != nil {
		return feedback.CorePortal{}, err
	}
	return toCorePortal(row), nil
}

func (r *Repo) ListPortals(ctx context.Context, workspaceID uuid.UUID) ([]feedback.CorePortal, error) {
	var rows []portalRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT fp.id, fp.workspace_id, w.name, w.slug, fp.is_public, fp.created_at, fp.updated_at
		FROM feedback_portals fp
		INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
		WHERE fp.workspace_id = $1
		ORDER BY fp.created_at ASC
	`, workspaceID); err != nil {
		return nil, err
	}
	result := make([]feedback.CorePortal, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCorePortal(row))
	}
	return result, nil
}

func (r *Repo) CreatePortal(ctx context.Context, input feedback.CorePortalInput) (feedback.CorePortal, error) {
	var row portalRow
	err := r.db.GetContext(ctx, &row, `
		WITH inserted AS (
			INSERT INTO feedback_portals (workspace_id, is_public)
			VALUES ($1, $2)
			RETURNING id, workspace_id, is_public, created_at, updated_at
		)
		SELECT inserted.id, inserted.workspace_id, w.name, w.slug, inserted.is_public, inserted.created_at, inserted.updated_at
		FROM inserted
		INNER JOIN workspaces w ON w.workspace_id = inserted.workspace_id
	`, input.WorkspaceID, input.IsPublic)
	if err != nil {
		return feedback.CorePortal{}, err
	}
	return toCorePortal(row), nil
}

func (r *Repo) UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input feedback.CorePortalInput) (feedback.CorePortal, error) {
	var row portalRow
	err := r.db.GetContext(ctx, &row, `
		WITH updated AS (
			UPDATE feedback_portals
			SET is_public = $3, updated_at = NOW()
			WHERE workspace_id = $1 AND id = $2
			RETURNING id, workspace_id, is_public, created_at, updated_at
		)
		SELECT updated.id, updated.workspace_id, w.name, w.slug, updated.is_public, updated.created_at, updated.updated_at
		FROM updated
		INNER JOIN workspaces w ON w.workspace_id = updated.workspace_id
	`, workspaceID, portalID, input.IsPublic)
	if err != nil {
		return feedback.CorePortal{}, err
	}
	return toCorePortal(row), nil
}

func (r *Repo) ListBoards(ctx context.Context, portalID uuid.UUID) ([]feedback.CoreBoard, error) {
	var rows []boardRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, workspace_id, portal_id, team_id, name, slug, color, order_index, created_at, updated_at
		FROM feedback_boards
		WHERE portal_id = $1
		ORDER BY order_index ASC, created_at ASC
	`, portalID); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreBoard, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCoreBoard(row))
	}
	return result, nil
}

func (r *Repo) GetBoard(ctx context.Context, portalID, boardID uuid.UUID) (feedback.CoreBoard, error) {
	var row boardRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, workspace_id, portal_id, team_id, name, slug, color, order_index, created_at, updated_at
		FROM feedback_boards
		WHERE portal_id = $1 AND id = $2
	`, portalID, boardID)
	if err != nil {
		return feedback.CoreBoard{}, err
	}
	return toCoreBoard(row), nil
}

func (r *Repo) CreateBoard(ctx context.Context, input feedback.CoreBoardInput) (feedback.CoreBoard, error) {
	var row boardRow
	err := r.db.GetContext(ctx, &row, `
		INSERT INTO feedback_boards (workspace_id, portal_id, team_id, name, slug, color, order_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, workspace_id, portal_id, team_id, name, slug, color, order_index, created_at, updated_at
	`, input.WorkspaceID, input.PortalID, input.TeamID, input.Name, input.Slug, input.Color, input.OrderIndex)
	if err != nil {
		return feedback.CoreBoard{}, err
	}
	return toCoreBoard(row), nil
}

func (r *Repo) ListItems(ctx context.Context, input feedback.CoreListItemsInput) (feedback.CoreItemsPage, error) {
	var rows []itemRow
	query, params := buildListItemsQuery(input)
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return feedback.CoreItemsPage{}, err
	}
	defer stmt.Close()
	if err := stmt.SelectContext(ctx, &rows, params); err != nil {
		return feedback.CoreItemsPage{}, err
	}
	hasMore := len(rows) > input.PageSize
	if hasMore {
		rows = rows[:input.PageSize]
	}
	result := make([]feedback.CoreItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCoreItem(row))
	}
	return feedback.CoreItemsPage{Items: result, HasMore: hasMore}, nil
}

func buildListItemsQuery(input feedback.CoreListItemsInput) (string, map[string]any) {
	where := make([]string, 0, 5)
	params := map[string]any{
		"limit":  input.PageSize + 1,
		"offset": (input.Page - 1) * input.PageSize,
	}
	if input.PortalID != uuid.Nil {
		where = append(where, "fi.portal_id = :portal_id")
		params["portal_id"] = input.PortalID
	}
	if input.TeamID != nil {
		where = append(where, "fi.workspace_id = :workspace_id", "fb.team_id = :team_id")
		params["workspace_id"] = input.WorkspaceID
		params["team_id"] = *input.TeamID
	}
	switch input.Status {
	case "", "all":
	case "active":
		where = append(where, projectedFeedbackStatus+" IN ('pending', 'reviewing')")
	default:
		where = append(where, projectedFeedbackStatus+" = :status")
		params["status"] = input.Status
	}
	if input.BoardID != nil {
		where = append(where, "fi.board_id = :board_id")
		params["board_id"] = *input.BoardID
	}
	if input.Search != "" {
		where = append(where, feedbackItemSearchVector+" @@ websearch_to_tsquery('english', :search)")
		params["search"] = input.Search
	}
	orderBy := "vote_count DESC, fi.created_at DESC"
	switch input.Sort {
	case "newest":
		orderBy = "fi.created_at DESC"
	case "oldest":
		orderBy = "fi.created_at ASC"
	case "top":
		orderBy = "vote_count DESC, fi.created_at DESC"
	}
	query := fmt.Sprintf("%s WHERE %s ORDER BY %s LIMIT :limit OFFSET :offset", itemSelectQuery(), strings.Join(where, " AND "), orderBy)
	return query, params
}

func (r *Repo) ListComments(ctx context.Context, portalID uuid.UUID) ([]feedback.CoreComment, error) {
	var rows []commentRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT fc.id, fc.workspace_id, fc.item_id, fc.author_id,
			COALESCE(u.full_name, u.email, 'Deleted user') AS author_name,
			u.avatar_url AS author_avatar,
			fc.body, fc.created_at, fc.updated_at
		FROM feedback_comments fc
		INNER JOIN feedback_items fi ON fi.id = fc.item_id
		LEFT JOIN users u ON u.user_id = fc.author_id
		WHERE fi.portal_id = $1
		ORDER BY fc.created_at ASC
	`, portalID); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreComment, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCoreComment(row))
	}
	return result, nil
}

func (r *Repo) ListItemComments(ctx context.Context, workspaceID, itemID uuid.UUID) ([]feedback.CoreComment, error) {
	var rows []commentRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT fc.id, fc.workspace_id, fc.item_id, fc.author_id,
			COALESCE(u.full_name, u.email, 'Deleted user') AS author_name,
			u.avatar_url AS author_avatar,
			fc.body, fc.created_at, fc.updated_at
		FROM feedback_comments fc
		LEFT JOIN users u ON u.user_id = fc.author_id
		WHERE fc.workspace_id = $1 AND fc.item_id = $2
		ORDER BY fc.created_at ASC
	`, workspaceID, itemID); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreComment, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCoreComment(row))
	}
	return result, nil
}

func (r *Repo) ListStoryLinks(ctx context.Context, portalID uuid.UUID) ([]feedback.CoreStoryLink, error) {
	var rows []storyLinkRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT fsl.id, fsl.workspace_id, fsl.item_id, fsl.story_id, fsl.relationship, fsl.is_primary, fsl.created_by_user_id, fsl.created_at
		FROM feedback_story_links fsl
		INNER JOIN feedback_items fi ON fi.id = fsl.item_id
		WHERE fi.portal_id = $1
		ORDER BY fsl.is_primary DESC, fsl.created_at ASC
	`, portalID); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreStoryLink, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCoreStoryLink(row))
	}
	return result, nil
}

func (r *Repo) ListItemStoryLinks(ctx context.Context, workspaceID, itemID uuid.UUID) ([]feedback.CoreStoryLink, error) {
	var rows []storyLinkRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, workspace_id, item_id, story_id, relationship, is_primary, created_by_user_id, created_at
		FROM feedback_story_links
		WHERE workspace_id = $1 AND item_id = $2
		ORDER BY is_primary DESC, created_at ASC
	`, workspaceID, itemID); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreStoryLink, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCoreStoryLink(row))
	}
	return result, nil
}

func (r *Repo) GetItem(ctx context.Context, workspaceID, itemID uuid.UUID) (feedback.CoreItem, error) {
	var row itemRow
	err := r.db.GetContext(ctx, &row, itemSelectQuery()+`
		WHERE fi.workspace_id = $1 AND fi.id = $2
	`, workspaceID, itemID)
	if err != nil {
		return feedback.CoreItem{}, err
	}
	return toCoreItem(row), nil
}

func (r *Repo) GetItemByPortal(ctx context.Context, portalID, itemID uuid.UUID) (feedback.CoreItem, error) {
	var row itemRow
	err := r.db.GetContext(ctx, &row, itemSelectQuery()+`
		WHERE fi.portal_id = $1 AND fi.id = $2
	`, portalID, itemID)
	if err != nil {
		return feedback.CoreItem{}, err
	}
	return toCoreItem(row), nil
}

func (r *Repo) CreateItem(ctx context.Context, input feedback.CoreItemInput) (feedback.CoreItem, error) {
	var row itemRow
	err := r.db.GetContext(ctx, &row, `
		WITH inserted AS (
			INSERT INTO feedback_items (workspace_id, portal_id, board_id, author_id, title, description, slug)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING *
		)
		SELECT inserted.id, inserted.workspace_id, inserted.portal_id, inserted.board_id, inserted.author_id,
			COALESCE(u.full_name, u.email, 'Deleted user') AS author_name,
			COALESCE(u.email, '') AS author_email,
			u.avatar_url AS author_avatar,
			inserted.title, inserted.description, inserted.slug, inserted.status,
			0 AS vote_count,
			0 AS comment_count,
			inserted.roadmap_summary,
			fb.team_id AS board_team_id, fb.name AS board_name, fb.slug AS board_slug,
			fb.color AS board_color, fb.order_index AS board_order_index,
			fb.created_at AS board_created_at, fb.updated_at AS board_updated_at,
			inserted.created_at, inserted.updated_at
		FROM inserted
		LEFT JOIN users u ON u.user_id = inserted.author_id
		INNER JOIN feedback_boards fb ON fb.id = inserted.board_id
	`, input.WorkspaceID, input.PortalID, input.BoardID, input.AuthorID, input.Title, input.Description, input.Slug)
	if err != nil {
		return feedback.CoreItem{}, err
	}
	return toCoreItem(row), nil
}

func (r *Repo) UpdateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, input feedback.CoreUpdateItemStatusInput) (feedback.CoreItem, bool, error) {
	var row itemRow
	err := r.db.GetContext(ctx, &row, `
		WITH previous AS (
			SELECT id, status
			FROM feedback_items
			WHERE workspace_id = $1 AND id = $2
				AND ($5 OR NOT EXISTS (
					SELECT 1
					FROM feedback_story_links fsl
					WHERE fsl.item_id = feedback_items.id AND fsl.is_primary = true
				))
			FOR UPDATE
		),
		updated AS (
			UPDATE feedback_items fi
			SET status = $3, roadmap_summary = COALESCE($4, fi.roadmap_summary), updated_at = NOW()
			FROM previous
			WHERE fi.id = previous.id
			RETURNING fi.*, previous.status AS previous_status
		)
		SELECT updated.id, updated.workspace_id, updated.portal_id, updated.board_id, updated.author_id,
			COALESCE(u.full_name, u.email, 'Deleted user') AS author_name,
			COALESCE(u.email, '') AS author_email,
			u.avatar_url AS author_avatar,
			updated.title, updated.description, updated.slug, updated.status,
			CAST(COALESCE((SELECT SUM(fv.direction) FROM feedback_votes fv WHERE fv.item_id = updated.id), 0) AS integer) AS vote_count,
			(SELECT COUNT(*) FROM feedback_comments fc WHERE fc.item_id = updated.id)::int AS comment_count,
			updated.roadmap_summary,
			fb.team_id AS board_team_id, fb.name AS board_name, fb.slug AS board_slug,
			fb.color AS board_color, fb.order_index AS board_order_index,
			fb.created_at AS board_created_at, fb.updated_at AS board_updated_at,
			updated.created_at, updated.updated_at,
			(updated.previous_status IS DISTINCT FROM updated.status) AS status_changed
		FROM updated
		LEFT JOIN users u ON u.user_id = updated.author_id
		INNER JOIN feedback_boards fb ON fb.id = updated.board_id
	`, workspaceID, itemID, input.Status, input.RoadmapSummary, input.AllowLinked)
	if err != nil {
		if err == sql.ErrNoRows && !input.AllowLinked {
			var storyManaged bool
			if checkErr := r.db.GetContext(ctx, &storyManaged, `
				SELECT EXISTS (
					SELECT 1
					FROM feedback_story_links
					WHERE workspace_id = $1 AND item_id = $2 AND is_primary = true
				)
			`, workspaceID, itemID); checkErr != nil {
				return feedback.CoreItem{}, false, checkErr
			}
			if storyManaged {
				return feedback.CoreItem{}, false, feedback.ErrStoryManaged
			}
		}
		return feedback.CoreItem{}, false, err
	}
	return toCoreItem(row), row.StatusChanged, nil
}

func (r *Repo) CreateComment(ctx context.Context, input feedback.CoreCommentInput) (feedback.CoreComment, error) {
	var row commentRow
	err := r.db.GetContext(ctx, &row, `
		WITH inserted AS (
			INSERT INTO feedback_comments (workspace_id, item_id, author_id, body)
			VALUES ($1, $2, $3, $4)
			RETURNING *
		)
		SELECT inserted.id, inserted.workspace_id, inserted.item_id, inserted.author_id,
			COALESCE(u.full_name, u.email, 'Deleted user') AS author_name,
			u.avatar_url AS author_avatar,
			inserted.body, inserted.created_at, inserted.updated_at
		FROM inserted
		LEFT JOIN users u ON u.user_id = inserted.author_id
	`, input.WorkspaceID, input.ItemID, input.AuthorID, input.Body)
	if err != nil {
		return feedback.CoreComment{}, err
	}
	return toCoreComment(row), nil
}

func (r *Repo) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (feedback.CoreVoteResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreVoteResult{}, err
	}
	defer tx.Rollback()

	var currentVote int
	if err := tx.GetContext(ctx, &currentVote, `
		SELECT COALESCE((
			SELECT direction FROM feedback_votes
			WHERE workspace_id = $1 AND item_id = $2 AND user_id = $3
		), 0)
	`, workspaceID, itemID, userID); err != nil {
		return feedback.CoreVoteResult{}, err
	}
	resultingVote := vote
	if currentVote == vote {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM feedback_votes
			WHERE workspace_id = $1 AND item_id = $2 AND user_id = $3
		`, workspaceID, itemID, userID); err != nil {
			return feedback.CoreVoteResult{}, err
		}
		resultingVote = 0
	} else {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO feedback_votes (workspace_id, item_id, user_id, direction)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (item_id, user_id)
			DO UPDATE SET direction = EXCLUDED.direction
		`, workspaceID, itemID, userID, vote); err != nil {
			return feedback.CoreVoteResult{}, err
		}
	}
	var count int
	if err := tx.GetContext(ctx, &count, `SELECT CAST(COALESCE(SUM(direction), 0) AS integer) FROM feedback_votes WHERE item_id = $1`, itemID); err != nil {
		return feedback.CoreVoteResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreVoteResult{}, err
	}
	return feedback.CoreVoteResult{Vote: resultingVote, VoteCount: count}, nil
}

func (r *Repo) LinkStory(ctx context.Context, input feedback.CoreStoryLinkInput) (feedback.CoreStoryLink, error) {
	var row storyLinkRow
	err := r.db.GetContext(ctx, &row, `
		INSERT INTO feedback_story_links (workspace_id, item_id, story_id, relationship, is_primary, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (item_id, story_id) DO UPDATE
		SET relationship = EXCLUDED.relationship, is_primary = EXCLUDED.is_primary
		RETURNING id, workspace_id, item_id, story_id, relationship, is_primary, created_by_user_id, created_at
	`, input.WorkspaceID, input.ItemID, input.StoryID, input.Relationship, input.IsPrimary, input.CreatedByUserID)
	if err != nil {
		if isPrimaryStoryConflict(err) {
			return feedback.CoreStoryLink{}, feedback.ErrAlreadyPlanned
		}
		return feedback.CoreStoryLink{}, err
	}
	return toCoreStoryLink(row), nil
}

func isPrimaryStoryConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "feedback_story_links_one_primary_per_item"
}

func (r *Repo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	var statusID uuid.UUID
	err := r.db.GetContext(ctx, &statusID, `
		SELECT status_id
		FROM statuses
		WHERE team_id = $1 AND category = $2
		ORDER BY order_index ASC
		LIMIT 1
	`, teamID, category)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &statusID, nil
}

func (r *Repo) GetStatusCategory(ctx context.Context, teamID, statusID uuid.UUID) (string, error) {
	var category string
	if err := r.db.GetContext(ctx, &category, `
		SELECT category
		FROM statuses
		WHERE team_id = $1 AND status_id = $2
	`, teamID, statusID); err != nil {
		return "", err
	}
	return category, nil
}

func itemSelectQuery() string {
	return `
		SELECT fi.id, fi.workspace_id, fi.portal_id, fi.board_id, fi.author_id,
			COALESCE(u.full_name, u.email, 'Deleted user') AS author_name,
			COALESCE(u.email, '') AS author_email,
			u.avatar_url AS author_avatar,
			fi.title, fi.description, fi.slug, ` + projectedFeedbackStatus + ` AS status,
			CAST(COALESCE((SELECT SUM(fv.direction) FROM feedback_votes fv WHERE fv.item_id = fi.id), 0) AS integer) AS vote_count,
			CAST((SELECT COUNT(*) FROM feedback_comments fc WHERE fc.item_id = fi.id) AS integer) AS comment_count,
			fi.roadmap_summary,
			fb.team_id AS board_team_id, fb.name AS board_name, fb.slug AS board_slug,
			fb.color AS board_color, fb.order_index AS board_order_index,
			fb.created_at AS board_created_at, fb.updated_at AS board_updated_at,
			primary_link.id AS primary_link_id, primary_link.story_id AS primary_story_id,
			primary_link.relationship AS primary_relationship,
			primary_link.created_by_user_id AS primary_created_by_user_id,
			primary_link.created_at AS primary_created_at,
			fi.created_at, fi.updated_at
		FROM feedback_items fi
		LEFT JOIN users u ON u.user_id = fi.author_id
		INNER JOIN feedback_boards fb ON fb.id = fi.board_id
		LEFT JOIN LATERAL (
			SELECT fsl.id, fsl.story_id, fsl.relationship, fsl.created_by_user_id, fsl.created_at
			FROM feedback_story_links fsl
			WHERE fsl.item_id = fi.id AND fsl.is_primary = true
			LIMIT 1
		) primary_link ON true
		LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
		LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
	`
}

func toCorePortal(row portalRow) feedback.CorePortal {
	return feedback.CorePortal{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		Name:        row.Name,
		Slug:        row.Slug,
		IsPublic:    row.IsPublic,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toCoreBoard(row boardRow) feedback.CoreBoard {
	return feedback.CoreBoard{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		PortalID:    row.PortalID,
		TeamID:      row.TeamID,
		Name:        row.Name,
		Slug:        row.Slug,
		Color:       row.Color,
		OrderIndex:  row.OrderIndex,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toCoreItem(row itemRow) feedback.CoreItem {
	authorID := uuid.Nil
	if row.AuthorID != nil {
		authorID = *row.AuthorID
	}
	item := feedback.CoreItem{
		ID:             row.ID,
		WorkspaceID:    row.WorkspaceID,
		PortalID:       row.PortalID,
		BoardID:        row.BoardID,
		AuthorID:       authorID,
		AuthorName:     row.AuthorName,
		AuthorEmail:    row.AuthorEmail,
		AuthorAvatar:   row.AuthorAvatar,
		Title:          row.Title,
		Description:    row.Description,
		Slug:           row.Slug,
		Status:         row.Status,
		VoteCount:      row.VoteCount,
		CommentCount:   row.CommentCount,
		RoadmapSummary: row.RoadmapSummary,
		Board: feedback.CoreBoard{
			ID:          row.BoardID,
			WorkspaceID: row.WorkspaceID,
			PortalID:    row.PortalID,
			TeamID:      row.BoardTeamID,
			Name:        row.BoardName,
			Slug:        row.BoardSlug,
			Color:       row.BoardColor,
			OrderIndex:  row.BoardOrder,
			CreatedAt:   row.BoardCreatedAt,
			UpdatedAt:   row.BoardUpdatedAt,
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.PrimaryLinkID != nil && row.PrimaryStoryID != nil && row.PrimaryRelation != nil && row.PrimaryCreatedAt != nil {
		createdBy := uuid.Nil
		if row.PrimaryCreator != nil {
			createdBy = *row.PrimaryCreator
		}
		item.StoryLinks = []feedback.CoreStoryLink{{
			ID:              *row.PrimaryLinkID,
			WorkspaceID:     row.WorkspaceID,
			ItemID:          row.ID,
			StoryID:         *row.PrimaryStoryID,
			Relationship:    *row.PrimaryRelation,
			IsPrimary:       true,
			CreatedByUserID: createdBy,
			CreatedAt:       *row.PrimaryCreatedAt,
		}}
	}
	return item
}

func toCoreComment(row commentRow) feedback.CoreComment {
	authorID := uuid.Nil
	if row.AuthorID != nil {
		authorID = *row.AuthorID
	}
	return feedback.CoreComment{
		ID:           row.ID,
		WorkspaceID:  row.WorkspaceID,
		ItemID:       row.ItemID,
		AuthorID:     authorID,
		AuthorName:   row.AuthorName,
		AuthorAvatar: row.AuthorAvatar,
		Body:         row.Body,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func toCoreStoryLink(row storyLinkRow) feedback.CoreStoryLink {
	createdBy := uuid.Nil
	if row.CreatedByUserID != nil {
		createdBy = *row.CreatedByUserID
	}
	return feedback.CoreStoryLink{
		ID:              row.ID,
		WorkspaceID:     row.WorkspaceID,
		ItemID:          row.ItemID,
		StoryID:         row.StoryID,
		Relationship:    row.Relationship,
		IsPrimary:       row.IsPrimary,
		CreatedByUserID: createdBy,
		CreatedAt:       row.CreatedAt,
	}
}
