package reports

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const pulseRiskLimit = 8

func (s *Service) GetPulseReport(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseReport, error) {
	s.log.Info(ctx, "business.core.reports.GetPulseReport")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetPulseReport")
	defer span.End()

	workload, err := s.GetWorkloadAnalysis(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CorePulseReport{}, fmt.Errorf("getting pulse workload: %w", err)
	}

	stories, err := s.repo.GetPulseStoryHealth(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CorePulseReport{}, fmt.Errorf("getting pulse story health: %w", err)
	}

	sprints, err := s.repo.GetPulseSprintHealth(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CorePulseReport{}, fmt.Errorf("getting pulse sprint health: %w", err)
	}

	objectives, err := s.repo.GetPulseObjectiveHealth(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CorePulseReport{}, fmt.Errorf("getting pulse objective health: %w", err)
	}

	requests, err := s.repo.GetPulseRequestHealth(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CorePulseReport{}, fmt.Errorf("getting pulse request health: %w", err)
	}

	summary := CorePulseSummary{
		OpenStories:       workload.Summary.TotalOpenStories,
		OverdueStories:    stories.OverdueStories,
		BlockedStories:    stories.BlockedStories,
		AtRiskSprints:     sprints.AtRiskSprints,
		AtRiskObjectives:  objectives.AtRiskObjectives + objectives.OffTrackObjectives,
		PendingRequests:   requests.PendingRequests,
		OverloadedMembers: len(workload.Risks.OverloadedMembers),
	}

	return CorePulseReport{
		WorkspaceID: workspaceID,
		ReportDate:  time.Now().UTC(),
		Filters:     filters,
		Summary:     summary,
		Stories:     stories,
		Sprints:     sprints,
		Objectives:  objectives,
		Requests:    requests,
		Workload:    workload,
		Risks:       derivePulseRisks(stories, sprints, objectives, requests, workload),
	}, nil
}

func derivePulseRisks(stories CorePulseStoryHealth, sprints CorePulseSprintHealth, objectives CorePulseObjectiveHealth, requests CorePulseRequestHealth, workload CoreWorkloadAnalysis) []CorePulseRisk {
	risks := make([]CorePulseRisk, 0, 7)

	addRisk := func(kind PulseRiskKind, severity PulseRiskSeverity, title string, description string, count int) {
		if count <= 0 {
			return
		}
		risks = append(risks, CorePulseRisk{
			Kind:        kind,
			Severity:    severity,
			Title:       title,
			Description: description,
			Count:       count,
		})
	}

	addRisk(PulseRiskKindOverdueStories, PulseRiskSeverityHigh, "Overdue stories", "Stories have passed their end date and are still open.", stories.OverdueStories)
	addRisk(PulseRiskKindBlockedStories, PulseRiskSeverityHigh, "Blocked stories", "Stories are linked to blockers and may need intervention.", stories.BlockedStories)
	addRisk(PulseRiskKindOverloadedMembers, PulseRiskSeverityHigh, "Overloaded members", "Members exceed the current open-story or estimate workload threshold.", len(workload.Risks.OverloadedMembers))
	addRisk(PulseRiskKindAtRiskSprints, PulseRiskSeverityMedium, "At-risk sprints", "Active sprints have overdue or incomplete work close to the end date.", sprints.AtRiskSprints)
	addRisk(PulseRiskKindAtRiskObjectives, PulseRiskSeverityMedium, "At-risk objectives", "Objectives are marked at risk or off track.", objectives.AtRiskObjectives+objectives.OffTrackObjectives)
	addRisk(PulseRiskKindPendingRequests, PulseRiskSeverityMedium, "Pending requests", "Integration requests are waiting to be accepted or declined.", requests.PendingRequests)
	addRisk(PulseRiskKindUnassignedStories, PulseRiskSeverityLow, "Unassigned stories", "Open stories do not have an owner yet.", workload.Risks.UnassignedStories)

	sort.SliceStable(risks, func(i, j int) bool {
		leftRank := pulseSeverityRank(risks[i].Severity)
		rightRank := pulseSeverityRank(risks[j].Severity)
		if leftRank == rightRank {
			return risks[i].Count > risks[j].Count
		}
		return leftRank < rightRank
	})

	if len(risks) > pulseRiskLimit {
		return risks[:pulseRiskLimit]
	}
	return risks
}

func pulseSeverityRank(severity PulseRiskSeverity) int {
	switch severity {
	case PulseRiskSeverityHigh:
		return 0
	case PulseRiskSeverityMedium:
		return 1
	case PulseRiskSeverityLow:
		return 2
	default:
		return 3
	}
}
