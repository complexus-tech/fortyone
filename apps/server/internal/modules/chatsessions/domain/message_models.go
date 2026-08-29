package chatsessionsdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrMessageWriteConflict     = errors.New("chat transcript changed before this write could be applied")
	ErrMessageWriteApprovalOpen = errors.New("chat transcript has an unresolved mutation approval")
	ErrMessageWriteInvalid      = errors.New("chat transcript write is invalid")
)

type MessageWriteOperation string

const (
	MessageWriteAppend     MessageWriteOperation = "append"
	MessageWriteRegenerate MessageWriteOperation = "regenerate"
	MessageWriteApproval   MessageWriteOperation = "approval"
)

type CoreChatSession struct {
	ID          string     `json:"id"`
	UserID      uuid.UUID  `json:"userId"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type CoreNewChatSession struct {
	ID          string    `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Title       string    `json:"title"`
	Messages    []any     `json:"messages"`
}

type BeginMessageWriteParams struct {
	Session         CoreChatSession
	Messages        []any
	Operation       MessageWriteOperation
	TargetMessageID string
}

type CoreMessageWriteReservation struct {
	Generation int64     `json:"generation"`
	Token      uuid.UUID `json:"token"`
	Messages   []any     `json:"messages,omitempty"`
}

type FinalizeMessageWriteParams struct {
	SessionID   string
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Messages    []any
	Generation  int64
	Token       uuid.UUID
}

type CoreMessageWriteResult struct {
	Applied bool `json:"applied"`
}

type RecoverMutationApprovalOutputParams struct {
	SessionID   string
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	ToolCallID  string
	Fingerprint string
}
