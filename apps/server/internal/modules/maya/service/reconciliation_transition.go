package maya

import (
	"sort"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func buildStoryScheduleTransition(
	story Story,
	userID uuid.UUID,
	previousBlocks []ScheduleBlock,
	segments []ScheduleSegmentInput,
	timezone string,
	status string,
	reason string,
) *events.StoryScheduleTransition {
	if userID == uuid.Nil {
		return nil
	}
	for _, block := range previousBlocks {
		if block.UserID != userID {
			// Assignment/reassignment has its own reason-aware StoryUpdated event.
			// Suppress a second schedule-move notification for the same decision.
			return nil
		}
	}
	previousStart, previousEnd := scheduleBlockBounds(previousBlocks)
	startAt, endAt := scheduleSegmentBounds(segments)
	previousState := events.StoryScheduleState(story.AutoSchedulingStatus)
	state := events.StoryScheduleState(status)
	kind := events.StoryScheduleTransitionKind("")
	wasLocked := false
	for _, block := range previousBlocks {
		wasLocked = wasLocked || block.IsLocked
	}

	location := time.UTC
	if parsed, err := time.LoadLocation(timezone); err == nil {
		location = parsed
	} else {
		timezone = "UTC"
	}
	previousLocalDate := localDate(previousStart, location)
	localDateValue := localDate(startAt, location)
	shiftMinutes := 0
	if change, ok := selectMeaningfulScheduleChange(previousBlocks, segments, location); ok {
		kind = change.Kind
		previousStart = &change.PreviousStartAt
		previousEnd = &change.PreviousEndAt
		startAt = &change.StartAt
		endAt = &change.EndAt
		previousLocalDate = change.PreviousLocalDate
		localDateValue = change.LocalDate
		shiftMinutes = change.ShiftMinutes
	}
	if kind == "" && story.AutoSchedulingLocked && !wasLocked && len(segments) > 0 {
		kind = events.StoryScheduleTransitionLocked
	} else if kind == "" && !story.AutoSchedulingLocked && wasLocked {
		kind = events.StoryScheduleTransitionUnlocked
	}
	if kind == "" && len(segments) > 0 && len(previousBlocks) == 0 {
		kind = events.StoryScheduleTransitionFirstSchedule
	}
	if kind == "" && previousState != state {
		kind = events.StoryScheduleTransitionStateChanged
	}
	if kind == "" && equalScheduleReasonChanged(story.AutoSchedulingReason, reason) &&
		(status == AutoSchedulingStatusNeedsTime || status == AutoSchedulingStatusCannotFit || status == AutoSchedulingStatusAtRisk) {
		kind = events.StoryScheduleTransitionStateChanged
	}
	if kind == "" {
		return nil
	}
	return &events.StoryScheduleTransition{
		Kind:              kind,
		UserID:            userID,
		PreviousState:     previousState,
		State:             state,
		PreviousStartAt:   previousStart,
		StartAt:           startAt,
		PreviousEndAt:     previousEnd,
		EndAt:             endAt,
		Timezone:          timezone,
		PreviousLocalDate: previousLocalDate,
		LocalDate:         localDateValue,
		ShiftMinutes:      shiftMinutes,
	}
}

type meaningfulScheduleChange struct {
	Kind              events.StoryScheduleTransitionKind
	PreviousStartAt   time.Time
	StartAt           time.Time
	PreviousEndAt     time.Time
	EndAt             time.Time
	PreviousLocalDate string
	LocalDate         string
	ShiftMinutes      int
}

func selectMeaningfulScheduleChange(
	previousBlocks []ScheduleBlock,
	segments []ScheduleSegmentInput,
	location *time.Location,
) (meaningfulScheduleChange, bool) {
	if scheduleTimesEqual(previousBlocks, segments) {
		return meaningfulScheduleChange{}, false
	}
	previousBySegment := make(map[int]ScheduleBlock, len(previousBlocks))
	for _, block := range previousBlocks {
		previousBySegment[block.SegmentIndex] = block
	}
	segmentsByIndex := make(map[int]ScheduleSegmentInput, len(segments))
	for _, segment := range segments {
		segmentsByIndex[segment.SegmentIndex] = segment
	}

	var selected meaningfulScheduleChange
	selectedSegmentIndex := 0
	selectedFound := false
	selectedDayChanged := false
	consider := func(previous ScheduleBlock, segment ScheduleSegmentInput) {
		previousLocalDate := previous.StartAt.In(location).Format(time.DateOnly)
		localDateValue := segment.StartAt.In(location).Format(time.DateOnly)
		startShiftMinutes := int(segment.StartAt.Sub(previous.StartAt).Minutes())
		endShiftMinutes := int(segment.EndAt.Sub(previous.EndAt).Minutes())
		dayChanged := previousLocalDate != localDateValue
		shiftMinutes := startShiftMinutes
		if !dayChanged && absoluteInt(startShiftMinutes) < 60 {
			if absoluteInt(endShiftMinutes) < 60 {
				return
			}
			shiftMinutes = endShiftMinutes
		}

		shouldSelect := !selectedFound ||
			(dayChanged && !selectedDayChanged) ||
			(dayChanged == selectedDayChanged && absoluteInt(shiftMinutes) > absoluteInt(selected.ShiftMinutes)) ||
			(dayChanged == selectedDayChanged && absoluteInt(shiftMinutes) == absoluteInt(selected.ShiftMinutes) && segment.SegmentIndex < selectedSegmentIndex)
		if !shouldSelect {
			return
		}

		kind := events.StoryScheduleTransitionMoved
		if dayChanged {
			kind = events.StoryScheduleTransitionDayChanged
		}
		selected = meaningfulScheduleChange{
			Kind:              kind,
			PreviousStartAt:   previous.StartAt,
			StartAt:           segment.StartAt,
			PreviousEndAt:     previous.EndAt,
			EndAt:             segment.EndAt,
			PreviousLocalDate: previousLocalDate,
			LocalDate:         localDateValue,
			ShiftMinutes:      shiftMinutes,
		}
		selectedSegmentIndex = segment.SegmentIndex
		selectedFound = true
		selectedDayChanged = dayChanged
	}

	for _, segment := range segments {
		previous, exists := previousBySegment[segment.SegmentIndex]
		if exists {
			consider(previous, segment)
		}
	}
	for _, segment := range segments {
		if _, matched := previousBySegment[segment.SegmentIndex]; matched || len(previousBlocks) == 0 {
			continue
		}
		consider(nearestScheduleBlock(previousBlocks, segment.SegmentIndex), segment)
	}
	for _, previous := range previousBlocks {
		if _, matched := segmentsByIndex[previous.SegmentIndex]; matched || len(segments) == 0 {
			continue
		}
		consider(previous, nearestScheduleSegment(segments, previous.SegmentIndex))
	}

	return selected, selectedFound
}

func scheduleTimesEqual(previousBlocks []ScheduleBlock, segments []ScheduleSegmentInput) bool {
	if len(previousBlocks) != len(segments) {
		return false
	}
	previousTimes := make([]ScheduleBlock, len(previousBlocks))
	copy(previousTimes, previousBlocks)
	segmentTimes := make([]ScheduleSegmentInput, len(segments))
	copy(segmentTimes, segments)
	sort.Slice(previousTimes, func(i, j int) bool {
		if !previousTimes[i].StartAt.Equal(previousTimes[j].StartAt) {
			return previousTimes[i].StartAt.Before(previousTimes[j].StartAt)
		}
		return previousTimes[i].EndAt.Before(previousTimes[j].EndAt)
	})
	sort.Slice(segmentTimes, func(i, j int) bool {
		if !segmentTimes[i].StartAt.Equal(segmentTimes[j].StartAt) {
			return segmentTimes[i].StartAt.Before(segmentTimes[j].StartAt)
		}
		return segmentTimes[i].EndAt.Before(segmentTimes[j].EndAt)
	})
	for index := range previousTimes {
		if !previousTimes[index].StartAt.Equal(segmentTimes[index].StartAt) || !previousTimes[index].EndAt.Equal(segmentTimes[index].EndAt) {
			return false
		}
	}
	return true
}

func nearestScheduleBlock(blocks []ScheduleBlock, segmentIndex int) ScheduleBlock {
	nearest := blocks[0]
	nearestDistance := absoluteInt(nearest.SegmentIndex - segmentIndex)
	for _, block := range blocks[1:] {
		distance := absoluteInt(block.SegmentIndex - segmentIndex)
		if distance < nearestDistance || (distance == nearestDistance && block.SegmentIndex < nearest.SegmentIndex) {
			nearest = block
			nearestDistance = distance
		}
	}
	return nearest
}

func nearestScheduleSegment(segments []ScheduleSegmentInput, segmentIndex int) ScheduleSegmentInput {
	nearest := segments[0]
	nearestDistance := absoluteInt(nearest.SegmentIndex - segmentIndex)
	for _, segment := range segments[1:] {
		distance := absoluteInt(segment.SegmentIndex - segmentIndex)
		if distance < nearestDistance || (distance == nearestDistance && segment.SegmentIndex < nearest.SegmentIndex) {
			nearest = segment
			nearestDistance = distance
		}
	}
	return nearest
}

func equalScheduleReasonChanged(previous *string, next string) bool {
	return previous == nil || *previous != next
}

func scheduleBlockBounds(blocks []ScheduleBlock) (*time.Time, *time.Time) {
	if len(blocks) == 0 {
		return nil, nil
	}
	startAt := blocks[0].StartAt
	endAt := blocks[0].EndAt
	for _, block := range blocks[1:] {
		if block.StartAt.Before(startAt) {
			startAt = block.StartAt
		}
		if block.EndAt.After(endAt) {
			endAt = block.EndAt
		}
	}
	return &startAt, &endAt
}

func scheduleSegmentBounds(segments []ScheduleSegmentInput) (*time.Time, *time.Time) {
	if len(segments) == 0 {
		return nil, nil
	}
	startAt := segments[0].StartAt
	endAt := segments[0].EndAt
	for _, segment := range segments[1:] {
		if segment.StartAt.Before(startAt) {
			startAt = segment.StartAt
		}
		if segment.EndAt.After(endAt) {
			endAt = segment.EndAt
		}
	}
	return &startAt, &endAt
}

func localDate(value *time.Time, location *time.Location) string {
	if value == nil {
		return ""
	}
	return value.In(location).Format(time.DateOnly)
}

func absoluteInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func uuidSliceContains(values []uuid.UUID, target uuid.UUID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
