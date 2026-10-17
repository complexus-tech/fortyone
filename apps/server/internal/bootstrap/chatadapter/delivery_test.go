package chatadapter

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	chat "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/google/uuid"
)

type eventStoreStub struct {
	mu                  sync.Mutex
	renews              int
	renewErr            error
	readLookups         int
	memberLookups       int
	claimLimits         []int
	events              []chat.Event
	recipients          []uuid.UUID
	targets             []chat.NotificationRecipient
	content             chat.NotificationContent
	contentErr          error
	contentLookups      int
	completed, released int
}

func (s *eventStoreStub) ClaimEvents(_ context.Context, limit int) ([]chat.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.claimLimits = append(s.claimLimits, limit)
	if len(s.events) == 0 {
		return nil, nil
	}
	count := min(limit, len(s.events))
	events := append([]chat.Event(nil), s.events[:count]...)
	s.events = s.events[count:]
	return events, nil
}
func (s *eventStoreStub) RenewEventLease(context.Context, uuid.UUID, uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.renews++
	return s.renewErr
}
func (s *eventStoreStub) ReadRecipientIDs(_ context.Context, event chat.Event) ([]uuid.UUID, error) {
	s.readLookups++
	for _, id := range s.recipients {
		if id == event.ActorID {
			return []uuid.UUID{id}, nil
		}
	}
	return nil, nil
}
func (s *eventStoreStub) CompleteEvent(context.Context, uuid.UUID, uuid.UUID) error {
	s.completed++
	return nil
}
func (s *eventStoreStub) ReleaseEvent(context.Context, uuid.UUID, uuid.UUID) error {
	s.released++
	return nil
}
func (s *eventStoreStub) RecipientIDs(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	s.memberLookups++
	return s.recipients, nil
}
func (s *eventStoreStub) NotificationRecipients(context.Context, chat.Event) ([]chat.NotificationRecipient, error) {
	return s.targets, nil
}
func (s *eventStoreStub) NotificationContent(context.Context, chat.Event) (chat.NotificationContent, error) {
	s.contentLookups++
	return s.content, s.contentErr
}

type deliveryStub struct {
	mu          sync.Mutex
	updates     []publisher.ChatUpdate
	notices     []notifications.NewNotification
	alerts      []publisher.ChatNotification
	receipts    map[string]uuid.UUID
	fail        bool
	alertErr    error
	createErr   error
	alertsMuted bool
	readAt      *time.Time
}

func (s *deliveryStub) PublishChatUpdated(_ context.Context, u publisher.ChatUpdate) error {
	if s.fail {
		return errors.New("redis unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updates = append(s.updates, u)
	return nil
}
func (s *deliveryStub) Create(_ context.Context, n notifications.NewNotification) (notifications.Notification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notices = append(s.notices, n)
	if s.createErr != nil {
		return notifications.Notification{}, s.createErr
	}
	if s.receipts == nil {
		s.receipts = make(map[string]uuid.UUID)
	}
	if s.receipts[n.DedupeKey] == uuid.Nil {
		s.receipts[n.DedupeKey] = uuid.New()
	}
	return notifications.Notification{ID: s.receipts[n.DedupeKey], Title: n.Title, Message: n.Message, InAppEnabled: !s.alertsMuted, CreatedAt: n.OccurredAt, ReadAt: s.readAt}, nil
}
func (s *deliveryStub) PublishChatNotification(_ context.Context, notification publisher.ChatNotification) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.alertErr != nil {
		return s.alertErr
	}
	s.alerts = append(s.alerts, notification)
	return nil
}
func TestDeliveryScopesHintsAndDeduplicatesTargets(t *testing.T) {
	actor, recipient := uuid.New(), uuid.New()
	event := chat.Event{ID: uuid.New(), LeaseToken: uuid.New(), WorkspaceID: uuid.New(), ConversationID: uuid.New(), MessageID: uuid.New(), ActorID: actor, Type: "message.created", Version: 1, CreatedAt: time.Now()}
	store := &eventStoreStub{events: []chat.Event{event}, recipients: []uuid.UUID{actor, recipient}, targets: []chat.NotificationRecipient{{UserID: recipient, Reason: "mention"}, {UserID: recipient, Reason: "reply"}, {UserID: actor, Reason: "dm"}}}
	delivery := &deliveryStub{}
	store.content = chat.NotificationContent{ConversationKind: "team", ConversationName: "Product", AuthorName: "Sarah", Text: "Please review the launch plan"}
	count, err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery}).Dispatch(t.Context(), 50)
	if err != nil || count != 1 || store.completed != 1 {
		t.Fatalf("dispatch = %d, %v", count, err)
	}
	if len(delivery.updates) != 2 || len(delivery.notices) != 1 || len(delivery.alerts) != 1 {
		t.Fatalf("unexpected delivery fanout: %+v", delivery)
	}
	for _, hint := range delivery.updates {
		if hint.WorkspaceID != event.WorkspaceID || hint.ConversationID != event.ConversationID || hint.EventID != event.ID {
			t.Fatal("hint escaped source scope")
		}
	}
	notice := delivery.notices[0]
	if notice.RecipientID != recipient || notice.EntityID != event.MessageID || notice.Type != notifications.NotificationTypeChatMessage {
		t.Fatal("incorrect notification scope")
	}
	if err := notice.Validate(); err != nil {
		t.Fatal(err)
	}
	if notice.DedupeKey == "" || notice.Title != "Product" || notice.Message.Template != "{message}" || notice.Message.Variables["message"] != (notifications.Variable{Value: "Sarah: Please review the launch plan", Type: "plain_text"}) {
		t.Fatal("notification receipt or source message missing")
	}
	alert := delivery.alerts[0]
	if alert.Title != notice.Title || alert.Body != notice.Message.Variables["message"].Value || store.contentLookups != 1 {
		t.Fatal("live alert must preserve the source and preview from one lookup per event")
	}
	if alert.EventID != event.ID || alert.WorkspaceID != event.WorkspaceID || alert.UserID != recipient || alert.ActorID != actor || alert.ConversationID != event.ConversationID || alert.MessageID != event.MessageID || !alert.OccurredAt.Equal(event.CreatedAt) || alert.NotificationID == uuid.Nil {
		t.Fatal("chat alert lost its source, recipient or receipt identity")
	}
}

func TestAllIncomingMessageKindsPublishChatAlerts(t *testing.T) {
	for _, reason := range []string{"message", "dm", "mention", "reply"} {
		t.Run(reason, func(t *testing.T) {
			thread := uuid.New()
			event := chat.Event{ID: uuid.New(), WorkspaceID: uuid.New(), ConversationID: uuid.New(), MessageID: uuid.New(), ActorID: uuid.New(), CreatedAt: time.Now().UTC(), Type: "message.created"}
			if reason == "reply" {
				event.ThreadRootID = &thread
			}
			store := &eventStoreStub{targets: []chat.NotificationRecipient{{UserID: uuid.New(), Reason: reason}}}
			delivery := &deliveryStub{}
			if err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery}).deliver(t.Context(), event); err != nil {
				t.Fatal(err)
			}
			if len(delivery.alerts) != 1 || delivery.alerts[0].ThreadRootID != event.ThreadRootID {
				t.Fatal("new message did not produce a correctly routed chat alert")
			}
			if delivery.notices[0].Message.EmailSuppressed != (reason == "message") {
				t.Fatal("new team-message interruptions must not expand the existing email audience")
			}
		})
	}
}

func TestChatAlertPublicationFailureRetriesTheSameReceipt(t *testing.T) {
	event := chat.Event{ID: uuid.New(), WorkspaceID: uuid.New(), ConversationID: uuid.New(), MessageID: uuid.New(), ActorID: uuid.New(), CreatedAt: time.Now().UTC(), Type: "message.created"}
	store := &eventStoreStub{events: []chat.Event{event}, targets: []chat.NotificationRecipient{{UserID: uuid.New(), Reason: "message"}}}
	delivery := &deliveryStub{alertErr: errors.New("redis unavailable")}
	dispatcher := Dispatcher{Store: store, Updates: delivery, Notifications: delivery}
	count, err := dispatcher.Dispatch(t.Context(), 1)
	if err == nil || count != 0 || store.completed != 0 || store.released != 1 {
		t.Fatal("failed live delivery must retain retryable outbox intent")
	}
	delivery.alertErr = nil
	store.events = []chat.Event{event}
	count, err = dispatcher.Dispatch(t.Context(), 1)
	if err != nil || count != 1 || len(delivery.alerts) != 1 || len(delivery.receipts) != 1 || len(delivery.notices) != 2 {
		t.Fatalf("retry did not recover one receipt: count=%d err=%v", count, err)
	}
	store.events = []chat.Event{event}
	if _, err := dispatcher.Dispatch(t.Context(), 1); err != nil {
		t.Fatal(err)
	}
	if delivery.alerts[0].NotificationID != delivery.alerts[1].NotificationID || delivery.alerts[0].EventID != delivery.alerts[1].EventID {
		t.Fatal("at-least-once delivery must retain stable client dedupe identity")
	}
}

func TestSuppressedChatReceiptsDoNotPublishAlerts(t *testing.T) {
	now := time.Now()
	for name, delivery := range map[string]*deliveryStub{
		"read": {readAt: &now}, "muted": {alertsMuted: true},
		"revoked": {createErr: notifications.ErrForbidden}, "deleted": {createErr: notifications.ErrNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			store := &eventStoreStub{targets: []chat.NotificationRecipient{{UserID: uuid.New(), Reason: "message"}}}
			if err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery}).deliver(t.Context(), chat.Event{}); err != nil {
				t.Fatal(err)
			}
			if len(delivery.alerts) != 0 {
				t.Fatal("suppressed receipt produced an interruption")
			}
		})
	}
}
func TestFailedDeliveryReleasesLeaseWithoutCompleting(t *testing.T) {
	store := &eventStoreStub{events: []chat.Event{{ID: uuid.New(), LeaseToken: uuid.New()}}, recipients: []uuid.UUID{uuid.New()}}
	delivery := &deliveryStub{fail: true}
	count, err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery}).Dispatch(t.Context(), 50)
	if err == nil || count != 0 || store.completed != 0 || store.released != 1 {
		t.Fatalf("failed delivery was not retryable: count=%d completed=%d released=%d err=%v", count, store.completed, store.released, err)
	}
}
func TestRevokedConversationHasNoDelivery(t *testing.T) {
	store := &eventStoreStub{events: []chat.Event{{ID: uuid.New(), LeaseToken: uuid.New()}}}
	delivery := &deliveryStub{}
	_, err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery}).Dispatch(t.Context(), 1)
	if err != nil || len(delivery.updates) != 0 || len(delivery.notices) != 0 || store.completed != 1 {
		t.Fatal("revoked event should retire without delivery")
	}
}

func TestReadDeliveryAuthorizesOnlyReaderAndSkipsNotifications(t *testing.T) {
	actor := uuid.New()
	store := &eventStoreStub{events: []chat.Event{{ID: uuid.New(), LeaseToken: uuid.New(), ActorID: actor, Type: "read.updated"}}, recipients: []uuid.UUID{actor, uuid.New()}}
	delivery := &deliveryStub{}
	count, err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery}).Dispatch(t.Context(), 50)
	if err != nil || count != 1 || store.readLookups != 1 || store.memberLookups != 0 || len(delivery.updates) != 1 || len(delivery.notices) != 0 {
		t.Fatalf("read fanout: count=%d read=%d members=%d hints=%d err=%v", count, store.readLookups, store.memberLookups, len(delivery.updates), err)
	}
}

type slowHints struct {
	deliveryStub
	active atomic.Int32
	peak   atomic.Int32
	delay  time.Duration
}

func (s *slowHints) PublishChatUpdated(ctx context.Context, update publisher.ChatUpdate) error {
	active := s.active.Add(1)
	defer s.active.Add(-1)
	for {
		peak := s.peak.Load()
		if active <= peak || s.peak.CompareAndSwap(peak, active) {
			break
		}
	}
	timer := time.NewTimer(s.delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	return s.deliveryStub.PublishChatUpdated(ctx, update)
}
func TestLargeFanoutIsBoundedAndKeepsOnlyActiveEventLeased(t *testing.T) {
	store := &eventStoreStub{events: []chat.Event{{ID: uuid.New(), LeaseToken: uuid.New()}, {ID: uuid.New(), LeaseToken: uuid.New()}}}
	for range 1000 {
		store.recipients = append(store.recipients, uuid.New())
	}
	delivery := &slowHints{delay: time.Millisecond}
	count, err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery, leaseInterval: 5 * time.Millisecond}).Dispatch(t.Context(), 50)
	if err != nil || count != 2 || len(delivery.updates) != 2000 {
		t.Fatalf("large delivery count=%d hints=%d err=%v", count, len(delivery.updates), err)
	}
	if delivery.peak.Load() > 16 || delivery.peak.Load() < 2 {
		t.Fatalf("fanout concurrency=%d", delivery.peak.Load())
	}
	if store.renews == 0 {
		t.Fatal("active lease must be renewed during long fanout")
	}
	for _, limit := range store.claimLimits {
		if limit != 1 {
			t.Fatal("waiting events must not age under an already-started lease")
		}
	}
}
func TestLostLeaseCancelsFanoutAndDoesNotComplete(t *testing.T) {
	store := &eventStoreStub{events: []chat.Event{{ID: uuid.New(), LeaseToken: uuid.New()}}, recipients: []uuid.UUID{uuid.New()}, renewErr: errors.New("lease replaced")}
	delivery := &slowHints{delay: time.Second}
	count, err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery, leaseInterval: time.Millisecond}).Dispatch(t.Context(), 1)
	if err == nil || count != 0 || store.completed != 0 || store.released != 1 || len(delivery.updates) != 0 {
		t.Fatalf("stale lease completed: count=%d completed=%d released=%d hints=%d err=%v", count, store.completed, store.released, len(delivery.updates), err)
	}
}
