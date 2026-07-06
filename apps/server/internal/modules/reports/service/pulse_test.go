package reports

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type pulseRepoStub struct {
	Repository

	workloadResult  CoreWorkloadAnalysis
	storyResult     CorePulseStoryHealth
	sprintResult    CorePulseSprintHealth
	objectiveResult CorePulseObjectiveHealth
	requestResult   CorePulseRequestHealth

	workloadCalls  int
	storyCalls     int
	sprintCalls    int
	objectiveCalls int
	requestCalls   int
	eventInputs    []CoreWorkspaceAnalyticsEventInput
}

func (s *pulseRepoStub) GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkloadAnalysis, error) {
	s.workloadCalls++
	return s.workloadResult, nil
}

func (s *pulseRepoStub) GetPulseStoryHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseStoryHealth, error) {
	s.storyCalls++
	return s.storyResult, nil
}

func (s *pulseRepoStub) GetPulseSprintHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseSprintHealth, error) {
	s.sprintCalls++
	return s.sprintResult, nil
}

func (s *pulseRepoStub) GetPulseObjectiveHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseObjectiveHealth, error) {
	s.objectiveCalls++
	return s.objectiveResult, nil
}

func (s *pulseRepoStub) GetPulseRequestHealth(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CorePulseRequestHealth, error) {
	s.requestCalls++
	return s.requestResult, nil
}

func (s *pulseRepoStub) CreateWorkspaceAnalyticsEvent(ctx context.Context, input CoreWorkspaceAnalyticsEventInput) error {
	s.eventInputs = append(s.eventInputs, input)
	return nil
}

func TestGetPulseReportComposesDeterministicSections(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	repo := &pulseRepoStub{
		workloadResult: CoreWorkloadAnalysis{
			Summary: CoreWorkloadSummary{
				TotalOpenStories:    14,
				OverdueStories:      3,
				HighPriorityStories: 2,
				UrgentStories:       1,
			},
			Members: []CoreMemberWorkload{
				{UserID: uuid.New(), Username: "maya", OpenStories: 9, EstimateTotal: 21},
			},
		},
		storyResult: CorePulseStoryHealth{
			OpenStories:    14,
			BlockedStories: 2,
			OverdueStories: 3,
		},
		sprintResult: CorePulseSprintHealth{
			ActiveSprints: 2,
			AtRiskSprints: 1,
		},
		objectiveResult: CorePulseObjectiveHealth{
			ActiveObjectives: 5,
			AtRiskObjectives: 2,
		},
		requestResult: CorePulseRequestHealth{
			PendingRequests: 4,
			UrgentRequests:  1,
		},
	}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)

	got, err := service.GetPulseReport(context.Background(), workspaceID, ReportFilters{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.workloadCalls != 1 || repo.storyCalls != 1 || repo.sprintCalls != 1 || repo.objectiveCalls != 1 || repo.requestCalls != 1 {
		t.Fatalf("expected each pulse dependency to be called once, got workload=%d story=%d sprint=%d objective=%d request=%d", repo.workloadCalls, repo.storyCalls, repo.sprintCalls, repo.objectiveCalls, repo.requestCalls)
	}
	if got.Summary.OpenStories != 14 || got.Summary.AtRiskSprints != 1 || got.Summary.PendingRequests != 4 {
		t.Fatalf("expected summary to include source counts, got %#v", got.Summary)
	}
	if got.Workload.Risks.HighPriorityStories != 3 {
		t.Fatalf("expected workload risks to include urgent and high priority stories, got %#v", got.Workload.Risks)
	}
}

func TestGetPulseReportBuildsOrderedRiskCards(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	repo := &pulseRepoStub{
		workloadResult: CoreWorkloadAnalysis{
			Summary: CoreWorkloadSummary{
				UnassignedStories: 5,
			},
			Members: []CoreMemberWorkload{
				{UserID: uuid.New(), Username: "maya", OpenStories: 9, EstimateTotal: 21},
			},
		},
		storyResult: CorePulseStoryHealth{
			OverdueStories: 3,
			BlockedStories: 2,
		},
		sprintResult: CorePulseSprintHealth{
			AtRiskSprints: 1,
		},
		objectiveResult: CorePulseObjectiveHealth{
			AtRiskObjectives: 2,
		},
		requestResult: CorePulseRequestHealth{
			PendingRequests: 4,
		},
	}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)

	got, err := service.GetPulseReport(context.Background(), workspaceID, ReportFilters{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got.Risks) == 0 {
		t.Fatal("expected risk cards")
	}
	first := got.Risks[0]
	if first.Kind != PulseRiskKindOverdueStories || first.Severity != PulseRiskSeverityHigh || first.Count != 3 {
		t.Fatalf("expected overdue stories to be the first high risk, got %#v", first)
	}
}

func TestTrackWorkspaceAnalyticsEventNormalizesSafeInput(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &pulseRepoStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)

	got, err := service.TrackWorkspaceAnalyticsEvent(context.Background(), CoreWorkspaceAnalyticsEventInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		EventName:   " analytics_report_viewed ",
		Surface:     " analytics ",
		Properties: map[string]any{
			"tab":        "pulse",
			"emptyValue": "",
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.EventName != "analytics_report_viewed" {
		t.Fatalf("expected normalized event name, got %q", got.EventName)
	}
	if got.Surface != "analytics" {
		t.Fatalf("expected normalized surface, got %q", got.Surface)
	}
	if len(repo.eventInputs) != 1 {
		t.Fatalf("expected one repository call, got %d", len(repo.eventInputs))
	}
	if repo.eventInputs[0].WorkspaceID != workspaceID || repo.eventInputs[0].UserID != userID {
		t.Fatalf("expected workspace/user ids to be preserved, got %#v", repo.eventInputs[0])
	}
}

func TestTrackWorkspaceAnalyticsEventRejectsBlankEventName(t *testing.T) {
	t.Parallel()

	repo := &pulseRepoStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)

	_, err := service.TrackWorkspaceAnalyticsEvent(context.Background(), CoreWorkspaceAnalyticsEventInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
		EventName:   " ",
	})

	if err == nil {
		t.Fatal("expected blank event name to fail")
	}
	if len(repo.eventInputs) != 0 {
		t.Fatalf("expected no repository calls, got %d", len(repo.eventInputs))
	}
}
