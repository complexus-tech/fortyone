package reports

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type commandCenterRepoStub struct {
	Repository

	overviewResult   CoreWorkspaceOverview
	storyResult      CoreStoryAnalytics
	objectiveResult  CoreObjectiveProgress
	teamResult       CoreTeamPerformance
	workloadResult   CoreWorkloadAnalysis
	sprintResult     CoreSprintAnalyticsWorkspace
	trendResult      CoreTimelineTrends
	requestResult    CoreRequestSourceAnalytics
	engagementResult CoreWorkspaceEngagementAnalytics
	storyErr         error
}

func (s *commandCenterRepoStub) GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkspaceOverview, error) {
	return s.overviewResult, nil
}

func (s *commandCenterRepoStub) GetStoryAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreStoryAnalytics, error) {
	if s.storyErr != nil {
		return CoreStoryAnalytics{}, s.storyErr
	}
	return s.storyResult, nil
}

func (s *commandCenterRepoStub) GetObjectiveProgress(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreObjectiveProgress, error) {
	return s.objectiveResult, nil
}

func (s *commandCenterRepoStub) GetTeamPerformance(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreTeamPerformance, error) {
	return s.teamResult, nil
}

func (s *commandCenterRepoStub) GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkloadAnalysis, error) {
	return s.workloadResult, nil
}

func (s *commandCenterRepoStub) GetPulseStoryHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseStoryHealth, error) {
	return CorePulseStoryHealth{
		OpenStories:    s.workloadResult.Summary.TotalOpenStories,
		OverdueStories: s.workloadResult.Summary.OverdueStories,
	}, nil
}

func (s *commandCenterRepoStub) GetPulseSprintHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseSprintHealth, error) {
	return CorePulseSprintHealth{}, nil
}

func (s *commandCenterRepoStub) GetPulseObjectiveHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseObjectiveHealth, error) {
	return CorePulseObjectiveHealth{}, nil
}

func (s *commandCenterRepoStub) GetPulseRequestHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseRequestHealth, error) {
	return CorePulseRequestHealth{}, nil
}

func (s *commandCenterRepoStub) GetSprintAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreSprintAnalyticsWorkspace, error) {
	return s.sprintResult, nil
}

func (s *commandCenterRepoStub) GetTimelineTrends(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreTimelineTrends, error) {
	return s.trendResult, nil
}

func (s *commandCenterRepoStub) GetRequestSourceAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreRequestSourceAnalytics, error) {
	return s.requestResult, nil
}

func (s *commandCenterRepoStub) GetWorkspaceEngagementAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkspaceEngagementAnalytics, error) {
	return s.engagementResult, nil
}

func (s *commandCenterRepoStub) CreateWorkspaceAnalyticsEvent(ctx context.Context, input CoreWorkspaceAnalyticsEventInput) error {
	return nil
}

func TestGetWorkspaceCommandCenterReportComposesDetailedSections(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	startDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
	repo := &commandCenterRepoStub{
		overviewResult: CoreWorkspaceOverview{
			WorkspaceID: workspaceID,
			Metrics: CoreWorkspaceMetrics{
				TotalStories:     42,
				CompletedStories: 18,
			},
		},
		workloadResult: CoreWorkloadAnalysis{
			Summary: CoreWorkloadSummary{
				TotalOpenStories: 24,
				OverdueStories:   5,
			},
		},
		requestResult: CoreRequestSourceAnalytics{
			Providers: []CoreRequestProviderPerformance{
				{Provider: "github", TotalRequests: 12, AcceptedRequests: 8},
			},
		},
		engagementResult: CoreWorkspaceEngagementAnalytics{
			TotalEvents: 27,
			UniqueUsers: 4,
		},
	}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)

	got, err := service.GetWorkspaceCommandCenterReport(context.Background(), workspaceID, ReportFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.WorkspaceID != workspaceID {
		t.Fatalf("expected workspace id %s, got %s", workspaceID, got.WorkspaceID)
	}
	if got.Overview.Metrics.TotalStories != 42 {
		t.Fatalf("expected overview metrics to be included, got %#v", got.Overview.Metrics)
	}
	if got.Pulse.Summary.OpenStories != 24 || got.Pulse.Summary.OverdueStories != 5 {
		t.Fatalf("expected pulse summary to reflect workload, got %#v", got.Pulse.Summary)
	}
	if got.Requests.Providers[0].Provider != "github" {
		t.Fatalf("expected request source analytics, got %#v", got.Requests)
	}
	if got.Engagement.TotalEvents != 27 || got.Engagement.UniqueUsers != 4 {
		t.Fatalf("expected engagement analytics, got %#v", got.Engagement)
	}
}

func TestGetWorkspaceCommandCenterReportKeepsSectionFailuresPartial(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	repo := &commandCenterRepoStub{
		storyErr: errors.New("story analytics unavailable"),
		workloadResult: CoreWorkloadAnalysis{
			Summary: CoreWorkloadSummary{
				TotalOpenStories: 7,
			},
			Members: []CoreMemberWorkload{},
			Teams:   []CoreTeamWorkloadSummary{},
			Risks: CoreWorkloadRisks{
				OverloadedMembers: []CoreMemberWorkload{},
				OverdueMembers:    []CoreMemberWorkload{},
			},
		},
	}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)

	got, err := service.GetWorkspaceCommandCenterReport(context.Background(), workspaceID, ReportFilters{})

	if err != nil {
		t.Fatalf("expected partial report, got error %v", err)
	}
	if len(got.SectionErrors) != 1 {
		t.Fatalf("expected one section error, got %#v", got.SectionErrors)
	}
	if got.SectionErrors[0].Section != "stories" {
		t.Fatalf("expected stories section error, got %#v", got.SectionErrors[0])
	}
	if got.Stories.StatusBreakdown == nil {
		t.Fatal("expected failed stories section to keep empty slices, got nil")
	}
	if got.Pulse.Summary.OpenStories != 7 {
		t.Fatalf("expected pulse summary to use workload data, got %#v", got.Pulse.Summary)
	}
}
