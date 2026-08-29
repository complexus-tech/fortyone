package mentionsrepository

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestValidatedMentionTargetsRejectsAmbiguousInputAndSorts(t *testing.T) {
	t.Parallel()
	workspaceID, commentID := uuid.New(), uuid.New()
	first := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	second := uuid.MustParse("22222222-2222-4222-8222-222222222222")

	targets, err := validatedMentionTargets(workspaceID, commentID, []uuid.UUID{second, first})
	if err != nil {
		t.Fatalf("validatedMentionTargets() error = %v", err)
	}
	if len(targets) != 2 || targets[0] != first || targets[1] != second {
		t.Fatalf("validatedMentionTargets() = %v", targets)
	}

	tooMany := make([]uuid.UUID, maximumMentionsPerComment+1)
	for index := range tooMany {
		tooMany[index] = uuid.New()
	}
	for _, input := range []struct {
		workspace uuid.UUID
		comment   uuid.UUID
		users     []uuid.UUID
	}{
		{workspace: uuid.Nil, comment: commentID},
		{workspace: workspaceID, comment: uuid.Nil},
		{workspace: workspaceID, comment: commentID, users: []uuid.UUID{uuid.Nil}},
		{workspace: workspaceID, comment: commentID, users: []uuid.UUID{first, first}},
		{workspace: workspaceID, comment: commentID, users: tooMany},
	} {
		if _, err := validatedMentionTargets(input.workspace, input.comment, input.users); !errors.Is(err, ErrInvalidMention) {
			t.Fatalf("validatedMentionTargets(%+v) error = %v", input, err)
		}
	}
}

func TestUnconfiguredRepositoryFailsClosed(t *testing.T) {
	t.Parallel()
	repository := New(nil)
	if err := repository.SaveMentions(t.Context(), uuid.New(), uuid.New(), nil); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("SaveMentions() error = %v", err)
	}
	if _, err := repository.GetMentions(t.Context(), uuid.New(), uuid.New()); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("GetMentions() error = %v", err)
	}
}
