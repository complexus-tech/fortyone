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
	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/workweek"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrNotConfigured = errors.New("maya agent is not configured")
	ErrPlanNotFound  = errors.New("maya plan not found")
)

const defaultCandidateLimit = 15

type StoriesService interface {
	Get(ctx context.Context, storyID, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	UpdateExternal(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any) error
	UpdateExternalWithReason(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any, reason string) error
}

type ReportsService interface {
	GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreWorkloadAnalysis, error)
}

type CalendarService interface {
	ListSchedule(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (calendar.CoreSchedule, error)
	CreateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error)
}

type UsersService interface {
	List(ctx context.Context, workspaceID uuid.UUID, filter users.CoreListUsersFilter) ([]users.CoreUser, error)
}

type TeamSettingsService interface {
	GetSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (teamsettings.CoreTeamSprintSettings, error)
}

type Dependencies struct {
	Repository   Repository
	Stories      StoriesService
	Reports      ReportsService
	Calendar     CalendarService
	Users        UsersService
	TeamSettings TeamSettingsService
	Planner      Planner
	MayaActorID  uuid.UUID
}

type Service struct {
	repo         Repository
	stories      StoriesService
	reports      ReportsService
	calendar     CalendarService
	users        UsersService
	teamSettings TeamSettingsService
	planner      Planner
	mayaActorID  uuid.UUID
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
		repo:         deps.Repository,
		stories:      deps.Stories,
		reports:      deps.Reports,
		calendar:     deps.Calendar,
		users:        deps.Users,
		teamSettings: deps.TeamSettings,
		planner:      planner,
		mayaActorID:  deps.MayaActorID,
	}
}

func (s *Service) CreateWorkPlan(ctx context.Context, input CreateWorkPlanInput) (WorkPlan, error) {
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
	workingDays, err := s.getWorkingDays(ctx, story.Team, input.WorkspaceID)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, err
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
		WorkingDays:      workingDays,
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

	completed, err := s.repo.CompleteRun(ctx, run.ID, RunStatusSucceeded, result.Summary, nil)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("complete maya run: %w", err)
	}
	return WorkPlan{Run: completed, Actions: actions}, nil
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
		if story.Team != input.TeamID || story.Assignee == nil || *story.Assignee != s.mayaActorID {
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

	workingDays, err := s.getWorkingDays(ctx, input.TeamID, input.WorkspaceID)
	if err != nil {
		span.RecordError(err)
		return ProcessAssignmentBatchResult{}, err
	}
	candidateRecommendations := candidateRecommendationsFromSchedules(candidates, input.WindowStart, input.WindowEnd, batchCandidateDurationMinutes(storiesForBatch, input.DurationMinutes), workingDays)
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
	duration := defaultDurationMinutes
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

func candidateRecommendationsFromSchedules(candidates []CandidateSchedule, windowStart, windowEnd time.Time, durationMinutes int, workingDays []int) []CandidateRecommendation {
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
		if slot, ok := planWorkWindow(candidate, windowStart, windowEnd, duration, workingDays); ok {
			recommendation.HasAvailableSlot = true
			recommendation.SlotStart = slot.start
			recommendation.SlotEnd = slot.end
		}
		recommendations = append(recommendations, recommendation)
	}
	return recommendations
}

func (s *Service) getWorkingDays(ctx context.Context, teamID, workspaceID uuid.UUID) ([]int, error) {
	if s.teamSettings == nil {
		return workweek.DefaultWorkingDays(), nil
	}
	settings, err := s.teamSettings.GetSprintSettings(ctx, teamID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("get team working days for maya plan: %w", err)
	}
	return workweek.Normalize(settings.WorkingDays), nil
}

func (s *Service) buildCandidates(ctx context.Context, input CreateWorkPlanInput, story stories.CoreSingleStory) ([]CandidateSchedule, json.RawMessage, error) {
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
		schedule, err := s.calendar.ListSchedule(ctx, input.WorkspaceID, userID, input.WindowStart, input.WindowEnd)
		if err != nil {
			return nil, nil, fmt.Errorf("list calendar schedule for candidate %s: %w", userID, err)
		}
		candidates = append(candidates, CandidateSchedule{
			Member:      member,
			BusyWindows: schedule.BusyWindows,
			Blocks:      schedule.Blocks,
		})
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
	for i, action := range applied {
		if err := s.applyAction(ctx, action); err != nil {
			message := err.Error()
			applied[i].Status = ActionStatusFailed
			applied[i].Error = &message
			_ = s.repo.MarkActionFailed(ctx, action.ID, message)
			continue
		}
		now := time.Now().UTC()
		applied[i].Status = ActionStatusApplied
		applied[i].AppliedAt = &now
		_ = s.repo.MarkActionApplied(ctx, action.ID)
	}
	return applied
}

func (s *Service) applyAction(ctx context.Context, action CoreAction) error {
	switch action.Type {
	case ActionTypeAssignStory:
		if action.Payload.AssignStory == nil {
			return fmt.Errorf("missing assign story payload")
		}
		return s.stories.UpdateExternalWithReason(ctx, s.mayaActorID, action.StoryID, action.WorkspaceID, map[string]any{
			"assignee_id": action.Payload.AssignStory.AssigneeID,
		}, action.Reason)
	case ActionTypeScheduleWorkBlock:
		if action.Payload.ScheduleBlock == nil {
			return fmt.Errorf("missing schedule block payload")
		}
		scheduleBlock := action.Payload.ScheduleBlock
		if _, err := s.calendar.CreateScheduleBlock(ctx, calendar.CoreScheduleBlockInput{
			WorkspaceID: action.WorkspaceID,
			UserID:      scheduleBlock.UserID,
			StoryID:     &action.StoryID,
			BlockType:   calendar.ScheduleBlockTypeWork,
			Title:       scheduleBlock.Title,
			StartAt:     scheduleBlock.StartAt,
			EndAt:       scheduleBlock.EndAt,
			IsLocked:    true,
			Source:      calendar.ScheduleBlockSourceMaya,
		}); err != nil {
			return err
		}
		updates := storyDateUpdatesFromSchedule(scheduleBlock)
		return s.stories.UpdateExternalWithReason(ctx, s.mayaActorID, action.StoryID, action.WorkspaceID, updates, action.Reason)
	case ActionTypeFlagScheduleRisk:
		return nil
	default:
		return fmt.Errorf("unsupported maya action type: %s", action.Type)
	}
}

func storyDateUpdatesFromSchedule(scheduleBlock *ScheduleBlockPayload) map[string]any {
	updates := make(map[string]any, 2)
	if scheduleBlock == nil {
		return updates
	}

	startDate := scheduleBlock.PlannedStartAt.UTC()
	endDate := scheduleBlock.PlannedEndAt.UTC()
	if startDate.IsZero() {
		startDate = scheduleBlock.StartAt.UTC()
	}
	if endDate.IsZero() {
		endDate = scheduleBlock.EndAt.UTC()
	}

	updates["start_date"] = startDate
	updates["end_date"] = endDate
	return updates
}
