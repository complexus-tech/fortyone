/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Objective } from "@/modules/objectives/types";
import {
  DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  getObjectiveGroupUpdate,
  getHiddenObjectiveGroupKeys,
  groupObjectives,
  hideObjectiveGroup,
  showObjectiveGroup,
} from "./objective-board-utils";

const objective = (overrides: Partial<Objective>): Objective => ({
  id: "objective-1",
  sequenceId: 1,
  name: "Improve reliability",
  description: "",
  shortSummary: null,
  leadUser: "member-1",
  teamId: "team-1",
  workspaceId: "workspace-1",
  startDate: "2026-01-01",
  endDate: "2026-03-31",
  isPrivate: false,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-02T00:00:00Z",
  createdBy: "member-1",
  statusId: "status-1",
  keyResultCount: 0,
  priority: "High",
  health: "On Track",
  ...overrides,
});

describe("objective board hidden columns", () => {
  it("keeps hidden columns scoped to the active grouping", () => {
    const statusOptions = hideObjectiveGroup(
      DEFAULT_OBJECTIVE_VIEW_OPTIONS,
      "status-1",
    );

    expect(getHiddenObjectiveGroupKeys(statusOptions)).toEqual(["status-1"]);

    const priorityOptions = {
      ...statusOptions,
      groupBy: "priority" as const,
    };
    expect(getHiddenObjectiveGroupKeys(priorityOptions)).toEqual([]);

    const restoredStatusOptions = showObjectiveGroup(
      { ...priorityOptions, groupBy: "status" },
      "status-1",
    );
    expect(getHiddenObjectiveGroupKeys(restoredStatusOptions)).toEqual([]);
  });
});

describe("groupObjectives", () => {
  it("keeps configured empty status groups and sorts their objectives", () => {
    const groups = groupObjectives({
      objectives: [
        objective({ id: "low", name: "Low", priority: "Low" }),
        objective({ id: "urgent", name: "Urgent", priority: "Urgent" }),
      ],
      statuses: [
        {
          id: "status-1",
          name: "In progress",
          category: "started",
          isDefault: true,
          color: "#fff",
          orderIndex: 0,
          workspaceId: "workspace-1",
          createdAt: "2026-01-01T00:00:00Z",
          updatedAt: "2026-01-01T00:00:00Z",
        },
        {
          id: "status-2",
          name: "Done",
          category: "completed",
          isDefault: false,
          color: "#fff",
          orderIndex: 1,
          workspaceId: "workspace-1",
          createdAt: "2026-01-01T00:00:00Z",
          updatedAt: "2026-01-01T00:00:00Z",
        },
      ],
      members: [],
      viewOptions: DEFAULT_OBJECTIVE_VIEW_OPTIONS,
    });

    expect(groups).toHaveLength(2);
    expect(groups[0].objectives.map(({ id }) => id)).toEqual(["urgent", "low"]);
    expect(groups[1].objectives).toEqual([]);
  });

  it("maps a lead drop onto an editable objective property", () => {
    expect(getObjectiveGroupUpdate("lead", "member-2")).toEqual({
      leadUser: "member-2",
    });
    expect(getObjectiveGroupUpdate("lead", "unassigned")).toEqual({
      leadUser: null,
    });
  });
});
