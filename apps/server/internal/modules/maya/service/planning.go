package maya

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
)

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
		candidate.PreemptibleBlockIDs = preemptibleBlockIDs(normalized.Story, candidate.Blocks, normalized.AsOf)
		candidateWindowStart, candidateWindowEnd := clampWindowToSprint(normalized, candidate)
		segmentPlan := planWorkSegments(candidate, candidateWindowStart, candidateWindowEnd, normalized.DurationMinutes, normalized.MinimumFocusBlockMinutes)
		if len(segmentPlan.segments) == 0 {
			continue
		}
		candidates = append(candidates, candidateChoice{
			candidate:         candidate,
			slot:              segmentPlan.segments[0],
			plan:              timeSlot{start: segmentPlan.segments[0].start, end: segmentPlan.segments[len(segmentPlan.segments)-1].end},
			segments:          segmentPlan.segments,
			remainingMinutes:  segmentPlan.remainingMinutes,
			preemptedBlockIDs: preemptedBlockIDsForSegments(candidate.Blocks, candidate.PreemptibleBlockIDs, segmentPlan.segments),
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
				Code:             "no_available_slot",
				Message:          "No candidate has free time in the selected planning window.",
				RequiredMinutes:  normalized.DurationMinutes,
				RemainingMinutes: normalized.DurationMinutes,
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
				Summary:          "Maya selected an owner, but no safe schedule slot was found for this work.",
				SelectedUserID:   &selectedUserID,
				Actions:          actions,
				DurationMinutes:  normalized.DurationMinutes,
				RemainingMinutes: normalized.DurationMinutes,
			}, nil
		}
		return PlanResult{
			Summary:          "No safe schedule slot was found for this work.",
			Actions:          []CoreAction{action},
			DurationMinutes:  normalized.DurationMinutes,
			RemainingMinutes: normalized.DurationMinutes,
		}, nil
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]
		if left.remainingMinutes != right.remainingMinutes {
			return left.remainingMinutes < right.remainingMinutes
		}
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
	scheduledMinutes := normalized.DurationMinutes - selected.remainingMinutes
	if selected.remainingMinutes > 0 {
		actions = append(actions, CoreAction{
			WorkspaceID: normalized.WorkspaceID,
			StoryID:     normalized.Story.ID,
			Type:        ActionTypeFlagScheduleRisk,
			Status:      ActionStatusProposed,
			Reason:      partialScheduleReason(scheduledMinutes, selected.remainingMinutes),
			Payload: ActionPayload{Risk: &RiskPayload{
				Code:             "no_available_slot",
				Message:          "Some focus time still needs a slot.",
				RequiredMinutes:  normalized.DurationMinutes,
				ScheduledMinutes: scheduledMinutes,
				RemainingMinutes: selected.remainingMinutes,
			}},
		})
	}

	return PlanResult{
		Summary:           planSummary(normalized.Story.Title, selected),
		SelectedUserID:    &selectedUserID,
		Actions:           actions,
		PreemptedBlockIDs: append([]uuid.UUID(nil), selected.preemptedBlockIDs...),
		DurationMinutes:   normalized.DurationMinutes,
		ScheduledMinutes:  scheduledMinutes,
		RemainingMinutes:  selected.remainingMinutes,
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

func candidateMember(candidates []CandidateSchedule, userID uuid.UUID) MemberWorkload {
	for _, candidate := range candidates {
		if candidate.Member.UserID == userID {
			return candidate.Member
		}
	}
	return MemberWorkload{UserID: userID}
}

func scheduleOwnershipRetentionAction(workspaceID uuid.UUID, story Story, userID uuid.UUID, reason string) CoreAction {
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
	bestRemainingMinutes := selected.remainingMinutes
	eligibleCandidates := make([]candidateChoice, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.remainingMinutes == bestRemainingMinutes {
			eligibleCandidates = append(eligibleCandidates, candidate)
		}
	}
	if p.advisor == nil || len(eligibleCandidates) == 1 {
		return selected, ""
	}

	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	recommendations := make([]CandidateRecommendation, 0, len(eligibleCandidates))
	for _, candidate := range eligibleCandidates {
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
	for _, candidate := range eligibleCandidates {
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
