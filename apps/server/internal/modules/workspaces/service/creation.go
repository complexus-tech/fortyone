package workspaces

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) Create(ctx context.Context, input CoreWorkspace, userID uuid.UUID) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.create")
	ctx, span := startSpan(ctx, "business.core.workspaces.Create")
	defer span.End()
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	if _, restricted := restrictedSlugs[input.Slug]; restricted {
		return CoreWorkspace{}, ErrRestrictedSlug
	}

	var workspace CoreWorkspace
	var team CreatedTeam
	var operationErr error
	err := s.unitOfWork.WithinTransaction(ctx, func(transaction Transaction) error {
		workspace, operationErr = transaction.CreateWorkspace(ctx, input, userID)
		if operationErr != nil {
			return operationErr
		}
		if operationErr = transaction.AddWorkspaceMember(ctx, workspace.ID, userID, "admin"); operationErr != nil {
			return operationErr
		}
		team, operationErr = transaction.CreateTeam(ctx, DefaultTeam{
			Name: "Team 1", Color: workspace.Color, Code: "TM", Workspace: workspace.ID,
		})
		if operationErr != nil {
			return operationErr
		}
		if operationErr = transaction.AddTeamMember(ctx, team.ID, userID, workspace.ID); operationErr != nil {
			return operationErr
		}

		// This convenience pointer is not an integrity boundary. A failure must
		// not invalidate the otherwise complete workspace membership graph.
		if updateErr := transaction.UpdateLastUsedWorkspace(ctx, userID, workspace.ID); updateErr != nil {
			s.log.Error(ctx, "failed to update user's last used workspace", "error", updateErr)
		}
		operationErr = transaction.InitializeWorkspaceSettings(ctx, workspace.ID)
		return operationErr
	})
	if err != nil {
		span.RecordError(err)
		if operationErr != nil {
			return CoreWorkspace{}, operationErr
		}
		return CoreWorkspace{}, ErrTx
	}

	if err := s.seedContent.CreateWorkspaceSeedContent(ctx, workspace.ID, team.ID, userID); err != nil {
		s.log.Error(ctx, "failed to create workspace seed content", "error", err)
	}
	s.enqueueTrialStart(ctx, workspace, userID)
	span.AddEvent("workspace created.", trace.WithAttributes(attribute.String("workspace_id", workspace.ID.String())))
	return workspace, nil
}

func (s *Service) enqueueTrialStart(ctx context.Context, workspace CoreWorkspace, userID uuid.UUID) {
	user, err := s.users.GetWorkspaceUser(ctx, userID)
	if err != nil {
		s.log.Error(ctx, "failed to get user details for trial task", "error", err, "user_id", userID)
		return
	}
	err = s.trials.ScheduleWorkspaceTrialStart(TrialStart{
		UserID: userID, Email: user.Email, FullName: user.FullName,
		WorkspaceSlug: workspace.Slug, WorkspaceName: workspace.Name,
	})
	if err != nil {
		s.log.Error(ctx, "failed to enqueue workspace trial task", "error", err)
	}
}
