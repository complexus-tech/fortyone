package taskhandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
)

func buildNotificationDigestSubject(workspaceName string, count int) string {
	if count == 1 {
		return fmt.Sprintf("New activity in %s", workspaceName)
	}
	return fmt.Sprintf("%d updates to review in %s", count, workspaceName)
}

func buildNotificationDigestCopyInput(data NotificationEmailDigestData, workspaceURL string) (notificationDigestCopyInput, error) {
	notificationsURL := workspaceURL + "/notifications"
	fallbackRows := make([]notificationDigestCopyRow, 0, maxNotificationDigestDetailRows+1)
	facts := []emailcopy.Fact{{
		ReferenceID: "digest_context",
		Text: fmt.Sprintf(
			"This email contains %d unread product %s from the workspace %s.",
			len(data.Items),
			pluralWord(len(data.Items), "update", "updates"),
			data.WorkspaceName,
		),
		EntityTokens: nonEmptyStrings(data.WorkspaceName),
		ProtectedTokens: nonEmptyStrings(fmt.Sprintf(
			"%d unread product %s",
			len(data.Items),
			pluralWord(len(data.Items), "update", "updates"),
		)),
	}}
	actions := make([]emailcopy.Action, 0, 1)
	actionURLs := make(map[string]string, len(data.Items)+1)
	factActions := make(map[string]string, len(data.Items))
	factLabels := make(map[string]string, len(data.Items))
	hasStrategySnapshot := false
	detailCount := 0
	omittedNotificationCount := 0

	for index, item := range data.Items {
		var message NotificationMessage
		if err := json.Unmarshal(item.Message, &message); err != nil {
			return notificationDigestCopyInput{}, fmt.Errorf("failed to unmarshal notification message %s: %w", item.NotificationID, err)
		}
		if message.Strategy != nil {
			hasStrategySnapshot = true
		}
		if detailCount >= maxNotificationDigestDetailRows {
			omittedNotificationCount++
			continue
		}

		if message.Strategy != nil && message.Strategy.Version == 1 && message.Strategy.Kind == "weekly_check_in" && message.Strategy.WeeklyCheckIn != nil {
			weeklyFacts, weeklyURLs, weeklyFactActions, weeklyFactLabels, weeklyFallbackRows := buildWeeklyStrategyDigestFacts(
				item,
				*message.Strategy.WeeklyCheckIn,
				workspaceURL,
				maxNotificationDigestDetailRows-detailCount,
			)
			facts = append(facts, weeklyFacts...)
			for referenceID, destination := range weeklyURLs {
				actionURLs[referenceID] = destination
			}
			for referenceID, actionReferenceID := range weeklyFactActions {
				factActions[referenceID] = actionReferenceID
			}
			for referenceID, label := range weeklyFactLabels {
				factLabels[referenceID] = label
			}
			fallbackRows = append(fallbackRows, weeklyFallbackRows...)
			detailCount += len(weeklyFallbackRows)
			continue
		}
		if message.Strategy != nil && message.Strategy.Version == 1 && message.Strategy.Kind == "planning_reminder" && message.Strategy.Planning != nil {
			fact, actionReferenceID, destination, label, fallbackRow := buildStrategyPlanningDigestFact(item, *message.Strategy.Planning, workspaceURL)
			facts = append(facts, fact)
			actionURLs[actionReferenceID] = destination
			factActions[fact.ReferenceID] = actionReferenceID
			factLabels[fact.ReferenceID] = label
			fallbackRows = append(fallbackRows, fallbackRow)
			detailCount++
			continue
		}
		if message.Strategy != nil && message.Strategy.Version == 1 && message.Strategy.Kind == "monthly_summary" && message.Strategy.MonthlySummary != nil {
			fact, actionReferenceID, destination, label, fallbackRow := buildStrategyMonthlyDigestFact(item, *message.Strategy, workspaceURL)
			facts = append(facts, fact)
			actionURLs[actionReferenceID] = destination
			factActions[fact.ReferenceID] = actionReferenceID
			factLabels[fact.ReferenceID] = label
			fallbackRows = append(fallbackRows, fallbackRow)
			detailCount++
			continue
		}
		parsed := parseNotificationMessage(message)
		factReferenceID := fmt.Sprintf("notification_%d", index+1)
		actionReferenceID := fmt.Sprintf("notification_action_%d", index+1)
		destination, _ := notificationEmailDestination(item.EntityType, item.EntityID, item.FeedbackSlug, item.NotificationID, workspaceURL)
		factText := strings.TrimSpace(fmt.Sprintf("%s: %s", notificationPlainText(item.Title, 180), parsed.Text))
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     factReferenceID,
			Text:            factText,
			EntityTokens:    nonEmptyStrings(notificationPlainText(item.Title, 180)),
			ProtectedTokens: notificationSemanticProtectedTokens(message, parsed.Text),
			Required:        true,
		})
		actionURLs[actionReferenceID] = destination
		factActions[factReferenceID] = actionReferenceID
		factLabels[factReferenceID] = notificationPlainText(item.Title, 180)
		fallbackRows = append(fallbackRows, notificationDigestCopyRow{Text: factText, Label: factLabels[factReferenceID], URL: destination})
		detailCount++
	}
	if omittedNotificationCount > 0 {
		const remainingActionReferenceID = "remaining_notifications_action"
		facts = append(facts, emailcopy.Fact{
			ReferenceID:  "remaining_updates",
			Text:         fmt.Sprintf("There are %d additional unread %s available in Notifications.", omittedNotificationCount, pluralWord(omittedNotificationCount, "update", "updates")),
			EntityTokens: []string{"Notifications"},
			ProtectedTokens: nonEmptyStrings(fmt.Sprintf(
				"%d additional unread %s available in Notifications",
				omittedNotificationCount,
				pluralWord(omittedNotificationCount, "update", "updates"),
			)),
			Required: true,
		})
		actionURLs[remainingActionReferenceID] = notificationsURL
		factActions["remaining_updates"] = remainingActionReferenceID
		factLabels["remaining_updates"] = "Notifications"
		fallbackRows = append(fallbackRows, notificationDigestCopyRow{
			Text:  fmt.Sprintf("%d additional unread %s are available in Notifications.", omittedNotificationCount, pluralWord(omittedNotificationCount, "update", "updates")),
			Label: "Notifications",
			URL:   notificationsURL,
		})
	}

	primaryActionReferenceID := digestActionPrimary
	primaryDestination := notificationsURL
	primaryDescription := "Open notifications"
	primaryLabel := "Open notifications"
	if feedbackOnlyDigest(data.Items) {
		primaryDestination = workspaceURL + "/feedback"
		primaryDescription = "Open feedback"
		primaryLabel = "Open feedback"
	}
	if hasStrategySnapshot {
		primaryActionReferenceID = digestActionStrategy
		primaryDestination = workspaceURL + "/strategy"
		primaryDescription = "Review strategy"
		primaryLabel = "Review strategy"
	}
	actions = append(actions, emailcopy.Action{ReferenceID: primaryActionReferenceID, Description: primaryDescription})
	actionURLs[primaryActionReferenceID] = primaryDestination

	fallbackSubject := buildNotificationDigestSubject(data.WorkspaceName, len(data.Items))
	fallbackIntro := "Review the activity below and choose the next useful action."
	if hasStrategySnapshot {
		fallbackSubject = "Your strategy check-in"
		fallbackIntro = "Here are the objectives, key results, and strategy updates that need your attention. I’m Maya, your AI agent. Reply to this email with what changed or what you want updated."
	}
	fallback := notificationDigestCopy{
		Subject:             fallbackSubject,
		Heading:             fallbackSubject,
		Intro:               fallbackIntro,
		Rows:                fallbackRows,
		CTA:                 notificationDigestCopyCTA{Label: primaryLabel, URL: primaryDestination},
		NotificationsURL:    notificationsURL,
		HasStrategySnapshot: hasStrategySnapshot,
	}
	if hasStrategySnapshot {
		fallback.Sender = mailer.SenderProfileMaya
	}

	safetyIdentifier := data.RecipientID.String()
	if data.RecipientID == uuid.Nil {
		safetyIdentifier = "unknown-notification-recipient"
	}

	return notificationDigestCopyInput{
		Request: emailcopy.Request{
			SafetyIdentifier:   safetyIdentifier,
			Purpose:            "persisted notification email digest",
			ProductVoice:       "Help the recipient understand what changed, why it matters, and choose the next useful action. Keep activity factual; make strategy guidance warm and calm.",
			Facts:              facts,
			Actions:            actions,
			IncludeSenderProse: hasStrategySnapshot,
			IncludeReplyPrompt: hasStrategySnapshot,
		},
		Actions:             actionURLs,
		FactActions:         factActions,
		FactLabels:          factLabels,
		Fallback:            fallback,
		HasStrategySnapshot: hasStrategySnapshot,
		NotificationsURL:    notificationsURL,
	}, nil
}

func notificationVariableValues(variables map[string]Variable) []string {
	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		variable := variables[key]
		values = append(values, notificationPlainText(variable.Value, maxNotificationMessageRunes))
	}
	return nonEmptyStrings(values...)
}

// notificationSemanticProtectedTokens keeps the factual roles in activity
// messages bound together. Requiring only the individual variable values
// would allow generated copy to swap an actor, field, assignee, or status while
// still passing literal-token validation. For long comment-style messages, the
// author/action prefix stays exact while the user-authored body remains free to
// be summarized.
func notificationSemanticProtectedTokens(message NotificationMessage, parsedText string) []string {
	values := notificationVariableValues(message.Variables)
	if len(values) == 0 {
		return values
	}

	const maxProtectedActivityRunes = 300
	activity := notificationPlainText(parsedText, 0)
	if len([]rune(activity)) <= maxProtectedActivityRunes {
		return nonEmptyStrings(activity)
	}
	if separator := strings.Index(activity, ":"); separator > 0 {
		semanticPrefix := strings.TrimSpace(activity[:separator])
		if len([]rune(semanticPrefix)) <= maxProtectedActivityRunes {
			return nonEmptyStrings(semanticPrefix)
		}
	}
	return values
}

func buildStrategyPlanningDigestFact(
	item NotificationEmailDigestItem,
	planning strategyPlanningSnapshot,
	workspaceURL string,
) (emailcopy.Fact, string, string, string, notificationDigestCopyRow) {
	ultimateGoalState := "does not yet have an ultimate goal"
	if planning.HasUltimateGoal {
		ultimateGoalState = "has an ultimate goal"
	}
	missingElements := strategyMissingElementLabels(planning.MissingElements)
	factText := fmt.Sprintf(
		"%s: %s starts on %s, in %d days. The strategy foundation %s, has %d strategic pillars, and has %d objectives for that period.",
		item.Title,
		planning.Period,
		planning.StartsAt.UTC().Format("January 2, 2006"),
		planning.DaysUntil,
		ultimateGoalState,
		planning.PillarCount,
		planning.ObjectiveCount,
	)
	if len(missingElements) > 0 {
		factText += " The missing elements are " + strings.Join(missingElements, ", ") + "."
	}
	factReferenceID := "strategy_planning_" + item.NotificationID.String()
	actionReferenceID := "strategy_planning_action_" + item.NotificationID.String()
	destination := workspaceURL + "/strategy"
	label := notificationPlainText(item.Title, 180)
	protectedTokens := nonEmptyStrings(
		fmt.Sprintf("%s starts on %s", planning.Period, planning.StartsAt.UTC().Format("January 2, 2006")),
		fmt.Sprintf("in %d days", planning.DaysUntil),
		"strategy foundation "+ultimateGoalState,
		fmt.Sprintf("has %d strategic pillars", planning.PillarCount),
		fmt.Sprintf("has %d objectives for that period", planning.ObjectiveCount),
	)
	if len(missingElements) > 0 {
		protectedTokens = append(protectedTokens, "missing elements are "+strings.Join(missingElements, ", "))
	}
	fact := emailcopy.Fact{
		ReferenceID:     factReferenceID,
		Text:            factText,
		EntityTokens:    nonEmptyStrings(label, planning.Period),
		ProtectedTokens: protectedTokens,
		Required:        true,
	}
	return fact, actionReferenceID, destination, label, notificationDigestCopyRow{Text: factText, Label: label, URL: destination}
}

func buildStrategyMonthlyDigestFact(
	item NotificationEmailDigestItem,
	snapshot strategyNotificationSnapshot,
	workspaceURL string,
) (emailcopy.Fact, string, string, string, notificationDigestCopyRow) {
	monthly := snapshot.MonthlySummary
	periodLabel := monthly.PeriodStart.UTC().Format("January 2006")
	generatedLabel := snapshot.GeneratedAt.UTC().Format("January 2, 2006")
	periodStart := monthly.PeriodStart.UTC().Format("January 2, 2006")
	periodEnd := monthly.PeriodEnd.UTC().Format("January 2, 2006")
	keyResultSummary := strategyMonthlyKeyResultSummary(*monthly)
	factText := fmt.Sprintf(
		"%s covers %s and was generated on %s. The current snapshot has %d strategic pillars, including %d that need review; %d objectives, including %d at risk or off track and %d without a strategic alignment; and %s. From %s up to %s, %d linked tasks were completed.",
		item.Title,
		periodLabel,
		generatedLabel,
		monthly.PillarCount,
		monthly.PillarsNeedingReview,
		monthly.ObjectiveCount,
		monthly.AtRiskObjectives,
		monthly.UnalignedObjectives,
		keyResultSummary,
		periodStart,
		periodEnd,
		monthly.CompletedStories,
	)
	factReferenceID := "strategy_monthly_" + item.NotificationID.String()
	actionReferenceID := "strategy_monthly_action_" + item.NotificationID.String()
	destination := workspaceURL + "/strategy"
	label := notificationPlainText(item.Title, 180)
	fact := emailcopy.Fact{
		ReferenceID:  factReferenceID,
		Text:         factText,
		EntityTokens: nonEmptyStrings(label, periodLabel),
		ProtectedTokens: nonEmptyStrings(
			"generated on "+generatedLabel,
			fmt.Sprintf("has %d strategic pillars", monthly.PillarCount),
			fmt.Sprintf("including %d that need review", monthly.PillarsNeedingReview),
			fmt.Sprintf("%d objectives", monthly.ObjectiveCount),
			fmt.Sprintf("including %d at risk or off track", monthly.AtRiskObjectives),
			fmt.Sprintf("%d without a strategic alignment", monthly.UnalignedObjectives),
			keyResultSummary,
			fmt.Sprintf("From %s up to %s", periodStart, periodEnd),
			fmt.Sprintf("%d linked tasks were completed", monthly.CompletedStories),
		),
		Required: true,
	}
	return fact, actionReferenceID, destination, label, notificationDigestCopyRow{Text: factText, Label: label, URL: destination}
}

func strategyMonthlyKeyResultSummary(monthly strategyMonthlySummarySnapshot) string {
	switch {
	case monthly.KeyResultProgress != nil && monthly.KeyResultCount > 0:
		return fmt.Sprintf(
			"%.0f%% average key-result progress across %d %s",
			*monthly.KeyResultProgress,
			monthly.KeyResultCount,
			pluralWord(monthly.KeyResultCount, "key result", "key results"),
		)
	case monthly.KeyResultProgress != nil:
		// Snapshots created before keyResultCount was added still carry valid
		// progress; preserve it during the additive schema transition.
		return fmt.Sprintf("%.0f%% average key-result progress", *monthly.KeyResultProgress)
	case monthly.KeyResultCount > 0:
		return fmt.Sprintf(
			"average progress is unavailable for %d %s",
			monthly.KeyResultCount,
			pluralWord(monthly.KeyResultCount, "key result", "key results"),
		)
	default:
		return "there are no key results in the current snapshot"
	}
}

func strategyMissingElementLabels(elements []string) []string {
	labels := make([]string, 0, len(elements))
	for _, element := range elements {
		switch element {
		case "ultimate_goal":
			labels = append(labels, "ultimate goal")
		case "strategic_pillars":
			labels = append(labels, "strategic pillars")
		case "objectives":
			labels = append(labels, "objectives")
		}
	}
	return labels
}

func buildWeeklyStrategyDigestFacts(item NotificationEmailDigestItem, weekly strategyWeeklyCheckInSnapshot, workspaceURL string, rowLimit int) ([]emailcopy.Fact, map[string]string, map[string]string, map[string]string, []notificationDigestCopyRow) {
	facts := make([]emailcopy.Fact, 0, rowLimit)
	actionURLs := make(map[string]string, len(weekly.Objectives)+len(weekly.KeyResults))
	factActions := make(map[string]string, len(weekly.Objectives)+len(weekly.KeyResults)+1)
	factLabels := make(map[string]string, len(weekly.Objectives)+len(weekly.KeyResults))
	fallbackRows := make([]notificationDigestCopyRow, 0, rowLimit)

	summaryReferenceID := fmt.Sprintf("strategy_summary_%s", item.NotificationID.String())
	summaryText := fmt.Sprintf(
		"The weekly strategy check-in has %d at-risk objectives, %d objectives without a recent update, and %d incomplete key results without a recent update; recent means within %d days.",
		weekly.Counts.AtRiskObjectives,
		weekly.Counts.StaleObjectives,
		weekly.Counts.StaleKeyResults,
		weekly.StaleAfterDays,
	)
	protectedSummaryTokens := nonEmptyStrings(
		fmt.Sprintf("%d at-risk objectives", weekly.Counts.AtRiskObjectives),
		fmt.Sprintf("%d objectives without a recent update", weekly.Counts.StaleObjectives),
		fmt.Sprintf("%d incomplete key results without a recent update", weekly.Counts.StaleKeyResults),
		fmt.Sprintf("within %d days", weekly.StaleAfterDays),
	)
	if weekly.OmittedDetails != nil {
		omitted := weekly.OmittedDetails.Objectives + weekly.OmittedDetails.KeyResults
		if omitted > 0 {
			summaryText += fmt.Sprintf(" The saved detail shows a prioritized selection; %d additional %s available in Strategy.", omitted, pluralWord(omitted, "detail is", "details are"))
			protectedSummaryTokens = append(protectedSummaryTokens, fmt.Sprintf("%d additional %s available in Strategy", omitted, pluralWord(omitted, "detail is", "details are")))
		}
	}
	facts = append(facts, emailcopy.Fact{
		ReferenceID:     summaryReferenceID,
		Text:            summaryText,
		ProtectedTokens: protectedSummaryTokens,
		Required:        true,
	})
	factActions[summaryReferenceID] = digestActionStrategy
	fallbackRows = append(fallbackRows, notificationDigestCopyRow{Text: summaryText, URL: workspaceURL + "/strategy"})

	selectedObjectives, selectedKeyResults := selectWeeklyStrategyDetails(weekly.Objectives, weekly.KeyResults, rowLimit-1)

	for index, objective := range selectedObjectives {
		factReferenceID := fmt.Sprintf("strategy_objective_%d_%s", index+1, objective.ID.String())
		actionReferenceID := fmt.Sprintf("strategy_objective_action_%d_%s", index+1, objective.ID.String())
		factText := strategyObjectiveFactText(objective, weekly.StaleAfterDays)
		destination := strategyObjectiveNotificationURL(workspaceURL, item.NotificationID, objective.ID)
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     factReferenceID,
			Text:            factText,
			EntityTokens:    nonEmptyStrings(objective.Name),
			ProtectedTokens: strategyObjectiveProtectedTokens(objective),
			Required:        true,
		})
		actionURLs[actionReferenceID] = destination
		factActions[factReferenceID] = actionReferenceID
		factLabels[factReferenceID] = objective.Name
		fallbackRows = append(fallbackRows, notificationDigestCopyRow{Text: factText, Label: objective.Name, URL: destination})
	}

	for index, keyResult := range selectedKeyResults {
		factReferenceID := fmt.Sprintf("strategy_key_result_%d_%s", index+1, keyResult.ID.String())
		actionReferenceID := fmt.Sprintf("strategy_key_result_action_%d_%s", index+1, keyResult.ID.String())
		factText := strategyKeyResultFactText(keyResult, weekly.StaleAfterDays)
		destination := strategyKeyResultNotificationURL(workspaceURL, item.NotificationID, keyResult.ID, keyResult.ObjectiveID)
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     factReferenceID,
			Text:            factText,
			EntityTokens:    nonEmptyStrings(keyResult.Name, keyResult.ObjectiveName),
			ProtectedTokens: strategyKeyResultProtectedTokens(keyResult),
			Required:        true,
		})
		actionURLs[actionReferenceID] = destination
		factActions[factReferenceID] = actionReferenceID
		factLabels[factReferenceID] = keyResult.Name
		fallbackRows = append(fallbackRows, notificationDigestCopyRow{Text: factText, Label: keyResult.Name, URL: destination})
	}

	return facts, actionURLs, factActions, factLabels, fallbackRows
}

func selectWeeklyStrategyDetails(
	objectives []strategyObjectiveSnapshot,
	keyResults []strategyKeyResultSnapshot,
	limit int,
) ([]strategyObjectiveSnapshot, []strategyKeyResultSnapshot) {
	if limit <= 0 {
		return nil, nil
	}
	objectiveLimit := min(len(objectives), (limit+1)/2)
	keyResultLimit := min(len(keyResults), limit-objectiveLimit)
	objectiveLimit = min(len(objectives), limit-keyResultLimit)
	if remaining := limit - objectiveLimit - keyResultLimit; remaining > 0 {
		keyResultLimit = min(len(keyResults), keyResultLimit+remaining)
	}
	return objectives[:objectiveLimit], keyResults[:keyResultLimit]
}

func strategyObjectiveFactText(objective strategyObjectiveSnapshot, staleAfterDays int) string {
	parts := []string{objective.Name}
	if objective.Health != nil && strings.TrimSpace(*objective.Health) != "" {
		parts = append(parts, "health is "+strings.TrimSpace(*objective.Health))
	}
	if objective.Status != nil && strings.TrimSpace(objective.Status.Name) != "" {
		parts = append(parts, "status is "+strings.TrimSpace(objective.Status.Name))
	}
	if containsValue(objective.Reasons, "stale") {
		parts = append(parts, fmt.Sprintf("has not had a recent update; recent means within %d days", staleAfterDays))
	}
	if !objective.UpdatedAt.IsZero() {
		parts = append(parts, "last updated on "+objective.UpdatedAt.UTC().Format("January 2, 2006"))
	}
	if objective.EndDate != nil {
		parts = append(parts, "ends on "+objective.EndDate.UTC().Format("January 2, 2006"))
	}
	return strings.Join(parts, "; ") + "."
}

func strategyObjectiveProtectedTokens(objective strategyObjectiveSnapshot) []string {
	values := make([]string, 0, 6)
	if objective.Health != nil {
		values = append(values, "health is "+strings.TrimSpace(*objective.Health))
	}
	if objective.Status != nil {
		values = append(values, "status is "+strings.TrimSpace(objective.Status.Name))
	}
	if containsValue(objective.Reasons, "stale") {
		values = append(values, "has not had a recent update")
	}
	if !objective.UpdatedAt.IsZero() {
		values = append(values, "last updated on "+objective.UpdatedAt.UTC().Format("January 2, 2006"))
	}
	if objective.EndDate != nil {
		values = append(values, "ends on "+objective.EndDate.UTC().Format("January 2, 2006"))
	}
	return nonEmptyStrings(values...)
}

func strategyKeyResultFactText(keyResult strategyKeyResultSnapshot, staleAfterDays int) string {
	parts := []string{fmt.Sprintf("%s under %s", keyResult.Name, keyResult.ObjectiveName)}
	if keyResult.ObjectiveHealth != nil && strings.TrimSpace(*keyResult.ObjectiveHealth) != "" {
		parts = append(parts, "objective health is "+strings.TrimSpace(*keyResult.ObjectiveHealth))
	}
	if keyResult.ObjectiveStatus != nil && strings.TrimSpace(keyResult.ObjectiveStatus.Name) != "" {
		parts = append(parts, "objective status is "+strings.TrimSpace(keyResult.ObjectiveStatus.Name))
	}
	if strings.TrimSpace(keyResult.MeasurementType) != "" {
		parts = append(parts, "measurement is "+strings.TrimSpace(keyResult.MeasurementType))
	}
	if keyResult.CurrentValue != nil && keyResult.TargetValue != nil {
		parts = append(parts, fmt.Sprintf("current value is %s and target value is %s", strategyValue(keyResult.CurrentValue), strategyValue(keyResult.TargetValue)))
	}
	parts = append(parts, fmt.Sprintf("is incomplete and has not had a recent update; recent means within %d days", staleAfterDays))
	if !keyResult.UpdatedAt.IsZero() {
		parts = append(parts, "last updated on "+keyResult.UpdatedAt.UTC().Format("January 2, 2006"))
	}
	if keyResult.EndDate != nil {
		parts = append(parts, "ends on "+keyResult.EndDate.UTC().Format("January 2, 2006"))
	}
	return strings.Join(parts, "; ") + "."
}

func strategyKeyResultProtectedTokens(keyResult strategyKeyResultSnapshot) []string {
	values := []string{"is incomplete", "has not had a recent update"}
	if keyResult.ObjectiveHealth != nil {
		values = append(values, "objective health is "+strings.TrimSpace(*keyResult.ObjectiveHealth))
	}
	if keyResult.ObjectiveStatus != nil {
		values = append(values, "objective status is "+strings.TrimSpace(keyResult.ObjectiveStatus.Name))
	}
	if measurementType := strings.TrimSpace(keyResult.MeasurementType); measurementType != "" {
		values = append(values, "measurement is "+measurementType)
	}
	if keyResult.CurrentValue != nil && keyResult.TargetValue != nil {
		values = append(values,
			"current value is "+strategyValue(keyResult.CurrentValue),
			"target value is "+strategyValue(keyResult.TargetValue),
		)
	}
	if !keyResult.UpdatedAt.IsZero() {
		values = append(values, "last updated on "+keyResult.UpdatedAt.UTC().Format("January 2, 2006"))
	}
	if keyResult.EndDate != nil {
		values = append(values, "ends on "+keyResult.EndDate.UTC().Format("January 2, 2006"))
	}
	return nonEmptyStrings(values...)
}

func strategyObjectiveNotificationURL(workspaceURL string, notificationID, objectiveID uuid.UUID) string {
	return fmt.Sprintf(
		"%s/notifications/%s?entityId=%s&entityType=objective",
		workspaceURL,
		notificationID.String(),
		url.QueryEscape(objectiveID.String()),
	)
}

func strategyKeyResultNotificationURL(workspaceURL string, notificationID, keyResultID, objectiveID uuid.UUID) string {
	return fmt.Sprintf(
		"%s/notifications/%s?entityId=%s&entityType=key_result&objectiveId=%s",
		workspaceURL,
		notificationID.String(),
		url.QueryEscape(keyResultID.String()),
		url.QueryEscape(objectiveID.String()),
	)
}

func buildGeneratedNotificationDigestCopy(input notificationDigestCopyInput, output emailcopy.Output) (notificationDigestCopy, error) {
	knownFacts := make(map[string]struct{}, len(input.Request.Facts))
	for _, fact := range input.Request.Facts {
		knownFacts[fact.ReferenceID] = struct{}{}
	}
	rows := make([]notificationDigestCopyRow, 0, len(output.Rows))
	for _, row := range output.Rows {
		if _, exists := knownFacts[row.ReferenceID]; !exists {
			return notificationDigestCopy{}, fmt.Errorf("email copy row %q cites an unknown fact", row.ReferenceID)
		}
		destination := ""
		if actionReferenceID := input.FactActions[row.ReferenceID]; actionReferenceID != "" {
			var ok bool
			destination, ok = input.Actions[actionReferenceID]
			if !ok || strings.TrimSpace(destination) == "" {
				return notificationDigestCopy{}, fmt.Errorf("email copy row %q has no trusted destination", row.ReferenceID)
			}
		}
		rows = append(rows, notificationDigestCopyRow{
			Text:  notificationPlainText(row.Text, 360),
			Label: input.FactLabels[row.ReferenceID],
			URL:   destination,
		})
	}
	if len(rows) == 0 {
		return notificationDigestCopy{}, errors.New("email copy has no rows")
	}

	primaryCTA := notificationDigestCopyCTA{}
	preferredReferenceID := digestActionPrimary
	if input.HasStrategySnapshot {
		preferredReferenceID = digestActionStrategy
	}
	for _, cta := range output.CTAs {
		if cta.ReferenceID != preferredReferenceID {
			continue
		}
		if destination, ok := input.Actions[cta.ReferenceID]; ok {
			primaryCTA = notificationDigestCopyCTA{Label: notificationPlainText(cta.Label, 48), URL: destination}
			break
		}
	}
	if primaryCTA.URL == "" {
		primaryCTA = input.Fallback.CTA
	}

	introParts := []string{notificationPlainText(output.Intro.Text, 420)}
	if output.SenderProse != nil {
		introParts = append(introParts, notificationPlainText(output.SenderProse.Text, 320))
	}
	if output.ReplyPrompt != nil {
		introParts = append(introParts, notificationPlainText(output.ReplyPrompt.Text, 320))
	}

	copy := notificationDigestCopy{
		Subject:             notificationPlainText(output.Subject.Text, 90),
		Heading:             notificationPlainText(output.H1.Text, 110),
		Intro:               strings.Join(nonEmptyStrings(introParts...), " "),
		Rows:                rows,
		CTA:                 primaryCTA,
		NotificationsURL:    input.NotificationsURL,
		HasStrategySnapshot: input.HasStrategySnapshot,
	}
	if input.HasStrategySnapshot {
		copy.Sender = mailer.SenderProfileMaya
	}
	if copy.Subject == "" || copy.Heading == "" || copy.Intro == "" {
		return notificationDigestCopy{}, errors.New("email copy has an empty visible field")
	}
	return copy, nil
}
