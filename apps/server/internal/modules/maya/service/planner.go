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
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/workweek"
	"github.com/google/uuid"
)

var (
	ErrInvalidPlanInput = errors.New("invalid maya plan input")
	ErrMissingDuration  = errors.New("story duration is required for scheduling")
)

const (
	fallbackMinimumFocusBlockMinutes = 60
	planningStartGranularityMinutes  = 5
	maxFocusBlockMinutes             = 120
	workdayStartHour                 = 9
	workdayEndHour                   = 17
	recentActivityDays               = 30
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
		if errors.Is(err, ErrMissingDuration) {
			actions := make([]CoreAction, 0, 3)
			var selectedUserID *uuid.UUID
			ownerID := uuid.Nil
			advisorReason := ""
			if input.Story.Assignee != nil && *input.Story.Assignee != uuid.Nil && hasCandidate(input.Candidates, *input.Story.Assignee) {
				ownerID = *input.Story.Assignee
			} else if selected, reason, ok := p.selectAssignmentCandidate(input, input.Candidates); ok {
				ownerID = selected.Member.UserID
				advisorReason = reason
			}
			if ownerID != uuid.Nil {
				selectedUserID = &ownerID
				if input.Story.Assignee == nil || *input.Story.Assignee != ownerID {
					reason := assignmentReasonForMember(candidateMember(input.Candidates, ownerID))
					if strings.TrimSpace(input.AssignmentReason) != "" {
						reason = input.AssignmentReason
					}
					if strings.TrimSpace(advisorReason) != "" {
						reason = advisorReason
					}
					actions = append(actions, CoreAction{
						WorkspaceID: input.WorkspaceID,
						StoryID:     input.Story.ID,
						Type:        ActionTypeAssignStory,
						Status:      ActionStatusProposed,
						Reason:      reason,
						Payload: ActionPayload{AssignStory: &AssignStoryPayload{
							AssigneeID:        ownerID,
							ExpectedUpdatedAt: input.Story.UpdatedAt,
						}},
					})
				}
				actions = append(actions, scheduleOwnershipRetentionAction(
					input.WorkspaceID,
					input.Story,
					ownerID,
					"Maya will keep watching this assigned work and schedule it after a time-needed estimate is added.",
				))
			}
			actions = append(actions, CoreAction{
				WorkspaceID: input.WorkspaceID,
				StoryID:     input.Story.ID,
				Type:        ActionTypeFlagScheduleRisk,
				Status:      ActionStatusProposed,
				Reason:      "No estimated duration is set, so Maya did not reserve calendar time from a complexity estimate.",
				Payload: ActionPayload{Risk: &RiskPayload{
					Code:    "missing_duration",
					Message: "Set the time needed for this work before asking Maya to schedule it.",
				}},
			})
			return PlanResult{
				Summary:        "Maya needs an estimated duration before this work can be scheduled.",
				SelectedUserID: selectedUserID,
				Actions:        actions,
			}, nil
		}
		return PlanResult{}, err
	}

	candidates := make([]candidateChoice, 0, len(normalized.Candidates))
	for _, candidate := range normalized.Candidates {
		if candidate.Member.UserID == uuid.Nil {
			continue
		}
		candidate.PreemptibleBlockIDs = preemptibleBlockIDs(normalized.Story, candidate.Blocks, time.Now().UTC())
		candidateWindowStart, candidateWindowEnd := clampWindowToSprint(normalized, candidate.Timezone)
		segments, ok := planWorkSegments(candidate, candidateWindowStart, candidateWindowEnd, normalized.DurationMinutes, normalized.MinimumFocusBlockMinutes, normalized.WorkingDays)
		if !ok {
			continue
		}
		candidates = append(candidates, candidateChoice{
			candidate:         candidate,
			slot:              segments[0],
			plan:              timeSlot{start: segments[0].start, end: segments[len(segments)-1].end},
			segments:          segments,
			preemptedBlockIDs: preemptedBlockIDsForSegments(candidate.Blocks, candidate.PreemptibleBlockIDs, segments),
		})
	}
	candidates = preferRecentlyActiveChoices(candidates)

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
			actions := make([]CoreAction, 0, 3)
			if normalized.Story.Assignee == nil || *normalized.Story.Assignee != selectedUserID {
				reason := assignmentReasonForMember(selected.Member)
				if strings.TrimSpace(normalized.AssignmentReason) != "" {
					reason = normalized.AssignmentReason
				}
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
						AssigneeID:        selectedUserID,
						ExpectedUpdatedAt: normalized.Story.UpdatedAt,
					}},
				})
			}
			actions = append(actions, scheduleOwnershipRetentionAction(
				normalized.WorkspaceID,
				normalized.Story,
				selectedUserID,
				"Maya will keep watching this work and retry placement when calendar availability changes.",
			))
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
	actions := make([]CoreAction, 0, 1+len(selected.segments))
	if normalized.Story.Assignee == nil || *normalized.Story.Assignee != selectedUserID {
		reason := assignmentReason(selected)
		if strings.TrimSpace(normalized.AssignmentReason) != "" {
			reason = normalized.AssignmentReason
		}
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
				AssigneeID:        selectedUserID,
				ExpectedUpdatedAt: normalized.Story.UpdatedAt,
			}},
		})
	}

	if !hasStoryScheduleBlock(selected.candidate.Blocks, normalized.Story.ID) {
		for segmentIndex, segment := range selected.segments {
			actions = append(actions, CoreAction{
				WorkspaceID: normalized.WorkspaceID,
				StoryID:     normalized.Story.ID,
				Type:        ActionTypeScheduleWorkBlock,
				Status:      ActionStatusProposed,
				Reason:      scheduleReason(selected, segmentIndex, len(selected.segments)),
				Payload: ActionPayload{ScheduleBlock: &ScheduleBlockPayload{
					UserID:                 selectedUserID,
					SegmentIndex:           segmentIndex,
					Title:                  normalized.Story.Title,
					StartAt:                segment.start,
					EndAt:                  segment.end,
					PlannedStartAt:         selected.plan.start,
					PlannedEndAt:           selected.plan.end,
					ExpectedStoryUpdatedAt: normalized.Story.UpdatedAt,
					PreemptBlockIDs:        append([]uuid.UUID(nil), selected.preemptedBlockIDs...),
				}},
			})
		}
	}

	return PlanResult{
		Summary:           planSummary(normalized.Story.Title, selected),
		SelectedUserID:    &selectedUserID,
		Actions:           actions,
		PreemptedBlockIDs: append([]uuid.UUID(nil), selected.preemptedBlockIDs...),
	}, nil
}

func hasCandidate(candidates []CandidateSchedule, userID uuid.UUID) bool {
	for _, candidate := range candidates {
		if candidate.Member.UserID == userID {
			return true
		}
	}
	return false
}

func candidateMember(candidates []CandidateSchedule, userID uuid.UUID) reports.CoreMemberWorkload {
	for _, candidate := range candidates {
		if candidate.Member.UserID == userID {
			return candidate.Member
		}
	}
	return reports.CoreMemberWorkload{UserID: userID}
}

func scheduleOwnershipRetentionAction(workspaceID uuid.UUID, story stories.CoreSingleStory, userID uuid.UUID, reason string) CoreAction {
	return CoreAction{
		WorkspaceID: workspaceID,
		StoryID:     story.ID,
		Type:        ActionTypeScheduleWorkBlock,
		Status:      ActionStatusProposed,
		Reason:      reason,
		Payload: ActionPayload{ScheduleBlock: &ScheduleBlockPayload{
			UserID:                 userID,
			Operation:              ScheduleBlockOperationRetain,
			Title:                  story.Title,
			ExpectedStoryUpdatedAt: story.UpdatedAt,
		}},
	}
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
			LastStoryActivityAt:   candidate.candidate.Member.LastStoryActivityAt,
			DaysSinceLastActivity: daysSinceLastActivity(candidate.candidate.Member.LastStoryActivityAt),
			RecentlyActive:        isRecentlyActive(candidate.candidate.Member.LastStoryActivityAt),
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
	assignable = preferRecentlyActiveSchedules(assignable)
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
			LastStoryActivityAt:   candidate.Member.LastStoryActivityAt,
			DaysSinceLastActivity: daysSinceLastActivity(candidate.Member.LastStoryActivityAt),
			RecentlyActive:        isRecentlyActive(candidate.Member.LastStoryActivityAt),
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
	input.Candidates = preferRecentlyActiveRecommendations(input.Candidates)
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

func preferRecentlyActiveChoices(candidates []candidateChoice) []candidateChoice {
	active := make([]candidateChoice, 0, len(candidates))
	for _, candidate := range candidates {
		if isRecentlyActive(candidate.candidate.Member.LastStoryActivityAt) {
			active = append(active, candidate)
		}
	}
	if len(active) == 0 {
		return candidates
	}
	return active
}

func preferRecentlyActiveSchedules(candidates []CandidateSchedule) []CandidateSchedule {
	active := make([]CandidateSchedule, 0, len(candidates))
	for _, candidate := range candidates {
		if isRecentlyActive(candidate.Member.LastStoryActivityAt) {
			active = append(active, candidate)
		}
	}
	if len(active) == 0 {
		return candidates
	}
	return active
}

func preferRecentlyActiveRecommendations(candidates []CandidateRecommendation) []CandidateRecommendation {
	active := make([]CandidateRecommendation, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.RecentlyActive {
			active = append(active, candidate)
		}
	}
	if len(active) == 0 {
		return candidates
	}
	return active
}

func isRecentlyActive(lastActivityAt *time.Time) bool {
	if lastActivityAt == nil {
		return false
	}
	return time.Since(lastActivityAt.UTC()) <= recentActivityDays*24*time.Hour
}

func daysSinceLastActivity(lastActivityAt *time.Time) *int {
	if lastActivityAt == nil {
		return nil
	}
	days := int(time.Since(lastActivityAt.UTC()).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return &days
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
	candidate         CandidateSchedule
	slot              timeSlot
	plan              timeSlot
	segments          []timeSlot
	preemptedBlockIDs []uuid.UUID
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
		input.DurationMinutes = estimatedWorkDurationMinutes(input.Story)
	}
	if input.DurationMinutes <= 0 {
		return PlanInput{}, ErrMissingDuration
	}
	if input.DurationMinutes > stories.MaximumEstimatedDurationMinutes {
		return PlanInput{}, fmt.Errorf("%w: duration exceeds %d minutes", ErrInvalidPlanInput, stories.MaximumEstimatedDurationMinutes)
	}
	if input.MinimumFocusBlockMinutes <= 0 && input.Story.MinimumFocusBlockMinutes != nil {
		input.MinimumFocusBlockMinutes = *input.Story.MinimumFocusBlockMinutes
	}
	if input.MinimumFocusBlockMinutes > input.DurationMinutes {
		input.MinimumFocusBlockMinutes = input.DurationMinutes
	}
	input.WindowStart = input.WindowStart.UTC()
	input.WindowEnd = input.WindowEnd.UTC()
	input.WorkingDays = workweek.Normalize(input.WorkingDays)
	return input, nil
}

func estimatedWorkDurationMinutes(story stories.CoreSingleStory) int {
	if story.EstimatedDurationMinutes != nil && *story.EstimatedDurationMinutes > 0 {
		return *story.EstimatedDurationMinutes
	}
	return 0
}

func clampWindowToSprint(input PlanInput, timezone string) (time.Time, time.Time) {
	if input.Story.SprintSummary == nil {
		return input.WindowStart, input.WindowEnd
	}
	location := calendarLocation(timezone)
	sprintStart := sprintWorkdayStart(input.Story.SprintSummary.StartDate, location)
	sprintEnd := sprintWorkdayEnd(input.Story.SprintSummary.EndDate, location)
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

func sprintWorkdayStart(value time.Time, location *time.Location) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), workdayStartHour, 0, 0, 0, location)
}

func sprintWorkdayEnd(value time.Time, location *time.Location) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), workdayEndHour, 0, 0, 0, location)
}

func planWorkWindow(candidate CandidateSchedule, startAt, endAt time.Time, duration time.Duration, workingDays []int) (timeSlot, bool) {
	if duration <= 0 {
		return timeSlot{}, false
	}
	durationMinutes := int(duration / time.Minute)
	segments, ok := planWorkSegments(candidate, startAt, endAt, durationMinutes, 0, workingDays)
	if !ok {
		return timeSlot{}, false
	}
	return timeSlot{start: segments[0].start, end: segments[len(segments)-1].end}, true
}

func planWorkSegments(candidate CandidateSchedule, startAt, endAt time.Time, durationMinutes, minimumFocusMinutes int, workingDays []int) ([]timeSlot, bool) {
	if minimumFocusMinutes <= 0 {
		// No chunks prefers one contiguous block. If conflicts make that
		// impossible, fall back to larger available windows so replanning can
		// split the remaining work around higher-priority commitments.
		if segments, ok := planWorkSegmentsWithLimits(candidate, startAt, endAt, durationMinutes, durationMinutes, durationMinutes, workingDays); ok {
			return segments, true
		}
		minimumFocusMinutes = fallbackMinimumFocusBlockMinutes
		if minimumFocusMinutes > durationMinutes {
			minimumFocusMinutes = durationMinutes
		}
		return planWorkSegmentsWithLimits(candidate, startAt, endAt, durationMinutes, minimumFocusMinutes, durationMinutes, workingDays)
	}

	return planWorkSegmentsWithLimits(candidate, startAt, endAt, durationMinutes, minimumFocusMinutes, maxFocusBlockMinutes, workingDays)
}

func planWorkSegmentsWithLimits(candidate CandidateSchedule, startAt, endAt time.Time, durationMinutes, minimumFocusMinutes, maximumFocusMinutes int, workingDays []int) ([]timeSlot, bool) {
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
		if !workweek.IsWorkingDay(cursor, workingDays) {
			cursor = nextPlanningWorkdayStart(cursor, candidate, workingDays)
			continue
		}

		dayStart := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), workdayStartHour, 0, 0, 0, location)
		dayEnd := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), workdayEndHour, 0, 0, 0, location)
		if cursor.Before(dayStart) {
			cursor = dayStart
		}
		if !cursor.Before(dayEnd) {
			cursor = nextPlanningWorkdayStart(cursor, candidate, workingDays)
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
			return segments, true
		}
		cursor = cursor.Add(take)
	}

	return nil, false
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

func nextWorkdayStart(value time.Time, workingDays []int) time.Time {
	next := time.Date(value.Year(), value.Month(), value.Day()+1, workdayStartHour, 0, 0, 0, value.Location())
	for !workweek.IsWorkingDay(next, workingDays) {
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
		Add(time.Duration(clampPreferredStartMinute(*candidate.PreferredStartMinute)) * time.Minute)
	if preferred.After(cursor) {
		return preferred
	}
	return cursor
}

func nextPlanningWorkdayStart(value time.Time, candidate CandidateSchedule, workingDays []int) time.Time {
	next := nextWorkdayStart(value, workingDays)
	if candidate.PreferredStartMinute == nil {
		return next
	}
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()).
		Add(time.Duration(clampPreferredStartMinute(*candidate.PreferredStartMinute)) * time.Minute)
}

func clampPreferredStartMinute(value int) int {
	minimum := workdayStartHour * 60
	maximum := workdayEndHour*60 - planningStartGranularityMinutes
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
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

func preemptibleBlockIDs(story stories.CoreSingleStory, blocks []calendar.CoreScheduleBlock, now time.Time) []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	for _, block := range blocks {
		if !shouldPreemptBlock(story, block, now) {
			continue
		}
		ids = append(ids, block.ID)
	}
	return ids
}

func preemptedBlockIDsForSegments(blocks []calendar.CoreScheduleBlock, preemptibleIDs []uuid.UUID, segments []timeSlot) []uuid.UUID {
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

func shouldPreemptBlock(story stories.CoreSingleStory, block calendar.CoreScheduleBlock, now time.Time) bool {
	if block.ID == uuid.Nil || block.Source != calendar.ScheduleBlockSourceMaya || block.StoryID == nil || *block.StoryID == story.ID {
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
	return fmt.Sprintf("Maya recommends assigning %q to %s and scheduling %d focus segment(s) from %s to %s.", storyTitle, name, len(choice.segments), choice.plan.start.Format(time.RFC3339), choice.plan.end.Format(time.RFC3339))
}
