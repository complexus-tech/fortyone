package feedbackrepository

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestCreateBoardRejectsOrderIndexOutsidePostgresIntegerRange(t *testing.T) {
	t.Parallel()

	repository := newWithQueries(nil, nil)
	_, err := repository.CreateBoard(t.Context(), feedback.CoreBoardInput{OrderIndex: math.MaxInt})

	require.ErrorIs(t, err, feedback.ErrInvalidInput)
}

func readFeedbackQueries(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("queries", name))
	require.NoError(t, err)
	return string(contents)
}

func TestAuthenticatedFeedbackQueriesFenceCurrentActorWorkspaceAndTeam(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"community.sql", "items.sql", "merges.sql", "story_links.sql"} {
		queries := readFeedbackQueries(t, name)
		require.Contains(t, queries, "workspace_members", name)
		require.Contains(t, queries, "team_members", name)
		require.Contains(t, queries, "current_actor.is_active = true", name)
		require.Contains(t, queries, "current_actor.is_system = false", name)
		require.Contains(t, queries, "credential_team_ids", name)
	}
}

func TestFeedbackMutationsRejectDeletedOrMergedItems(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"community.sql", "items.sql", "merges.sql", "story_links.sql", "updates.sql"} {
		queries := readFeedbackQueries(t, name)
		require.Contains(t, queries, "deleted_at IS NULL", name)
		require.Contains(t, queries, "merged_into_item_id IS NULL", name)
	}
}

func TestPublicIdentityQueriesMaskGuestContactDetails(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"community.sql", "items.sql"} {
		queries := readFeedbackQueries(t, name)
		require.Contains(t, queries, "guest_identity_policy = 'always_mask_guests'", name)
		require.Contains(t, queries, "contributor.kind IN ('verified_guest', 'external')", name)
		require.Contains(t, queries, "THEN 'Anonymous'", name)
		require.NotContains(t, queries, "THEN CAST(NULL AS text)", name)
	}
}

func TestFeedbackApplicationQueriesAvoidPostgresShorthandCasts(t *testing.T) {
	t.Parallel()

	entries, err := filepath.Glob(filepath.Join("queries", "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	for _, entry := range entries {
		contents, readErr := os.ReadFile(entry)
		require.NoError(t, readErr)
		require.NotContains(t, string(contents), "::", filepath.Base(entry))
	}
}

func TestAnonymousParticipationMigrationPreservesIdentityLifecycle(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "000119_feedback_anonymous_participation.up.sql"))
	require.NoError(t, err)
	migration := string(contents)
	require.Contains(t, migration, "one unlinkable anonymous contributor each")
	require.Contains(t, migration, "ON DELETE SET NULL")
	require.Contains(t, migration, "ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED")
}

func TestConstraintErrorsMapToFiniteFeedbackErrors(t *testing.T) {
	t.Parallel()

	primaryErr := fmt.Errorf("insert primary feedback story: %w", &pgconn.PgError{
		Code: uniqueViolation, ConstraintName: "feedback_story_links_one_primary_per_item",
	})
	require.True(t, isPrimaryStoryConflict(primaryErr))
	require.False(t, isPrimaryStoryConflict(errors.New("database unavailable")))

	boardErr := &pgconn.PgError{Code: uniqueViolation, ConstraintName: "feedback_boards_workspace_team_unique"}
	require.ErrorIs(t, normalizeBoardWriteError(boardErr), feedback.ErrBoardExists)
}

func TestItemProjectionIncludesPrimaryLinkAndExplicitIdentity(t *testing.T) {
	t.Parallel()

	itemID, workspaceID, portalID, boardID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	linkID, storyID := uuid.New(), uuid.New()
	createdAt := time.Now().UTC()
	relationship := feedback.RelationshipCreatedFrom
	title := "Repair traffic signals"
	avatar := "profiles/ada.webp"
	item := (itemProjection{ID: itemID, WorkspaceID: workspaceID, PortalID: portalID, BoardID: boardID,
		AuthorName: "Ada", AuthorEmail: "ada@example.com", AuthorAvatar: &avatar,
		PrimaryLinkID: &linkID, PrimaryStoryID: &storyID, PrimaryStoryTitle: &title,
		PrimaryRelationship: &relationship, PrimaryCreatedAt: &createdAt}).core()

	require.Equal(t, "Ada", item.AuthorName)
	require.Equal(t, "ada@example.com", item.AuthorEmail)
	require.Equal(t, avatar, *item.AuthorAvatar)
	require.Len(t, item.StoryLinks, 1)
	require.Equal(t, storyID, item.StoryLinks[0].StoryID)
}

func TestFeedbackRepositoryContainsNoRawApplicationSQL(t *testing.T) {
	t.Parallel()

	entries, err := filepath.Glob("*.go")
	require.NoError(t, err)
	for _, entry := range entries {
		if strings.HasSuffix(entry, "_test.go") {
			continue
		}
		contents, readErr := os.ReadFile(entry)
		require.NoError(t, readErr)
		source := string(contents)
		require.NotContains(t, source, "github.com/jmoiron/sqlx", entry)
		require.NotContains(t, source, "database/sql", entry)
	}
}
