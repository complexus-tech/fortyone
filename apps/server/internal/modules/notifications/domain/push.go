package notifications

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PushPlatform string

const (
	PushPlatformIOS     PushPlatform = "ios"
	PushPlatformAndroid PushPlatform = "android"
)

func (platform PushPlatform) Valid() bool {
	return platform == PushPlatformIOS || platform == PushPlatformAndroid
}

type RegisterPushDevice struct {
	UserID   uuid.UUID
	Token    string
	Platform PushPlatform
}

func (command RegisterPushDevice) Validate() error {
	if command.UserID == uuid.Nil || !command.Platform.Valid() {
		return fmt.Errorf("%w: user and supported platform are required", ErrInvalid)
	}
	token := strings.TrimSpace(command.Token)
	if len(token) < 20 || len(token) > 512 ||
		(!strings.HasPrefix(token, "ExponentPushToken[") && !strings.HasPrefix(token, "ExpoPushToken[")) {
		return fmt.Errorf("%w: invalid Expo push token", ErrInvalid)
	}
	return nil
}

type PushDevice struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	Platform  PushPlatform
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PushDelivery struct {
	NotificationID uuid.UUID
	RecipientID    uuid.UUID
	WorkspaceID    uuid.UUID
	WorkspaceSlug  string
	EntityType     EntityType
	EntityID       uuid.UUID
	Title          string
	Message        json.RawMessage
	Tokens         []string
}
