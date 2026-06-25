package reports

import (
	"context"
	"sync"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (s *Service) GetWorkspaceCommandCenterReport(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkspaceCommandCenterReport, error) {
	s.log.Info(ctx, "business.core.reports.GetWorkspaceCommandCenterReport")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetWorkspaceCommandCenterReport")
	defer span.End()

	reportDate := time.Now().UTC()
	overview := emptyWorkspaceOverview(workspaceID, reportDate, filters)
	stories := emptyStoryAnalytics()
	objectives := emptyObjectiveProgress()
	teams := emptyTeamPerformance()
	workload := emptyWorkloadAnalysis()
	sprints := emptySprintAnalytics()
	trends := emptyTimelineTrends()
	requests := CoreRequestSourceAnalytics{Providers: []CoreRequestProviderPerformance{}}
	engagement := emptyWorkspaceEngagementAnalytics()
	pulseStories := CorePulseStoryHealth{}
	pulseSprints := CorePulseSprintHealth{}
	pulseObjectives := CorePulseObjectiveHealth{}
	pulseRequests := CorePulseRequestHealth{}

	var errorsMu sync.Mutex
	sectionErrors := []CoreWorkspaceCommandCenterSectionError{}
	recordSectionError := func(section string, err error) {
		if err == nil {
			return
		}

		span.RecordError(err)
		s.log.Error(ctx, "failed to build command center section", "section", section, "error", err)

		errorsMu.Lock()
		defer errorsMu.Unlock()
		sectionErrors = append(sectionErrors, CoreWorkspaceCommandCenterSectionError{
			Section: section,
			Message: err.Error(),
		})
	}

	var wg sync.WaitGroup
	run := func(section string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			recordSectionError(section, fn())
		}()
	}

	run("overview", func() error {
		result, err := s.GetWorkspaceOverview(ctx, workspaceID, filters)
		if err == nil {
			overview = result
		}
		return err
	})
	run("stories", func() error {
		result, err := s.GetStoryAnalytics(ctx, workspaceID, filters)
		if err == nil {
			stories = result
		}
		return err
	})
	run("objectives", func() error {
		result, err := s.GetObjectiveProgress(ctx, workspaceID, filters)
		if err == nil {
			objectives = result
		}
		return err
	})
	run("teams", func() error {
		result, err := s.GetTeamPerformance(ctx, workspaceID, filters)
		if err == nil {
			teams = result
		}
		return err
	})
	run("workload", func() error {
		result, err := s.GetWorkloadAnalysis(ctx, workspaceID, filters)
		if err == nil {
			workload = result
		}
		return err
	})
	run("sprints", func() error {
		result, err := s.GetSprintAnalytics(ctx, workspaceID, filters)
		if err == nil {
			sprints = result
		}
		return err
	})
	run("trends", func() error {
		result, err := s.GetTimelineTrends(ctx, workspaceID, filters)
		if err == nil {
			trends = result
		}
		return err
	})
	run("requests", func() error {
		result, err := s.repo.GetRequestSourceAnalytics(ctx, workspaceID, filters)
		if err == nil {
			requests = result
		}
		return err
	})
	run("engagement", func() error {
		result, err := s.repo.GetWorkspaceEngagementAnalytics(ctx, workspaceID, filters)
		if err == nil {
			engagement = result
		}
		return err
	})
	run("pulse_stories", func() error {
		result, err := s.repo.GetPulseStoryHealth(ctx, workspaceID, filters)
		if err == nil {
			pulseStories = result
		}
		return err
	})
	run("pulse_sprints", func() error {
		result, err := s.repo.GetPulseSprintHealth(ctx, workspaceID, filters)
		if err == nil {
			pulseSprints = result
		}
		return err
	})
	run("pulse_objectives", func() error {
		result, err := s.repo.GetPulseObjectiveHealth(ctx, workspaceID, filters)
		if err == nil {
			pulseObjectives = result
		}
		return err
	})
	run("pulse_requests", func() error {
		result, err := s.repo.GetPulseRequestHealth(ctx, workspaceID, filters)
		if err == nil {
			pulseRequests = result
		}
		return err
	})

	wg.Wait()
	if err := ctx.Err(); err != nil {
		span.RecordError(err)
		return CoreWorkspaceCommandCenterReport{}, err
	}

	pulse := CorePulseReport{
		WorkspaceID: workspaceID,
		ReportDate:  reportDate,
		Filters:     filters,
		Summary: CorePulseSummary{
			OpenStories:       workload.Summary.TotalOpenStories,
			OverdueStories:    pulseStories.OverdueStories,
			BlockedStories:    pulseStories.BlockedStories,
			AtRiskSprints:     pulseSprints.AtRiskSprints,
			AtRiskObjectives:  pulseObjectives.AtRiskObjectives + pulseObjectives.OffTrackObjectives,
			PendingRequests:   pulseRequests.PendingRequests,
			OverloadedMembers: len(workload.Risks.OverloadedMembers),
		},
		Stories:    pulseStories,
		Sprints:    pulseSprints,
		Objectives: pulseObjectives,
		Requests:   pulseRequests,
		Workload:   workload,
		Risks:      derivePulseRisks(pulseStories, pulseSprints, pulseObjectives, pulseRequests, workload),
	}

	return CoreWorkspaceCommandCenterReport{
		WorkspaceID:   workspaceID,
		ReportDate:    reportDate,
		Filters:       filters,
		SectionErrors: sectionErrors,
		Overview:      overview,
		Pulse:         pulse,
		Stories:       stories,
		Objectives:    objectives,
		Teams:         teams,
		Workload:      workload,
		Sprints:       sprints,
		Trends:        trends,
		Requests:      requests,
		Engagement:    engagement,
	}, nil
}

func emptyWorkspaceOverview(workspaceID uuid.UUID, reportDate time.Time, filters ReportFilters) CoreWorkspaceOverview {
	return CoreWorkspaceOverview{
		WorkspaceID:     workspaceID,
		ReportDate:      reportDate,
		Filters:         filters,
		CompletionTrend: []CoreCompletionTrendPoint{},
		VelocityTrend:   []CoreVelocityTrendPoint{},
	}
}

func emptyStoryAnalytics() CoreStoryAnalytics {
	return CoreStoryAnalytics{
		StatusBreakdown:      []CoreStatusBreakdownItem{},
		PriorityDistribution: []CorePriorityDistributionItem{},
		CompletionByTeam:     []CoreTeamCompletionItem{},
		Burndown:             []CoreBurndownPoint{},
	}
}

func emptyObjectiveProgress() CoreObjectiveProgress {
	return CoreObjectiveProgress{
		HealthDistribution: []CoreHealthDistributionItem{},
		StatusBreakdown:    []CoreObjectiveStatusItem{},
		KeyResultsProgress: []CoreKeyResultProgressItem{},
		ProgressByTeam:     []CoreObjectiveTeamProgressItem{},
	}
}

func emptyTeamPerformance() CoreTeamPerformance {
	return CoreTeamPerformance{
		TeamWorkload:        []CoreTeamWorkloadItem{},
		MemberContributions: []CoreMemberContributionItem{},
		VelocityByTeam:      []CoreTeamVelocityItem{},
		WorkloadTrend:       []CoreWorkloadTrendPoint{},
	}
}

func emptyWorkloadAnalysis() CoreWorkloadAnalysis {
	return CoreWorkloadAnalysis{
		Members: []CoreMemberWorkload{},
		Teams:   []CoreTeamWorkloadSummary{},
		Risks: CoreWorkloadRisks{
			OverloadedMembers: []CoreMemberWorkload{},
			OverdueMembers:    []CoreMemberWorkload{},
		},
	}
}

func emptySprintAnalytics() CoreSprintAnalyticsWorkspace {
	return CoreSprintAnalyticsWorkspace{
		SprintProgress:   []CoreSprintProgressItem{},
		CombinedBurndown: []CoreCombinedBurndownPoint{},
		TeamAllocation:   []CoreSprintTeamAllocation{},
		SprintHealth:     []CoreSprintHealthItem{},
	}
}

func emptyTimelineTrends() CoreTimelineTrends {
	return CoreTimelineTrends{
		StoryCompletion:   []CoreStoryCompletionPoint{},
		ObjectiveProgress: []CoreObjectiveProgressPoint{},
		TeamVelocity:      []CoreTeamVelocityPoint{},
		KeyMetricsTrend:   []CoreKeyMetricsTrendPoint{},
	}
}

func emptyWorkspaceEngagementAnalytics() CoreWorkspaceEngagementAnalytics {
	return CoreWorkspaceEngagementAnalytics{
		EventsByName:    []CoreWorkspaceEngagementCount{},
		EventsBySurface: []CoreWorkspaceEngagementCount{},
		TopUsers:        []CoreWorkspaceEngagementUser{},
	}
}
