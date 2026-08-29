package chatsessions

import chatsessionsdomain "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"

func ReserveMessageWrite(current, incoming []any, operation MessageWriteOperation) ([]any, error) {
	return chatsessionsdomain.ReserveMessageWrite(current, incoming, operation)
}

func CanonicalMessageWriteResponse(persisted, request []any) ([]any, bool, error) {
	return chatsessionsdomain.CanonicalMessageWriteResponse(persisted, request)
}

func ReserveMessageWriteForTarget(current, incoming []any, operation MessageWriteOperation, targetMessageID string) ([]any, error) {
	return chatsessionsdomain.ReserveMessageWriteForTarget(current, incoming, operation, targetMessageID)
}
