package maya

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *Service) recoveryScheduleStateMatches(
	ctx context.Context,
	scheduleCalendar ScheduleCalendarService,
	ref ScheduleStoryRef,
	owners []uuid.UUID,
	desiredOwner uuid.UUID,
	desiredSegments []ScheduleSegmentInput,
	locked bool,
) (bool, error) {
	if desiredOwner != uuid.Nil {
		if len(owners) != 1 || owners[0] != desiredOwner {
			return false, nil
		}
		blocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, desiredOwner, ref.StoryID)
		if err != nil {
			return false, err
		}
		return mayaScheduleSegmentsMatchBlocks(desiredSegments, blocks, locked), nil
	}
	for _, ownerID := range owners {
		blocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, ownerID, ref.StoryID)
		if err != nil {
			return false, err
		}
		if len(blocks) > 0 {
			return false, nil
		}
	}
	return true, nil
}

func mayaScheduleSegmentsMatchBlocks(segments []ScheduleSegmentInput, blocks []ScheduleBlock, locked bool) bool {
	if len(segments) != len(blocks) {
		return false
	}
	blocksByIndex := make(map[int]ScheduleBlock, len(blocks))
	for _, block := range blocks {
		blocksByIndex[block.SegmentIndex] = block
	}
	for _, segment := range segments {
		block, ok := blocksByIndex[segment.SegmentIndex]
		if !ok || block.Title != segment.Title || !block.StartAt.Equal(segment.StartAt) || !block.EndAt.Equal(segment.EndAt) || block.IsLocked != locked {
			return false
		}
	}
	return true
}

func (s *Service) planAssignedStory(ctx context.Context, story Story, userID uuid.UUID, asOf time.Time) (PlanResult, error) {
	windowStart := asOf
	if story.StartDate != nil && story.StartDate.After(windowStart) {
		windowStart = story.StartDate.UTC()
	}
	windowEnd := windowStart.Add(90 * 24 * time.Hour)
	if story.EndDate != nil {
		deadlineEnd := story.EndDate.UTC().Add(24 * time.Hour)
		if !deadlineEnd.After(windowStart) {
			return PlanResult{
				Summary:        "Maya could not place this work because its deadline has already passed.",
				SelectedUserID: &userID,
				Actions: []CoreAction{
					scheduleOwnershipRetentionAction(story.Workspace, story, userID, "Maya will keep watching this overdue work for a changed deadline."),
					{
						WorkspaceID: story.Workspace,
						StoryID:     story.ID,
						Type:        ActionTypeFlagScheduleRisk,
						Status:      ActionStatusProposed,
						Reason:      "The story deadline passed before enough work time could be reserved.",
						Payload: ActionPayload{Risk: &RiskPayload{
							Code:    "deadline_passed",
							Message: "Move the deadline or adjust the time needed before Maya can place this work.",
						}},
					},
				},
			}, nil
		}
		if deadlineEnd.Before(windowEnd) {
			windowEnd = deadlineEnd
		}
	}
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return PlanResult{}, err
	}
	schedule, err := scheduleCalendar.ListSchedulingAvailability(ctx, story.Workspace, userID, windowStart, windowEnd)
	if err != nil {
		return PlanResult{}, err
	}
	blocks := make([]ScheduleBlock, 0, len(schedule.Blocks))
	var activeBlock *ScheduleBlock
	consumedOrCommittedMinutes := 0
	hasUnfinishedScheduledTime := false
	for _, block := range schedule.Blocks {
		if block.WorkspaceID == story.Workspace && block.StoryID != nil && *block.StoryID == story.ID && block.Source == ScheduleBlockSourceMaya {
			consumedOrCommittedMinutes += consumedOrCommittedScheduleMinutes(block, asOf)
			hasUnfinishedScheduledTime = hasUnfinishedScheduledTime || block.EndAt.After(asOf)
			if !block.IsLocked && block.StartAt.Before(asOf) && block.EndAt.After(asOf) {
				current := block
				activeBlock = &current
			}
			continue
		}
		blocks = append(blocks, block)
	}
	if activeBlock != nil {
		// Keep the in-progress block as an anonymous occupied interval. The
		// planner must not move it, while the reconciliation layer will restore
		// its story metadata and segment index in the desired schedule below.
		current := *activeBlock
		current.StoryID = nil
		current.StoryTitle = nil
		current.StoryCode = nil
		blocks = append(blocks, current)
	}
	workSchedule, err := s.getUserWorkSchedule(ctx, story.Workspace, userID)
	if err != nil {
		return PlanResult{}, err
	}
	if !hasUnfinishedScheduledTime {
		// A fully elapsed unlocked schedule is treated as unfinished work by
		// the existing recovery contract; reserve the full estimate again.
		consumedOrCommittedMinutes = 0
	}
	durationMinutes := effectiveRemainingDurationMinutes(story, consumedOrCommittedMinutes)
	if durationMinutes == 0 {
		result := PlanResult{
			Summary:        "Maya retained the work already in progress without reserving more time.",
			SelectedUserID: &userID,
		}
		if activeBlock != nil {
			result = retainActiveScheduleBlock(result, story, userID, *activeBlock)
		}
		result.Timezone = schedule.Timezone
		return result, nil
	}
	candidate := CandidateSchedule{
		Member: MemberWorkload{UserID: userID}, Timezone: schedule.Timezone,
		WorkingDays: workSchedule.WorkingDays, WorkingStartMinute: workSchedule.StartMinute, WorkingEndMinute: workSchedule.EndMinute,
		BusyWindows: schedule.BusyWindows, Blocks: blocks,
	}
	if feedbackService, ok := s.calendar.(ScheduleFeedbackService); ok {
		preference, preferenceErr := feedbackService.ListManualSchedulePreference(ctx, story.Workspace, userID)
		if preferenceErr != nil {
			return PlanResult{}, fmt.Errorf("list calendar schedule preference for %s: %w", userID, preferenceErr)
		}
		if preference.Confidence > 0 {
			candidate.PreferredStartMinute = preference.PreferredStartMinute
		}
	}
	result, err := s.planner.Plan(PlanInput{
		Context: ctx, AsOf: asOf, WorkspaceID: story.Workspace, Story: story,
		DurationMinutes: durationMinutes,
		WindowStart:     windowStart, WindowEnd: windowEnd,
		MinimumFocusBlockMinutes: valueOrZero(story.MinimumFocusBlockMinutes),
		Candidates:               []CandidateSchedule{candidate},
	})
	if activeBlock != nil {
		result = retainActiveScheduleBlock(result, story, userID, *activeBlock)
	}
	result.Timezone = schedule.Timezone
	return result, err
}

func consumedOrCommittedScheduleMinutes(block ScheduleBlock, now time.Time) int {
	if !block.StartAt.Before(now) {
		return 0
	}
	if block.EndAt.After(now) {
		// Reconciliation retains an in-progress block in full. Count the full
		// reservation before planning replacement segments so its remaining
		// time cannot be allocated a second time.
		return max(0, int(block.EndAt.Sub(block.StartAt)/time.Minute))
	}
	endAt := block.EndAt
	return max(0, int(endAt.Sub(block.StartAt)/time.Minute))
}

func effectiveRemainingDurationMinutes(story Story, elapsedMinutes int) int {
	if story.EstimatedDurationMinutes == nil || *story.EstimatedDurationMinutes <= 0 {
		return 0
	}
	remaining := *story.EstimatedDurationMinutes - elapsedMinutes
	if remaining < 0 {
		return 0
	}
	return remaining
}

func retainActiveScheduleBlock(result PlanResult, story Story, userID uuid.UUID, block ScheduleBlock) PlanResult {
	if story.EstimatedDurationMinutes != nil && *story.EstimatedDurationMinutes > 0 {
		maximumEndAt := block.StartAt.Add(time.Duration(*story.EstimatedDurationMinutes) * time.Minute)
		if block.EndAt.After(maximumEndAt) {
			block.EndAt = maximumEndAt
		}
	}
	activeAction := CoreAction{
		WorkspaceID: story.Workspace,
		StoryID:     story.ID,
		Type:        ActionTypeScheduleWorkBlock,
		Status:      ActionStatusProposed,
		Reason:      "Maya retained the work already in progress while rebalancing future focus blocks.",
		Payload: ActionPayload{ScheduleBlock: &ScheduleBlockPayload{
			UserID: userID, SegmentIndex: 0, Title: story.Title,
			StartAt: block.StartAt, EndAt: block.EndAt,
			PlannedStartAt: block.StartAt, PlannedEndAt: block.EndAt,
			ExpectedStoryUpdatedAt: story.UpdatedAt,
		}},
	}
	shifted := make([]CoreAction, 0, len(result.Actions)+1)
	shifted = append(shifted, activeAction)
	for _, action := range result.Actions {
		if action.Payload.ScheduleBlock != nil && action.Type == ActionTypeScheduleWorkBlock && action.Payload.ScheduleBlock.Operation != ScheduleBlockOperationRetain {
			action.Payload.ScheduleBlock.SegmentIndex++
		}
		shifted = append(shifted, action)
	}
	result.Actions = shifted
	return result
}

func autoSchedulingOutcome(result PlanResult, segments []ScheduleSegmentInput) (string, string) {
	if result.RemainingMinutes > 0 {
		if result.ScheduledMinutes > 0 {
			return AutoSchedulingStatusCannotFit, partialScheduleReason(result.ScheduledMinutes, result.RemainingMinutes)
		}
		return AutoSchedulingStatusCannotFit, fmt.Sprintf("%s left to schedule.", formatMinutes(result.RemainingMinutes))
	}
	if len(segments) > 0 {
		return AutoSchedulingStatusScheduled, "Maya scheduled this story around the assignee's availability."
	}
	for _, action := range result.Actions {
		if action.Payload.Risk == nil {
			continue
		}
		switch action.Payload.Risk.Code {
		case "missing_duration":
			return AutoSchedulingStatusNeedsTime, "Add time needed so Maya can reserve focused work."
		case "deadline_passed":
			return AutoSchedulingStatusAtRisk, "The deadline has passed before Maya could reserve enough focused time."
		case "no_available_slot":
			return AutoSchedulingStatusCannotFit, "Maya could not fit the required focus time into the current planning window."
		}
	}
	return AutoSchedulingStatusCannotFit, "Maya could not place this work safely in the current planning window."
}

func refineScheduleOutcomeReason(previousBlocks []ScheduleBlock, segments []ScheduleSegmentInput, status, fallback string) string {
	if status != AutoSchedulingStatusScheduled || len(previousBlocks) == 0 {
		return fallback
	}
	if scheduleSegmentsChanged(previousBlocks, segments) {
		return "The assignee's availability or this story's scheduling constraints changed, so Maya moved it to the next safe slot."
	}
	return fallback
}

func scheduleSegmentsChanged(previousBlocks []ScheduleBlock, segments []ScheduleSegmentInput) bool {
	if len(previousBlocks) != len(segments) {
		return true
	}
	previousBySegment := make(map[int]ScheduleBlock, len(previousBlocks))
	for _, block := range previousBlocks {
		previousBySegment[block.SegmentIndex] = block
	}
	for _, segment := range segments {
		previous, exists := previousBySegment[segment.SegmentIndex]
		if !exists || !previous.StartAt.Equal(segment.StartAt) || !previous.EndAt.Equal(segment.EndAt) {
			return true
		}
	}
	return false
}
