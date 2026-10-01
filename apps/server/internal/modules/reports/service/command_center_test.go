package reports

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"sync"
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
	filtersMu        sync.Mutex
	sectionFilters   map[string]ReportFilters
}

func (s *commandCenterRepoStub) recordFilters(section string, filters ReportFilters) {
	s.filtersMu.Lock()
	defer s.filtersMu.Unlock()
	if s.sectionFilters == nil {
		s.sectionFilters = make(map[string]ReportFilters)
	}
	s.sectionFilters[section] = filters
}

func (s *commandCenterRepoStub) GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkspaceOverview, error) {
	s.recordFilters("overview", filters)
	return s.overviewResult, nil
}

func (s *commandCenterRepoStub) GetStoryAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreStoryAnalytics, error) {
	s.recordFilters("stories", filters)
	if s.storyErr != nil {
		return CoreStoryAnalytics{}, s.storyErr
	}
	return s.storyResult, nil
}

func (s *commandCenterRepoStub) GetObjectiveProgress(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreObjectiveProgress, error) {
	s.recordFilters("objectives", filters)
	return s.objectiveResult, nil
}

func (s *commandCenterRepoStub) GetTeamPerformance(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreTeamPerformance, error) {
	s.recordFilters("teams", filters)
	return s.teamResult, nil
}

func (s *commandCenterRepoStub) GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkloadAnalysis, error) {
	s.recordFilters("workload", filters)
	return s.workloadResult, nil
}

func (s *commandCenterRepoStub) GetPulseStoryHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseStoryHealth, error) {
	s.recordFilters("pulse_stories", filters)
	return CorePulseStoryHealth{
		OpenStories:    s.workloadResult.Summary.TotalOpenStories,
		OverdueStories: s.workloadResult.Summary.OverdueStories,
	}, nil
}

func (s *commandCenterRepoStub) GetPulseSprintHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseSprintHealth, error) {
	s.recordFilters("pulse_sprints", filters)
	return CorePulseSprintHealth{}, nil
}

func (s *commandCenterRepoStub) GetPulseObjectiveHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseObjectiveHealth, error) {
	s.recordFilters("pulse_objectives", filters)
	return CorePulseObjectiveHealth{}, nil
}

func (s *commandCenterRepoStub) GetPulseRequestHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseRequestHealth, error) {
	s.recordFilters("pulse_requests", filters)
	return CorePulseRequestHealth{}, nil
}

func (s *commandCenterRepoStub) GetSprintAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreSprintAnalyticsWorkspace, error) {
	s.recordFilters("sprints", filters)
	return s.sprintResult, nil
}

func (s *commandCenterRepoStub) GetTimelineTrends(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreTimelineTrends, error) {
	s.recordFilters("trends", filters)
	return s.trendResult, nil
}

func (s *commandCenterRepoStub) GetRequestSourceAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreRequestSourceAnalytics, error) {
	s.recordFilters("requests", filters)
	return s.requestResult, nil
}

func (s *commandCenterRepoStub) GetWorkspaceEngagementAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkspaceEngagementAnalytics, error) {
	s.recordFilters("engagement", filters)
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

	got, err := service.GetWorkspaceCommandCenterReport(reportTestContext(workspaceID), workspaceID, ReportFilters{
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

func TestGetWorkspaceCommandCenterSeparatesCurrentRiskFromReportingPeriod(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actorID := uuid.New()
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 7)
	periodFilters := ReportFilters{
		ActorID:      actorID,
		TeamIDs:      []uuid.UUID{uuid.New()},
		AssigneeIDs:  []uuid.UUID{uuid.New()},
		ObjectiveIDs: []uuid.UUID{uuid.New()},
		SprintIDs:    []uuid.UUID{uuid.New()},
		StartDate:    &startDate,
		EndDate:      &endDate,
	}
	repo := &commandCenterRepoStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)
	report, err := service.GetWorkspaceCommandCenterReport(reportTestContext(actorID), workspaceID, periodFilters)
	if err != nil {
		t.Fatalf("get command center: %v", err)
	}

	currentFilters := periodFilters
	currentFilters.StartDate = nil
	currentFilters.EndDate = nil
	for _, section := range []string{"workload", "pulse_stories", "pulse_sprints", "pulse_objectives", "pulse_requests"} {
		if !reflect.DeepEqual(repo.sectionFilters[section], currentFilters) {
			t.Errorf("%s scope = %#v, want current work with unchanged item/access filters", section, repo.sectionFilters[section])
		}
	}
	for _, section := range []string{"overview", "stories", "objectives", "teams", "sprints", "trends", "requests", "engagement"} {
		if !reflect.DeepEqual(repo.sectionFilters[section], periodFilters) {
			t.Errorf("%s scope = %#v, want selected reporting period", section, repo.sectionFilters[section])
		}
	}
	if !reflect.DeepEqual(report.Filters, periodFilters) || !reflect.DeepEqual(report.Pulse.Filters, currentFilters) {
		t.Fatalf("response filter metadata does not describe each report's actual scope: period=%#v pulse=%#v", report.Filters, report.Pulse.Filters)
	}
}

func TestGetWorkspaceCommandCenterReportKeepsSectionFailuresPartial(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	startDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
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

	got, err := service.GetWorkspaceCommandCenterReport(reportTestContext(workspaceID), workspaceID, ReportFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})

	if err != nil {
		t.Fatalf("expected partial report, got error %v", err)
	}
	if len(got.SectionErrors) != 1 {
		t.Fatalf("expected one section error, got %#v", got.SectionErrors)
	}
	if got.SectionErrors[0].Section != "stories" {
		t.Fatalf("expected stories section error, got %#v", got.SectionErrors[0])
	}
	if got.SectionErrors[0].Message != commandCenterSectionUnavailable {
		t.Fatalf("expected sanitized section error, got %#v", got.SectionErrors[0])
	}
	if got.Stories.StatusBreakdown == nil {
		t.Fatal("expected failed stories section to keep empty slices, got nil")
	}
	if got.Pulse.Summary.OpenStories != 7 {
		t.Fatalf("expected pulse summary to use workload data, got %#v", got.Pulse.Summary)
	}
}

func TestGetWorkspaceCommandCenterReportFailsClosedOnAccessDenied(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	startDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), &commandCenterRepoStub{
		storyErr: ErrReportsAccessDenied,
	})

	_, err := service.GetWorkspaceCommandCenterReport(reportTestContext(workspaceID), workspaceID, ReportFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if !errors.Is(err, ErrReportsAccessDenied) {
		t.Fatalf("command center error = %v, want access denied", err)
	}
}
