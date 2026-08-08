package storiesrepository

import (
	"context"
	"errors"
	"reflect"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeStoryCreateTransaction struct {
	sequence  int
	created   dbStory
	nextErr   error
	insertErr error
	labelsErr error
	commitErr error

	calls          []string
	insertedStory  *stories.CoreSingleStory
	insertedLabels []uuid.UUID
}

func (tx *fakeStoryCreateTransaction) NextSequence(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	tx.calls = append(tx.calls, "sequence")
	return tx.sequence, tx.nextErr
}

func (tx *fakeStoryCreateTransaction) InsertStory(_ context.Context, story *stories.CoreSingleStory) (dbStory, error) {
	tx.calls = append(tx.calls, "story")
	copy := *story
	tx.insertedStory = &copy
	return tx.created, tx.insertErr
}

func (tx *fakeStoryCreateTransaction) InsertLabels(
	_ context.Context,
	_, _, _ uuid.UUID,
	labelIDs []uuid.UUID,
) error {
	tx.calls = append(tx.calls, "labels")
	tx.insertedLabels = append([]uuid.UUID(nil), labelIDs...)
	return tx.labelsErr
}

func (tx *fakeStoryCreateTransaction) Commit() error {
	tx.calls = append(tx.calls, "commit")
	return tx.commitErr
}

func (tx *fakeStoryCreateTransaction) Rollback() error {
	tx.calls = append(tx.calls, "rollback")
	return nil
}

func TestExecuteStoryCreateTransactionCommitsSequenceStoryAndLabelsTogether(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	labelIDs := []uuid.UUID{uuid.New(), uuid.New()}
	tx := &fakeStoryCreateTransaction{
		sequence: 41,
		created: dbStory{
			ID:        storyID,
			Workspace: workspaceID,
			Team:      teamID,
		},
	}
	story := &stories.CoreSingleStory{
		Workspace: workspaceID,
		Team:      teamID,
		Title:     "Create Slack story",
	}

	created, err := executeStoryCreateTransaction(context.Background(), tx, story, labelIDs)
	if err != nil {
		t.Fatalf("execute story creation: %v", err)
	}
	if created.ID != storyID {
		t.Fatalf("created story ID = %s, want %s", created.ID, storyID)
	}
	if tx.insertedStory == nil || tx.insertedStory.SequenceID != 42 {
		t.Fatalf("inserted sequence = %v, want 42", tx.insertedStory)
	}
	if !reflect.DeepEqual(tx.insertedLabels, labelIDs) {
		t.Fatalf("inserted labels = %v, want %v", tx.insertedLabels, labelIDs)
	}
	wantCalls := []string{"sequence", "story", "labels", "commit"}
	if !reflect.DeepEqual(tx.calls, wantCalls) {
		t.Fatalf("transaction calls = %v, want %v", tx.calls, wantCalls)
	}
}

func TestExecuteStoryCreateTransactionRollsBackStoryAndSequenceWhenLabelsFail(t *testing.T) {
	tx := &fakeStoryCreateTransaction{
		sequence:  7,
		created:   dbStory{ID: uuid.New()},
		labelsErr: stories.ErrInvalidStoryLabels,
	}
	story := &stories.CoreSingleStory{
		Workspace: uuid.New(),
		Team:      uuid.New(),
	}

	_, err := executeStoryCreateTransaction(context.Background(), tx, story, []uuid.UUID{uuid.New()})
	if !errors.Is(err, stories.ErrInvalidStoryLabels) {
		t.Fatalf("error = %v, want ErrInvalidStoryLabels", err)
	}
	wantCalls := []string{"sequence", "story", "labels", "rollback"}
	if !reflect.DeepEqual(tx.calls, wantCalls) {
		t.Fatalf("transaction calls = %v, want %v", tx.calls, wantCalls)
	}
}

func TestExecuteStoryCreateTransactionRollsBackOnCommitFailure(t *testing.T) {
	commitErr := errors.New("commit failed")
	tx := &fakeStoryCreateTransaction{
		sequence:  2,
		created:   dbStory{ID: uuid.New()},
		commitErr: commitErr,
	}
	story := &stories.CoreSingleStory{
		Workspace: uuid.New(),
		Team:      uuid.New(),
	}

	_, err := executeStoryCreateTransaction(context.Background(), tx, story, nil)
	if !errors.Is(err, commitErr) {
		t.Fatalf("error = %v, want commit failure", err)
	}
	wantCalls := []string{"sequence", "story", "labels", "commit", "rollback"}
	if !reflect.DeepEqual(tx.calls, wantCalls) {
		t.Fatalf("transaction calls = %v, want %v", tx.calls, wantCalls)
	}
}

func TestDeduplicateLabelIDsPreservesFirstOccurrence(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	got := deduplicateLabelIDs([]uuid.UUID{first, second, first, uuid.Nil, uuid.Nil})
	want := []uuid.UUID{first, second, uuid.Nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("deduplicated labels = %v, want %v", got, want)
	}
}

func TestIsStorySequenceConflict(t *testing.T) {
	conflict := &pgconn.PgError{Code: "23505", ConstraintName: "unique_team_sequence"}
	if !isStorySequenceConflict(errors.Join(errors.New("insert story"), conflict)) {
		t.Fatal("expected sequence conflict to remain retryable")
	}
	legacyConflict := errors.New(`duplicate key value violates unique constraint "unique_team_sequence"`)
	if !isStorySequenceConflict(legacyConflict) {
		t.Fatal("expected legacy sequence conflict text to remain retryable")
	}
	if isStorySequenceConflict(errors.New("different constraint")) {
		t.Fatal("expected unrelated database error not to be retryable")
	}
}

func TestIsExternalCreationKeyConflict(t *testing.T) {
	conflict := &pgconn.PgError{Code: "23505", ConstraintName: "stories_external_creation_key_key"}
	if !isExternalCreationKeyConflict(errors.Join(errors.New("insert story"), conflict)) {
		t.Fatal("expected creation-key conflict to remain discoverable through wrapping")
	}
	if isExternalCreationKeyConflict(&pgconn.PgError{Code: "23505", ConstraintName: "unique_team_sequence"}) {
		t.Fatal("expected unrelated unique conflict not to match creation key")
	}
}
