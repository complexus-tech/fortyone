package reportshttp

import (
	"context"
	"fmt"
	"net/http"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *Handlers) GetWorkspaceOverview(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetWorkspaceOverview")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	filters, err := parseReportFilters(r.URL.Query(), h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	overview, err := h.reports.GetWorkspaceOverview(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting workspace overview: %w", err))
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

	filters, err := parseReportFilters(r.URL.Query(), h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	analytics, err := h.reports.GetStoryAnalytics(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting story analytics: %w", err))
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

	filters, err := parseReportFilters(r.URL.Query(), h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	progress, err := h.reports.GetObjectiveProgress(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting objective progress: %w", err))
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

	filters, err := parseReportFilters(r.URL.Query(), h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	performance, err := h.reports.GetTeamPerformance(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting team performance: %w", err))
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

	filters, err := parseWorkloadAnalysisFilters(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	analysis, err := h.reports.GetWorkloadAnalysis(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting workload analysis: %w", err))
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

	filters, err := parseWorkloadAnalysisFilters(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	report, err := h.reports.GetPulseReport(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting pulse report: %w", err))
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

	filters, err := parseReportFilters(r.URL.Query(), h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	analytics, err := h.reports.GetSprintAnalytics(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting sprint analytics: %w", err))
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

	filters, err := parseReportFilters(r.URL.Query(), h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	trends, err := h.reports.GetTimelineTrends(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting timeline trends: %w", err))
	}

	return web.Respond(ctx, w, toAppTimelineTrends(trends), http.StatusOK)
}
