package notifications

import notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"

// Service aliases keep existing first-party callers source-compatible while
// domain values remain transport-neutral and repository adapters import the
// domain package directly.
type (
	NotificationType = notificationsdomain.NotificationType
	EntityType       = notificationsdomain.EntityType
	PreferenceType   = notificationsdomain.PreferenceType

	NotificationMessage                         = notificationsdomain.NotificationMessage
	Variable                                    = notificationsdomain.Variable
	StrategyNotificationSnapshot                = notificationsdomain.StrategyNotificationSnapshot
	StrategyPlanningSnapshot                    = notificationsdomain.StrategyPlanningSnapshot
	StrategyWeeklyCheckInSnapshot               = notificationsdomain.StrategyWeeklyCheckInSnapshot
	StrategyWeeklyCheckInTeamCountsSnapshot     = notificationsdomain.StrategyWeeklyCheckInTeamCountsSnapshot
	StrategyWeeklyCheckInOmittedDetailsSnapshot = notificationsdomain.StrategyWeeklyCheckInOmittedDetailsSnapshot
	StrategyWeeklyCheckInCounts                 = notificationsdomain.StrategyWeeklyCheckInCounts
	StrategyObjectiveSnapshot                   = notificationsdomain.StrategyObjectiveSnapshot
	StrategyObjectiveStatusSnapshot             = notificationsdomain.StrategyObjectiveStatusSnapshot
	StrategyKeyResultSnapshot                   = notificationsdomain.StrategyKeyResultSnapshot
	StrategyMonthlySummarySnapshot              = notificationsdomain.StrategyMonthlySummarySnapshot
	CoreNewNotification                         = notificationsdomain.NewNotification
	CoreNotification                            = notificationsdomain.Notification
	CoreNotificationActor                       = notificationsdomain.NotificationActor
	CorePortalNotification                      = notificationsdomain.PortalNotification
	CoreNotificationPreferences                 = notificationsdomain.Preferences
	NotificationChannels                        = notificationsdomain.Channels
	NotificationPreferenceSet                   = notificationsdomain.PreferenceSet
	NotificationChannelPatch                    = notificationsdomain.ChannelPatch
)

const (
	StrategyNotificationKindPlanningReminder = notificationsdomain.StrategyNotificationKindPlanningReminder
	StrategyNotificationKindWeeklyCheckIn    = notificationsdomain.StrategyNotificationKindWeeklyCheckIn
	StrategyNotificationKindMonthlySummary   = notificationsdomain.StrategyNotificationKindMonthlySummary

	StrategySignalReasonAtRisk     = notificationsdomain.StrategySignalReasonAtRisk
	StrategySignalReasonStale      = notificationsdomain.StrategySignalReasonStale
	StrategySignalReasonIncomplete = notificationsdomain.StrategySignalReasonIncomplete

	StrategyMissingElementUltimateGoal = notificationsdomain.StrategyMissingElementUltimateGoal
	StrategyMissingElementPillars      = notificationsdomain.StrategyMissingElementPillars
	StrategyMissingElementObjectives   = notificationsdomain.StrategyMissingElementObjectives

	NotificationTypeStoryUpdate             = notificationsdomain.NotificationTypeStoryUpdate
	NotificationTypeStoryComment            = notificationsdomain.NotificationTypeStoryComment
	NotificationTypeCommentReply            = notificationsdomain.NotificationTypeCommentReply
	NotificationTypeObjectiveUpdate         = notificationsdomain.NotificationTypeObjectiveUpdate
	NotificationTypeKeyResultUpdate         = notificationsdomain.NotificationTypeKeyResultUpdate
	NotificationTypeMention                 = notificationsdomain.NotificationTypeMention
	NotificationTypeFeedbackComment         = notificationsdomain.NotificationTypeFeedbackComment
	NotificationTypeFeedbackStatusUpdate    = notificationsdomain.NotificationTypeFeedbackStatusUpdate
	NotificationTypeFeedbackUpdatePublished = notificationsdomain.NotificationTypeFeedbackUpdatePublished
	NotificationTypeFeedbackItemMerged      = notificationsdomain.NotificationTypeFeedbackItemMerged
	NotificationTypeStrategyUpdate          = notificationsdomain.NotificationTypeStrategyUpdate

	EntityTypeStory     = notificationsdomain.EntityTypeStory
	EntityTypeComment   = notificationsdomain.EntityTypeComment
	EntityTypeObjective = notificationsdomain.EntityTypeObjective
	EntityTypeKeyResult = notificationsdomain.EntityTypeKeyResult
	EntityTypeFeedback  = notificationsdomain.EntityTypeFeedback
	EntityTypeStrategy  = notificationsdomain.EntityTypeStrategy
)

func SupportsInAppDelivery[T ~string](notificationType T) bool {
	parsed, err := notificationsdomain.ParseNotificationType(string(notificationType))
	return err == nil && notificationsdomain.SupportsInAppDelivery(parsed)
}
