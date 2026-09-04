package workspacebootstrap

import (
	"context"
	"errors"

	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

type userDirectory struct {
	service *users.Service
}

func NewUserDirectory(service *users.Service) workspaces.UserDirectory {
	return userDirectory{service: service}
}

func (directory userDirectory) GetWorkspaceUser(ctx context.Context, userID uuid.UUID) (workspaces.WorkspaceUser, error) {
	user, err := directory.service.GetUser(ctx, userID)
	if err != nil {
		return workspaces.WorkspaceUser{}, err
	}
	return workspaces.WorkspaceUser{
		Email: user.Email, FullName: user.FullName, Username: user.Username,
	}, nil
}

func (directory userDirectory) UpdateLastUsedWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return directory.service.UpdateUserWorkspace(ctx, userID, workspaceID)
}

type subscriptionManager struct {
	service interface {
		UpdateSubscriptionSeats(context.Context, uuid.UUID) error
		CancelSubscription(context.Context, uuid.UUID) error
	}
}

func NewSubscriptionManager(service *subscriptions.Service) workspaces.SubscriptionManager {
	return subscriptionManager{service: service}
}

func (manager subscriptionManager) UpdateWorkspaceSeats(ctx context.Context, workspaceID uuid.UUID) error {
	return manager.service.UpdateSubscriptionSeats(ctx, workspaceID)
}

func (manager subscriptionManager) CancelWorkspaceSubscription(ctx context.Context, workspaceID uuid.UUID) error {
	err := manager.service.CancelSubscription(ctx, workspaceID)
	if errors.Is(err, subscriptions.ErrNoActiveSubscriptionToChange) ||
		errors.Is(err, subscriptions.ErrSubscriptionAlreadyCanceled) {
		return nil
	}
	return err
}

type trialScheduler struct {
	service *tasks.Service
}

func NewTrialScheduler(service *tasks.Service) workspaces.TrialScheduler {
	return trialScheduler{service: service}
}

func (scheduler trialScheduler) ScheduleWorkspaceTrialStart(input workspaces.TrialStart) error {
	_, err := scheduler.service.EnqueueWorkspaceTrialStart(tasks.WorkspaceTrialStartPayload{
		UserID: input.UserID.String(), Email: input.Email, FullName: input.FullName,
		WorkspaceSlug: input.WorkspaceSlug, WorkspaceName: input.WorkspaceName,
	})
	return err
}
