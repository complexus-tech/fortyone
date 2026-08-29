package slackadapter

import (
	"context"
	"errors"

	messagingbudget "github.com/complexus-tech/projects-api/internal/modules/messaging/budget"
	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/google/uuid"
)

type CallLimiterBackend interface {
	Admit(context.Context, messagingbudget.AdmissionInput) (messagingbudget.AdmissionDecision, error)
}

type CallLimiter struct {
	backend CallLimiterBackend
}

func NewCallLimiter(backend CallLimiterBackend) *CallLimiter {
	if backend == nil {
		return nil
	}
	return &CallLimiter{backend: backend}
}

func (adapter *CallLimiter) Admit(ctx context.Context, input slack.AssistantAdmissionInput) (slack.AssistantAdmissionDecision, error) {
	decision, err := adapter.backend.Admit(ctx, messagingbudget.AdmissionInput{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalEventID: input.ExternalEventID,
	})
	return slack.AssistantAdmissionDecision{
		Allowed: decision.Allowed, Duplicate: decision.Duplicate,
		LimitedScope: decision.LimitedScope, RetryAfter: decision.RetryAfter,
		UserCount: decision.UserCount, WorkspaceCount: decision.WorkspaceCount,
	}, err
}

type UsageBudgetBackend interface {
	Check(context.Context, uuid.UUID, int64) (messagingrepository.DailyUsageSnapshot, error)
	Record(context.Context, messagingrepository.DailyUsageRecordInput, int64) (messagingrepository.DailyUsageSnapshot, error)
}

type UsageBudget struct {
	backend UsageBudgetBackend
}

func NewUsageBudget(backend UsageBudgetBackend) *UsageBudget {
	if backend == nil {
		return nil
	}
	return &UsageBudget{backend: backend}
}

func (adapter *UsageBudget) Check(ctx context.Context, workspaceID uuid.UUID, limit int64) (slack.DailyUsageSnapshot, error) {
	snapshot, err := adapter.backend.Check(ctx, workspaceID, limit)
	return mapUsageSnapshot(snapshot), mapUsageError(err)
}

func (adapter *UsageBudget) Record(ctx context.Context, input slack.DailyUsageRecordInput, limit int64) (slack.DailyUsageSnapshot, error) {
	snapshot, err := adapter.backend.Record(ctx, messagingrepository.DailyUsageRecordInput{
		InboundEventID: input.InboundEventID, WorkspaceID: input.WorkspaceID,
		Provider: input.Provider, ExternalWorkspaceID: input.ExternalWorkspaceID,
		ExternalEventID: input.ExternalEventID, AttemptCount: input.AttemptCount,
		Usage: messagingdomain.Usage{
			InputTokens: input.Usage.InputTokens, OutputTokens: input.Usage.OutputTokens,
			TotalTokens: input.Usage.TotalTokens,
		},
	}, limit)
	return mapUsageSnapshot(snapshot), mapUsageError(err)
}

func mapUsageSnapshot(snapshot messagingrepository.DailyUsageSnapshot) slack.DailyUsageSnapshot {
	return slack.DailyUsageSnapshot{
		InputTokens: snapshot.InputTokens, OutputTokens: snapshot.OutputTokens,
		TotalTokens: snapshot.TotalTokens, RequestCount: snapshot.RequestCount,
		Limit: snapshot.Limit, Remaining: snapshot.Remaining, Allowed: snapshot.Allowed,
	}
}

func mapUsageError(err error) error {
	if errors.Is(err, messagingrepository.ErrDailyWorkspaceTokenLimit) {
		return errors.Join(slack.ErrDailyWorkspaceTokenLimit, err)
	}
	return err
}

var (
	_ slack.AssistantCallLimiter = (*CallLimiter)(nil)
	_ slack.AssistantUsageBudget = (*UsageBudget)(nil)
)
