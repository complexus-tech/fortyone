package objectivesrepository

import (
	"context"
	"fmt"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type objectiveKeyResultWriter interface {
	AllocateSequences(ctx context.Context, workspaceID, teamID uuid.UUID, count int) (int, error)
	InsertKeyResult(ctx context.Context, keyResult dbKeyResult) (dbKeyResult, error)
	InsertContributor(ctx context.Context, keyResultID, contributorID uuid.UUID) error
}

type sqlObjectiveKeyResultWriter struct {
	tx                   *sqlx.Tx
	keyResultStatement   *sqlx.NamedStmt
	contributorStatement *sqlx.Stmt
}

func newSQLObjectiveKeyResultWriter(ctx context.Context, tx *sqlx.Tx) (*sqlObjectiveKeyResultWriter, error) {
	const insertKeyResult = `
		INSERT INTO key_results (
			objective_id, team_id, sequence_id, name, measurement_type,
			start_value, current_value, target_value,
			lead, start_date, end_date, created_by
		) VALUES (
			:objective_id, :team_id, :sequence_id, :name, :measurement_type,
			:start_value, :current_value, :target_value,
			:lead, :start_date, :end_date, :created_by
		)
		RETURNING id, sequence_id, objective_id, team_id, name, measurement_type,
			start_value, current_value, target_value, lead, start_date, end_date,
			created_at, updated_at, created_by
	`
	keyResultStatement, err := tx.PrepareNamedContext(ctx, insertKeyResult)
	if err != nil {
		return nil, fmt.Errorf("prepare objective key result insert: %w", err)
	}

	const insertContributor = `
		INSERT INTO key_result_contributors (key_result_id, user_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (key_result_id, user_id) DO NOTHING
	`
	contributorStatement, err := tx.PreparexContext(ctx, insertContributor)
	if err != nil {
		keyResultStatement.Close()
		return nil, fmt.Errorf("prepare objective key result contributor insert: %w", err)
	}

	return &sqlObjectiveKeyResultWriter{
		tx:                   tx,
		keyResultStatement:   keyResultStatement,
		contributorStatement: contributorStatement,
	}, nil
}

func (w *sqlObjectiveKeyResultWriter) Close() {
	w.keyResultStatement.Close()
	w.contributorStatement.Close()
}

func (w *sqlObjectiveKeyResultWriter) AllocateSequences(
	ctx context.Context,
	workspaceID, teamID uuid.UUID,
	count int,
) (int, error) {
	var finalSequenceID int
	if err := w.tx.GetContext(ctx, &finalSequenceID, `
		INSERT INTO team_key_result_sequences (
			workspace_id,
			team_id,
			current_sequence
		) VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id, team_id)
		DO UPDATE SET current_sequence =
			team_key_result_sequences.current_sequence + EXCLUDED.current_sequence
		RETURNING current_sequence
	`, workspaceID, teamID, count); err != nil {
		return 0, fmt.Errorf("allocate objective key result sequences: %w", err)
	}

	return finalSequenceID, nil
}

func (w *sqlObjectiveKeyResultWriter) InsertKeyResult(ctx context.Context, keyResult dbKeyResult) (dbKeyResult, error) {
	var created dbKeyResult
	if err := w.keyResultStatement.GetContext(ctx, &created, keyResult); err != nil {
		return dbKeyResult{}, err
	}
	return created, nil
}

func (w *sqlObjectiveKeyResultWriter) InsertContributor(
	ctx context.Context,
	keyResultID, contributorID uuid.UUID,
) error {
	_, err := w.contributorStatement.ExecContext(ctx, keyResultID, contributorID)
	return err
}

func createObjectiveKeyResults(
	ctx context.Context,
	writer objectiveKeyResultWriter,
	objectiveID, workspaceID, teamID uuid.UUID,
	keyResults []keyresults.CoreNewKeyResult,
) ([]keyresults.CoreKeyResult, error) {
	if len(keyResults) == 0 {
		return []keyresults.CoreKeyResult{}, nil
	}

	finalSequenceID, err := writer.AllocateSequences(ctx, workspaceID, teamID, len(keyResults))
	if err != nil {
		return nil, err
	}
	firstSequenceID := finalSequenceID - len(keyResults) + 1

	createdKeyResults := make([]keyresults.CoreKeyResult, 0, len(keyResults))
	for i, keyResult := range keyResults {
		keyResult.ObjectiveID = objectiveID
		params := toDBKeyResult(keyResult, keyResult.CreatedBy)
		params.TeamID = teamID
		params.SequenceID = firstSequenceID + i

		created, err := writer.InsertKeyResult(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("insert objective key result %d: %w", i, err)
		}

		contributors := deduplicateUUIDs(keyResult.Contributors)
		for _, contributorID := range contributors {
			if err := writer.InsertContributor(ctx, created.ID, contributorID); err != nil {
				return nil, fmt.Errorf("insert objective key result contributor: %w", err)
			}
		}

		coreKeyResult := toCoreKeyResult(created)
		coreKeyResult.Contributors = contributors
		createdKeyResults = append(createdKeyResults, coreKeyResult)
	}

	return createdKeyResults, nil
}

func deduplicateUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	unique := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}
