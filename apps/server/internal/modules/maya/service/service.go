package maya

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/workschedule"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrNotConfigured    = errors.New("maya agent is not configured")
	ErrPlanNotFound     = errors.New("maya plan not found")
	ErrMayaAccessDenied = errors.New("workspace does not have access to Maya auto-scheduling")
)

const defaultCandidateLimit = 15

type StoriesService interface {
	Get(ctx context.Context, storyID, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	UpdateExternal(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any) error
	UpdateExternalWithReason(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any, reason string) error
	UpdateExternalWithReasonIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any, reason string) error
	UpdateAutomationIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any, reason string) error
	UpdateAutomationStateIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, status string, reason *string, locked *bool, schedule *events.StoryScheduleTransition) error
}

type ReportsService interface {
	GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreWorkloadAnalysis, error)
}

type CalendarService interface {
	ListSchedule(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (calendar.CoreSchedule, error)
	CreateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error)
}

type ScheduleCalendarService interface {
	ListSchedulingAvailability(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (calendar.CoreSchedule, error)
	ListMayaScheduleBlocksForStory(ctx context.Context, workspaceID, userID, storyID uuid.UUID) ([]calendar.CoreScheduleBlock, error)
	MayaScheduleOwnershipExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error)
	ReconcileMayaScheduleBlocks(ctx context.Context, input calendar.MayaScheduleReconcileInput) (calendar.CoreScheduleReconcileResult, error)
	DispatchScheduleEventOutbox(ctx context.Context, userID uuid.UUID) error
}

type ScheduleFeedbackService interface {
	ListManualSchedulePreference(ctx context.Context, workspaceID, userID uuid.UUID) (calendar.CoreSchedulePreference, error)
}

type UsersService interface {
	List(ctx context.Context, workspaceID uuid.UUID, filter users.CoreListUsersFilter) ([]users.CoreUser, error)
}

type UserScheduleReader interface {
	GetUser(ctx context.Context, userID uuid.UUID) (users.CoreUser, error)
}

type WorkspaceSettingsService interface {
	GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (workspaces.CoreWorkspaceSettings, error)
}

type Dependencies struct {
	Repository        Repository
	Stories           StoriesService
	Reports           ReportsService
	Calendar          CalendarService
	Users             UsersService
	WorkspaceSettings WorkspaceSettingsService
	Planner           Planner
	MayaActorID       uuid.UUID
}

type Service struct {
	repo              Repository
	stories           StoriesService
	reports           ReportsService
	calendar          CalendarService
	users             UsersService
	workspaceSettings WorkspaceSettingsService
	planner           Planner
	mayaActorID       uuid.UUID
}

type CreateWorkPlanInput struct {
	WorkspaceID      uuid.UUID   `json:"workspaceId"`
	StoryID          uuid.UUID   `json:"storyId"`
	TriggeredBy      uuid.UUID   `json:"triggeredBy"`
	Trigger          RunTrigger  `json:"trigger"`
	WindowStart      time.Time   `json:"windowStart"`
	WindowEnd        time.Time   `json:"windowEnd"`
	DurationMinutes  int         `json:"durationMinutes"`
	CandidateUserIDs []uuid.UUID `json:"candidateUserIds"`
	AutoApply        bool        `json:"autoApply"`
	AssignmentReason string      `json:"-"`
}

type WorkPlan struct {
	Run     CoreRun      `json:"run"`
	Actions []CoreAction `json:"actions"`
}

type ProcessAssignmentBatchInput struct {
	WorkspaceID     uuid.UUID
	TeamID          uuid.UUID
	StoryIDs        []uuid.UUID
	TriggeredBy     uuid.UUID
	WindowStart     time.Time
	WindowEnd       time.Time
	DurationMinutes int
	AutoApply       bool
}

type ProcessAssignmentBatchResult struct {
	Processed int
	Skipped   int
	Plans     []WorkPlan
}

func New(deps Dependencies) *Service {
	planner := deps.Planner
	return &Service{
		repo:              deps.Repository,
		stories:           deps.Stories,
		reports:           deps.Reports,
		calendar:          deps.Calendar,
		users:             deps.Users,
		workspaceSettings: deps.WorkspaceSettings,
		planner:           planner,
		mayaActorID:       deps.MayaActorID,
	}
}

func (s *Service) CreateWorkPlan(ctx context.Context, input CreateWorkPlanInput) (WorkPlan, error) {
	if !input.AutoApply {
		return s.createWorkPlan(ctx, input)
	}
	if err := s.validate(); err != nil {
		return WorkPlan{}, err
	}
	if input.WorkspaceID == uuid.Nil || input.StoryID == uuid.Nil {
		return WorkPlan{}, fmt.Errorf("%w: story and workspace are required", ErrInvalidPlanInput)
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return WorkPlan{}, err
	}
	var plan WorkPlan
	err = scheduleRepo.WithScheduleStoryLock(ctx, input.WorkspaceID, input.StoryID, func() error {
		var planErr error
		plan, planErr = s.createWorkPlan(ctx, input)
		return planErr
	})
	return plan, err
}

func (s *Service) createWorkPlan(ctx context.Context, input CreateWorkPlanInput) (WorkPlan, error) {
	ctx, span := web.AddSpan(ctx, "business.core.maya.CreateWorkPlan")
	defer span.End()

	if err := s.validate(); err != nil {
		span.RecordError(err)
		return WorkPlan{}, err
	}
	if input.Trigger == "" {
		input.Trigger = RunTriggerManual
	}

	story, err := s.stories.Get(ctx, input.StoryID, input.WorkspaceID)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("get story for maya plan: %w", err)
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return WorkPlan{}, err
	}
	hasAccess, err := scheduleRepo.WorkspaceCanUseMaya(ctx, input.WorkspaceID)
	if err != nil {
		return WorkPlan{}, err
	}
	if !hasAccess {
		return WorkPlan{}, ErrMayaAccessDenied
	}
	if input.AutoApply {
		active, activeErr := scheduleRepo.StoryIsActiveForAutoScheduling(ctx, input.WorkspaceID, input.StoryID)
		if activeErr != nil {
			return WorkPlan{}, activeErr
		}
		if !active {
			return WorkPlan{}, fmt.Errorf("%w: story is not active for auto-scheduling", ErrInvalidPlanInput)
		}
	}
	if input.AutoApply && story.AutoSchedulingLocked {
		return WorkPlan{}, stories.ErrAutoSchedulingOwnerLocked
	}
	prePlanUpdates := make(map[string]any)
	if input.AutoApply && !story.AutoSchedulingEnabled {
		prePlanUpdates["auto_scheduling_enabled"] = true
		prePlanUpdates["auto_scheduling_locked"] = false
	}
	if input.AutoApply && input.DurationMinutes > 0 && (story.EstimatedDurationMinutes == nil || *story.EstimatedDurationMinutes != input.DurationMinutes) {
		prePlanUpdates["estimated_duration_minutes"] = input.DurationMinutes
		if story.MinimumFocusBlockMinutes != nil && *story.MinimumFocusBlockMinutes > input.DurationMinutes {
			prePlanUpdates["minimum_focus_block_minutes"] = input.DurationMinutes
		}
	}
	if len(prePlanUpdates) > 0 {
		if err := s.stories.UpdateAutomationIfUnchanged(
			ctx,
			s.mayaActorID,
			story.ID,
			story.Workspace,
			story.UpdatedAt,
			prePlanUpdates,
			"Maya saved the confirmed scheduling preferences for this work plan.",
		); err != nil {
			return WorkPlan{}, fmt.Errorf("enable Maya auto-scheduling: %w", err)
		}
		story, err = s.stories.Get(ctx, input.StoryID, input.WorkspaceID)
		if err != nil {
			return WorkPlan{}, fmt.Errorf("reload story after enabling Maya auto-scheduling: %w", err)
		}
	}
	candidates, contextPayload, err := s.buildCandidates(ctx, input, story)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, err
	}

	run, err := s.repo.CreateRun(ctx, CreateRunInput{
		WorkspaceID: input.WorkspaceID,
		StoryID:     input.StoryID,
		TriggeredBy: input.TriggeredBy,
		Trigger:     input.Trigger,
		Context:     contextPayload,
	})
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("create maya run: %w", err)
	}

	result, err := s.planner.Plan(PlanInput{
		Context:          ctx,
		WorkspaceID:      input.WorkspaceID,
		Story:            story,
		DurationMinutes:  input.DurationMinutes,
		WindowStart:      input.WindowStart,
		WindowEnd:        input.WindowEnd,
		Candidates:       candidates,
		AssignmentReason: input.AssignmentReason,
	})
	if err != nil {
		message := err.Error()
		completed, completeErr := s.repo.CompleteRun(ctx, run.ID, RunStatusFailed, "", &message)
		if completeErr != nil {
			return WorkPlan{}, fmt.Errorf("complete failed maya run: %w", completeErr)
		}
		return WorkPlan{Run: completed}, err
	}
	if result.SelectedUserID != nil {
		for _, candidate := range candidates {
			if candidate.Member.UserID == *result.SelectedUserID {
				result.Timezone = candidate.Timezone
				break
			}
		}
	}
	previousBlocks := []calendar.CoreScheduleBlock{}
	if input.AutoApply && result.SelectedUserID != nil {
		ownerIDs, ownerErr := scheduleRepo.ListMayaScheduleOwners(ctx, story.Workspace, story.ID)
		if ownerErr != nil {
			return WorkPlan{}, ownerErr
		}
		if !slices.Contains(ownerIDs, *result.SelectedUserID) {
			ownerIDs = append(ownerIDs, *result.SelectedUserID)
		}
		scheduleCalendar, calendarErr := s.scheduleCalendarService()
		if calendarErr != nil {
			return WorkPlan{}, calendarErr
		}
		for _, ownerID := range ownerIDs {
			blocks, listErr := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, story.Workspace, ownerID, story.ID)
			if listErr != nil {
				return WorkPlan{}, listErr
			}
			previousBlocks = append(previousBlocks, blocks...)
		}
	}

	for i := range result.Actions {
		result.Actions[i].RunID = run.ID
		result.Actions[i].WorkspaceID = input.WorkspaceID
		result.Actions[i].StoryID = input.StoryID
	}
	actions, err := s.repo.CreateActions(ctx, result.Actions)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("create maya actions: %w", err)
	}
	if input.AutoApply {
		actions = s.applyActions(ctx, actions)
	}

	runStatus := RunStatusSucceeded
	var runError *string
	for _, action := range actions {
		if action.Status == ActionStatusFailed {
			runStatus = RunStatusFailed
			if action.Error != nil {
				message := *action.Error
				runError = &message
			}
			break
		}
	}
	completed, err := s.repo.CompleteRun(ctx, run.ID, runStatus, result.Summary, runError)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("complete maya run: %w", err)
	}
	if input.AutoApply && runStatus == RunStatusSucceeded {
		if err := s.finalizeAppliedWorkPlan(ctx, story, result, previousBlocks); err != nil {
			return WorkPlan{Run: completed, Actions: actions}, fmt.Errorf("finalize Maya auto-scheduling state: %w", err)
		}
	}
	return WorkPlan{Run: completed, Actions: actions}, nil
}

func (s *Service) finalizeAppliedWorkPlan(ctx context.Context, plannedStory stories.CoreSingleStory, result PlanResult, previousBlocks []calendar.CoreScheduleBlock) error {
	story, err := s.stories.Get(ctx, plannedStory.ID, plannedStory.Workspace)
	if err != nil {
		return err
	}
	if !story.AutoSchedulingEnabled {
		return nil
	}
	if result.SelectedUserID == nil || *result.SelectedUserID == uuid.Nil {
		reason := "Maya could not select an eligible teammate yet."
		return s.stories.UpdateAutomationStateIfUnchanged(
			ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
			stories.AutoSchedulingStatusNeedsOwner, &reason, nil, nil,
		)
	}
	ownerID := *result.SelectedUserID
	if story.Assignee == nil || *story.Assignee != ownerID {
		return errors.New("applied schedule owner does not match the current story assignee")
	}
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return err
	}
	blocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, story.Workspace, ownerID, story.ID)
	if err != nil {
		return err
	}
	segments := mayaSegmentsFromBlocks(blocks)
	status, reason := autoSchedulingOutcome(result, segments)
	if story.AutoSchedulingLocked && len(segments) > 0 {
		status = stories.AutoSchedulingStatusLocked
		reason = "Maya retained the locked schedule without moving its time."
	}
	reason = refineScheduleOutcomeReason(previousBlocks, segments, status, reason)
	transition := buildStoryScheduleTransition(story, ownerID, previousBlocks, segments, result.Timezone, status, reason)
	return s.stories.UpdateAutomationStateIfUnchanged(
		ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
		status, &reason, nil, transition,
	)
}

func (s *Service) ProcessAssignmentBatch(ctx context.Context, input ProcessAssignmentBatchInput) (ProcessAssignmentBatchResult, error) {
	ctx, span := web.AddSpan(ctx, "business.core.maya.ProcessAssignmentBatch")
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
	storiesForBatch := make([]stories.CoreSingleStory, 0, len(input.StoryIDs))
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
	storyByID := make(map[uuid.UUID]stories.CoreSingleStory, len(storiesForBatch))
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

func effectiveWorkDurationMinutes(story stories.CoreSingleStory, requestedDurationMinutes int) int {
	if requestedDurationMinutes > 0 {
		return requestedDurationMinutes
	}
	return estimatedWorkDurationMinutes(story)
}

func batchCandidateDurationMinutes(storiesForBatch []stories.CoreSingleStory, requestedDurationMinutes int) int {
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

func (s *Service) buildCandidates(ctx context.Context, input CreateWorkPlanInput, story stories.CoreSingleStory) ([]CandidateSchedule, json.RawMessage, error) {
	workspaceSchedule, err := s.getWorkspaceWorkSchedule(ctx, input.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	usersFilter := users.CoreListUsersFilter{TeamID: &story.Team}
	if len(input.CandidateUserIDs) == 0 {
		usersFilter.Limit = defaultCandidateLimit
	}
	members, err := s.users.List(ctx, input.WorkspaceID, usersFilter)
	if err != nil {
		return nil, nil, fmt.Errorf("list team members for maya plan: %w", err)
	}
	workload, err := s.reports.GetWorkloadAnalysis(ctx, input.WorkspaceID, reports.ReportFilters{TeamIDs: []uuid.UUID{story.Team}})
	if err != nil {
		return nil, nil, fmt.Errorf("get workload for maya plan: %w", err)
	}
	workloadByUserID := make(map[uuid.UUID]reports.CoreMemberWorkload, len(workload.Members))
	for _, member := range workload.Members {
		workloadByUserID[member.UserID] = member
	}
	memberByUserID := make(map[uuid.UUID]users.CoreUser, len(members))
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
			member = reports.CoreMemberWorkload{UserID: userID}
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

	contextPayload, err := json.Marshal(map[string]any{
		"storyId":          story.ID,
		"teamId":           story.Team,
		"windowStart":      input.WindowStart,
		"windowEnd":        input.WindowEnd,
		"durationMinutes":  input.DurationMinutes,
		"candidateCount":   len(candidates),
		"candidateUserIds": input.CandidateUserIDs,
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

func (s *Service) applyActions(ctx context.Context, actions []CoreAction) []CoreAction {
	applied := make([]CoreAction, len(actions))
	copy(applied, actions)
	scheduleIndexes := make([]int, 0, len(applied))
	for index := range applied {
		if applied[index].Type == ActionTypeScheduleWorkBlock {
			scheduleIndexes = append(scheduleIndexes, index)
		}
	}
	scheduleApplied := len(scheduleIndexes) == 0
	var scheduleState *appliedScheduleState
	if len(scheduleIndexes) > 0 {
		scheduleActions := make([]CoreAction, 0, len(scheduleIndexes))
		for _, index := range scheduleIndexes {
			scheduleActions = append(scheduleActions, applied[index])
		}
		state, err := s.applyScheduleActionsAtomically(ctx, scheduleActions)
		if err != nil {
			for _, index := range scheduleIndexes {
				s.markActionFailed(ctx, &applied[index], err.Error())
			}
		} else {
			scheduleApplied = true
			scheduleState = &state
		}
	}
	dispatchSchedule := false
	for i, action := range applied {
		if action.Type == ActionTypeScheduleWorkBlock {
			continue
		}
		if action.Type == ActionTypeAssignStory && !scheduleApplied {
			s.markActionFailed(ctx, &applied[i], "assignment was not applied because the schedule could not be committed")
			continue
		}
		if err := s.applyAction(ctx, action); err != nil {
			s.markActionFailed(ctx, &applied[i], err.Error())
			if action.Type == ActionTypeAssignStory && scheduleState != nil {
				rollbackErr := s.restoreScheduleState(ctx, *scheduleState)
				scheduleApplied = false
				message := "schedule was restored because the dependent story assignment failed"
				if rollbackErr != nil {
					message = errors.Join(errors.New(message), rollbackErr).Error()
				} else {
					dispatchSchedule = true
				}
				for _, index := range scheduleIndexes {
					s.markActionFailed(ctx, &applied[index], message)
				}
			}
			continue
		}
		if action.Type == ActionTypeAssignStory && scheduleState != nil {
			if err := s.refreshAppliedScheduleAfterAssignment(ctx, *scheduleState); err != nil {
				// The assignment is already committed and must never be reported as
				// rolled back. Leave provider delivery fenced by the stale ownership
				// watermark; the assignment event/recovery sweep will retry the exact
				// schedule against the current story version.
				s.markActionApplied(ctx, &applied[i])
				scheduleApplied = false
				for _, index := range scheduleIndexes {
					s.markActionFailed(ctx, &applied[index], "schedule ownership refresh failed after assignment: "+err.Error())
				}
				continue
			}
		}
		s.markActionApplied(ctx, &applied[i])
	}
	if scheduleApplied {
		for _, index := range scheduleIndexes {
			s.markActionApplied(ctx, &applied[index])
		}
		dispatchSchedule = scheduleState != nil
	}
	if dispatchSchedule && scheduleState != nil {
		// The database outbox is the delivery contract. A transient provider
		// failure must not change the result of the durable local mutation.
		_ = scheduleState.calendar.DispatchScheduleEventOutbox(ctx, scheduleState.userID)
	}
	return applied
}

type appliedScheduleState struct {
	calendar       ScheduleCalendarService
	workspaceID    uuid.UUID
	userID         uuid.UUID
	storyID        uuid.UUID
	previousOwners []scheduleOwnerState
	segments       []calendar.MayaScheduleSegmentInput
}

type scheduleOwnerState struct {
	userID        uuid.UUID
	segments      []calendar.MayaScheduleSegmentInput
	keepOwnership bool
}

func (s *Service) applyScheduleActionsAtomically(ctx context.Context, actions []CoreAction) (appliedScheduleState, error) {
	if len(actions) == 0 {
		return appliedScheduleState{}, nil
	}
	first := actions[0]
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return appliedScheduleState{}, err
	}
	if first.Payload.ScheduleBlock == nil {
		return appliedScheduleState{}, fmt.Errorf("missing schedule block payload")
	}
	userID := first.Payload.ScheduleBlock.UserID
	expectedStoryUpdatedAt := first.Payload.ScheduleBlock.ExpectedStoryUpdatedAt
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return appliedScheduleState{}, err
	}
	ownerIDs, err := scheduleRepo.ListMayaScheduleOwners(ctx, first.WorkspaceID, first.StoryID)
	if err != nil {
		return appliedScheduleState{}, err
	}
	if !slices.Contains(ownerIDs, userID) {
		ownerIDs = append(ownerIDs, userID)
	}
	previousOwners := make([]scheduleOwnerState, 0, len(ownerIDs))
	for _, ownerID := range ownerIDs {
		currentBlocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, first.WorkspaceID, ownerID, first.StoryID)
		if err != nil {
			return appliedScheduleState{}, err
		}
		previousOwnership, err := scheduleCalendar.MayaScheduleOwnershipExists(ctx, first.WorkspaceID, ownerID, first.StoryID)
		if err != nil {
			return appliedScheduleState{}, err
		}
		previousOwners = append(previousOwners, scheduleOwnerState{
			userID: ownerID, segments: mayaSegmentsFromBlocks(currentBlocks), keepOwnership: previousOwnership,
		})
	}
	segments := make([]calendar.MayaScheduleSegmentInput, 0, len(actions))
	preemptBlockIDs := []uuid.UUID(nil)
	for _, action := range actions {
		payload := action.Payload.ScheduleBlock
		if payload == nil || action.WorkspaceID != first.WorkspaceID || action.StoryID != first.StoryID || payload.UserID != userID {
			return appliedScheduleState{}, fmt.Errorf("schedule actions do not belong to one story and user")
		}
		if !payload.ExpectedStoryUpdatedAt.Equal(expectedStoryUpdatedAt) {
			return appliedScheduleState{}, fmt.Errorf("schedule actions were planned from different story versions")
		}
		if len(payload.PreemptBlockIDs) > 0 {
			if len(preemptBlockIDs) == 0 {
				preemptBlockIDs = append([]uuid.UUID(nil), payload.PreemptBlockIDs...)
			} else if !slices.Equal(preemptBlockIDs, payload.PreemptBlockIDs) {
				return appliedScheduleState{}, fmt.Errorf("schedule actions were planned with different preemption sets")
			}
		}
		if payload.Operation == ScheduleBlockOperationRetain {
			continue
		}
		if payload.Operation != "" && payload.Operation != ScheduleBlockOperationUpsert {
			return appliedScheduleState{}, fmt.Errorf("unsupported initial schedule operation %q", payload.Operation)
		}
		segments = append(segments, calendar.MayaScheduleSegmentInput{
			SegmentIndex: payload.SegmentIndex,
			Title:        payload.Title,
			StartAt:      payload.StartAt,
			EndAt:        payload.EndAt,
		})
	}
	if _, err := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
		WorkspaceID: first.WorkspaceID, UserID: userID, StoryID: first.StoryID,
		ExpectedStoryUpdatedAt: &expectedStoryUpdatedAt, Segments: segments, PreemptBlockIDs: preemptBlockIDs, KeepOwnership: true,
	}); err != nil {
		return appliedScheduleState{}, err
	}
	state := appliedScheduleState{
		calendar: scheduleCalendar, workspaceID: first.WorkspaceID, userID: userID,
		storyID: first.StoryID, previousOwners: previousOwners, segments: segments,
	}
	for _, previousOwner := range previousOwners {
		if previousOwner.userID == userID {
			continue
		}
		if _, err := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
			WorkspaceID: first.WorkspaceID, UserID: previousOwner.userID, StoryID: first.StoryID,
		}); err != nil {
			return appliedScheduleState{}, errors.Join(err, s.restoreScheduleState(ctx, state))
		}
	}
	return state, nil
}

func (s *Service) refreshAppliedScheduleAfterAssignment(ctx context.Context, state appliedScheduleState) error {
	story, err := s.stories.Get(ctx, state.storyID, state.workspaceID)
	if err != nil {
		return err
	}
	if !story.AutoSchedulingEnabled || story.Assignee == nil || *story.Assignee != state.userID {
		return errors.New("story assignment no longer matches the committed Maya schedule")
	}
	_, err = state.calendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
		WorkspaceID:            state.workspaceID,
		UserID:                 state.userID,
		StoryID:                state.storyID,
		ExpectedStoryUpdatedAt: &story.UpdatedAt,
		Segments:               state.segments,
		KeepOwnership:          true,
		Locked:                 story.AutoSchedulingLocked,
	})
	return err
}

func (s *Service) restoreScheduleState(ctx context.Context, state appliedScheduleState) error {
	previousByOwner := make(map[uuid.UUID]scheduleOwnerState, len(state.previousOwners))
	for _, previous := range state.previousOwners {
		previousByOwner[previous.userID] = previous
	}
	selected, existed := previousByOwner[state.userID]
	if !existed {
		selected = scheduleOwnerState{userID: state.userID}
	}
	var restoreErr error
	if _, err := state.calendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
		WorkspaceID: state.workspaceID, UserID: selected.userID, StoryID: state.storyID,
		Segments: selected.segments, KeepOwnership: selected.keepOwnership,
	}); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}
	for _, previous := range state.previousOwners {
		if previous.userID == state.userID {
			continue
		}
		if _, err := state.calendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
			WorkspaceID: state.workspaceID, UserID: previous.userID, StoryID: state.storyID,
			Segments: previous.segments, KeepOwnership: previous.keepOwnership,
		}); err != nil {
			restoreErr = errors.Join(restoreErr, err)
		}
	}
	return restoreErr
}

func mayaSegmentsFromBlocks(blocks []calendar.CoreScheduleBlock) []calendar.MayaScheduleSegmentInput {
	segments := make([]calendar.MayaScheduleSegmentInput, 0, len(blocks))
	for _, block := range blocks {
		segments = append(segments, calendar.MayaScheduleSegmentInput{
			SegmentIndex: block.SegmentIndex, Title: block.Title, StartAt: block.StartAt, EndAt: block.EndAt,
		})
	}
	return segments
}

func (s *Service) markActionApplied(ctx context.Context, action *CoreAction) {
	now := time.Now().UTC()
	action.Status = ActionStatusApplied
	action.AppliedAt = &now
	action.Error = nil
	_ = s.repo.MarkActionApplied(ctx, action.ID)
}

func (s *Service) markActionFailed(ctx context.Context, action *CoreAction, message string) {
	action.Status = ActionStatusFailed
	action.Error = &message
	_ = s.repo.MarkActionFailed(ctx, action.ID, message)
}

func (s *Service) applyAction(ctx context.Context, action CoreAction) error {
	switch action.Type {
	case ActionTypeAssignStory:
		if action.Payload.AssignStory == nil {
			return fmt.Errorf("missing assign story payload")
		}
		return s.stories.UpdateAutomationIfUnchanged(ctx, s.mayaActorID, action.StoryID, action.WorkspaceID, action.Payload.AssignStory.ExpectedUpdatedAt, map[string]any{
			"assignee_id": action.Payload.AssignStory.AssigneeID,
		}, action.Reason)
	case ActionTypeScheduleWorkBlock:
		return fmt.Errorf("schedule work blocks must be applied as one atomic set")
	case ActionTypeFlagScheduleRisk:
		return nil
	default:
		return fmt.Errorf("unsupported maya action type: %s", action.Type)
	}
}
