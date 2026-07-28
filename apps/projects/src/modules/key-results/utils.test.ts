/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  formatKeyResultValue,
  getKeyResultProgress,
  groupKeyResultsByObjective,
} from "./utils";

describe("key result utilities", () => {
  describe("getKeyResultProgress", () => {
    it("calculates progress relative to the start and target values", () => {
      expect(
        getKeyResultProgress({
          measurementType: "number",
          startValue: 10,
          currentValue: 15,
          targetValue: 20,
        }),
      ).toBe(50);
    });

    it("supports decreasing targets and clamps progress", () => {
      expect(
        getKeyResultProgress({
          measurementType: "number",
          startValue: 100,
          currentValue: 40,
          targetValue: 50,
        }),
      ).toBe(100);
    });

    it("handles boolean and unchanged targets", () => {
      expect(
        getKeyResultProgress({
          measurementType: "boolean",
          startValue: 0,
          currentValue: 1,
          targetValue: 1,
        }),
      ).toBe(100);
      expect(
        getKeyResultProgress({
          measurementType: "number",
          startValue: 5,
          currentValue: 5,
          targetValue: 5,
        }),
      ).toBe(100);
    });
  });

  describe("formatKeyResultValue", () => {
    it("formats percentage, numeric, and boolean values", () => {
      expect(formatKeyResultValue(42.5, "percentage")).toBe("42.5%");
      expect(formatKeyResultValue(15000, "number")).toBe("15,000");
      expect(formatKeyResultValue(1, "boolean")).toBe("Complete");
    });
  });

  describe("groupKeyResultsByObjective", () => {
    it("groups key results by objective and calculates average progress", () => {
      const baseKeyResult = {
        contributors: [],
        createdAt: "2026-01-01",
        createdBy: "user-1",
        currentValue: 5,
        endDate: "2026-03-31",
        lead: null,
        measurementType: "number" as const,
        objectiveId: "objective-1",
        objectiveName: "Improve customer growth",
        sequenceId: 1,
        startDate: "2026-01-01",
        startValue: 0,
        targetValue: 10,
        teamCode: "GROW",
        teamId: "team-1",
        teamName: "Growth",
        updatedAt: "2026-01-01",
        workspaceId: "workspace-1",
      };

      const groups = groupKeyResultsByObjective([
        { ...baseKeyResult, id: "kr-1", name: "Increase conversion" },
        {
          ...baseKeyResult,
          currentValue: 10,
          id: "kr-2",
          name: "Increase activation",
          sequenceId: 2,
        },
      ]);

      expect(groups).toHaveLength(1);
      expect(groups[0]).toMatchObject({
        averageProgress: 75,
        objectiveId: "objective-1",
        objectiveName: "Improve customer growth",
      });
      expect(groups[0].keyResults).toHaveLength(2);
    });
  });
});
