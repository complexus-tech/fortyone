package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	strategyCommunicationHour       = 9
	strategyCommunicationBatchSize  = 100
	strategyCommunicationMaxBatches = 100
	strategyWeeklySignalBatchSize   = 100
	strategyWeeklySignalMaxBatches  = 100
	strategySnapshotVersion         = 1
	strategyStaleAfterDays          = 7
	strategyWeeklyDetailLimit       = 10
	strategyPlanningLeadDays        = 7
)

var errStrategyCommunicationBacklog = errors.New("strategy communication backlog remains")

// StrategyNotificationCreator is the notification boundary used by strategy jobs.
// The concrete service persists the notification, publishes it in-app, and queues
// the existing coalesced email digest. Strategy notifications are email-only.
type StrategyNotificationCreator interface {
	Create(context.Context, notifications.CoreNewNotification) (notifications.CoreNotification, error)
}

// StrategyCommunicationsStore is the worker-owned persistence capability used
// to page current recipients and load their strategy evidence. Every timestamp
// is captured by the application once and passed through retries and pages.
type StrategyCommunicationsStore interface {
	ListStrategyCommunicationAdministrators(
		context.Context,
		*objectivesdomain.StrategyCommunicationCursor,
		int,
	) (objectivesdomain.StrategyCommunicationRecipientPage, error)
	GetStrategyCommunicationFoundation(
		context.Context,
		uuid.UUID,
		time.Time,
		time.Time,
	) (objectivesdomain.StrategyCommunicationFoundation, error)
	GetStrategyCommunicationMonthlySummary(
		context.Context,
		uuid.UUID,
		time.Time,
		time.Time,
	) (objectivesdomain.StrategyCommunicationMonthlySummary, error)
	ListStrategyWeeklyCommunicationRecipients(
		context.Context,
		*objectivesdomain.StrategyCommunicationCursor,
		int,
	) (objectivesdomain.StrategyCommunicationRecipientPage, error)
	ListStrategyWeeklyCommunicationSignals(
		context.Context,
		time.Time,
		uuid.UUID,
		uuid.UUID,
		*objectivesdomain.StrategyWeeklySignalCursor,
		int,
	) (objectivesdomain.StrategyWeeklySignalPage, error)
}

type strategyRecipient struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Timezone    string
}

type strategyFoundation struct {
	HasUltimateGoal bool
	PillarCount     int
	ObjectiveCount  int
}

type strategyCheckIn struct {
	strategyRecipient
	StaleObjectives  int
	AtRiskObjectives int
	StaleKeyResults  int
	Objectives       []notifications.StrategyObjectiveSnapshot
	KeyResults       []notifications.StrategyKeyResultSnapshot
}

type strategyCheckInRecord struct {
	strategyRecipient
	ObjectiveID              uuid.UUID
	TeamID                   uuid.UUID
	ObjectiveName            string
	ObjectiveHealth          *string
	ObjectiveStatusID        *uuid.UUID
	ObjectiveStatusName      *string
	ObjectiveStatusCategory  *string
	ObjectiveStartDate       *time.Time
	ObjectiveEndDate         *time.Time
	ObjectiveUpdatedAt       time.Time
	IsStaleObjective         bool
	IsAtRiskObjective        bool
	KeyResultID              *uuid.UUID
	KeyResultName            *string
	KeyResultMeasurementType *string
	KeyResultStartValue      *float64
	KeyResultCurrentValue    *float64
	KeyResultTargetValue     *float64
	KeyResultStartDate       *time.Time
	KeyResultEndDate         *time.Time
	KeyResultUpdatedAt       *time.Time
}

type strategyMonthlySummary struct {
	PillarCount          int
	PillarsNeedingReview int
	ObjectiveCount       int
	AtRiskObjectives     int
	UnalignedObjectives  int
	KeyResultCount       int
	KeyResultProgress    *float64
	CompletedStories     int
}

// ProcessStrategyCommunications creates due local-time planning reminders,
// weekly owner check-ins, and monthly leadership summaries.
func ProcessStrategyCommunications(
	ctx context.Context,
	store StrategyCommunicationsStore,
	log *logger.Logger,
	notifier StrategyNotificationCreator,
	systemUserID uuid.UUID,
) error {
	return processStrategyCommunicationsAt(ctx, store, log, notifier, systemUserID, time.Now().UTC())
}

func processStrategyCommunicationsAt(
	ctx context.Context,
	store StrategyCommunicationsStore,
	log *logger.Logger,
	notifier StrategyNotificationCreator,
	systemUserID uuid.UUID,
	now time.Time,
) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessStrategyCommunications")
	defer span.End()

	if store == nil {
		return errors.New("strategy communications store is required")
	}
	if log == nil {
		return errors.New("strategy communications logger is required")
	}
	if notifier == nil {
		return errors.New("strategy communications notifier is required")
	}
	if systemUserID == uuid.Nil {
		return errors.New("strategy communications system user is required")
	}
	if now.IsZero() {
		return errors.New("strategy communications as-of time is required")
	}
	now = now.UTC()
	var processingErrors []error

	if err := processStrategyPlanningReminders(ctx, store, notifier, systemUserID, now); err != nil {
		processingErrors = append(processingErrors, fmt.Errorf("planning reminders: %w", err))
	}
	if err := processStrategyWeeklyCheckIns(ctx, store, notifier, systemUserID, now); err != nil {
		processingErrors = append(processingErrors, fmt.Errorf("weekly check-ins: %w", err))
	}
	if err := processStrategyMonthlySummaries(ctx, store, notifier, systemUserID, now); err != nil {
		processingErrors = append(processingErrors, fmt.Errorf("monthly summaries: %w", err))
	}

	if len(processingErrors) > 0 {
		err := errors.Join(processingErrors...)
		span.RecordError(err)
		log.Error(ctx, "strategy communications completed with errors", "error", err)
		return err
	}

	return nil
}

type strategyRecipientPageLoader func(
	context.Context,
	*objectivesdomain.StrategyCommunicationCursor,
	int,
) (objectivesdomain.StrategyCommunicationRecipientPage, error)

func processStrategyRecipientPages(
	ctx context.Context,
	load strategyRecipientPageLoader,
	process func(objectivesdomain.StrategyCommunicationRecipient) error,
) error {
	if load == nil || process == nil {
		return errors.New("strategy recipient page processor is not configured")
	}

	var processingErrors []error
	var cursor *objectivesdomain.StrategyCommunicationCursor
	for batch := 0; batch < strategyCommunicationMaxBatches; batch++ {
		page, err := load(ctx, cursor, strategyCommunicationBatchSize)
		if err != nil {
			processingErrors = append(processingErrors, err)
			return errors.Join(processingErrors...)
		}
		if len(page.Recipients) > strategyCommunicationBatchSize {
			processingErrors = append(processingErrors, fmt.Errorf("strategy recipient page exceeds limit: got %d, limit %d", len(page.Recipients), strategyCommunicationBatchSize))
			return errors.Join(processingErrors...)
		}
		for _, recipient := range page.Recipients {
			if err := process(recipient); err != nil {
				processingErrors = append(processingErrors, err)
			}
		}
		if !page.HasMore {
			return errors.Join(processingErrors...)
		}
		if len(page.Recipients) == 0 {
			processingErrors = append(processingErrors, errors.New("strategy recipient page reported more work without a cursor row"))
			return errors.Join(processingErrors...)
		}

		last := page.Recipients[len(page.Recipients)-1]
		next := &objectivesdomain.StrategyCommunicationCursor{
			WorkspaceID: last.WorkspaceID,
			UserID:      last.UserID,
		}
		if cursor != nil && !strategyCommunicationCursorAdvances(*cursor, *next) {
			processingErrors = append(processingErrors, errors.New("strategy recipient cursor did not advance"))
			return errors.Join(processingErrors...)
		}
		cursor = next
	}

	processingErrors = append(processingErrors, errStrategyCommunicationBacklog)
	return errors.Join(processingErrors...)
}

func strategyCommunicationCursorAdvances(previous, next objectivesdomain.StrategyCommunicationCursor) bool {
	if previous.WorkspaceID != next.WorkspaceID {
		return previous.WorkspaceID.String() < next.WorkspaceID.String()
	}
	return previous.UserID.String() < next.UserID.String()
}

func processStrategyPlanningReminders(ctx context.Context, store StrategyCommunicationsStore, notifier StrategyNotificationCreator, systemUserID uuid.UUID, now time.Time) error {
	return processStrategyRecipientPages(ctx, store.ListStrategyCommunicationAdministrators, func(candidate objectivesdomain.StrategyCommunicationRecipient) error {
		recipient := strategyRecipientFromDomain(candidate)
		localNow := now.In(strategyLocation(recipient.Timezone))
		if localNow.Hour() != strategyCommunicationHour {
			return nil
		}

		quarterStart := nextQuarterStart(localNow)
		daysUntil := calendarDaysBetween(localNow, quarterStart)
		if !isStrategyPlanningReminderDue(daysUntil) {
			return nil
		}

		foundation, loadErr := getStrategyFoundation(ctx, store, recipient.WorkspaceID, quarterStart)
		if loadErr != nil {
			return loadErr
		}
		if foundation.HasUltimateGoal && foundation.PillarCount > 0 && foundation.ObjectiveCount > 0 {
			return nil
		}

		period := fmt.Sprintf("Q%d", quarterForMonth(quarterStart.Month()))
		if quarterStart.Month() == time.January {
			period = fmt.Sprintf("%d strategy", quarterStart.Year())
		}
		missing := missingStrategyElements(foundation)
		notification := notifications.CoreNewNotification{
			DedupeKey:   fmt.Sprintf("strategy:planning:%s:%s:%s:%d", recipient.WorkspaceID, recipient.UserID, quarterStart.Format("2006-01-02"), daysUntil),
			RecipientID: recipient.UserID,
			WorkspaceID: recipient.WorkspaceID,
			Type:        "strategy_update",
			EntityType:  "strategy",
			EntityID:    recipient.WorkspaceID,
			ActorID:     systemUserID,
			Title:       fmt.Sprintf("Plan your %s", period),
			Message: notifications.NotificationMessage{
				Template: "Your next planning period starts in {days}. Review {missing} while there is still time to align the team.",
				Variables: map[string]notifications.Variable{
					"days":    {Value: fmt.Sprintf("%d days", daysUntil), Type: "date"},
					"missing": {Value: missing, Type: "value"},
				},
				Strategy: &notifications.StrategyNotificationSnapshot{
					Version:     strategySnapshotVersion,
					Kind:        notifications.StrategyNotificationKindPlanningReminder,
					GeneratedAt: now.UTC(),
					Planning: &notifications.StrategyPlanningSnapshot{
						Period:          period,
						StartsAt:        quarterStart,
						DaysUntil:       daysUntil,
						HasUltimateGoal: foundation.HasUltimateGoal,
						PillarCount:     foundation.PillarCount,
						ObjectiveCount:  foundation.ObjectiveCount,
						MissingElements: missingStrategyElementKeys(foundation),
					},
				},
			},
		}
		if _, createErr := notifier.Create(ctx, notification); createErr != nil {
			return createErr
		}
		return nil
	})
}

func isStrategyPlanningReminderDue(daysUntil int) bool {
	return daysUntil == strategyPlanningLeadDays
}

func processStrategyMonthlySummaries(ctx context.Context, store StrategyCommunicationsStore, notifier StrategyNotificationCreator, systemUserID uuid.UUID, now time.Time) error {
	return processStrategyRecipientPages(ctx, store.ListStrategyCommunicationAdministrators, func(candidate objectivesdomain.StrategyCommunicationRecipient) error {
		recipient := strategyRecipientFromDomain(candidate)
		localNow := now.In(strategyLocation(recipient.Timezone))
		if localNow.Hour() != strategyCommunicationHour || localNow.Day() != 1 {
			return nil
		}

		monthStart := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, localNow.Location())
		previousMonthStart := monthStart.AddDate(0, -1, 0)
		summary, loadErr := getStrategyMonthlySummary(ctx, store, recipient.WorkspaceID, previousMonthStart.UTC(), monthStart.UTC())
		if loadErr != nil {
			return loadErr
		}
		if !summary.hasActionableSignal() {
			return nil
		}

		title := fmt.Sprintf("%s strategy summary", previousMonthStart.Format("January"))
		summaryText := strategyMonthlySummaryText(summary)
		if summary.PillarsNeedingReview > 0 {
			summaryText = fmt.Sprintf("%d pillars need review; %s", summary.PillarsNeedingReview, summaryText)
		}

		notification := notifications.CoreNewNotification{
			DedupeKey:   fmt.Sprintf("strategy:monthly:%s:%s:%s", recipient.WorkspaceID, recipient.UserID, previousMonthStart.Format("2006-01")),
			RecipientID: recipient.UserID,
			WorkspaceID: recipient.WorkspaceID,
			Type:        "strategy_update",
			EntityType:  "strategy",
			EntityID:    recipient.WorkspaceID,
			ActorID:     systemUserID,
			Title:       title,
			Message: notifications.NotificationMessage{
				Template: "Here is the current strategy snapshot and last month's linked delivery: {summary}.",
				Variables: map[string]notifications.Variable{
					"summary": {Value: summaryText, Type: "value"},
				},
				Strategy: &notifications.StrategyNotificationSnapshot{
					Version:     strategySnapshotVersion,
					Kind:        notifications.StrategyNotificationKindMonthlySummary,
					GeneratedAt: now.UTC(),
					MonthlySummary: &notifications.StrategyMonthlySummarySnapshot{
						PeriodStart:          previousMonthStart,
						PeriodEnd:            monthStart,
						PillarCount:          summary.PillarCount,
						PillarsNeedingReview: summary.PillarsNeedingReview,
						ObjectiveCount:       summary.ObjectiveCount,
						AtRiskObjectives:     summary.AtRiskObjectives,
						UnalignedObjectives:  summary.UnalignedObjectives,
						KeyResultCount:       summary.KeyResultCount,
						KeyResultProgress:    summary.KeyResultProgress,
						CompletedStories:     summary.CompletedStories,
					},
				},
			},
		}
		if _, createErr := notifier.Create(ctx, notification); createErr != nil {
			return createErr
		}
		return nil
	})
}

func (summary strategyMonthlySummary) hasActionableSignal() bool {
	return summary.PillarsNeedingReview > 0 ||
		summary.AtRiskObjectives > 0 ||
		summary.UnalignedObjectives > 0
}

func strategyMonthlySummaryText(summary strategyMonthlySummary) string {
	parts := make([]string, 0, 4)
	switch {
	case summary.KeyResultProgress != nil && summary.KeyResultCount > 0:
		parts = append(parts, fmt.Sprintf(
			"%.0f%% average progress across %d %s",
			*summary.KeyResultProgress,
			summary.KeyResultCount,
			pluralize(summary.KeyResultCount, "key result", "key results"),
		))
	case summary.KeyResultProgress != nil:
		// Older persisted snapshots did not carry a count. Preserve their real
		// progress instead of treating a missing additive field as no data.
		parts = append(parts, fmt.Sprintf("%.0f%% average key-result progress", *summary.KeyResultProgress))
	case summary.KeyResultCount > 0:
		parts = append(parts, fmt.Sprintf(
			"progress is unavailable for %d %s",
			summary.KeyResultCount,
			pluralize(summary.KeyResultCount, "key result", "key results"),
		))
	default:
		parts = append(parts, "no key results in the current snapshot")
	}
	parts = append(parts,
		fmt.Sprintf("%d objectives needing attention", summary.AtRiskObjectives),
		fmt.Sprintf("%d unaligned objectives", summary.UnalignedObjectives),
		fmt.Sprintf("%d linked stories completed last month", summary.CompletedStories),
	)
	return strings.Join(parts, ", ")
}

func strategyRecipientFromDomain(recipient objectivesdomain.StrategyCommunicationRecipient) strategyRecipient {
	return strategyRecipient{
		UserID:      recipient.UserID,
		WorkspaceID: recipient.WorkspaceID,
		Timezone:    recipient.Timezone,
	}
}

func getStrategyFoundation(ctx context.Context, store StrategyCommunicationsStore, workspaceID uuid.UUID, quarterStart time.Time) (strategyFoundation, error) {
	periodStart := strategyCalendarDateUTC(quarterStart)
	periodEnd := strategyCalendarDateUTC(quarterStart.AddDate(0, 3, 0))
	foundation, err := store.GetStrategyCommunicationFoundation(ctx, workspaceID, periodStart, periodEnd)
	if err != nil {
		return strategyFoundation{}, fmt.Errorf("load strategy foundation for workspace %s: %w", workspaceID, err)
	}
	return strategyFoundation{
		HasUltimateGoal: foundation.HasUltimateGoal,
		PillarCount:     foundation.PillarCount,
		ObjectiveCount:  foundation.ObjectiveCount,
	}, nil
}

func strategyCalendarDateUTC(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func getStrategyMonthlySummary(ctx context.Context, store StrategyCommunicationsStore, workspaceID uuid.UUID, periodStart, periodEnd time.Time) (strategyMonthlySummary, error) {
	summary, err := store.GetStrategyCommunicationMonthlySummary(ctx, workspaceID, periodStart.UTC(), periodEnd.UTC())
	if err != nil {
		return strategyMonthlySummary{}, fmt.Errorf("load monthly strategy summary for workspace %s: %w", workspaceID, err)
	}
	return strategyMonthlySummary{
		PillarCount:          summary.PillarCount,
		PillarsNeedingReview: summary.PillarsNeedingReview,
		ObjectiveCount:       summary.ObjectiveCount,
		AtRiskObjectives:     summary.AtRiskObjectives,
		UnalignedObjectives:  summary.UnalignedObjectives,
		KeyResultCount:       summary.KeyResultCount,
		KeyResultProgress:    summary.KeyResultProgress,
		CompletedStories:     summary.CompletedStories,
	}, nil
}

func strategyLocation(timezone string) *time.Location {
	location, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		return time.UTC
	}
	return location
}

func nextQuarterStart(now time.Time) time.Time {
	currentQuarter := (int(now.Month()) - 1) / 3
	nextMonth := time.Month((currentQuarter+1)*3 + 1)
	year := now.Year()
	if nextMonth > time.December {
		nextMonth = time.January
		year++
	}
	return time.Date(year, nextMonth, 1, 0, 0, 0, 0, now.Location())
}

func quarterForMonth(month time.Month) int {
	return (int(month)-1)/3 + 1
}

func calendarDaysBetween(from, to time.Time) int {
	// UTC is intentional here: only the local calendar components matter, and
	// using the local offset would make DST boundaries appear 23 or 25 hours long.
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	return int(toDate.Sub(fromDate).Hours() / 24)
}

func missingStrategyElements(foundation strategyFoundation) string {
	missing := make([]string, 0, 3)
	if !foundation.HasUltimateGoal {
		missing = append(missing, "your ultimate goal")
	}
	if foundation.PillarCount == 0 {
		missing = append(missing, "strategic pillars")
	}
	if foundation.ObjectiveCount == 0 {
		missing = append(missing, "the next period's objectives")
	}
	return strings.Join(missing, ", ")
}

func missingStrategyElementKeys(foundation strategyFoundation) []string {
	missing := make([]string, 0, 3)
	if !foundation.HasUltimateGoal {
		missing = append(missing, notifications.StrategyMissingElementUltimateGoal)
	}
	if foundation.PillarCount == 0 {
		missing = append(missing, notifications.StrategyMissingElementPillars)
	}
	if foundation.ObjectiveCount == 0 {
		missing = append(missing, notifications.StrategyMissingElementObjectives)
	}
	return missing
}

func strategyCheckInSummary(checkIn strategyCheckIn) string {
	parts := make([]string, 0, 3)
	if checkIn.AtRiskObjectives > 0 {
		parts = append(parts, fmt.Sprintf("%d at-risk objectives", checkIn.AtRiskObjectives))
	}
	if checkIn.StaleObjectives > 0 {
		parts = append(parts, fmt.Sprintf("%d objectives without a recent update", checkIn.StaleObjectives))
	}
	if checkIn.StaleKeyResults > 0 {
		parts = append(parts, fmt.Sprintf("%d stalled key results", checkIn.StaleKeyResults))
	}
	return strings.Join(parts, ", ")
}
