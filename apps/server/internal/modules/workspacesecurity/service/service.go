package workspacesecurity

import (
	"context"
	"strings"
	"unicode/utf8"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	"github.com/google/uuid"
)

type Repository interface {
	CheckSession(context.Context, domain.Scope, domain.SessionIdentity) error
	Policy(context.Context, domain.Scope) (domain.Policy, error)
	UpdatePolicy(context.Context, domain.Scope, domain.PolicyUpdate) (domain.Policy, error)
	Sessions(context.Context, domain.Scope, *uuid.UUID, bool) (domain.SessionList, error)
	RevokeSession(context.Context, domain.Scope, uuid.UUID, string) error
	RevokeMember(context.Context, domain.Scope, uuid.UUID, string) error
	Audit(context.Context, domain.Scope, domain.AuditFilter) ([]domain.AuditEvent, error)
	RecordExport(context.Context, domain.Scope, int) error
	RecordAuditExport(context.Context, domain.Scope, int) error
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
func validScope(scope domain.Scope) error {
	if scope.ActorID == uuid.Nil || scope.WorkspaceID == uuid.Nil {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) CheckSession(ctx context.Context, scope domain.Scope, session domain.SessionIdentity) error {
	if err := validScope(scope); err != nil {
		return err
	}
	return s.repository.CheckSession(ctx, scope, session)
}
func (s *Service) Policy(ctx context.Context, scope domain.Scope) (domain.Policy, error) {
	if err := validScope(scope); err != nil {
		return domain.Policy{}, err
	}
	return s.repository.Policy(ctx, scope)
}
func (s *Service) UpdatePolicy(ctx context.Context, scope domain.Scope, input domain.PolicyUpdate) (domain.Policy, error) {
	if err := validScope(scope); err != nil {
		return domain.Policy{}, err
	}
	input, err := domain.NormalizePolicy(input)
	if err != nil {
		return domain.Policy{}, err
	}
	return s.repository.UpdatePolicy(ctx, scope, input)
}
func (s *Service) Sessions(ctx context.Context, scope domain.Scope, userID *uuid.UUID, includeRevoked bool) (domain.SessionList, error) {
	if err := validScope(scope); err != nil {
		return domain.SessionList{}, err
	}
	if userID != nil && *userID == uuid.Nil {
		return domain.SessionList{}, domain.ErrInvalid
	}
	return s.repository.Sessions(ctx, scope, userID, includeRevoked)
}
func (s *Service) RevokeSession(ctx context.Context, scope domain.Scope, id uuid.UUID, reason string) error {
	if err := validScope(scope); err != nil {
		return err
	}
	if id == uuid.Nil {
		return domain.ErrInvalid
	}
	reason = strings.TrimSpace(reason)
	if err := domain.ValidateReason(reason); err != nil {
		return err
	}
	return s.repository.RevokeSession(ctx, scope, id, reason)
}
func (s *Service) RevokeMember(ctx context.Context, scope domain.Scope, id uuid.UUID, reason string) error {
	if err := validScope(scope); err != nil {
		return err
	}
	if id == uuid.Nil {
		return domain.ErrInvalid
	}
	reason = strings.TrimSpace(reason)
	if err := domain.ValidateReason(reason); err != nil {
		return err
	}
	return s.repository.RevokeMember(ctx, scope, id, reason)
}
func (s *Service) Audit(ctx context.Context, scope domain.Scope, filter domain.AuditFilter) ([]domain.AuditEvent, error) {
	if err := validScope(scope); err != nil {
		return nil, err
	}
	if filter.Limit < 1 || filter.Limit > 10001 || !utf8.ValidString(filter.ResourceType) || len(filter.ResourceType) > 64 || (filter.ActorID != nil && *filter.ActorID == uuid.Nil) || (filter.ResourceID != nil && *filter.ResourceID == uuid.Nil) || (filter.From != nil && filter.To != nil && !filter.From.Before(*filter.To)) {
		return nil, domain.ErrInvalid
	}
	return s.repository.Audit(ctx, scope, filter)
}
func (s *Service) RecordExport(ctx context.Context, actorID, workspaceID uuid.UUID, taskCount int) error {
	scope := domain.Scope{ActorID: actorID, WorkspaceID: workspaceID}
	if err := validScope(scope); err != nil {
		return err
	}
	if taskCount < 0 || taskCount > 10000 {
		return domain.ErrInvalid
	}
	return s.repository.RecordExport(ctx, scope, taskCount)
}
func (s *Service) RecordAuditExport(ctx context.Context, scope domain.Scope, count int) error {
	if err := validScope(scope); err != nil {
		return err
	}
	if count < 0 || count > 10000 {
		return domain.ErrInvalid
	}
	return s.repository.RecordAuditExport(ctx, scope, count)
}
