package workspaces

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mime/multipart"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Service errors
var (
	ErrNotFound               = errors.New("workspace not found")
	ErrMemberNotFound         = errors.New("member not found")
	ErrSlugTaken              = errors.New("workspace with this url already exists")
	ErrTx                     = errors.New("failed to create a workspace")
	ErrAlreadyWorkspaceMember = errors.New("user is already a member of this workspace")
)

var restrictedSlugs = []string{
	"admin", "internal", "qa", "staging", "ops", "team", "complexus",
	"dev", "test", "prod", "staging", "development", "testing",
	"production", "staff", "hr", "finance", "legal", "marketing",
	"sales", "support", "it", "security", "engineering", "design",
	"product", "auth", "fortyone", "forty-one",
}

// AttachmentsService provides workspace logo operations.
type AttachmentsService interface {
	UploadWorkspaceLogo(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, workspaceID uuid.UUID) (string, error)
	DeleteWorkspaceLogo(ctx context.Context, logoURL string) error
}

// Repository provides access to the users storage.
type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]CoreWorkspace, error)
	Create(ctx context.Context, tx *sqlx.Tx, newWorkspace CoreWorkspace) (CoreWorkspace, error)
	Update(ctx context.Context, workspaceID uuid.UUID, updates CoreWorkspace) (CoreWorkspace, error)
	Delete(ctx context.Context, workspaceID, deletedBy uuid.UUID) error
	Restore(ctx context.Context, workspaceID, restoredBy uuid.UUID) error
	GetWorkspaceAdminEmails(ctx context.Context, workspaceID, actorID uuid.UUID) ([]string, error)
	AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error
	AddMemberTx(ctx context.Context, tx *sqlx.Tx, workspaceID, userID uuid.UUID, role string) error
	Get(ctx context.Context, workspaceID, userID uuid.UUID) (CoreWorkspace, error)
	GetBySlug(ctx context.Context, slug string, userID uuid.UUID) (CoreWorkspace, error)
	RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error
	CheckSlugAvailability(ctx context.Context, slug string) (bool, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error
	GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSettings, error)
	UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, settings CoreWorkspaceSettings) (CoreWorkspaceSettings, error)
	InitializeWorkspaceSettings(ctx context.Context, tx *sqlx.Tx, workspaceID uuid.UUID) error
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
	subscriptions   *subscriptions.Service
	attachments     AttachmentsService
	cache           *cache.Service
	systemUserID    uuid.UUID
	publisher       *publisher.Publisher
}

// New constructs a new users service instance with the provided repository.
func New(log *logger.Logger, repo Repository, db *sqlx.DB, teams *teams.Service, stories *stories.Service, statuses *states.Service, users *users.Service, objectivestatus *objectivestatus.Service, subscriptions *subscriptions.Service, attachments AttachmentsService, cache *cache.Service, systemUserID uuid.UUID, publisher *publisher.Publisher) *Service {
	return &Service{
		repo:            repo,
		log:             log,
		db:              db,
		teams:           teams,
		stories:         stories,
		statuses:        statuses,
		users:           users,
		objectivestatus: objectivestatus,
		subscriptions:   subscriptions,
		attachments:     attachments,
		cache:           cache,
		systemUserID:    systemUserID,
		publisher:       publisher,
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

	// Add system user as member of the workspace
	if err := s.repo.AddMemberTx(ctx, tx, workspace.ID, s.systemUserID, "system"); err != nil {
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

	// Initialize workspace settings
	if err := s.repo.InitializeWorkspaceSettings(ctx, tx, workspace.ID); err != nil {
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

func (s *Service) Delete(ctx context.Context, workspaceID, deletedBy uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.delete")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, workspaceID, deletedBy); err != nil {
		span.RecordError(err)
		return err
	}

	// Publish notification events
	if err := s.publishWorkspaceDeletionEvents(ctx, workspaceID, deletedBy); err != nil {
		s.log.Error(ctx, "failed to publish workspace deletion events", "error", err, "workspace_id", workspaceID)
		// Don't fail the deletion if event publishing fails
	}

	span.AddEvent("workspace scheduled for deletion.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("deleted_by", deletedBy.String()),
	))
	return nil
}

func (s *Service) Restore(ctx context.Context, workspaceID, restoredBy uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.restore")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Restore")
	defer span.End()

	if err := s.repo.Restore(ctx, workspaceID, restoredBy); err != nil {
		span.RecordError(err)
		return err
	}

	// Publish notification events
	if err := s.publishWorkspaceRestoreEvents(ctx, workspaceID, restoredBy); err != nil {
		s.log.Error(ctx, "failed to publish workspace restore events", "error", err, "workspace_id", workspaceID)
		// Don't fail the restore if event publishing fails
	}

	span.AddEvent("workspace restored.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("restored_by", restoredBy.String()),
	))
	return nil
}

// publishWorkspaceDeletionEvents publishes both confirmation and notification events for workspace deletion
func (s *Service) publishWorkspaceDeletionEvents(ctx context.Context, workspaceID uuid.UUID, actorID uuid.UUID) error {
	// Get workspace details
	workspace, err := s.repo.Get(ctx, workspaceID, actorID)
	if err != nil {
		return fmt.Errorf("failed to get workspace details: %w", err)
	}

	// Get actor details
	actor, err := s.users.GetUser(ctx, actorID)
	if err != nil {
		return fmt.Errorf("failed to get actor details: %w", err)
	}

	actorName := actor.FullName
	if actorName == "" {
		actorName = actor.Username
	}

	// Publish confirmation event for the actor
	confirmationEvent := events.Event{
		Type: events.WorkspaceDeletionScheduledConfirmation,
		Payload: events.WorkspaceDeletionScheduledConfirmationPayload{
			WorkspaceID:   workspaceID,
			WorkspaceName: workspace.Name,
			WorkspaceSlug: workspace.Slug,
			ActorEmail:    actor.Email,
			ActorName:     actorName,
		},
		Timestamp: time.Now(),
		ActorID:   actorID,
	}

	if err := s.publisher.Publish(ctx, confirmationEvent); err != nil {
		s.log.Error(ctx, "failed to publish workspace deletion confirmation event", "error", err, "workspace_id", workspaceID)
		return fmt.Errorf("failed to publish workspace deletion confirmation event: %w", err)
	}

	// Get admin emails for notification
	adminEmails, err := s.repo.GetWorkspaceAdminEmails(ctx, workspaceID, actorID)
	if err != nil {
		s.log.Error(ctx, "failed to get workspace admin emails", "error", err, "workspace_id", workspaceID)
		// Continue without admin emails - the consumer will handle empty list
		adminEmails = []string{}
	}

	// Publish notification event for other admins
	notificationEvent := events.Event{
		Type: events.WorkspaceDeletionScheduledNotification,
		Payload: events.WorkspaceDeletionScheduledNotificationPayload{
			WorkspaceID:   workspaceID,
			WorkspaceName: workspace.Name,
			WorkspaceSlug: workspace.Slug,
			ActorID:       actorID,
			ActorName:     actorName,
			ActorEmail:    actor.Email,
			AdminEmails:   adminEmails,
		},
		Timestamp: time.Now(),
		ActorID:   actorID,
	}

	if err := s.publisher.Publish(ctx, notificationEvent); err != nil {
		s.log.Error(ctx, "failed to publish workspace deletion notification event", "error", err, "workspace_id", workspaceID)
		return fmt.Errorf("failed to publish workspace deletion notification event: %w", err)
	}

	return nil
}

// publishWorkspaceRestoreEvents publishes both confirmation and notification events for workspace restore
func (s *Service) publishWorkspaceRestoreEvents(ctx context.Context, workspaceID uuid.UUID, actorID uuid.UUID) error {
	// Get workspace details
	workspace, err := s.repo.Get(ctx, workspaceID, actorID)
	if err != nil {
		return fmt.Errorf("failed to get workspace details: %w", err)
	}

	// Get actor details
	actor, err := s.users.GetUser(ctx, actorID)
	if err != nil {
		return fmt.Errorf("failed to get actor details: %w", err)
	}

	actorName := actor.FullName
	if actorName == "" {
		actorName = actor.Username
	}

	// Publish confirmation event for the actor
	confirmationEvent := events.Event{
		Type: events.WorkspaceRestoredConfirmation,
		Payload: events.WorkspaceRestoredConfirmationPayload{
			WorkspaceID:   workspaceID,
			WorkspaceName: workspace.Name,
			WorkspaceSlug: workspace.Slug,
			ActorEmail:    actor.Email,
			ActorName:     actorName,
		},
		Timestamp: time.Now(),
		ActorID:   actorID,
	}

	if err := s.publisher.Publish(ctx, confirmationEvent); err != nil {
		s.log.Error(ctx, "failed to publish workspace restore confirmation event", "error", err, "workspace_id", workspaceID)
		return fmt.Errorf("failed to publish workspace restore confirmation event: %w", err)
	}

	// Get admin emails for notification
	adminEmails, err := s.repo.GetWorkspaceAdminEmails(ctx, workspaceID, actorID)
	if err != nil {
		s.log.Error(ctx, "failed to get workspace admin emails", "error", err, "workspace_id", workspaceID)
		// Continue without admin emails - the consumer will handle empty list
		adminEmails = []string{}
	}

	// Publish notification event for other admins
	notificationEvent := events.Event{
		Type: events.WorkspaceRestoredNotification,
		Payload: events.WorkspaceRestoredNotificationPayload{
			WorkspaceID:   workspaceID,
			WorkspaceName: workspace.Name,
			WorkspaceSlug: workspace.Slug,
			ActorID:       actorID,
			ActorName:     actorName,
			ActorEmail:    actor.Email,
			AdminEmails:   adminEmails,
		},
		Timestamp: time.Now(),
		ActorID:   actorID,
	}

	if err := s.publisher.Publish(ctx, notificationEvent); err != nil {
		s.log.Error(ctx, "failed to publish workspace restore notification event", "error", err, "workspace_id", workspaceID)
		return fmt.Errorf("failed to publish workspace restore notification event: %w", err)
	}

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

	// update subscription seats
	if err := s.subscriptions.UpdateSubscriptionSeats(ctx, workspaceID); err != nil {
		s.log.Error(ctx, "failed to update subscription seats", err)
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

func (s *Service) GetBySlug(ctx context.Context, slug string, userID uuid.UUID) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.getBySlug")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.GetBySlug")
	defer span.End()

	workspace, err := s.repo.GetBySlug(ctx, slug, userID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}

	span.AddEvent("workspace retrieved by slug.", trace.WithAttributes(
		attribute.String("slug", slug),
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("user_id", userID.String()),
	))
	return workspace, nil
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.removeMember")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.RemoveMember")
	defer span.End()

	// Get workspace before deletion to get the slug for cache invalidation
	workspace, err := s.repo.Get(ctx, workspaceID, userID)
	if err != nil {
		s.log.Error(ctx, "failed to get workspace for cache invalidation", "error", err)
	}

	if err := s.repo.RemoveMember(ctx, workspaceID, userID); err != nil {
		span.RecordError(err)
		return err
	}

	// Invalidate cache for this user if we got the workspace
	if workspace.Slug != "" {
		cacheKey := fmt.Sprintf("workspace:%s:user:%s", workspace.Slug, userID)
		if err := s.cache.Delete(ctx, cacheKey); err != nil {
			s.log.Error(ctx, "failed to invalidate workspace cache", "error", err)
		}
	}

	// update subscription seats
	if err := s.subscriptions.UpdateSubscriptionSeats(ctx, workspaceID); err != nil {
		s.log.Error(ctx, "failed to update subscription seats", err)
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

	// Invalidate cache for this user
	workspace, err := s.repo.Get(ctx, workspaceID, userID)
	if err != nil {
		s.log.Error(ctx, "failed to get workspace for cache invalidation", "error", err)
	} else {
		cacheKey := fmt.Sprintf("workspace:%s:user:%s", workspace.Slug, userID)
		if err := s.cache.Delete(ctx, cacheKey); err != nil {
			s.log.Error(ctx, "failed to invalidate workspace cache", "error", err)
		}
	}

	// update subscription seats
	if err := s.subscriptions.UpdateSubscriptionSeats(ctx, workspaceID); err != nil {
		s.log.Error(ctx, "failed to update subscription seats", err)
	}

	span.AddEvent("workspace member role updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role", role),
	))
	return nil
}

// GetWorkspaceSettings retrieves the settings for a workspace.
func (s *Service) GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSettings, error) {
	s.log.Info(ctx, "business.core.workspaces.getSettings")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.GetWorkspaceSettings")
	defer span.End()

	settings, err := s.repo.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceSettings{}, err
	}

	span.AddEvent("settings retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return settings, nil
}

// GetOrCreateWorkspaceSettings retrieves settings or creates default settings if none exist.
func (s *Service) GetOrCreateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSettings, error) {
	s.log.Info(ctx, "business.core.workspaces.getOrCreateSettings")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.GetOrCreateWorkspaceSettings")
	defer span.End()

	settings, err := s.repo.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Start transaction to create default settings
			tx, err := s.db.BeginTxx(ctx, nil)
			if err != nil {
				span.RecordError(err)
				return CoreWorkspaceSettings{}, err
			}
			defer tx.Rollback()

			// Initialize default settings
			if err := s.repo.InitializeWorkspaceSettings(ctx, tx, workspaceID); err != nil {
				span.RecordError(err)
				return CoreWorkspaceSettings{}, err
			}

			// Commit the transaction
			if err := tx.Commit(); err != nil {
				span.RecordError(err)
				return CoreWorkspaceSettings{}, err
			}

			// Retrieve the newly created settings
			settings, err = s.repo.GetWorkspaceSettings(ctx, workspaceID)
			if err != nil {
				span.RecordError(err)
				return CoreWorkspaceSettings{}, err
			}

			span.AddEvent("default settings created and retrieved.", trace.WithAttributes(
				attribute.String("workspace_id", workspaceID.String()),
			))
			return settings, nil
		}

		span.RecordError(err)
		return CoreWorkspaceSettings{}, err
	}

	span.AddEvent("settings retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return settings, nil
}

// UpdateWorkspaceSettings updates the settings for a workspace.
func (s *Service) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, settings CoreWorkspaceSettings) (CoreWorkspaceSettings, error) {
	s.log.Info(ctx, "business.core.workspaces.updateSettings")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.UpdateWorkspaceSettings")
	defer span.End()

	// Ensure workspaceID is set correctly
	settings.WorkspaceID = workspaceID

	updated, err := s.repo.UpdateWorkspaceSettings(ctx, workspaceID, settings)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceSettings{}, err
	}

	span.AddEvent("settings updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return updated, nil
}

// UploadWorkspaceLogo uploads a new logo for a workspace
func (s *Service) UploadWorkspaceLogo(ctx context.Context, workspaceID uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader, attachmentsService AttachmentsService) error {
	s.log.Info(ctx, "business.core.workspaces.uploadWorkspaceLogo")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.UploadWorkspaceLogo")
	defer span.End()

	// Get current workspace to check for existing logo
	workspace, err := s.repo.Get(ctx, workspaceID, s.systemUserID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Upload new logo
	logoURL, err := attachmentsService.UploadWorkspaceLogo(ctx, file, fileHeader, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Delete old logo if exists
	if workspace.AvatarURL != nil && *workspace.AvatarURL != "" {
		_ = attachmentsService.DeleteWorkspaceLogo(ctx, *workspace.AvatarURL)
	}

	// Update workspace's avatar URL using pointer-based update
	updates := CoreWorkspace{
		AvatarURL: &logoURL,
	}

	_, err = s.repo.Update(ctx, workspaceID, updates)
	if err != nil {
		span.RecordError(err)
		// Try to cleanup uploaded logo since DB update failed
		_ = attachmentsService.DeleteWorkspaceLogo(ctx, logoURL)
		return err
	}

	span.AddEvent("workspace logo updated", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("logo_url", logoURL),
	))

	return nil
}

// DeleteWorkspaceLogo removes the current workspace logo
func (s *Service) DeleteWorkspaceLogo(ctx context.Context, workspaceID uuid.UUID, attachmentsService AttachmentsService) error {
	s.log.Info(ctx, "business.core.workspaces.deleteWorkspaceLogo")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.DeleteWorkspaceLogo")
	defer span.End()

	// Get current workspace
	workspace, err := s.repo.Get(ctx, workspaceID, s.systemUserID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Delete from Azure if exists
	if workspace.AvatarURL != nil && *workspace.AvatarURL != "" {
		_ = attachmentsService.DeleteWorkspaceLogo(ctx, *workspace.AvatarURL)
	}

	// Clear avatar URL in database using pointer-based update
	avatarURL := ""
	updates := CoreWorkspace{
		AvatarURL: &avatarURL,
	}

	_, err = s.repo.Update(ctx, workspaceID, updates)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("workspace logo deleted", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))

	return nil
}
