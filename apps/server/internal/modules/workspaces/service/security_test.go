package workspaces

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type securityRepositoryStub struct {
	Repository
	addMember func(context.Context, uuid.UUID, uuid.UUID, string) error
}

func (stub securityRepositoryStub) AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	return stub.addMember(ctx, workspaceID, userID, role)
}

func TestMembershipRolesFailClosedBeforePersistence(t *testing.T) {
	t.Parallel()
	persisted := false
	service := &Service{
		log: logger.NewWithText(io.Discard, slog.LevelError, "workspace-service-test"),
		repo: securityRepositoryStub{addMember: func(context.Context, uuid.UUID, uuid.UUID, string) error {
			persisted = true
			return nil
		}},
	}
	for _, role := range []string{"owner", "system", "ADMIN", " member "} {
		if err := service.AddMember(context.Background(), uuid.New(), uuid.New(), role); !errors.Is(err, authorization.ErrInvalidWorkspaceRole) {
			t.Fatalf("role %q error = %v, want ErrInvalidWorkspaceRole", role, err)
		}
	}
	if persisted {
		t.Fatal("invalid role reached persistence")
	}
}

func TestRestrictedSlugsAreServiceInvariants(t *testing.T) {
	t.Parallel()
	service := &Service{
		log:        logger.NewWithText(io.Discard, slog.LevelError, "workspace-service-test"),
		unitOfWork: unavailableUnitOfWork{},
	}
	if _, err := service.Create(context.Background(), CoreWorkspace{Name: "Reserved", Slug: " ADMIN "}, uuid.New()); !errors.Is(err, ErrRestrictedSlug) {
		t.Fatalf("create restricted slug error = %v, want ErrRestrictedSlug", err)
	}
}

func TestSeedStoriesRequireTypedStatusInput(t *testing.T) {
	t.Parallel()
	if _, err := BuildSeedStories(uuid.New(), uuid.New(), nil); !errors.Is(err, ErrNoSeedStatuses) {
		t.Fatalf("empty seed statuses error = %v, want ErrNoSeedStatuses", err)
	}
}
