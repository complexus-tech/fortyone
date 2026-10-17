package chatrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) ListMembers(ctx context.Context, query domain.MemberSearch, mayaID uuid.UUID) (domain.MemberPage, error) {
	result := domain.MemberPage{Items: []domain.Person{}}
	if _, err := r.authorize(ctx, query.Scope, query.ConversationID, false); err != nil {
		return result, err
	}
	rows, err := r.query(ctx).ListConversationMembers(ctx, chatsql.ListConversationMembersParams{WorkspaceID: query.Scope.WorkspaceID, ConversationID: query.ConversationID, UserID: query.Scope.Actor.PrincipalID, MayaID: mayaID, Search: query.Query, PageLimit: int32(query.Limit)})
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		person, err := decode[domain.Person](row)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, person)
	}
	return result, nil
}
