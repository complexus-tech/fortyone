package chatsessions

import chatsessionsdomain "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"

func FinalizeMessageWriteTransition(current, incoming []any, operation MessageWriteOperation) ([]any, error) {
	return chatsessionsdomain.FinalizeMessageWriteTransition(current, incoming, operation)
}
