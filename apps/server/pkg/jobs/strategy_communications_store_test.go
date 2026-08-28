package jobs

import (
	"context"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type strategyCommunicationsStoreStub struct {
	adminPages            []objectivesdomain.StrategyCommunicationRecipientPage
	adminCursors          []*objectivesdomain.StrategyCommunicationCursor
	weeklyPages           []objectivesdomain.StrategyCommunicationRecipientPage
	weeklyCursors         []*objectivesdomain.StrategyCommunicationCursor
	weeklySignals         map[uuid.UUID][]objectivesdomain.StrategyWeeklySignalPage
	weeklySignalCalls     []strategyWeeklySignalCall
	foundation            objectivesdomain.StrategyCommunicationFoundation
	foundationPeriodStart time.Time
	foundationPeriodEnd   time.Time
	monthlySummary        objectivesdomain.StrategyCommunicationMonthlySummary
}

type strategyWeeklySignalCall struct {
	StaleBefore time.Time
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Cursor      *objectivesdomain.StrategyWeeklySignalCursor
	Limit       int
}

func (stub *strategyCommunicationsStoreStub) ListStrategyCommunicationAdministrators(
	_ context.Context,
	cursor *objectivesdomain.StrategyCommunicationCursor,
	_ int,
) (objectivesdomain.StrategyCommunicationRecipientPage, error) {
	stub.adminCursors = append(stub.adminCursors, copyStrategyCommunicationCursor(cursor))
	if len(stub.adminPages) == 0 {
		return objectivesdomain.StrategyCommunicationRecipientPage{}, nil
	}
	page := stub.adminPages[0]
	stub.adminPages = stub.adminPages[1:]
	return page, nil
}

func (stub *strategyCommunicationsStoreStub) GetStrategyCommunicationFoundation(
	_ context.Context,
	_ uuid.UUID,
	periodStart time.Time,
	periodEnd time.Time,
) (objectivesdomain.StrategyCommunicationFoundation, error) {
	stub.foundationPeriodStart = periodStart
	stub.foundationPeriodEnd = periodEnd
	return stub.foundation, nil
}

func (stub *strategyCommunicationsStoreStub) GetStrategyCommunicationMonthlySummary(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ time.Time,
) (objectivesdomain.StrategyCommunicationMonthlySummary, error) {
	return stub.monthlySummary, nil
}

func (stub *strategyCommunicationsStoreStub) ListStrategyWeeklyCommunicationRecipients(
	_ context.Context,
	cursor *objectivesdomain.StrategyCommunicationCursor,
	_ int,
) (objectivesdomain.StrategyCommunicationRecipientPage, error) {
	stub.weeklyCursors = append(stub.weeklyCursors, copyStrategyCommunicationCursor(cursor))
	if len(stub.weeklyPages) == 0 {
		return objectivesdomain.StrategyCommunicationRecipientPage{}, nil
	}
	page := stub.weeklyPages[0]
	stub.weeklyPages = stub.weeklyPages[1:]
	return page, nil
}

func (stub *strategyCommunicationsStoreStub) ListStrategyWeeklyCommunicationSignals(
	_ context.Context,
	staleBefore time.Time,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	cursor *objectivesdomain.StrategyWeeklySignalCursor,
	limit int,
) (objectivesdomain.StrategyWeeklySignalPage, error) {
	stub.weeklySignalCalls = append(stub.weeklySignalCalls, strategyWeeklySignalCall{
		StaleBefore: staleBefore,
		UserID:      userID,
		WorkspaceID: workspaceID,
		Cursor:      copyStrategyWeeklySignalCursor(cursor),
		Limit:       limit,
	})
	pages := stub.weeklySignals[userID]
	if len(pages) == 0 {
		return objectivesdomain.StrategyWeeklySignalPage{}, nil
	}
	page := pages[0]
	stub.weeklySignals[userID] = pages[1:]
	return page, nil
}

func copyStrategyCommunicationCursor(cursor *objectivesdomain.StrategyCommunicationCursor) *objectivesdomain.StrategyCommunicationCursor {
	if cursor == nil {
		return nil
	}
	copy := *cursor
	return &copy
}

func copyStrategyWeeklySignalCursor(cursor *objectivesdomain.StrategyWeeklySignalCursor) *objectivesdomain.StrategyWeeklySignalCursor {
	if cursor == nil {
		return nil
	}
	copy := *cursor
	return &copy
}

type strategyNotificationCreatorStub struct {
	notifications []notifications.CoreNewNotification
	err           error
}

func (stub *strategyNotificationCreatorStub) Create(
	_ context.Context,
	notification notifications.CoreNewNotification,
) (notifications.CoreNotification, error) {
	stub.notifications = append(stub.notifications, notification)
	return notifications.CoreNotification{}, stub.err
}

func TestStrategyWeeklyCheckInsAdvancePastRecipientsWithoutSignals(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	firstUserID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	secondUserID := uuid.MustParse("20000000-0000-0000-0000-000000000002")
	objectiveID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	teamID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	firstRecipient := objectivesdomain.StrategyCommunicationRecipient{UserID: firstUserID, WorkspaceID: workspaceID, Timezone: "UTC"}
	secondRecipient := objectivesdomain.StrategyCommunicationRecipient{UserID: secondUserID, WorkspaceID: workspaceID, Timezone: "UTC"}
	now := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)
	stub := &strategyCommunicationsStoreStub{
		weeklyPages: []objectivesdomain.StrategyCommunicationRecipientPage{
			{Recipients: []objectivesdomain.StrategyCommunicationRecipient{firstRecipient}, HasMore: true},
			{Recipients: []objectivesdomain.StrategyCommunicationRecipient{secondRecipient}},
		},
		weeklySignals: map[uuid.UUID][]objectivesdomain.StrategyWeeklySignalPage{
			firstUserID: {{Records: []objectivesdomain.StrategyWeeklySignalRecord{}}},
			secondUserID: {{Records: []objectivesdomain.StrategyWeeklySignalRecord{{
				Recipient:          secondRecipient,
				ObjectiveID:        objectiveID,
				TeamID:             teamID,
				ObjectiveName:      "Protect launch readiness",
				ObjectiveUpdatedAt: now.Add(-24 * time.Hour),
				IsAtRiskObjective:  true,
			}}}},
		},
	}
	notifier := &strategyNotificationCreatorStub{}

	err := processStrategyWeeklyCheckIns(context.Background(), stub, notifier, uuid.New(), now)

	require.NoError(t, err)
	require.Len(t, stub.weeklyCursors, 2)
	require.Nil(t, stub.weeklyCursors[0])
	require.Equal(t, &objectivesdomain.StrategyCommunicationCursor{WorkspaceID: workspaceID, UserID: firstUserID}, stub.weeklyCursors[1])
	require.Len(t, stub.weeklySignalCalls, 2)
	for _, call := range stub.weeklySignalCalls {
		require.Equal(t, now.Add(-7*24*time.Hour), call.StaleBefore)
		require.Equal(t, strategyWeeklySignalBatchSize, call.Limit)
	}
	require.Len(t, notifier.notifications, 1)
	require.Equal(t, secondUserID, notifier.notifications[0].RecipientID)
	require.Equal(t, notifications.StrategyNotificationKindWeeklyCheckIn, notifier.notifications[0].Message.Strategy.Kind)
}

func TestGetStrategyFoundationUsesLocalCalendarDatesAcrossUTCOffsets(t *testing.T) {
	t.Parallel()

	location, err := time.LoadLocation("Pacific/Kiritimati")
	require.NoError(t, err)
	quarterStart := time.Date(2027, time.January, 1, 0, 0, 0, 0, location)
	stub := &strategyCommunicationsStoreStub{}

	_, err = getStrategyFoundation(context.Background(), stub, uuid.New(), quarterStart)

	require.NoError(t, err)
	require.Equal(t, time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC), stub.foundationPeriodStart)
	require.Equal(t, time.Date(2027, time.April, 1, 0, 0, 0, 0, time.UTC), stub.foundationPeriodEnd)
}

func TestStrategyRecipientPagesStopAtBoundedBacklog(t *testing.T) {
	t.Parallel()

	calls := 0
	load := func(_ context.Context, cursor *objectivesdomain.StrategyCommunicationCursor, _ int) (objectivesdomain.StrategyCommunicationRecipientPage, error) {
		calls++
		userID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
		workspaceID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
		if cursor != nil {
			workspaceID = uuid.MustParse("10000000-0000-0000-0000-" + strategyCommunicationTestSequence(calls))
		}
		return objectivesdomain.StrategyCommunicationRecipientPage{
			Recipients: []objectivesdomain.StrategyCommunicationRecipient{{UserID: userID, WorkspaceID: workspaceID, Timezone: "UTC"}},
			HasMore:    true,
		}, nil
	}

	err := processStrategyRecipientPages(context.Background(), load, func(objectivesdomain.StrategyCommunicationRecipient) error { return nil })

	require.ErrorIs(t, err, errStrategyCommunicationBacklog)
	require.Equal(t, strategyCommunicationMaxBatches, calls)
}

func strategyCommunicationTestSequence(value int) string {
	const digits = "000000000000"
	formatted := []byte(digits)
	for index := len(formatted) - 1; value > 0; index-- {
		formatted[index] = byte('0' + value%10)
		value /= 10
	}
	return string(formatted)
}
