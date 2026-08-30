/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { State } from "@/types/states";
import {
  buildFilterSignature,
  buildObjectiveProgressChartData,
  buildProviderChartData,
  buildStatusBreakdownData,
  chartPalette,
} from "./model";

describe("command center report model", () => {
  it("creates a stable signature for every supported filter", () => {
    expect(
      buildFilterSignature({
        assigneeIds: ["member-1", "member-2"],
        endDate: "2026-08-30",
        objectiveIds: ["objective-1"],
        sprintIds: ["sprint-1"],
        startDate: "2026-08-01",
        teamIds: ["team-1", "team-2"],
      }),
    ).toBe(
      "2026-08-01|2026-08-30|team-1,team-2|member-1,member-2|sprint-1|objective-1",
    );
  });

  it("uses team-specific status colors before workspace-wide matches", () => {
    const statuses: Pick<State, "category" | "color" | "name" | "teamId">[] = [
      {
        category: "started",
        color: "#workspace",
        name: "In progress",
        teamId: "",
      },
      {
        category: "started",
        color: "#team-a",
        name: "In progress",
        teamId: "team-a",
      },
    ];

    expect(
      buildStatusBreakdownData(
        [
          { count: 4, statusName: "In Progress", teamId: "team-a" },
          { count: 2, statusName: "Unknown", teamId: null },
        ],
        statuses,
      ),
    ).toEqual([
      { color: "#team-a", label: "In Progress", value: 4 },
      { color: chartPalette.primary, label: "Unknown", value: 2 },
    ]);
  });

  it("caps planning chart rows and never renders a negative remainder", () => {
    const objectives = Array.from({ length: 50 }, (_, index) => ({
      avgProgress: 0,
      completed: index === 0 ? 12 : index,
      objectiveId: `objective-${index + 1}`,
      objectiveName: `Objective ${index + 1}`,
      total: index === 0 ? 8 : 10,
    }));

    const chartData = buildObjectiveProgressChartData(objectives);

    expect(chartData).toHaveLength(8);
    expect(chartData[0]).toEqual({
      completed: 12,
      label: "Objective 1",
      remaining: 0,
    });
  });

  it("sorts provider rows without mutating the API order", () => {
    const providers = [
      {
        acceptanceRate: 0,
        acceptedRequests: 1,
        declinedRequests: 0,
        highRequests: 0,
        pendingRequests: 1,
        provider: "linear",
        staleRequests: 0,
        totalRequests: 2,
        urgentRequests: 0,
      },
      {
        acceptanceRate: 0,
        acceptedRequests: 3,
        declinedRequests: 0,
        highRequests: 0,
        pendingRequests: 2,
        provider: "github",
        staleRequests: 1,
        totalRequests: 6,
        urgentRequests: 0,
      },
    ];

    const chartData = buildProviderChartData(providers);

    expect(providers.map((provider) => provider.provider)).toEqual([
      "linear",
      "github",
    ]);
    expect(chartData[0]).toMatchObject({
      acceptedShare: 50,
      pendingShare: expect.closeTo(33.33333333333333),
      provider: "Github",
      staleShare: expect.closeTo(16.666666666666664),
    });
  });
});
