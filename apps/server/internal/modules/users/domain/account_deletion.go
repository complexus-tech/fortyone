package usersdomain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DeletedUserID is a shared, permanently inactive attribution identity. It is
// never a replacement account: the original account and its identities are
// erased, while organization-owned records retain a valid non-personal author.
var DeletedUserID = uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")

var ErrAccountDeletionUnavailable = errors.New("account deletion is temporarily unavailable")

type AccountDeletion struct {
	UserID      uuid.UUID
	RequestedAt time.Time
}

func (command AccountDeletion) Validate() error {
	if command.UserID == uuid.Nil || command.UserID == DeletedUserID || command.RequestedAt.IsZero() {
		return errors.New("account deletion requires a human account and request time")
	}
	return nil
}

type AccountDeletionConflict struct {
	Workspaces []string
}

func (conflict *AccountDeletionConflict) Error() string {
	return fmt.Sprintf("Before deleting your account, transfer administrator access or delete these workspaces: %s. Then try again.", strings.Join(conflict.Workspaces, ", "))
}
