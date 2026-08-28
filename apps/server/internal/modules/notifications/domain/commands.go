package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaximumPageSize    = 100
	MaximumSearchBytes = 512
)

type WorkspaceAccess struct {
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
}

func (access WorkspaceAccess) Validate() error {
	if access.ActorID == uuid.Nil || access.WorkspaceID == uuid.Nil {
		return fmt.Errorf("%w: actor and workspace are required", ErrInvalid)
	}
	return nil
}

type ListQuery struct {
	Access WorkspaceAccess
	Search string
	Limit  int
	Offset int
}

func (query ListQuery) Validate() error {
	if err := query.Access.Validate(); err != nil {
		return err
	}
	if query.Limit < 1 || query.Limit > MaximumPageSize+1 || query.Offset < 0 {
		return fmt.Errorf("%w: notification page is outside supported bounds", ErrInvalid)
	}
	if len(query.Search) > MaximumSearchBytes {
		return fmt.Errorf("%w: notification search is too long", ErrInvalid)
	}
	return nil
}

func (query ListQuery) Normalized() ListQuery {
	query.Search = strings.TrimSpace(query.Search)
	return query
}

type NotificationMutation struct {
	Access         WorkspaceAccess
	NotificationID uuid.UUID
	Kind           NotificationMutationKind
	At             time.Time
}

type NotificationMutationKind string

const (
	NotificationMutationRead   NotificationMutationKind = "read"
	NotificationMutationUnread NotificationMutationKind = "unread"
	NotificationMutationDelete NotificationMutationKind = "delete"
)

func (kind NotificationMutationKind) Valid() bool {
	switch kind {
	case NotificationMutationRead, NotificationMutationUnread, NotificationMutationDelete:
		return true
	default:
		return false
	}
}

func (command NotificationMutation) Validate() error {
	if err := command.Access.Validate(); err != nil {
		return err
	}
	if command.NotificationID == uuid.Nil {
		return fmt.Errorf("%w: notification ID is required", ErrInvalid)
	}
	if !command.Kind.Valid() {
		return fmt.Errorf("%w: notification mutation kind is required", ErrInvalid)
	}
	if command.At.IsZero() {
		return fmt.Errorf("%w: notification mutation time is required", ErrInvalid)
	}
	return nil
}

type WorkspaceMutation struct {
	Access WorkspaceAccess
	Kind   WorkspaceMutationKind
	At     time.Time
}

func (command WorkspaceMutation) Validate() error {
	if err := command.Access.Validate(); err != nil {
		return err
	}
	if !command.Kind.Valid() {
		return fmt.Errorf("%w: workspace notification mutation kind is required", ErrInvalid)
	}
	if command.At.IsZero() {
		return fmt.Errorf("%w: workspace notification mutation time is required", ErrInvalid)
	}
	return nil
}

type WorkspaceMutationKind string

const (
	WorkspaceMutationReadAll    WorkspaceMutationKind = "read_all"
	WorkspaceMutationDeleteAll  WorkspaceMutationKind = "delete_all"
	WorkspaceMutationDeleteRead WorkspaceMutationKind = "delete_read"
)

func (kind WorkspaceMutationKind) Valid() bool {
	switch kind {
	case WorkspaceMutationReadAll, WorkspaceMutationDeleteAll, WorkspaceMutationDeleteRead:
		return true
	default:
		return false
	}
}

type UpdatePreference struct {
	Access WorkspaceAccess
	Type   PreferenceType
	Patch  ChannelPatch
	At     time.Time
}

func (command UpdatePreference) Validate() error {
	if err := command.Access.Validate(); err != nil {
		return err
	}
	if !command.Type.Valid() {
		return fmt.Errorf("%w: unsupported notification preference type", ErrInvalid)
	}
	if command.Patch.Empty() {
		return fmt.Errorf("%w: at least one notification channel is required", ErrInvalid)
	}
	if value, specified := command.Patch.Email.Value(); specified && value == nil {
		return fmt.Errorf("%w: email preference cannot be cleared", ErrInvalid)
	}
	if value, specified := command.Patch.InApp.Value(); specified && value == nil {
		return fmt.Errorf("%w: in-app preference cannot be cleared", ErrInvalid)
	}
	if command.At.IsZero() {
		return fmt.Errorf("%w: preference update time is required", ErrInvalid)
	}
	return nil
}

type PortalAccess struct {
	ActorID    uuid.UUID
	PortalSlug string
}

func (access PortalAccess) Normalized() PortalAccess {
	access.PortalSlug = strings.TrimSpace(access.PortalSlug)
	return access
}

func (access PortalAccess) Validate() error {
	if access.ActorID == uuid.Nil || strings.TrimSpace(access.PortalSlug) == "" {
		return fmt.Errorf("%w: actor and portal are required", ErrInvalid)
	}
	return nil
}

type PortalListQuery struct {
	Access     PortalAccess
	UnreadOnly bool
	Limit      int
	Offset     int
}

func (query PortalListQuery) Validate() error {
	if err := query.Access.Validate(); err != nil {
		return err
	}
	if query.Limit < 1 || query.Limit > MaximumPageSize+1 || query.Offset < 0 {
		return fmt.Errorf("%w: portal notification page is outside supported bounds", ErrInvalid)
	}
	return nil
}

type PortalNotificationMutation struct {
	Access         PortalAccess
	NotificationID uuid.UUID
	At             time.Time
}

func (command PortalNotificationMutation) Validate() error {
	if err := command.Access.Validate(); err != nil {
		return err
	}
	if command.NotificationID == uuid.Nil {
		return fmt.Errorf("%w: notification ID is required", ErrInvalid)
	}
	if command.At.IsZero() {
		return fmt.Errorf("%w: portal notification mutation time is required", ErrInvalid)
	}
	return nil
}

type PortalMutation struct {
	Access PortalAccess
	At     time.Time
}

func (command PortalMutation) Validate() error {
	if err := command.Access.Validate(); err != nil {
		return err
	}
	if command.At.IsZero() {
		return fmt.Errorf("%w: portal notification mutation time is required", ErrInvalid)
	}
	return nil
}
