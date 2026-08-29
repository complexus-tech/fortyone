package maya

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/workschedule"
	"github.com/complexus-tech/projects-api/internal/platform/workweek"
	"github.com/google/uuid"
)

func normalizePlanInput(input PlanInput) (PlanInput, error) {
	if input.WorkspaceID == uuid.Nil || input.Story.ID == uuid.Nil || input.Story.Workspace != input.WorkspaceID {
		return PlanInput{}, ErrInvalidPlanInput
	}
	if !input.WindowEnd.After(input.WindowStart) {
		return PlanInput{}, fmt.Errorf("%w: planning window end must be after start", ErrInvalidPlanInput)
	}
	if len(input.Candidates) == 0 {
		return PlanInput{}, fmt.Errorf("%w: at least one candidate is required", ErrInvalidPlanInput)
	}
	if input.DurationMinutes <= 0 {
		input.DurationMinutes = estimatedWorkDurationMinutes(input.Story)
	}
	if input.DurationMinutes <= 0 {
		return PlanInput{}, ErrMissingDuration
	}
	if input.DurationMinutes > MaximumEstimatedDurationMinutes {
		return PlanInput{}, fmt.Errorf("%w: duration exceeds %d minutes", ErrInvalidPlanInput, MaximumEstimatedDurationMinutes)
	}
	if input.MinimumFocusBlockMinutes <= 0 && input.Story.MinimumFocusBlockMinutes != nil {
		input.MinimumFocusBlockMinutes = *input.Story.MinimumFocusBlockMinutes
	}
	if input.MinimumFocusBlockMinutes > input.DurationMinutes {
		input.MinimumFocusBlockMinutes = input.DurationMinutes
	}
	input.WindowStart = input.WindowStart.UTC()
	input.WindowEnd = input.WindowEnd.UTC()
	if input.AsOf.IsZero() {
		input.AsOf = input.WindowStart
	} else {
		input.AsOf = input.AsOf.UTC()
	}
	input.WorkingDays = workweek.Normalize(input.WorkingDays)
	for index := range input.Candidates {
		candidate := &input.Candidates[index]
		workingDays := candidate.WorkingDays
		if len(workingDays) == 0 {
			workingDays = input.WorkingDays
		}
		schedule := workschedule.Normalize(workingDays, candidate.WorkingStartMinute, candidate.WorkingEndMinute)
		candidate.WorkingDays = schedule.WorkingDays
		candidate.WorkingStartMinute = schedule.StartMinute
		candidate.WorkingEndMinute = schedule.EndMinute
	}
	return input, nil
}

func estimatedWorkDurationMinutes(story Story) int {
	if story.EstimatedDurationMinutes != nil && *story.EstimatedDurationMinutes > 0 {
		return *story.EstimatedDurationMinutes
	}
	return 0
}

func clampWindowToSprint(input PlanInput, candidate CandidateSchedule) (time.Time, time.Time) {
	if input.Story.SprintSummary == nil {
		return input.WindowStart, input.WindowEnd
	}
	location := calendarLocation(candidate.Timezone)
	sprintStart := workdayBoundary(input.Story.SprintSummary.StartDate, candidate.WorkingStartMinute, location)
	sprintEnd := workdayBoundary(input.Story.SprintSummary.EndDate, candidate.WorkingEndMinute, location)
	startAt := input.WindowStart
	endAt := input.WindowEnd
	if startAt.Before(sprintStart) {
		startAt = sprintStart
	}
	if endAt.After(sprintEnd) {
		endAt = sprintEnd
	}
	return startAt, endAt
}

func planWorkWindow(candidate CandidateSchedule, startAt, endAt time.Time, duration time.Duration) (timeSlot, bool) {
	if duration <= 0 {
		return timeSlot{}, false
	}
	durationMinutes := int(duration / time.Minute)
	plan := planWorkSegments(candidate, startAt, endAt, durationMinutes, 0)
	if plan.remainingMinutes > 0 || len(plan.segments) == 0 {
		return timeSlot{}, false
	}
	return timeSlot{start: plan.segments[0].start, end: plan.segments[len(plan.segments)-1].end}, true
}

func planWorkSegments(candidate CandidateSchedule, startAt, endAt time.Time, durationMinutes, minimumFocusMinutes int) segmentPlan {
	if minimumFocusMinutes <= 0 {
		// Automatic fills the earliest useful windows, keeping each segment as
		// large as the calendar allows while retaining only a small lower bound.
		minimumFocusMinutes = automaticMinimumFocusBlockMinutes
		if minimumFocusMinutes > durationMinutes {
			minimumFocusMinutes = durationMinutes
		}
		return planWorkSegmentsWithLimits(candidate, startAt, endAt, durationMinutes, minimumFocusMinutes, durationMinutes)
	}

	return planWorkSegmentsWithLimits(candidate, startAt, endAt, durationMinutes, minimumFocusMinutes, maxFocusBlockMinutes)
}

func planWorkSegmentsWithLimits(candidate CandidateSchedule, startAt, endAt time.Time, durationMinutes, minimumFocusMinutes, maximumFocusMinutes int) segmentPlan {
	occupied := occupiedSlots(candidate)
	location := calendarLocation(candidate.Timezone)
	cursor := preferredPlanningCursor(candidate, startAt, location)
	endAt = endAt.In(location)
	remaining := time.Duration(durationMinutes) * time.Minute
	minimumFocus := time.Duration(minimumFocusMinutes) * time.Minute
	maximumFocus := time.Duration(maximumFocusMinutes) * time.Minute
	if maximumFocus < minimumFocus {
		maximumFocus = minimumFocus
	}
	segments := make([]timeSlot, 0, durationMinutes/minimumFocusMinutes+1)

	for cursor.Before(endAt) {
		if !workweek.IsWorkingDay(cursor, candidate.WorkingDays) {
			cursor = nextPlanningWorkdayStart(cursor, candidate)
			continue
		}

		dayStart := workdayBoundary(cursor, candidate.WorkingStartMinute, location)
		dayEnd := workdayBoundary(cursor, candidate.WorkingEndMinute, location)
		if cursor.Before(dayStart) {
			cursor = dayStart
		}
		if !cursor.Before(dayEnd) {
			cursor = nextPlanningWorkdayStart(cursor, candidate)
			continue
		}

		if nextCursor, blocked := advancePastOccupiedSlot(cursor, occupied); blocked {
			cursor = nextCursor
			continue
		}

		nextBoundary := minTime(dayEnd, endAt)
		for _, slot := range occupied {
			if slot.start.After(cursor) && slot.start.Before(nextBoundary) {
				nextBoundary = slot.start
				break
			}
		}
		if !nextBoundary.After(cursor) {
			cursor = cursor.Add(planningStartGranularityMinutes * time.Minute)
			continue
		}
		available := nextBoundary.Sub(cursor)
		if available < minimumFocus && remaining > available {
			cursor = nextBoundary
			continue
		}
		take := minDuration(remaining, available, maximumFocus)
		remainingAfter := remaining - take
		if remainingAfter > 0 && remainingAfter < minimumFocus {
			adjustment := minimumFocus - remainingAfter
			if take-adjustment >= minimumFocus {
				take -= adjustment
			} else if available >= remaining {
				take = remaining
			} else {
				cursor = nextBoundary
				continue
			}
		}
		if take < minimumFocus && take != remaining {
			cursor = nextBoundary
			continue
		}
		segments = append(segments, timeSlot{start: cursor.UTC(), end: cursor.Add(take).UTC()})
		remaining -= take
		if remaining == 0 {
			return segmentPlan{segments: segments}
		}
		cursor = cursor.Add(take)
	}

	return segmentPlan{segments: segments, remainingMinutes: int(remaining / time.Minute)}
}

func minDuration(values ...time.Duration) time.Duration {
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}

func advancePastOccupiedSlot(cursor time.Time, occupied []timeSlot) (time.Time, bool) {
	for _, slot := range occupied {
		if cursor.Before(slot.end) && !cursor.Before(slot.start) {
			return alignToPlanningGranularity(slot.end.In(cursor.Location())), true
		}
	}
	return cursor, false
}

func nextWorkdayStart(value time.Time, candidate CandidateSchedule) time.Time {
	next := workdayBoundary(value.AddDate(0, 0, 1), candidate.WorkingStartMinute, value.Location())
	for !workweek.IsWorkingDay(next, candidate.WorkingDays) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func preferredPlanningCursor(candidate CandidateSchedule, startAt time.Time, location *time.Location) time.Time {
	cursor := alignToPlanningGranularity(startAt.In(location))
	if candidate.PreferredStartMinute == nil {
		return cursor
	}
	preferred := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, location).
		Add(time.Duration(clampPreferredStartMinute(*candidate.PreferredStartMinute, candidate)) * time.Minute)
	if preferred.After(cursor) {
		return preferred
	}
	return cursor
}

func nextPlanningWorkdayStart(value time.Time, candidate CandidateSchedule) time.Time {
	next := nextWorkdayStart(value, candidate)
	if candidate.PreferredStartMinute == nil {
		return next
	}
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()).
		Add(time.Duration(clampPreferredStartMinute(*candidate.PreferredStartMinute, candidate)) * time.Minute)
}

func clampPreferredStartMinute(value int, candidate CandidateSchedule) int {
	minimum := candidate.WorkingStartMinute
	maximum := candidate.WorkingEndMinute - planningStartGranularityMinutes
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func workdayBoundary(value time.Time, minute int, location *time.Location) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, location).
		Add(time.Duration(minute) * time.Minute)
}

func minTime(left, right time.Time) time.Time {
	if left.Before(right) {
		return left
	}
	return right
}

func occupiedSlots(candidate CandidateSchedule) []timeSlot {
	location := calendarLocation(candidate.Timezone)
	slots := make([]timeSlot, 0, len(candidate.BusyWindows)+len(candidate.Blocks))
	preemptible := make(map[uuid.UUID]struct{}, len(candidate.PreemptibleBlockIDs))
	for _, blockID := range candidate.PreemptibleBlockIDs {
		preemptible[blockID] = struct{}{}
	}
	for _, window := range candidate.BusyWindows {
		slots = append(slots, timeSlot{start: window.StartAt.In(location), end: window.EndAt.In(location)})
	}
	for _, block := range candidate.Blocks {
		if _, ok := preemptible[block.ID]; ok {
			continue
		}
		slots = append(slots, timeSlot{start: block.StartAt.In(location), end: block.EndAt.In(location)})
	}
	sort.SliceStable(slots, func(i, j int) bool {
		return slots[i].start.Before(slots[j].start)
	})
	return slots
}

const (
	priorityUrgent = iota
	priorityHigh
	priorityMedium
	priorityLow
	priorityNone
)

func storyPriorityRank(priority string) int {
	switch strings.TrimSpace(priority) {
	case "Urgent":
		return priorityUrgent
	case "High":
		return priorityHigh
	case "Medium":
		return priorityMedium
	case "Low":
		return priorityLow
	default:
		return priorityNone
	}
}

func preemptibleBlockIDs(story Story, blocks []ScheduleBlock, now time.Time) []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	for _, block := range blocks {
		if !shouldPreemptBlock(story, block, now) {
			continue
		}
		ids = append(ids, block.ID)
	}
	return ids
}

func preemptedBlockIDsForSegments(blocks []ScheduleBlock, preemptibleIDs []uuid.UUID, segments []timeSlot) []uuid.UUID {
	preemptible := make(map[uuid.UUID]struct{}, len(preemptibleIDs))
	for _, blockID := range preemptibleIDs {
		preemptible[blockID] = struct{}{}
	}
	ids := make([]uuid.UUID, 0)
	for _, block := range blocks {
		if _, ok := preemptible[block.ID]; !ok {
			continue
		}
		for _, segment := range segments {
			if block.StartAt.Before(segment.end) && block.EndAt.After(segment.start) {
				ids = append(ids, block.ID)
				break
			}
		}
	}
	return ids
}

func shouldPreemptBlock(story Story, block ScheduleBlock, now time.Time) bool {
	if block.ID == uuid.Nil || block.Source != ScheduleBlockSourceMaya || block.StoryID == nil || *block.StoryID == story.ID {
		return false
	}
	if block.WorkspaceID != story.Workspace || block.IsLocked || !block.StartAt.After(now) {
		return false
	}
	if storyPriorityRank(story.Priority) >= storyPriorityRank(block.StoryPriority) {
		return false
	}
	if story.EndDate == nil {
		return block.StoryEndDate == nil
	}
	if block.StoryEndDate == nil {
		return true
	}
	return !calendarDateOnly(story.EndDate.UTC()).After(calendarDateOnly(block.StoryEndDate.UTC()))
}

func calendarDateOnly(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func alignToPlanningGranularity(value time.Time) time.Time {
	value = value.Truncate(time.Minute)
	minute := value.Minute()
	remainder := minute % planningStartGranularityMinutes
	if remainder == 0 && value.Second() == 0 && value.Nanosecond() == 0 {
		return value
	}
	return value.Add(time.Duration(planningStartGranularityMinutes-remainder) * time.Minute).Truncate(time.Minute)
}

func calendarLocation(timezone string) *time.Location {
	location, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		return time.UTC
	}
	return location
}

func hasStoryScheduleBlock(blocks []ScheduleBlock, storyID uuid.UUID) bool {
	for _, block := range blocks {
		if block.StoryID != nil && *block.StoryID == storyID {
			return true
		}
	}
	return false
}

func assignmentReason(choice candidateChoice) string {
	return assignmentReasonForMember(choice.candidate.Member)
}

func assignmentReasonForMember(member MemberWorkload) string {
	name := strings.TrimSpace(member.FullName)
	if name == "" {
		name = strings.TrimSpace(member.Username)
	}
	if name == "" {
		name = "this teammate"
	}
	if !isRecentlyActive(member.LastStoryActivityAt) {
		return fmt.Sprintf("Maya selected %s because no recently active alternative was available and they have %d open items with %d estimate points.", name, member.OpenStories, member.EstimateTotal)
	}
	return fmt.Sprintf("Maya selected %s because they have the strongest available fit and currently have %d open items with %d estimate points.", name, member.OpenStories, member.EstimateTotal)
}

func scheduleReason(choice candidateChoice, segmentIndex, segmentCount int) string {
	segment := choice.segments[segmentIndex]
	return fmt.Sprintf("Maya scheduled focus segment %d of %d from %s to %s without overlapping existing busy time or locked work.", segmentIndex+1, segmentCount, segment.start.Format(time.RFC3339), segment.end.Format(time.RFC3339))
}

func planSummary(storyTitle string, choice candidateChoice) string {
	name := strings.TrimSpace(choice.candidate.Member.FullName)
	if name == "" {
		name = strings.TrimSpace(choice.candidate.Member.Username)
	}
	if name == "" {
		name = "the selected teammate"
	}
	if choice.remainingMinutes > 0 {
		return fmt.Sprintf("Maya recommends assigning %q to %s, scheduling %d focus segment(s), and finding another slot for %s.", storyTitle, name, len(choice.segments), formatMinutes(choice.remainingMinutes))
	}
	return fmt.Sprintf("Maya recommends assigning %q to %s and scheduling %d focus segment(s) from %s to %s.", storyTitle, name, len(choice.segments), choice.plan.start.Format(time.RFC3339), choice.plan.end.Format(time.RFC3339))
}

func partialScheduleReason(scheduledMinutes, remainingMinutes int) string {
	return fmt.Sprintf("Maya scheduled %s. %s still needs a slot.", formatMinutes(scheduledMinutes), formatMinutes(remainingMinutes))
}

func formatMinutes(minutes int) string {
	hours := minutes / 60
	remainder := minutes % 60
	if hours == 0 {
		return fmt.Sprintf("%dm", remainder)
	}
	if remainder == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, remainder)
}
