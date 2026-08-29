package maya

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/workschedule"
	"github.com/google/uuid"
)

func (s *Service) ProcessAssignmentBatch(ctx context.Context, input ProcessAssignmentBatchInput) (ProcessAssignmentBatchResult, error) {
	ctx, span := mayaServiceTracer.Start(ctx, "business.core.maya.ProcessAssignmentBatch")
	defer span.End()

	if err := s.validate(); err != nil {
		span.RecordError(err)
		return ProcessAssignmentBatchResult{}, err
	}
	if input.WorkspaceID == uuid.Nil || input.TeamID == uuid.Nil || input.TriggeredBy == uuid.Nil {
		return ProcessAssignmentBatchResult{}, fmt.Errorf("%w: assignment batch identity is required", ErrInvalidPlanInput)
	}
	if !input.WindowEnd.After(input.WindowStart) {
		return ProcessAssignmentBatchResult{}, fmt.Errorf("%w: planning window end must be after start", ErrInvalidPlanInput)
	}
	storiesForBatch := make([]Story, 0, len(input.StoryIDs))
	for _, storyID := range input.StoryIDs {
		if storyID == uuid.Nil {
			continue
		}
		story, err := s.stories.Get(ctx, storyID, input.WorkspaceID)
		if err != nil {
			continue
		}
		if story.Team != input.TeamID || !story.AutoSchedulingEnabled || story.Assignee == nil || *story.Assignee != s.mayaActorID {
			continue
		}
		storiesForBatch = append(storiesForBatch, story)
	}
	if len(storiesForBatch) == 0 {
		return ProcessAssignmentBatchResult{Skipped: len(input.StoryIDs)}, nil
	}

	candidates, _, err := s.buildCandidates(ctx, CreateWorkPlanInput{
		WorkspaceID:     input.WorkspaceID,
		WindowStart:     input.WindowStart,
		WindowEnd:       input.WindowEnd,
		DurationMinutes: input.DurationMinutes,
	}, storiesForBatch[0])
	if err != nil {
		span.RecordError(err)
		return ProcessAssignmentBatchResult{}, err
	}

	batchStories := make([]BatchAssignmentStory, 0, len(storiesForBatch))
	storyByID := make(map[uuid.UUID]Story, len(storiesForBatch))
	for _, story := range storiesForBatch {
		description := ""
		if story.Description != nil {
			description = *story.Description
		}
		batchStories = append(batchStories, BatchAssignmentStory{
			ID:              story.ID,
			Title:           story.Title,
			Description:     description,
			Priority:        story.Priority,
			EstimateValue:   story.EstimateValue,
			EstimateLabel:   story.EstimateLabel,
			DurationMinutes: effectiveWorkDurationMinutes(story, input.DurationMinutes),
		})
		storyByID[story.ID] = story
	}

	candidateRecommendations := candidateRecommendationsFromSchedules(candidates, input.WindowStart, input.WindowEnd, batchCandidateDurationMinutes(storiesForBatch, input.DurationMinutes))
	recommendations, err := s.planner.RecommendAssignments(ctx, BatchAssignmentRecommendationInput{
		WorkspaceID: input.WorkspaceID,
		Stories:     batchStories,
		Candidates:  candidateRecommendations,
	})
	if err != nil {
		span.RecordError(err)
		return ProcessAssignmentBatchResult{}, fmt.Errorf("recommend Maya assignment batch: %w", err)
	}

	candidateIDs := make(map[uuid.UUID]struct{}, len(candidateRecommendations))
	for _, candidate := range candidateRecommendations {
		candidateIDs[candidate.UserID] = struct{}{}
	}

	result := ProcessAssignmentBatchResult{Plans: make([]WorkPlan, 0, len(recommendations.Assignments))}
	seenStoryIDs := make(map[uuid.UUID]struct{}, len(recommendations.Assignments))
	for _, recommendation := range recommendations.Assignments {
		story, ok := storyByID[recommendation.StoryID]
		if !ok {
			result.Skipped++
			continue
		}
		if _, ok := candidateIDs[recommendation.AssigneeID]; !ok {
			result.Skipped++
			continue
		}
		if _, seen := seenStoryIDs[story.ID]; seen {
			result.Skipped++
			continue
		}
		seenStoryIDs[story.ID] = struct{}{}

		plan, err := s.CreateWorkPlan(ctx, CreateWorkPlanInput{
			WorkspaceID:      input.WorkspaceID,
			StoryID:          story.ID,
			TriggeredBy:      input.TriggeredBy,
			Trigger:          RunTriggerEvent,
			WindowStart:      input.WindowStart,
			WindowEnd:        input.WindowEnd,
			DurationMinutes:  effectiveWorkDurationMinutes(story, input.DurationMinutes),
			CandidateUserIDs: []uuid.UUID{recommendation.AssigneeID},
			AutoApply:        input.AutoApply,
			AssignmentReason: recommendation.Reason,
		})
		if err != nil {
			result.Skipped++
			continue
		}
		result.Processed++
		result.Plans = append(result.Plans, plan)
	}
	result.Skipped += len(storiesForBatch) - len(seenStoryIDs)
	return result, nil
}

func effectiveWorkDurationMinutes(story Story, requestedDurationMinutes int) int {
	if requestedDurationMinutes > 0 {
		return requestedDurationMinutes
	}
	return estimatedWorkDurationMinutes(story)
}

func batchCandidateDurationMinutes(storiesForBatch []Story, requestedDurationMinutes int) int {
	if requestedDurationMinutes > 0 {
		return requestedDurationMinutes
	}
	duration := 0
	for _, story := range storiesForBatch {
		if candidate := estimatedWorkDurationMinutes(story); candidate > duration {
			duration = candidate
		}
	}
	return duration
}

func (s *Service) validate() error {
	if s == nil || s.repo == nil || s.stories == nil || s.reports == nil || s.calendar == nil || s.users == nil || s.mayaActorID == uuid.Nil {
		return ErrNotConfigured
	}
	return nil
}

func candidateRecommendationsFromSchedules(candidates []CandidateSchedule, windowStart, windowEnd time.Time, durationMinutes int) []CandidateRecommendation {
	recommendations := make([]CandidateRecommendation, 0, len(candidates))
	duration := time.Duration(durationMinutes) * time.Minute
	for _, candidate := range candidates {
		recommendation := CandidateRecommendation{
			UserID:                candidate.Member.UserID,
			FullName:              candidate.Member.FullName,
			Username:              candidate.Member.Username,
			TeamAIRoleTitle:       candidate.Member.TeamAIRoleTitle,
			TeamAIRoleDescription: candidate.Member.TeamAIRoleDescription,
			OpenStories:           candidate.Member.OpenStories,
			EstimateTotal:         candidate.Member.EstimateTotal,
			LastStoryActivityAt:   candidate.Member.LastStoryActivityAt,
			DaysSinceLastActivity: daysSinceLastActivity(candidate.Member.LastStoryActivityAt),
			RecentlyActive:        isRecentlyActive(candidate.Member.LastStoryActivityAt),
		}
		if slot, ok := planWorkWindow(candidate, windowStart, windowEnd, duration); ok {
			recommendation.HasAvailableSlot = true
			recommendation.SlotStart = slot.start
			recommendation.SlotEnd = slot.end
		}
		recommendations = append(recommendations, recommendation)
	}
	return recommendations
}

func (s *Service) getWorkspaceWorkSchedule(ctx context.Context, workspaceID uuid.UUID) (workschedule.Schedule, error) {
	if s.workspaceSettings == nil {
		return workschedule.Default(), nil
	}
	settings, err := s.workspaceSettings.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		return workschedule.Schedule{}, fmt.Errorf("get workspace work schedule for Maya plan: %w", err)
	}
	return workschedule.Normalize(settings.WorkingDays, settings.WorkingStartMinute, settings.WorkingEndMinute), nil
}

func (s *Service) getUserWorkSchedule(ctx context.Context, workspaceID, userID uuid.UUID) (workschedule.Schedule, error) {
	workspaceSchedule, err := s.getWorkspaceWorkSchedule(ctx, workspaceID)
	if err != nil {
		return workschedule.Schedule{}, err
	}
	reader, ok := s.users.(UserScheduleReader)
	if !ok {
		return workspaceSchedule, nil
	}
	user, err := reader.GetUser(ctx, userID)
	if err != nil {
		return workschedule.Schedule{}, fmt.Errorf("get user work schedule for Maya plan: %w", err)
	}
	return workschedule.Resolve(
		workspaceSchedule,
		user.WorkingDays,
		user.WorkingStartMinute,
		user.WorkingEndMinute,
	), nil
}

func (s *Service) buildCandidates(ctx context.Context, input CreateWorkPlanInput, story Story) ([]CandidateSchedule, json.RawMessage, error) {
	workspaceSchedule, err := s.getWorkspaceWorkSchedule(ctx, input.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	usersFilter := UserListFilter{TeamID: &story.Team}
	if len(input.CandidateUserIDs) == 0 {
		usersFilter.Limit = defaultCandidateLimit
	}
	members, err := s.users.List(ctx, input.WorkspaceID, usersFilter)
	if err != nil {
		return nil, nil, fmt.Errorf("list team members for maya plan: %w", err)
	}
	workload, err := s.reports.GetWorkloadAnalysis(ctx, input.WorkspaceID, WorkloadReportFilters{TeamIDs: []uuid.UUID{story.Team}})
	if err != nil {
		return nil, nil, fmt.Errorf("get workload for maya plan: %w", err)
	}
	workloadByUserID := make(map[uuid.UUID]MemberWorkload, len(workload.Members))
	for _, member := range workload.Members {
		workloadByUserID[member.UserID] = member
	}
	memberByUserID := make(map[uuid.UUID]User, len(members))
	for _, member := range members {
		memberByUserID[member.ID] = member
	}

	candidateIDs := make(map[uuid.UUID]struct{})
	orderedCandidateIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		if shouldIncludeCandidate(member.ID, input.CandidateUserIDs, s.mayaActorID) {
			if _, exists := candidateIDs[member.ID]; !exists {
				candidateIDs[member.ID] = struct{}{}
				orderedCandidateIDs = append(orderedCandidateIDs, member.ID)
			}
		}
	}
	if len(orderedCandidateIDs) == 0 {
		for _, member := range workload.Members {
			if !shouldIncludeCandidate(member.UserID, input.CandidateUserIDs, s.mayaActorID) {
				continue
			}
			if _, exists := candidateIDs[member.UserID]; exists {
				continue
			}
			candidateIDs[member.UserID] = struct{}{}
			orderedCandidateIDs = append(orderedCandidateIDs, member.UserID)
			if len(input.CandidateUserIDs) == 0 && len(orderedCandidateIDs) >= defaultCandidateLimit {
				break
			}
		}
	}

	candidates := make([]CandidateSchedule, 0, len(orderedCandidateIDs))
	for _, userID := range orderedCandidateIDs {
		member := workloadByUserID[userID]
		if member.UserID == uuid.Nil {
			member = MemberWorkload{UserID: userID}
		}
		if user, ok := memberByUserID[userID]; ok {
			if member.FullName == "" {
				member.FullName = user.FullName
			}
			if member.Username == "" {
				member.Username = user.Username
			}
			if member.AvatarURL == "" {
				member.AvatarURL = user.AvatarURL
			}
			member.LastStoryActivityAt = user.LastStoryActivityAt
			member.TeamAIRoleTitle = user.TeamAIRoleTitle
			member.TeamAIRoleDescription = user.TeamAIRoleDescription
			if member.TeamAIRoleTitle == "" {
				member.TeamAIRoleTitle = user.InferredTeamAIRoleTitle
			}
			if member.TeamAIRoleDescription == "" {
				member.TeamAIRoleDescription = user.InferredTeamAIRoleDescription
			}
		}
		scheduleCalendar, ok := s.calendar.(ScheduleCalendarService)
		if !ok {
			return nil, nil, ErrNotConfigured
		}
		schedule, err := scheduleCalendar.ListSchedulingAvailability(ctx, input.WorkspaceID, userID, input.WindowStart, input.WindowEnd)
		if err != nil {
			return nil, nil, fmt.Errorf("list calendar schedule for candidate %s: %w", userID, err)
		}
		candidate := CandidateSchedule{
			Member:      member,
			Timezone:    schedule.Timezone,
			BusyWindows: schedule.BusyWindows,
			Blocks:      schedule.Blocks,
		}
		user := memberByUserID[userID]
		resolvedSchedule := workschedule.Resolve(
			workspaceSchedule,
			user.WorkingDays,
			user.WorkingStartMinute,
			user.WorkingEndMinute,
		)
		candidate.WorkingDays = resolvedSchedule.WorkingDays
		candidate.WorkingStartMinute = resolvedSchedule.StartMinute
		candidate.WorkingEndMinute = resolvedSchedule.EndMinute
		if feedbackService, ok := s.calendar.(ScheduleFeedbackService); ok {
			preference, preferenceErr := feedbackService.ListManualSchedulePreference(ctx, input.WorkspaceID, userID)
			if preferenceErr != nil {
				return nil, nil, fmt.Errorf("list calendar schedule preference for candidate %s: %w", userID, preferenceErr)
			}
			if preference.Confidence > 0 {
				candidate.PreferredStartMinute = preference.PreferredStartMinute
			}
		}
		candidates = append(candidates, candidate)
	}

	candidateTimezones := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		candidateTimezones[candidate.Member.UserID.String()] = candidate.Timezone
	}
	contextPayload, err := json.Marshal(map[string]any{
		"storyId":            story.ID,
		"storyUpdatedAt":     story.UpdatedAt,
		"teamId":             story.Team,
		"windowStart":        input.WindowStart,
		"windowEnd":          input.WindowEnd,
		"durationMinutes":    input.DurationMinutes,
		"candidateCount":     len(candidates),
		"candidateUserIds":   input.CandidateUserIDs,
		"candidateTimezones": candidateTimezones,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("marshal maya plan context: %w", err)
	}
	return candidates, contextPayload, nil
}

func shouldIncludeCandidate(userID uuid.UUID, candidateUserIDs []uuid.UUID, mayaActorID uuid.UUID) bool {
	if userID == uuid.Nil || userID == mayaActorID {
		return false
	}
	if len(candidateUserIDs) == 0 {
		return true
	}
	return slices.Contains(candidateUserIDs, userID)
}

type actionApplicationOptions struct {
	ExpectedStoryUpdatedAt time.Time
	StoryUpdates           map[string]any
	StoryID                uuid.UUID
	WorkspaceID            uuid.UUID
}
