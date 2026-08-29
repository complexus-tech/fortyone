package messaging

import (
	"context"
	"encoding/json"
	"testing"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

func TestFortyOneToolExecutorWorkloadSummaryReportsOverloadedMembers(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	team := teams.CoreTeam{
		ID:        uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001"),
		Name:      "Product",
		Code:      "PRO",
		Workspace: scope.WorkspaceID,
	}
	scope.SharedTeamIDs = []uuid.UUID{team.ID}
	reportsReader := &workloadServiceStub{analysis: reports.CoreWorkloadAnalysis{
		Summary: reports.CoreWorkloadSummary{TotalOpenStories: 11, TotalEstimate: 45, OverdueStories: 2, UnassignedStories: 1, UnestimatedStories: 3},
		Members: []reports.CoreMemberWorkload{
			{FullName: "John Smith", Username: "john", OpenStories: 8, EstimateTotal: 12, StartedStories: 4},
			{FullName: "Jane Doe", Username: "jane", OpenStories: 3, EstimateTotal: 25, StartedStories: 2},
		},
		Risks: reports.CoreWorkloadRisks{OverloadedMembers: []reports.CoreMemberWorkload{
			{FullName: "John Smith", Username: "john", OpenStories: 8, EstimateTotal: 12, StartedStories: 4},
			{FullName: "Jane Doe", Username: "jane", OpenStories: 3, EstimateTotal: 25, StartedStories: 2},
		}},
	}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		&storiesServiceStub{},
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States:   &statesServiceStub{},
			Users:    &usersServiceStub{},
			Workload: reportsReader,
		}),
	)

	raw, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolGetWorkloadSummary,
		Arguments: json.RawMessage(`{"team_name":"Product","member_name":"John","limit":null}`),
	})
	if err != nil {
		t.Fatalf("execute workload summary: %v", err)
	}
	var result workloadSummaryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode workload summary: %v", err)
	}
	if result.Access != "granted" || result.OverloadedCount != 2 || result.Member == nil || !result.Member.Overloaded || result.Member.Name != "John Smith" {
		t.Fatalf("unexpected workload result: %#v", result)
	}
	if len(reportsReader.filters.TeamIDs) != 1 || reportsReader.filters.TeamIDs[0] != team.ID {
		t.Fatalf("workload query was not scoped to the selected shared team: %#v", reportsReader.filters)
	}
}

func TestFortyOneToolExecutorWorkloadSummaryDeniesUnsharedTeams(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	team := teams.CoreTeam{ID: uuid.New(), Name: "Product", Code: "PRO", Workspace: scope.WorkspaceID}
	reportsReader := &workloadServiceStub{}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		&storiesServiceStub{},
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States:   &statesServiceStub{},
			Users:    &usersServiceStub{},
			Workload: reportsReader,
		}),
	)

	raw, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolGetWorkloadSummary,
		Arguments: json.RawMessage(`{"team_name":"Product","member_name":null,"limit":null}`),
	})
	if err != nil {
		t.Fatalf("execute denied workload summary: %v", err)
	}
	var result workloadSummaryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode denied workload summary: %v", err)
	}
	if result.Access != "denied" || result.AccessReason != "shared_team_scope_required" || reportsReader.called {
		t.Fatalf("expected denied workload access without a report query: %#v", result)
	}
}

type workloadServiceStub struct {
	analysis reports.CoreWorkloadAnalysis
	filters  reports.ReportFilters
	called   bool
}

func (s *workloadServiceStub) GetWorkloadAnalysis(_ context.Context, _ uuid.UUID, filters reports.ReportFilters) (reports.CoreWorkloadAnalysis, error) {
	s.called = true
	s.filters = filters
	return s.analysis, nil
}
