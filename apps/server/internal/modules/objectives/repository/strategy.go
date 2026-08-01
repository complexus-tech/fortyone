package objectivesrepository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/google/uuid"
)

type dbStrategy struct {
	UltimateGoal string  `db:"ultimate_goal"`
	Description  *string `db:"description"`
}

type dbStrategicPillar struct {
	ID          uuid.UUID `db:"pillar_id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	OrderIndex  int       `db:"order_index"`
}

type dbObjectiveAlignment struct {
	PillarID    uuid.UUID `db:"pillar_id"`
	ObjectiveID uuid.UUID `db:"objective_id"`
}

func (r *repo) GetStrategyMap(ctx context.Context, workspaceID uuid.UUID) (objectives.CoreStrategyMap, error) {
	strategy := dbStrategy{}
	err := r.db.GetContext(ctx, &strategy, `
		SELECT ultimate_goal, description
		FROM workspace_strategies
		WHERE workspace_id = $1
	`, workspaceID)
	if err != nil && err != sql.ErrNoRows {
		return objectives.CoreStrategyMap{}, fmt.Errorf("get workspace strategy: %w", err)
	}

	var dbPillars []dbStrategicPillar
	if err := r.db.SelectContext(ctx, &dbPillars, `
		SELECT pillar_id, name, description, order_index
		FROM strategic_pillars
		WHERE workspace_id = $1
		ORDER BY order_index, created_at
	`, workspaceID); err != nil {
		return objectives.CoreStrategyMap{}, fmt.Errorf("list strategic pillars: %w", err)
	}

	var alignments []dbObjectiveAlignment
	if err := r.db.SelectContext(ctx, &alignments, `
		SELECT pillar_id, objective_id
		FROM strategy_objective_alignments
		WHERE workspace_id = $1
	`, workspaceID); err != nil {
		return objectives.CoreStrategyMap{}, fmt.Errorf("list strategy alignments: %w", err)
	}

	objectiveIDs := make(map[uuid.UUID][]uuid.UUID, len(dbPillars))
	for _, alignment := range alignments {
		objectiveIDs[alignment.PillarID] = append(objectiveIDs[alignment.PillarID], alignment.ObjectiveID)
	}

	pillars := make([]objectives.CoreStrategicPillar, 0, len(dbPillars))
	for _, pillar := range dbPillars {
		pillars = append(pillars, objectives.CoreStrategicPillar{
			ID: pillar.ID, Name: pillar.Name, Description: pillar.Description,
			OrderIndex: pillar.OrderIndex, ObjectiveIDs: objectiveIDs[pillar.ID],
		})
	}

	return objectives.CoreStrategyMap{UltimateGoal: strategy.UltimateGoal, Description: strategy.Description, Pillars: pillars}, nil
}

func (r *repo) UpdateStrategy(ctx context.Context, workspaceID uuid.UUID, strategy objectives.CoreStrategyUpdate) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO workspace_strategies (workspace_id, ultimate_goal, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id) DO UPDATE SET
			ultimate_goal = EXCLUDED.ultimate_goal,
			description = EXCLUDED.description,
			updated_at = NOW()
	`, workspaceID, strategy.UltimateGoal, strategy.Description)
	if err != nil {
		return fmt.Errorf("update workspace strategy: %w", err)
	}
	return nil
}

func (r *repo) CreateStrategicPillar(ctx context.Context, workspaceID uuid.UUID, pillar objectives.CoreNewStrategicPillar) (objectives.CoreStrategicPillar, error) {
	var created dbStrategicPillar
	err := r.db.GetContext(ctx, &created, `
		INSERT INTO strategic_pillars (workspace_id, name, description, order_index)
		VALUES ($1, $2, $3, $4)
		RETURNING pillar_id, name, description, order_index
	`, workspaceID, pillar.Name, pillar.Description, pillar.OrderIndex)
	if err != nil {
		return objectives.CoreStrategicPillar{}, fmt.Errorf("create strategic pillar: %w", err)
	}
	return objectives.CoreStrategicPillar{ID: created.ID, Name: created.Name, Description: created.Description, OrderIndex: created.OrderIndex, ObjectiveIDs: []uuid.UUID{}}, nil
}

func (r *repo) UpdateStrategicPillar(ctx context.Context, workspaceID, pillarID uuid.UUID, pillar objectives.CoreUpdateStrategicPillar) (objectives.CoreStrategicPillar, error) {
	setClauses := []string{"updated_at = NOW()"}
	params := map[string]any{"workspace_id": workspaceID, "pillar_id": pillarID}
	if pillar.Name != nil {
		setClauses = append(setClauses, "name = :name")
		params["name"] = *pillar.Name
	}
	if pillar.Description != nil {
		setClauses = append(setClauses, "description = :description")
		params["description"] = *pillar.Description
	}
	if pillar.OrderIndex != nil {
		setClauses = append(setClauses, "order_index = :order_index")
		params["order_index"] = *pillar.OrderIndex
	}

	query := `UPDATE strategic_pillars SET ` + strings.Join(setClauses, ", ") + `
		WHERE workspace_id = :workspace_id AND pillar_id = :pillar_id
		RETURNING pillar_id, name, description, order_index`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return objectives.CoreStrategicPillar{}, fmt.Errorf("prepare strategic pillar update: %w", err)
	}
	defer stmt.Close()
	var updated dbStrategicPillar
	if err := stmt.GetContext(ctx, &updated, params); err != nil {
		return objectives.CoreStrategicPillar{}, fmt.Errorf("update strategic pillar: %w", err)
	}
	return objectives.CoreStrategicPillar{ID: updated.ID, Name: updated.Name, Description: updated.Description, OrderIndex: updated.OrderIndex, ObjectiveIDs: []uuid.UUID{}}, nil
}

func (r *repo) DeleteStrategicPillar(ctx context.Context, workspaceID, pillarID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM strategic_pillars WHERE workspace_id = $1 AND pillar_id = $2`, workspaceID, pillarID)
	if err != nil {
		return fmt.Errorf("delete strategic pillar: %w", err)
	}
	return nil
}

func (r *repo) AlignObjective(ctx context.Context, workspaceID, objectiveID uuid.UUID, pillarID *uuid.UUID) error {
	if pillarID == nil {
		_, err := r.db.ExecContext(ctx, `DELETE FROM strategy_objective_alignments WHERE workspace_id = $1 AND objective_id = $2`, workspaceID, objectiveID)
		return err
	}
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO strategy_objective_alignments (workspace_id, objective_id, pillar_id)
		SELECT $1, o.objective_id, p.pillar_id
		FROM objectives o
		JOIN strategic_pillars p ON p.pillar_id = $3 AND p.workspace_id = $1
		WHERE o.objective_id = $2 AND o.workspace_id = $1
		ON CONFLICT (objective_id) DO UPDATE SET pillar_id = EXCLUDED.pillar_id, workspace_id = EXCLUDED.workspace_id, updated_at = NOW()
	`, workspaceID, objectiveID, *pillarID)
	if err != nil {
		return fmt.Errorf("align objective: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return objectives.ErrNotFound
	}
	return nil
}
