package notifications

import (
	"errors"
	"testing"
	"time"

	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
)

func TestNotificationTypeAndEntityCompatibility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		notificationType NotificationType
		entityType       EntityType
	}{
		{NotificationTypeStoryUpdate, EntityTypeStory},
		{NotificationTypeStoryComment, EntityTypeComment},
		{NotificationTypeCommentReply, EntityTypeStory},
		{NotificationTypeObjectiveUpdate, EntityTypeObjective},
		{NotificationTypeKeyResultUpdate, EntityTypeKeyResult},
		{NotificationTypeMention, EntityTypeComment},
		{NotificationTypeFeedbackComment, EntityTypeFeedback},
		{NotificationTypeFeedbackStatusUpdate, EntityTypeFeedback},
		{NotificationTypeFeedbackUpdatePublished, EntityTypeFeedback},
		{NotificationTypeFeedbackItemMerged, EntityTypeFeedback},
		{NotificationTypeStrategyUpdate, EntityTypeStrategy},
	}
	for _, test := range tests {
		parsedType, err := ParseNotificationType("  " + string(test.notificationType) + "  ")
		if err != nil || parsedType != test.notificationType {
			t.Errorf("ParseNotificationType(%q) = %q, %v", test.notificationType, parsedType, err)
		}
		parsedEntity, err := ParseEntityType(" " + string(test.entityType) + " ")
		if err != nil || parsedEntity != test.entityType {
			t.Errorf("ParseEntityType(%q) = %q, %v", test.entityType, parsedEntity, err)
		}
		if !test.notificationType.SupportsEntity(test.entityType) {
			t.Errorf("%q should support %q", test.notificationType, test.entityType)
		}
	}
	if _, err := ParseNotificationType("arbitrary"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown notification type error = %v, want ErrInvalid", err)
	}
	if _, err := ParseEntityType("arbitrary"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown entity type error = %v, want ErrInvalid", err)
	}
	if NotificationTypeStoryUpdate.SupportsEntity(EntityTypeFeedback) {
		t.Fatal("story update unexpectedly supports feedback entity")
	}
}

func TestPreferenceDefaultsBackfillAndTypedPatch(t *testing.T) {
	t.Parallel()

	defaults := DefaultPreferences()
	for _, preferenceType := range preferenceTypes {
		channels, exists := defaults[preferenceType]
		if !exists || !channels.Email {
			t.Fatalf("default %q = %#v, want email enabled", preferenceType, channels)
		}
		if channels.InApp != preferenceType.SupportsInAppDelivery() {
			t.Fatalf("default %q in-app = %t, want %t", preferenceType, channels.InApp, preferenceType.SupportsInAppDelivery())
		}
	}

	partial := PreferenceSet{
		PreferenceTypeMention:        {Email: false, InApp: false},
		PreferenceTypeStrategyUpdate: {Email: true, InApp: true},
		PreferenceType("unknown"):    {Email: false, InApp: false},
	}.WithDefaults()
	if partial[PreferenceTypeMention].Email || partial[PreferenceTypeMention].InApp {
		t.Fatalf("explicit mention preference was not preserved: %#v", partial[PreferenceTypeMention])
	}
	if partial[PreferenceTypeStrategyUpdate].InApp {
		t.Fatal("email-only strategy preference enabled in-app delivery")
	}
	if _, exists := partial[PreferenceType("unknown")]; exists {
		t.Fatal("unknown persisted preference survived normalization")
	}
	if _, exists := partial[PreferenceTypeWeeklyDigest]; !exists {
		t.Fatal("missing preference was not backfilled")
	}

	patch := ChannelPatch{Email: platformpatch.Set(false)}
	email, specified := patch.Email.Value()
	if !specified || email == nil || *email {
		t.Fatalf("typed email patch = %v/%t", email, specified)
	}
	if patch.InApp.Specified() {
		t.Fatal("omitted in-app channel became specified")
	}
	normalized := ChannelPatch{InApp: platformpatch.Set(true)}.Normalized(PreferenceTypeWeeklyDigest)
	inApp, specified := normalized.InApp.Value()
	if !specified || inApp == nil || *inApp {
		t.Fatalf("email-only in-app patch = %v/%t, want explicit false", inApp, specified)
	}
}

func TestDomainCommandsRejectAmbiguousOrUnboundedIntent(t *testing.T) {
	t.Parallel()

	actorID, workspaceID, notificationID := uuid.New(), uuid.New(), uuid.New()
	access := WorkspaceAccess{ActorID: actorID, WorkspaceID: workspaceID}
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	validList := ListQuery{Access: access, Limit: MaximumPageSize + 1}
	if err := validList.Validate(); err != nil {
		t.Fatalf("valid lookahead page: %v", err)
	}
	for _, query := range []ListQuery{
		{Access: access, Limit: 0},
		{Access: access, Limit: MaximumPageSize + 2},
		{Access: access, Limit: 1, Offset: -1},
		{Access: access, Search: string(make([]byte, MaximumSearchBytes+1)), Limit: 1},
	} {
		if err := query.Validate(); !errors.Is(err, ErrInvalid) {
			t.Errorf("ListQuery(%#v) error = %v, want ErrInvalid", query, err)
		}
	}

	validMutation := NotificationMutation{
		Access: access, NotificationID: notificationID, Kind: NotificationMutationRead, At: now,
	}
	if err := validMutation.Validate(); err != nil {
		t.Fatalf("valid notification mutation: %v", err)
	}
	invalidMutation := validMutation
	invalidMutation.Kind = NotificationMutationKind("toggle")
	if err := invalidMutation.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("free-form notification mutation error = %v, want ErrInvalid", err)
	}

	validPreference := UpdatePreference{
		Access: access, Type: PreferenceTypeMention,
		Patch: ChannelPatch{Email: platformpatch.Set(false)}, At: now,
	}
	if err := validPreference.Validate(); err != nil {
		t.Fatalf("valid preference patch: %v", err)
	}
	validPreference.Patch = ChannelPatch{}
	if err := validPreference.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty preference patch error = %v, want ErrInvalid", err)
	}
}

func TestNewNotificationValidationRequiresTypedScopedResource(t *testing.T) {
	t.Parallel()

	valid := NewNotification{
		RecipientID: uuid.New(), WorkspaceID: uuid.New(),
		Type: NotificationTypeStoryUpdate, EntityType: EntityTypeStory,
		EntityID: uuid.New(), ActorID: uuid.New(), Title: "Story updated",
		Message: NotificationMessage{Template: "updated", Variables: map[string]Variable{}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid notification: %v", err)
	}
	for _, mutate := range []func(*NewNotification){
		func(notification *NewNotification) { notification.RecipientID = uuid.Nil },
		func(notification *NewNotification) { notification.WorkspaceID = uuid.Nil },
		func(notification *NewNotification) { notification.EntityID = uuid.Nil },
		func(notification *NewNotification) { notification.ActorID = uuid.Nil },
		func(notification *NewNotification) { notification.EntityType = EntityTypeFeedback },
		func(notification *NewNotification) { notification.Title = " " },
	} {
		input := valid
		mutate(&input)
		if err := input.Validate(); !errors.Is(err, ErrInvalid) {
			t.Errorf("invalid notification %#v error = %v, want ErrInvalid", input, err)
		}
	}
}
