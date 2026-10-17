package chatservice

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
)

type memberRepository interface {
	ListMembers(context.Context, domain.MemberSearch, uuid.UUID) (domain.MemberPage, error)
}

func (s *Service) ListMembers(ctx context.Context, query domain.MemberSearch) (domain.MemberPage, error) {
	if err := s.validateScope(ctx, query.Scope, false); err != nil {
		return domain.MemberPage{}, err
	}
	query.Query = strings.TrimSpace(query.Query)
	if query.ConversationID == uuid.Nil || utf8.RuneCountInString(query.Query) > 100 {
		return domain.MemberPage{}, domain.ErrInvalid
	}
	var err error
	query.Limit, err = boundedLimit(query.Limit, 15, 50)
	if err != nil {
		return domain.MemberPage{}, err
	}
	repository, ok := s.repo.(memberRepository)
	if !ok {
		return domain.MemberPage{}, errors.New("chat member search is unavailable")
	}
	return repository.ListMembers(ctx, query, s.mayaID)
}
