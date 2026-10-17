package chatservice

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
)

func (s *Service) Unread(ctx context.Context, scope domain.ActorScope) (domain.UnreadSummary, error) {
	if err := s.validateScope(ctx, scope, false); err != nil {
		return domain.UnreadSummary{}, err
	}
	return s.repo.Unread(ctx, scope)
}

func (s *Service) UnreadThreads(ctx context.Context, query domain.UnreadThreadQuery) (domain.UnreadThreadPage, error) {
	if err := s.validateScope(ctx, query.Scope, false); err != nil {
		return domain.UnreadThreadPage{}, err
	}
	if query.ConversationID != nil && *query.ConversationID == uuid.Nil {
		return domain.UnreadThreadPage{}, domain.ErrInvalid
	}
	var err error
	query.Limit, err = boundedLimit(query.Limit, 50, 100)
	if err != nil {
		return domain.UnreadThreadPage{}, err
	}
	return s.repo.UnreadThreads(ctx, query)
}

func (s *Service) MarkAllThreadsRead(ctx context.Context, scope domain.ActorScope) error {
	if err := s.validateScope(ctx, scope, true); err != nil {
		return err
	}
	// Keep the original boundaries across transaction retries so replies arriving
	// while this request runs remain unread. MarkRead rechecks each target's access.
	targets, err := s.repo.UnreadThreadReadTargets(ctx, scope)
	if err != nil || len(targets) == 0 {
		return err
	}
	return s.mutate(ctx, func(unit context.Context) error {
		for _, target := range targets {
			if _, err := s.repo.MarkRead(unit, domain.ReadCommand{
				Scope: scope, ConversationID: target.ConversationID,
				ThreadRootID: &target.ThreadRootID, ThroughMessageID: target.ThroughMessageID,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}
