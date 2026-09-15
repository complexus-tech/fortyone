package teamuow

import (
	"context"
	"errors"
	"strings"
	"time"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type teamBinder interface {
	BindDeletionTransaction(pgx.Tx) (teamsrepository.DeletionTransaction, error)
}

type storyDeletion interface {
	LockTeamDeletionStories(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
	DeleteTeamStories(context.Context, pgx.Tx, platformauth.Actor, []uuid.UUID, time.Time) error
}

type attachmentDeletion interface {
	LockTeamDeletionAttachments(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
	RetireTeamDeletionAttachments(context.Context, pgx.Tx, []uuid.UUID, uuid.UUID, string, string, time.Time) error
}

// Manager composes repository-owned capabilities on one transaction. It never
// exposes a transaction to a service or performs remote work before commit.
type Manager struct {
	transactions platformdatabase.Transactor
	teams        teamBinder
	stories      storyDeletion
	attachments  attachmentDeletion
	provider     string
	container    string
}

func New(beginner platformdatabase.Beginner, teams teamBinder, stories storyDeletion, attachments attachmentDeletion, provider, container string) (*Manager, error) {
	provider, container = strings.TrimSpace(provider), strings.TrimSpace(container)
	if beginner == nil || teams == nil || stories == nil || attachments == nil ||
		provider == "" || len(provider) > 32 || container == "" || len(container) > 255 {
		return nil, errors.New("team deletion repositories, transaction beginner and storage route are required")
	}
	return &Manager{
		transactions: platformdatabase.NewTransactor(beginner), teams: teams, stories: stories,
		attachments: attachments, provider: provider, container: container,
	}, nil
}

func (manager *Manager) Delete(ctx context.Context, command teamsdomain.Deletion) error {
	if err := command.Validate(); err != nil {
		return err
	}
	return manager.transactions.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		teams, err := manager.teams.BindDeletionTransaction(tx)
		if err != nil {
			return err
		}
		if err := teams.Lock(ctx, command); err != nil {
			return err
		}
		storyIDs, err := manager.stories.LockTeamDeletionStories(ctx, tx, command.TeamID, command.WorkspaceID)
		if err != nil {
			return err
		}
		attachmentIDs, err := manager.attachments.LockTeamDeletionAttachments(ctx, tx, command.TeamID, command.WorkspaceID)
		if err != nil {
			return err
		}
		if err := teams.RemoveEntityReferences(ctx, command); err != nil {
			return err
		}
		if err := manager.stories.DeleteTeamStories(ctx, tx, command.Actor, storyIDs, command.DeletedAt); err != nil {
			return err
		}
		if err := teams.Delete(ctx, command); err != nil {
			return err
		}
		return manager.attachments.RetireTeamDeletionAttachments(
			ctx, tx, attachmentIDs, command.WorkspaceID, manager.provider, manager.container, command.DeletedAt,
		)
	})
}
