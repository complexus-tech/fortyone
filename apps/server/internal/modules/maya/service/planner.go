package maya

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

var ErrInvalidPlanInput = errors.New("invalid maya plan input")

const (
	defaultDurationMinutes = 60
	minimumSlotMinutes     = 30
	workdayStartHour       = 9
	workdayEndHour         = 17
)

type Planner struct{}

func NewPlanner() Planner {
	return Planner{}
}

func (p Planner) Plan(input PlanInput) (PlanResult, error) {
	normalized, err := normalizePlanInput(input)
	if err != nil {
		return PlanResult{}, err
	}

	candidates := make([]candidateChoice, 0, len(normalized.Candidates))
	for _, candidate := range normalized.Candidates {
		if candidate.Member.UserID == uuid.Nil {
			continue
		}
		slot, ok := earliestSlot(candidate, normalized.WindowStart, normalized.WindowEnd, time.Duration(normalized.DurationMinutes)*time.Minute)
		if !ok {
			continue
		}
		candidates = append(candidates, candidateChoice{
			candidate: candidate,
			slot:      slot,
		})
	}

	if len(candidates) == 0 {
		action := CoreAction{
			WorkspaceID: normalized.WorkspaceID,
			StoryID:     normalized.Story.ID,
			Type:        ActionTypeFlagScheduleRisk,
			Status:      ActionStatusProposed,
			Reason:      "Maya could not find enough available calendar time for this work in the selected planning window.",
			Payload: ActionPayload{Risk: &RiskPayload{
				Code:    "no_available_slot",
				Message: "No candidate has enough free time in the selected planning window.",
			}},
		}
		return PlanResult{
			Summary: "No safe schedule slot was found for this work.",
			Actions: []CoreAction{action},
		}, nil
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]
		if !left.slot.start.Equal(right.slot.start) {
			return left.slot.start.Before(right.slot.start)
		}
		if left.candidate.Member.EstimateTotal != right.candidate.Member.EstimateTotal {
			return left.candidate.Member.EstimateTotal < right.candidate.Member.EstimateTotal
		}
		if left.candidate.Member.OpenStories != right.candidate.Member.OpenStories {
			return left.candidate.Member.OpenStories < right.candidate.Member.OpenStories
		}
		return left.candidate.Member.FullName < right.candidate.Member.FullName
	})

	selected := candidates[0]
	selectedUserID := selected.candidate.Member.UserID
	actions := make([]CoreAction, 0, 2)
	if normalized.Story.Assignee == nil || *normalized.Story.Assignee != selectedUserID {
		actions = append(actions, CoreAction{
			WorkspaceID: normalized.WorkspaceID,
			StoryID:     normalized.Story.ID,
			Type:        ActionTypeAssignStory,
			Status:      ActionStatusProposed,
			Reason:      assignmentReason(selected),
			Payload: ActionPayload{AssignStory: &AssignStoryPayload{
				AssigneeID: selectedUserID,
			}},
		})
	}

	if !hasStoryScheduleBlock(selected.candidate.Blocks, normalized.Story.ID) {
		actions = append(actions, CoreAction{
			WorkspaceID: normalized.WorkspaceID,
			StoryID:     normalized.Story.ID,
			Type:        ActionTypeScheduleWorkBlock,
			Status:      ActionStatusProposed,
			Reason:      scheduleReason(selected),
			Payload: ActionPayload{ScheduleBlock: &ScheduleBlockPayload{
				UserID:  selectedUserID,
				Title:   normalized.Story.Title,
				StartAt: selected.slot.start,
				EndAt:   selected.slot.end,
			}},
		})
	}

	return PlanResult{
		Summary:        planSummary(normalized.Story.Title, selected),
		SelectedUserID: &selectedUserID,
		Actions:        actions,
	}, nil
}

type candidateChoice struct {
	candidate CandidateSchedule
	slot      timeSlot
}

type timeSlot struct {
	start time.Time
	end   time.Time
}

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
		input.DurationMinutes = defaultDurationMinutes
	}
	if input.DurationMinutes < minimumSlotMinutes {
		input.DurationMinutes = minimumSlotMinutes
	}
	input.WindowStart = input.WindowStart.UTC()
	input.WindowEnd = input.WindowEnd.UTC()
	return input, nil
}

func earliestSlot(candidate CandidateSchedule, startAt, endAt time.Time, duration time.Duration) (timeSlot, bool) {
	occupied := occupiedSlots(candidate)
	cursor := alignToNextHalfHour(startAt.UTC())

	for cursor.Before(endAt) {
		dayStart := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), workdayStartHour, 0, 0, 0, time.UTC)
		dayEnd := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), workdayEndHour, 0, 0, 0, time.UTC)
		if cursor.Before(dayStart) {
			cursor = dayStart
		}
		if !cursor.Before(dayEnd) {
			cursor = time.Date(cursor.Year(), cursor.Month(), cursor.Day()+1, workdayStartHour, 0, 0, 0, time.UTC)
			continue
		}

		slotEnd := cursor.Add(duration)
		if slotEnd.After(dayEnd) || slotEnd.After(endAt) {
			cursor = time.Date(cursor.Year(), cursor.Month(), cursor.Day()+1, workdayStartHour, 0, 0, 0, time.UTC)
			continue
		}
		if !overlapsAny(cursor, slotEnd, occupied) {
			return timeSlot{start: cursor, end: slotEnd}, true
		}
		cursor = cursor.Add(minimumSlotMinutes * time.Minute)
	}

	return timeSlot{}, false
}

func occupiedSlots(candidate CandidateSchedule) []timeSlot {
	slots := make([]timeSlot, 0, len(candidate.BusyWindows)+len(candidate.Blocks))
	for _, window := range candidate.BusyWindows {
		slots = append(slots, timeSlot{start: window.StartAt.UTC(), end: window.EndAt.UTC()})
	}
	for _, block := range candidate.Blocks {
		slots = append(slots, timeSlot{start: block.StartAt.UTC(), end: block.EndAt.UTC()})
	}
	sort.SliceStable(slots, func(i, j int) bool {
		return slots[i].start.Before(slots[j].start)
	})
	return slots
}

func alignToNextHalfHour(value time.Time) time.Time {
	value = value.Truncate(time.Minute)
	minute := value.Minute()
	remainder := minute % minimumSlotMinutes
	if remainder == 0 && value.Second() == 0 && value.Nanosecond() == 0 {
		return value
	}
	return value.Add(time.Duration(minimumSlotMinutes-remainder) * time.Minute).Truncate(time.Minute)
}

func overlapsAny(startAt, endAt time.Time, slots []timeSlot) bool {
	for _, slot := range slots {
		if startAt.Before(slot.end) && endAt.After(slot.start) {
			return true
		}
	}
	return false
}

func hasStoryScheduleBlock(blocks []calendar.CoreScheduleBlock, storyID uuid.UUID) bool {
	for _, block := range blocks {
		if block.StoryID != nil && *block.StoryID == storyID {
			return true
		}
	}
	return false
}

func assignmentReason(choice candidateChoice) string {
	name := strings.TrimSpace(choice.candidate.Member.FullName)
	if name == "" {
		name = strings.TrimSpace(choice.candidate.Member.Username)
	}
	if name == "" {
		name = "this teammate"
	}
	return fmt.Sprintf("Maya selected %s because they have the earliest available slot and currently have %d open items with %d estimate points.", name, choice.candidate.Member.OpenStories, choice.candidate.Member.EstimateTotal)
}

func scheduleReason(choice candidateChoice) string {
	return fmt.Sprintf("Maya found an available calendar slot from %s to %s without overlapping existing busy time or scheduled work.", choice.slot.start.Format(time.RFC3339), choice.slot.end.Format(time.RFC3339))
}

func planSummary(storyTitle string, choice candidateChoice) string {
	name := strings.TrimSpace(choice.candidate.Member.FullName)
	if name == "" {
		name = strings.TrimSpace(choice.candidate.Member.Username)
	}
	if name == "" {
		name = "the selected teammate"
	}
	return fmt.Sprintf("Maya recommends assigning %q to %s and scheduling it from %s to %s.", storyTitle, name, choice.slot.start.Format(time.RFC3339), choice.slot.end.Format(time.RFC3339))
}
