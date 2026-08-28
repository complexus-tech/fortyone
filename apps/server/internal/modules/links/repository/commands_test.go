package linksrepository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	linksdomain "github.com/complexus-tech/projects-api/internal/modules/links/domain"
	linksql "github.com/complexus-tech/projects-api/internal/modules/links/repository/sqlc"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeLinkQueries struct {
	create func(context.Context, linksql.CreateLinkForWorkspaceParams) (linksql.CreateLinkForWorkspaceRow, error)
	update func(context.Context, linksql.UpdateLinkForWorkspaceParams) (int64, error)
	delete func(context.Context, linksql.DeleteLinkForWorkspaceParams) (int64, error)
}

func (f fakeLinkQueries) CreateLinkForWorkspace(
	ctx context.Context,
	params linksql.CreateLinkForWorkspaceParams,
) (linksql.CreateLinkForWorkspaceRow, error) {
	if f.create == nil {
		panic("unexpected CreateLinkForWorkspace call")
	}
	return f.create(ctx, params)
}

func (f fakeLinkQueries) UpdateLinkForWorkspace(
	ctx context.Context,
	params linksql.UpdateLinkForWorkspaceParams,
) (int64, error) {
	if f.update == nil {
		panic("unexpected UpdateLinkForWorkspace call")
	}
	return f.update(ctx, params)
}

func (f fakeLinkQueries) DeleteLinkForWorkspace(
	ctx context.Context,
	params linksql.DeleteLinkForWorkspaceParams,
) (int64, error) {
	if f.delete == nil {
		panic("unexpected DeleteLinkForWorkspace call")
	}
	return f.delete(ctx, params)
}

func TestCreateLinkMapsTypedParamsAndRow(t *testing.T) {
	title := "Design"
	storyID := uuid.New()
	workspaceID := uuid.New()
	actorID := uuid.New()
	linkID := uuid.New()
	createdAt := time.Date(2026, time.August, 27, 8, 30, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	repository := newTestRepository(fakeLinkQueries{
		create: func(_ context.Context, params linksql.CreateLinkForWorkspaceParams) (linksql.CreateLinkForWorkspaceRow, error) {
			if params.Title != &title || params.URL != "https://example.com" {
				t.Fatalf("content params = %#v", params)
			}
			if params.StoryID != storyID || params.WorkspaceID != workspaceID || params.ActorID != actorID {
				t.Fatalf("scope params = %s/%s/%s, want %s/%s/%s", params.StoryID, params.WorkspaceID, params.ActorID, storyID, workspaceID, actorID)
			}
			return linksql.CreateLinkForWorkspaceRow{
				LinkID:    linkID,
				Title:     &title,
				URL:       params.URL,
				StoryID:   params.StoryID,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}, nil
		},
	})

	got, err := repository.CreateLink(context.Background(), actorID, linksdomain.CreateLink{
		Title:       &title,
		URL:         "https://example.com",
		StoryID:     storyID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		t.Fatalf("create link: %v", err)
	}
	if got.LinkID != linkID || got.StoryID != storyID || got.CreatedAt != createdAt || got.UpdatedAt != updatedAt {
		t.Fatalf("mapped link = %#v", got)
	}
}

func TestCreateLinkMapsMissingScopedStoryToNotFound(t *testing.T) {
	repository := newTestRepository(fakeLinkQueries{
		create: func(context.Context, linksql.CreateLinkForWorkspaceParams) (linksql.CreateLinkForWorkspaceRow, error) {
			return linksql.CreateLinkForWorkspaceRow{}, pgx.ErrNoRows
		},
	})

	_, err := repository.CreateLink(context.Background(), uuid.New(), linksdomain.CreateLink{})
	if !errors.Is(err, linksdomain.ErrNotFound) {
		t.Fatalf("error = %v, want linksdomain.ErrNotFound", err)
	}
}

func TestUpdateLinkPreservesPatchAndScopeParams(t *testing.T) {
	url := "https://example.com/updated"
	linkID := uuid.New()
	workspaceID := uuid.New()
	actorID := uuid.New()
	repository := newTestRepository(fakeLinkQueries{
		update: func(_ context.Context, params linksql.UpdateLinkForWorkspaceParams) (int64, error) {
			if params.Title != nil || params.URL != &url {
				t.Fatalf("patch params = %#v", params)
			}
			if params.LinkID != linkID || params.WorkspaceID != workspaceID || params.ActorID != actorID {
				t.Fatalf("scope params = %s/%s/%s, want %s/%s/%s", params.LinkID, params.WorkspaceID, params.ActorID, linkID, workspaceID, actorID)
			}
			return 1, nil
		},
	})

	if err := repository.UpdateLink(
		context.Background(),
		actorID,
		linkID,
		workspaceID,
		linksdomain.UpdateLink{URL: &url},
	); err != nil {
		t.Fatalf("update link: %v", err)
	}
}

func TestUpdateAndDeleteMapZeroRowsToNotFound(t *testing.T) {
	repository := newTestRepository(fakeLinkQueries{
		update: func(context.Context, linksql.UpdateLinkForWorkspaceParams) (int64, error) {
			return 0, nil
		},
		delete: func(context.Context, linksql.DeleteLinkForWorkspaceParams) (int64, error) {
			return 0, nil
		},
	})

	if err := repository.UpdateLink(context.Background(), uuid.New(), uuid.New(), uuid.New(), linksdomain.UpdateLink{}); !errors.Is(err, linksdomain.ErrNotFound) {
		t.Fatalf("update error = %v, want linksdomain.ErrNotFound", err)
	}
	if err := repository.DeleteLink(context.Background(), uuid.New(), uuid.New(), uuid.New()); !errors.Is(err, linksdomain.ErrNotFound) {
		t.Fatalf("delete error = %v, want linksdomain.ErrNotFound", err)
	}
}

func TestTypedQueryErrorsRemainDiscoverable(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	repository := newTestRepository(fakeLinkQueries{
		create: func(context.Context, linksql.CreateLinkForWorkspaceParams) (linksql.CreateLinkForWorkspaceRow, error) {
			return linksql.CreateLinkForWorkspaceRow{}, databaseErr
		},
		update: func(context.Context, linksql.UpdateLinkForWorkspaceParams) (int64, error) {
			return 0, databaseErr
		},
		delete: func(context.Context, linksql.DeleteLinkForWorkspaceParams) (int64, error) {
			return 0, databaseErr
		},
	})

	if _, err := repository.CreateLink(context.Background(), uuid.New(), linksdomain.CreateLink{}); !errors.Is(err, databaseErr) {
		t.Fatalf("create error = %v, want wrapped database error", err)
	}
	if err := repository.UpdateLink(context.Background(), uuid.New(), uuid.New(), uuid.New(), linksdomain.UpdateLink{}); !errors.Is(err, databaseErr) {
		t.Fatalf("update error = %v, want wrapped database error", err)
	}
	if err := repository.DeleteLink(context.Background(), uuid.New(), uuid.New(), uuid.New()); !errors.Is(err, databaseErr) {
		t.Fatalf("delete error = %v, want wrapped database error", err)
	}
}

func newTestRepository(queries linksql.Querier) *repo {
	return newWithQueries(
		logger.NewWithText(io.Discard, slog.LevelError, "links-repository-test"),
		queries,
	)
}
