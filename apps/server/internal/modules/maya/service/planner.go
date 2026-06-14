package maya

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

var ErrInvalidPlanInput = errors.New("invalid maya plan input")

const (
	defaultDurationMinutes = 60
	minimumSlotMinutes     = 30
	workdayStartHour       = 9
	workdayEndHour         = 17
)

type Planner struct {
	advisor CandidateAdvisor
}

func NewPlanner() Planner {
	return Planner{}
}

func NewPlannerWithAdvisor(advisor CandidateAdvisor) Planner {
	return Planner{advisor: advisor}
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
		selected, advisorReason, ok := p.selectAssignmentCandidate(normalized, normalized.Candidates)
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
		if ok {
			selectedUserID := selected.Member.UserID
			actions := make([]CoreAction, 0, 2)
			if normalized.Story.Assignee == nil || *normalized.Story.Assignee != selectedUserID {
				reason := assignmentReasonForMember(selected.Member)
				if strings.TrimSpace(advisorReason) != "" {
					reason = advisorReason
				}
				actions = append(actions, CoreAction{
					WorkspaceID: normalized.WorkspaceID,
					StoryID:     normalized.Story.ID,
					Type:        ActionTypeAssignStory,
					Status:      ActionStatusProposed,
					Reason:      reason,
					Payload: ActionPayload{AssignStory: &AssignStoryPayload{
						AssigneeID: selectedUserID,
					}},
				})
			}
			actions = append(actions, action)
			return PlanResult{
				Summary:        "Maya selected an owner, but no safe schedule slot was found for this work.",
				SelectedUserID: &selectedUserID,
				Actions:        actions,
			}, nil
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

	selected, advisorReason := p.selectCandidate(normalized, candidates)
	selectedUserID := selected.candidate.Member.UserID
	actions := make([]CoreAction, 0, 2)
	if normalized.Story.Assignee == nil || *normalized.Story.Assignee != selectedUserID {
		reason := assignmentReason(selected)
		if strings.TrimSpace(advisorReason) != "" {
			reason = advisorReason
		}
		actions = append(actions, CoreAction{
			WorkspaceID: normalized.WorkspaceID,
			StoryID:     normalized.Story.ID,
			Type:        ActionTypeAssignStory,
			Status:      ActionStatusProposed,
			Reason:      reason,
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

func (p Planner) selectCandidate(input PlanInput, candidates []candidateChoice) (candidateChoice, string) {
	selected := candidates[0]
	if p.advisor == nil || len(candidates) == 1 {
		return selected, ""
	}

	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	recommendations := make([]CandidateRecommendation, 0, len(candidates))
	for _, candidate := range candidates {
		recommendations = append(recommendations, CandidateRecommendation{
			UserID:                candidate.candidate.Member.UserID,
			FullName:              candidate.candidate.Member.FullName,
			Username:              candidate.candidate.Member.Username,
			TeamAIRoleTitle:       candidate.candidate.Member.TeamAIRoleTitle,
			TeamAIRoleDescription: candidate.candidate.Member.TeamAIRoleDescription,
			OpenStories:           candidate.candidate.Member.OpenStories,
			EstimateTotal:         candidate.candidate.Member.EstimateTotal,
			HasAvailableSlot:      true,
			SlotStart:             candidate.slot.start,
			SlotEnd:               candidate.slot.end,
		})
	}
	result, err := p.advisor.RecommendCandidate(ctx, CandidateRecommendationInput{
		WorkspaceID:     input.WorkspaceID,
		Story:           input.Story,
		DurationMinutes: input.DurationMinutes,
		WindowStart:     input.WindowStart,
		WindowEnd:       input.WindowEnd,
		Candidates:      recommendations,
	})
	if err != nil || result.UserID == uuid.Nil {
		return selected, ""
	}
	for _, candidate := range candidates {
		if candidate.candidate.Member.UserID == result.UserID {
			return candidate, result.Reason
		}
	}
	return selected, ""
}

func (p Planner) selectAssignmentCandidate(input PlanInput, candidates []CandidateSchedule) (CandidateSchedule, string, bool) {
	assignable := make([]CandidateSchedule, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Member.UserID != uuid.Nil {
			assignable = append(assignable, candidate)
		}
	}
	if len(assignable) == 0 {
		return CandidateSchedule{}, "", false
	}
	sort.SliceStable(assignable, func(i, j int) bool {
		left := assignable[i].Member
		right := assignable[j].Member
		if left.EstimateTotal != right.EstimateTotal {
			return left.EstimateTotal < right.EstimateTotal
		}
		if left.OpenStories != right.OpenStories {
			return left.OpenStories < right.OpenStories
		}
		return left.FullName < right.FullName
	})
	selected := assignable[0]
	if p.advisor == nil || len(assignable) == 1 {
		return selected, "", true
	}

	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	recommendations := make([]CandidateRecommendation, 0, len(assignable))
	for _, candidate := range assignable {
		recommendations = append(recommendations, CandidateRecommendation{
			UserID:                candidate.Member.UserID,
			FullName:              candidate.Member.FullName,
			Username:              candidate.Member.Username,
			TeamAIRoleTitle:       candidate.Member.TeamAIRoleTitle,
			TeamAIRoleDescription: candidate.Member.TeamAIRoleDescription,
			OpenStories:           candidate.Member.OpenStories,
			EstimateTotal:         candidate.Member.EstimateTotal,
			HasAvailableSlot:      false,
		})
	}
	result, err := p.advisor.RecommendCandidate(ctx, CandidateRecommendationInput{
		WorkspaceID:     input.WorkspaceID,
		Story:           input.Story,
		DurationMinutes: input.DurationMinutes,
		WindowStart:     input.WindowStart,
		WindowEnd:       input.WindowEnd,
		Candidates:      recommendations,
	})
	if err != nil || result.UserID == uuid.Nil {
		return selected, "", true
	}
	for _, candidate := range assignable {
		if candidate.Member.UserID == result.UserID {
			return candidate, result.Reason, true
		}
	}
	return selected, "", true
}

func (p Planner) RecommendAssignments(ctx context.Context, input BatchAssignmentRecommendationInput) (BatchAssignmentRecommendationResult, error) {
	if p.advisor != nil {
		if batchAdvisor, ok := p.advisor.(BatchAssignmentAdvisor); ok {
			result, err := batchAdvisor.RecommendAssignments(ctx, input)
			if err == nil && len(result.Assignments) > 0 {
				return result, nil
			}
		}
	}
	return deterministicBatchAssignments(input), nil
}

func deterministicBatchAssignments(input BatchAssignmentRecommendationInput) BatchAssignmentRecommendationResult {
	if len(input.Candidates) == 0 || len(input.Stories) == 0 {
		return BatchAssignmentRecommendationResult{}
	}
	candidates := append([]CandidateRecommendation(nil), input.Candidates...)
	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]
		if left.EstimateTotal != right.EstimateTotal {
			return left.EstimateTotal < right.EstimateTotal
		}
		if left.OpenStories != right.OpenStories {
			return left.OpenStories < right.OpenStories
		}
		return left.FullName < right.FullName
	})
	assignments := make([]BatchAssignmentRecommendation, 0, len(input.Stories))
	for index, story := range input.Stories {
		candidate := candidates[index%len(candidates)]
		assignments = append(assignments, BatchAssignmentRecommendation{
			StoryID:    story.ID,
			AssigneeID: candidate.UserID,
			Reason:     assignmentReasonForCandidate(candidate),
		})
	}
	return BatchAssignmentRecommendationResult{Assignments: assignments}
}

func assignmentReasonForCandidate(candidate CandidateRecommendation) string {
	if strings.TrimSpace(candidate.TeamAIRoleTitle) != "" {
		return fmt.Sprintf("Maya selected %s because their work focus is %s and their current workload is lighter than the alternatives.", displayCandidateName(candidate), candidate.TeamAIRoleTitle)
	}
	return fmt.Sprintf("Maya selected %s based on current workload and availability.", displayCandidateName(candidate))
}

func displayCandidateName(candidate CandidateRecommendation) string {
	if strings.TrimSpace(candidate.FullName) != "" {
		return candidate.FullName
	}
	if strings.TrimSpace(candidate.Username) != "" {
		return candidate.Username
	}
	return candidate.UserID.String()
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
	return assignmentReasonForMember(choice.candidate.Member)
}

func assignmentReasonForMember(member reports.CoreMemberWorkload) string {
	name := strings.TrimSpace(member.FullName)
	if name == "" {
		name = strings.TrimSpace(member.Username)
	}
	if name == "" {
		name = "this teammate"
	}
	return fmt.Sprintf("Maya selected %s because they have the strongest available fit and currently have %d open items with %d estimate points.", name, member.OpenStories, member.EstimateTotal)
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
