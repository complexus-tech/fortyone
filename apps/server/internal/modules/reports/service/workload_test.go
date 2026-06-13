package reports

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type workloadRepoStub struct {
	Repository

	gotWorkspaceID uuid.UUID
	gotFilters     ReportFilters
	result         CoreWorkloadAnalysis
}

func (s *workloadRepoStub) GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkloadAnalysis, error) {
	s.gotWorkspaceID = workspaceID
	s.gotFilters = filters

	return s.result, nil
}

func TestGetWorkloadAnalysisDelegatesToRepository(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.June, 13, 0, 0, 0, 0, time.UTC)

	expected := CoreWorkloadAnalysis{
		Summary: CoreWorkloadSummary{
			TotalOpenStories: 3,
			TotalEstimate:    13,
		},
		Members: []CoreMemberWorkload{
			{
				UserID:        uuid.New(),
				Username:      "maya",
				OpenStories:   2,
				EstimateTotal: 8,
			},
		},
	}
	repo := &workloadRepoStub{result: expected}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)

	got, err := service.GetWorkloadAnalysis(context.Background(), workspaceID, ReportFilters{
		TeamIDs:   []uuid.UUID{teamID},
		StartDate: &startDate,
		EndDate:   &endDate,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.gotWorkspaceID != workspaceID {
		t.Fatalf("expected workspace %s, got %s", workspaceID, repo.gotWorkspaceID)
	}
	if len(repo.gotFilters.TeamIDs) != 1 || repo.gotFilters.TeamIDs[0] != teamID {
		t.Fatalf("expected team filter %s, got %#v", teamID, repo.gotFilters.TeamIDs)
	}
	if got.Summary.TotalOpenStories != expected.Summary.TotalOpenStories {
		t.Fatalf("expected %d open stories, got %d", expected.Summary.TotalOpenStories, got.Summary.TotalOpenStories)
	}
	if len(got.Members) != 1 || got.Members[0].EstimateTotal != expected.Members[0].EstimateTotal {
		t.Fatalf("expected member workload estimate %d, got %#v", expected.Members[0].EstimateTotal, got.Members)
	}
}

func TestGetWorkloadAnalysisDerivesRiskSections(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	overloadedMember := CoreMemberWorkload{
		UserID:        uuid.New(),
		Username:      "maya",
		OpenStories:   9,
		EstimateTotal: 21,
	}
	overdueMember := CoreMemberWorkload{
		UserID:         uuid.New(),
		Username:       "noah",
		OpenStories:    2,
		OverdueStories: 1,
	}
	repo := &workloadRepoStub{
		result: CoreWorkloadAnalysis{
			Summary: CoreWorkloadSummary{
				UnassignedStories:   4,
				UnestimatedStories:  3,
				HighPriorityStories: 2,
			},
			Members: []CoreMemberWorkload{overdueMember, overloadedMember},
		},
	}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-test"), repo)

	got, err := service.GetWorkloadAnalysis(context.Background(), workspaceID, ReportFilters{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got.Risks.OverloadedMembers) != 1 || got.Risks.OverloadedMembers[0].UserID != overloadedMember.UserID {
		t.Fatalf("expected overloaded member %s, got %#v", overloadedMember.UserID, got.Risks.OverloadedMembers)
	}
	if len(got.Risks.OverdueMembers) != 1 || got.Risks.OverdueMembers[0].UserID != overdueMember.UserID {
		t.Fatalf("expected overdue member %s, got %#v", overdueMember.UserID, got.Risks.OverdueMembers)
	}
	if got.Risks.UnassignedStories != 4 || got.Risks.UnestimatedStories != 3 || got.Risks.HighPriorityStories != 2 {
		t.Fatalf("expected summary risks to mirror summary, got %#v", got.Risks)
	}
}
