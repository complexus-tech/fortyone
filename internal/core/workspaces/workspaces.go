package workspaces

import (
	"context"
	"errors"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Service errors
var (
	ErrNotFound               = errors.New("workspace not found")
	ErrSlugTaken              = errors.New("workspace with this url already exists")
	ErrTx                     = errors.New("failed to create a workspace")
	ErrAlreadyWorkspaceMember = errors.New("user is already a member of this workspace")
)

var restrictedSlugs = []string{
	"admin", "internal", "qa", "staging", "ops", "team", "complexus",
	"dev", "test", "prod", "staging", "development", "testing",
	"production", "staff", "hr", "finance", "legal", "marketing",
	"sales", "support", "it", "security", "engineering", "design",
	"product", "auth",
}

// Repository provides access to the users storage.
type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]CoreWorkspace, error)
	Create(ctx context.Context, tx *sqlx.Tx, newWorkspace CoreWorkspace) (CoreWorkspace, error)
	Update(ctx context.Context, workspaceID uuid.UUID, updates CoreWorkspace) (CoreWorkspace, error)
	Delete(ctx context.Context, workspaceID uuid.UUID) error
	AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error
	AddMemberTx(ctx context.Context, tx *sqlx.Tx, workspaceID, userID uuid.UUID, role string) error
	Get(ctx context.Context, workspaceID, userID uuid.UUID) (CoreWorkspace, error)
	RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error
	CheckSlugAvailability(ctx context.Context, slug string) (bool, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error
	GetWorkspaceTerminology(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceTerminology, error)
	UpdateWorkspaceTerminology(ctx context.Context, workspaceID uuid.UUID, terminology CoreWorkspaceTerminology) (CoreWorkspaceTerminology, error)
	InitializeWorkspaceTerminology(ctx context.Context, tx *sqlx.Tx, workspaceID uuid.UUID) error
}

// Service provides user-related operations.
type Service struct {
	repo            Repository
	log             *logger.Logger
	db              *sqlx.DB
	teams           *teams.Service
	stories         *stories.Service
	statuses        *states.Service
	users           *users.Service
	objectivestatus *objectivestatus.Service
}

// New constructs a new users service instance with the provided repository.
func New(log *logger.Logger, repo Repository, db *sqlx.DB, teams *teams.Service, stories *stories.Service, statuses *states.Service, users *users.Service, objectivestatus *objectivestatus.Service) *Service {
	return &Service{
		repo:            repo,
		log:             log,
		db:              db,
		teams:           teams,
		stories:         stories,
		statuses:        statuses,
		users:           users,
		objectivestatus: objectivestatus,
	}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.list")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.List")
	defer span.End()

	workspaces, err := s.repo.List(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("workspaces retrieved.", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
	))
	return workspaces, nil
}

func (s *Service) Create(ctx context.Context, newWorkspace CoreWorkspace, userID uuid.UUID) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.create")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Create")
	defer span.End()

	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return CoreWorkspace{}, ErrTx
	}
	defer tx.Rollback()

	// Critical operations in transaction
	workspace, err := s.repo.Create(ctx, tx, newWorkspace)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}

	// Add creator as member of the workspace
	if err := s.repo.AddMemberTx(ctx, tx, workspace.ID, userID, "admin"); err != nil {
		return CoreWorkspace{}, err
	}

	// Create a default team
	team, err := s.teams.CreateTx(ctx, tx, teams.CoreTeam{
		Name:      "Team 1",
		Color:     workspace.Color,
		Code:      "TM",
		Workspace: workspace.ID,
	})
	if err != nil {
		return CoreWorkspace{}, err
	}

	// Add creator as member of the default team
	if err := s.teams.AddMemberTx(ctx, tx, team.ID, userID); err != nil {
		return CoreWorkspace{}, err
	}

	// Update user's last used workspace
	if err := s.users.UpdateUserWorkspaceWithTx(ctx, tx, userID, workspace.ID); err != nil {
		s.log.Error(ctx, "failed to update user's last used workspace", err)
	}

	// Initialize workspace terminology
	if err := s.repo.InitializeWorkspaceTerminology(ctx, tx, workspace.ID); err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			s.log.Error(ctx, "error rolling back tx", "method", "Create")
		}
		return CoreWorkspace{}, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return CoreWorkspace{}, ErrTx
	}

	// Create seed stories after the transaction is committed
	statuses, err := s.statuses.TeamList(ctx, workspace.ID, team.ID)
	if err != nil {
		s.log.Error(ctx, "failed to get statuses for the team", err)
		// Non-critical, continue
	} else {
		seedStoryData := seedStories(team.ID, userID, statuses)
		for _, newStory := range seedStoryData {
			if _, err := s.stories.Create(ctx, newStory, workspace.ID); err != nil {
				s.log.Error(ctx, "failed to create seed story", err)
				// Non-critical, continue
			}
		}
	}

	span.AddEvent("workspace created.", trace.WithAttributes(
		attribute.String("workspace_id", workspace.ID.String()),
	))
	return workspace, nil
}

func (s *Service) Update(ctx context.Context, workspaceID uuid.UUID, updates CoreWorkspace) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.update")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Update")
	defer span.End()

	workspace, err := s.repo.Update(ctx, workspaceID, updates)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}

	span.AddEvent("workspace updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return workspace, nil
}

func (s *Service) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.delete")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("workspace deleted.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return nil
}

func (s *Service) AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	s.log.Info(ctx, "business.core.workspaces.addMember")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.AddMember")
	defer span.End()

	if role == "" {
		role = "member"
	}

	if err := s.repo.AddMember(ctx, workspaceID, userID, role); err != nil {
		span.RecordError(err)
		return err
	}

	// switch the user's the last workspace to the new workspace
	if err := s.users.UpdateUserWorkspace(ctx, userID, workspaceID); err != nil {
		s.log.Error(ctx, "failed to update user workspace", err)
		// no need to return error this is not a critical operation
	}

	span.AddEvent("workspace member added.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role", role),
	))
	return nil
}

func (s *Service) Get(ctx context.Context, workspaceID, userID uuid.UUID) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.get")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Get")
	defer span.End()

	workspace, err := s.repo.Get(ctx, workspaceID, userID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}

	span.AddEvent("workspace retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
	))
	return workspace, nil
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.removeMember")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.RemoveMember")
	defer span.End()

	if err := s.repo.RemoveMember(ctx, workspaceID, userID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("workspace member removed.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
	))
	return nil
}

func (s *Service) CheckSlugAvailability(ctx context.Context, slug string) (bool, error) {
	s.log.Info(ctx, "business.core.workspaces.checkSlugAvailability")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.CheckSlugAvailability")
	defer span.End()

	// Check against restricted slugs first
	for _, restrictedSlug := range restrictedSlugs {
		if slug == restrictedSlug {
			return false, nil
		}
	}

	available, err := s.repo.CheckSlugAvailability(ctx, slug)
	if err != nil {
		span.RecordError(err)
		return false, err
	}

	span.AddEvent("slug availability checked.", trace.WithAttributes(
		attribute.String("slug", slug),
		attribute.Bool("available", available),
	))
	return available, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	s.log.Info(ctx, "business.core.workspaces.updateMemberRole")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.UpdateMemberRole")
	defer span.End()

	if role == "" {
		role = "member"
	}

	if err := s.repo.UpdateMemberRole(ctx, workspaceID, userID, role); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("workspace member role updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role", role),
	))
	return nil
}

// GetWorkspaceTerminology retrieves the terminology settings for a workspace.
func (s *Service) GetWorkspaceTerminology(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceTerminology, error) {
	s.log.Info(ctx, "business.core.workspaces.getTerminology")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.GetWorkspaceTerminology")
	defer span.End()

	terminology, err := s.repo.GetWorkspaceTerminology(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceTerminology{}, err
	}

	span.AddEvent("terminology retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return terminology, nil
}

// GetOrCreateWorkspaceTerminology retrieves terminology settings or creates default settings if none exist.
func (s *Service) GetOrCreateWorkspaceTerminology(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceTerminology, error) {
	s.log.Info(ctx, "business.core.workspaces.getOrCreateTerminology")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.GetOrCreateWorkspaceTerminology")
	defer span.End()

	terminology, err := s.repo.GetWorkspaceTerminology(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Start transaction to create default terminology
			tx, err := s.db.BeginTxx(ctx, nil)
			if err != nil {
				span.RecordError(err)
				return CoreWorkspaceTerminology{}, err
			}
			defer tx.Rollback()

			// Initialize default terminology
			if err := s.repo.InitializeWorkspaceTerminology(ctx, tx, workspaceID); err != nil {
				span.RecordError(err)
				return CoreWorkspaceTerminology{}, err
			}

			// Commit the transaction
			if err := tx.Commit(); err != nil {
				span.RecordError(err)
				return CoreWorkspaceTerminology{}, err
			}

			// Retrieve the newly created terminology
			terminology, err = s.repo.GetWorkspaceTerminology(ctx, workspaceID)
			if err != nil {
				span.RecordError(err)
				return CoreWorkspaceTerminology{}, err
			}

			span.AddEvent("default terminology created and retrieved.", trace.WithAttributes(
				attribute.String("workspace_id", workspaceID.String()),
			))
			return terminology, nil
		}

		span.RecordError(err)
		return CoreWorkspaceTerminology{}, err
	}

	span.AddEvent("terminology retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return terminology, nil
}

// UpdateWorkspaceTerminology updates the terminology settings for a workspace.
func (s *Service) UpdateWorkspaceTerminology(ctx context.Context, workspaceID uuid.UUID, terminology CoreWorkspaceTerminology) (CoreWorkspaceTerminology, error) {
	s.log.Info(ctx, "business.core.workspaces.updateTerminology")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.UpdateWorkspaceTerminology")
	defer span.End()

	// Ensure workspaceID is set correctly
	terminology.WorkspaceID = workspaceID

	updated, err := s.repo.UpdateWorkspaceTerminology(ctx, workspaceID, terminology)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceTerminology{}, err
	}

	span.AddEvent("terminology updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return updated, nil
}
