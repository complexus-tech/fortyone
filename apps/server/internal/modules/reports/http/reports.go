package reportshttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("invalid workspace id")
	ErrInvalidDate        = errors.New("invalid date parameter")
	avatarAccessURLExpiry = 24 * time.Hour
)

type Handlers struct {
	reports     *reports.Service
	log         *logger.Logger
	attachments *attachments.Service
}

func New(log *logger.Logger, reports *reports.Service, attachments *attachments.Service) *Handlers {
	return &Handlers{
		reports:     reports,
		log:         log,
		attachments: attachments,
	}
}

func (h *Handlers) resolveUserAvatarURL(ctx context.Context, avatar string) string {
	if h.attachments == nil {
		return avatar
	}
	resolved, err := h.attachments.ResolveProfileImageURL(ctx, avatar, avatarAccessURLExpiry)
	if err != nil {
		return ""
	}
	return resolved
}

func (h *Handlers) GetStoryStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	startDate, endDate, err := date.RangeFromQuery(r.URL.Query(), 30)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}

	stats, err := h.reports.GetStoryStats(ctx, workspace.ID, reports.StoryStatsFilters{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppStoryStats(stats), http.StatusOK)
}

func (h *Handlers) GetContributionStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	startDate, endDate, err := date.RangeFromQuery(r.URL.Query(), 30)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}

	stats, err := h.reports.GetContributionStats(ctx, userID, workspace.ID, startDate, endDate)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppContributionsStats(stats), http.StatusOK)
}

func (h *Handlers) GetUserStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	stats, err := h.reports.GetUserStats(ctx, userID, workspace.ID)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppUserStats(stats), http.StatusOK)
}

func (h *Handlers) GetStatusStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetStatusStats")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppStatsFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	coreFilters := reports.StatsFilters{}
	if filters["teamId"] != nil {
		teamID, err := uuid.Parse(filters["teamId"].(string))
		if err == nil {
			coreFilters.TeamID = &teamID
		}
	}
	if filters["sprintId"] != nil {
		sprintID, err := uuid.Parse(filters["sprintId"].(string))
		if err == nil {
			coreFilters.SprintID = &sprintID
		}
	}
	if filters["objectiveId"] != nil {
		objectiveID, err := uuid.Parse(filters["objectiveId"].(string))
		if err == nil {
			coreFilters.ObjectiveID = &objectiveID
		}
	}

	startDate, endDate, err := date.RangeFromQuery(r.URL.Query(), 30)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}
	coreFilters.StartDate = startDate
	coreFilters.EndDate = endDate

	stats, err := h.reports.GetStatusStats(ctx, workspace.ID, coreFilters)
	if err != nil {
		return fmt.Errorf("getting status stats: %w", err)
	}

	return web.Respond(ctx, w, toAppStatusStats(stats), http.StatusOK)
}

func (h *Handlers) GetPriorityStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetPriorityStats")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppStatsFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	coreFilters := reports.StatsFilters{}
	if filters["teamId"] != nil {
		teamID, err := uuid.Parse(filters["teamId"].(string))
		if err == nil {
			coreFilters.TeamID = &teamID
		}
	}
	if filters["sprintId"] != nil {
		sprintID, err := uuid.Parse(filters["sprintId"].(string))
		if err == nil {
			coreFilters.SprintID = &sprintID
		}
	}
	if filters["objectiveId"] != nil {
		objectiveID, err := uuid.Parse(filters["objectiveId"].(string))
		if err == nil {
			coreFilters.ObjectiveID = &objectiveID
		}
	}

	startDate, endDate, err := date.RangeFromQuery(r.URL.Query(), 30)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}
	coreFilters.StartDate = startDate
	coreFilters.EndDate = endDate

	stats, err := h.reports.GetPriorityStats(ctx, workspace.ID, coreFilters)
	if err != nil {
		return fmt.Errorf("getting priority stats: %w", err)
	}

	return web.Respond(ctx, w, toAppPriorityStats(stats), http.StatusOK)
}

func parseReportFilters(query map[string]interface{}) (reports.ReportFilters, error) {
	filters := reports.ReportFilters{}

	if teamIds, ok := query["teamIds"].(string); ok {
		filters.TeamIDs = parseCommaSeparatedUUIDs(teamIds)
	}
	if assigneeIds, ok := query["assigneeIds"].(string); ok {
		filters.AssigneeIDs = parseCommaSeparatedUUIDs(assigneeIds)
	}
	if sprintIds, ok := query["sprintIds"].(string); ok {
		filters.SprintIDs = parseCommaSeparatedUUIDs(sprintIds)
	}
	if objectiveIds, ok := query["objectiveIds"].(string); ok {
		filters.ObjectiveIDs = parseCommaSeparatedUUIDs(objectiveIds)
	}

	now := time.Now()
	defaultEndDate := now
	defaultStartDate := now.AddDate(0, 0, -60)

	if startDate, ok := query["startDate"]; ok {
		if startDateStr, ok := startDate.(string); ok {
			parsedDate, err := parseReportDate(startDateStr)
			if err != nil {
				return reports.ReportFilters{}, ErrInvalidDate
			}
			filters.StartDate = &parsedDate
		}
	} else {
		filters.StartDate = &defaultStartDate
	}

	if endDate, ok := query["endDate"]; ok {
		if endDateStr, ok := endDate.(string); ok {
			parsedDate, err := parseReportDate(endDateStr)
			if err != nil {
				return reports.ReportFilters{}, ErrInvalidDate
			}
			filters.EndDate = &parsedDate
		}
	} else {
		filters.EndDate = &defaultEndDate
	}

	return filters, nil
}

func parseWorkloadAnalysisFilters(query map[string]interface{}) (reports.ReportFilters, error) {
	filters := reports.ReportFilters{}

	if teamIds, ok := query["teamIds"].(string); ok {
		filters.TeamIDs = parseCommaSeparatedUUIDs(teamIds)
	}
	if assigneeIds, ok := query["assigneeIds"].(string); ok {
		filters.AssigneeIDs = parseCommaSeparatedUUIDs(assigneeIds)
	}
	if sprintIds, ok := query["sprintIds"].(string); ok {
		filters.SprintIDs = parseCommaSeparatedUUIDs(sprintIds)
	}
	if objectiveIds, ok := query["objectiveIds"].(string); ok {
		filters.ObjectiveIDs = parseCommaSeparatedUUIDs(objectiveIds)
	}
	if startDate, ok := query["startDate"].(string); ok && startDate != "" {
		parsedDate, err := parseReportDate(startDate)
		if err != nil {
			return reports.ReportFilters{}, ErrInvalidDate
		}
		filters.StartDate = &parsedDate
	}
	if endDate, ok := query["endDate"].(string); ok && endDate != "" {
		parsedDate, err := parseReportDate(endDate)
		if err != nil {
			return reports.ReportFilters{}, ErrInvalidDate
		}
		filters.EndDate = &parsedDate
	}

	return filters, nil
}

func parseReportDate(value string) (time.Time, error) {
	if parsedDate, err := time.Parse(time.RFC3339, value); err == nil {
		return parsedDate, nil
	}

	return time.Parse("2006-01-02", value)
}

func parseCommaSeparatedUUIDs(value string) []uuid.UUID {
	if value == "" {
		return nil
	}

	ids := strings.Split(value, ",")
	result := make([]uuid.UUID, 0, len(ids))
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		if id, err := uuid.Parse(idStr); err == nil {
			result = append(result, id)
		}
	}

	return result
}

func (h *Handlers) GetWorkspaceOverview(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetWorkspaceOverview")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	overview, err := h.reports.GetWorkspaceOverview(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting workspace overview: %w", err)
	}

	return web.Respond(ctx, w, toAppWorkspaceOverview(overview), http.StatusOK)
}

func (h *Handlers) GetStoryAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetStoryAnalytics")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	analytics, err := h.reports.GetStoryAnalytics(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting story analytics: %w", err)
	}

	return web.Respond(ctx, w, toAppStoryAnalytics(analytics), http.StatusOK)
}

func (h *Handlers) GetObjectiveProgress(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetObjectiveProgress")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	progress, err := h.reports.GetObjectiveProgress(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting objective progress: %w", err)
	}

	return web.Respond(ctx, w, toAppObjectiveProgress(progress), http.StatusOK)
}

func (h *Handlers) GetTeamPerformance(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetTeamPerformance")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	performance, err := h.reports.GetTeamPerformance(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting team performance: %w", err)
	}

	for i := range performance.MemberContributions {
		performance.MemberContributions[i].AvatarURL = h.resolveUserAvatarURL(ctx, performance.MemberContributions[i].AvatarURL)
	}

	return web.Respond(ctx, w, toAppTeamPerformance(performance), http.StatusOK)
}

func (h *Handlers) GetWorkloadAnalysis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetWorkloadAnalysis")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseWorkloadAnalysisFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	analysis, err := h.reports.GetWorkloadAnalysis(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting workload analysis: %w", err)
	}

	for i := range analysis.Members {
		analysis.Members[i].AvatarURL = h.resolveUserAvatarURL(ctx, analysis.Members[i].AvatarURL)
	}
	for i := range analysis.Risks.OverloadedMembers {
		analysis.Risks.OverloadedMembers[i].AvatarURL = h.resolveUserAvatarURL(ctx, analysis.Risks.OverloadedMembers[i].AvatarURL)
	}
	for i := range analysis.Risks.OverdueMembers {
		analysis.Risks.OverdueMembers[i].AvatarURL = h.resolveUserAvatarURL(ctx, analysis.Risks.OverdueMembers[i].AvatarURL)
	}

	return web.Respond(ctx, w, toAppWorkloadAnalysis(analysis), http.StatusOK)
}

func (h *Handlers) GetPulseReport(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetPulseReport")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseWorkloadAnalysisFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	report, err := h.reports.GetPulseReport(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting pulse report: %w", err)
	}

	for i := range report.Workload.Members {
		report.Workload.Members[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Workload.Members[i].AvatarURL)
	}
	for i := range report.Workload.Risks.OverloadedMembers {
		report.Workload.Risks.OverloadedMembers[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Workload.Risks.OverloadedMembers[i].AvatarURL)
	}
	for i := range report.Workload.Risks.OverdueMembers {
		report.Workload.Risks.OverdueMembers[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Workload.Risks.OverdueMembers[i].AvatarURL)
	}

	return web.Respond(ctx, w, toAppPulseReport(report), http.StatusOK)
}

func (h *Handlers) GetSprintAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetSprintAnalytics")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	analytics, err := h.reports.GetSprintAnalytics(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting sprint analytics: %w", err)
	}

	return web.Respond(ctx, w, toAppSprintAnalyticsWorkspace(analytics), http.StatusOK)
}

func (h *Handlers) GetTimelineTrends(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetTimelineTrends")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	trends, err := h.reports.GetTimelineTrends(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting timeline trends: %w", err)
	}

	return web.Respond(ctx, w, toAppTimelineTrends(trends), http.StatusOK)
}

func (h *Handlers) GetWorkspaceCommandCenterReport(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetWorkspaceCommandCenterReport")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	report, err := h.reports.GetWorkspaceCommandCenterReport(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting workspace command center report: %w", err)
	}
	h.resolveCommandCenterAvatarURLs(ctx, &report)

	return web.Respond(ctx, w, report, http.StatusOK)
}

func (h *Handlers) resolveCommandCenterAvatarURLs(ctx context.Context, report *reports.CoreWorkspaceCommandCenterReport) {
	for i := range report.Teams.MemberContributions {
		report.Teams.MemberContributions[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Teams.MemberContributions[i].AvatarURL)
	}
	for i := range report.Workload.Members {
		report.Workload.Members[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Workload.Members[i].AvatarURL)
	}
	for i := range report.Workload.Risks.OverloadedMembers {
		report.Workload.Risks.OverloadedMembers[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Workload.Risks.OverloadedMembers[i].AvatarURL)
	}
	for i := range report.Workload.Risks.OverdueMembers {
		report.Workload.Risks.OverdueMembers[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Workload.Risks.OverdueMembers[i].AvatarURL)
	}
	for i := range report.Pulse.Workload.Members {
		report.Pulse.Workload.Members[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Pulse.Workload.Members[i].AvatarURL)
	}
	for i := range report.Pulse.Workload.Risks.OverloadedMembers {
		report.Pulse.Workload.Risks.OverloadedMembers[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Pulse.Workload.Risks.OverloadedMembers[i].AvatarURL)
	}
	for i := range report.Pulse.Workload.Risks.OverdueMembers {
		report.Pulse.Workload.Risks.OverdueMembers[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Pulse.Workload.Risks.OverdueMembers[i].AvatarURL)
	}
	for i := range report.Engagement.TopUsers {
		report.Engagement.TopUsers[i].AvatarURL = h.resolveUserAvatarURL(ctx, report.Engagement.TopUsers[i].AvatarURL)
	}
}

func (h *Handlers) TrackWorkspaceAnalyticsEvent(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.TrackWorkspaceAnalyticsEvent")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req AppTrackWorkspaceAnalyticsEventRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	input := reports.CoreWorkspaceAnalyticsEventInput{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		EventName:   req.EventName,
		Surface:     req.Surface,
		TeamID:      req.TeamID,
		StoryID:     req.StoryID,
		ObjectiveID: req.ObjectiveID,
		SprintID:    req.SprintID,
		KeyResultID: req.KeyResultID,
		Properties:  req.Properties,
	}
	if req.OccurredAt != nil {
		input.OccurredAt = *req.OccurredAt
	}

	event, err := h.reports.TrackWorkspaceAnalyticsEvent(ctx, input)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	return web.Respond(ctx, w, AppTrackWorkspaceAnalyticsEventResponse{
		EventName:  event.EventName,
		Surface:    event.Surface,
		OccurredAt: event.OccurredAt,
	}, http.StatusCreated)
}
