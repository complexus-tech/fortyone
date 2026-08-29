package reportshttp

import (
	"context"
	"fmt"
	"net/http"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *Handlers) GetWorkspaceCommandCenterReport(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetWorkspaceCommandCenterReport")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	filters, err := parseReportFilters(r.URL.Query(), h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	report, err := h.reports.GetWorkspaceCommandCenterReport(ctx, workspace.ID, filters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting workspace command center report: %w", err))
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
