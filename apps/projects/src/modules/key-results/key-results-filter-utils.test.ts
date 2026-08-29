/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getActiveKeyResultFilterCount } from "./key-results-filter-utils";

describe("getActiveKeyResultFilterCount", () => {
  it("counts each active filter category once", () => {
    expect(
      getActiveKeyResultFilterCount({
        endDateAfter: "2026-07-01T00:00:00.000Z",
        endDateBefore: "2026-07-31T00:00:00.000Z",
        leadIds: ["user-1", "user-2"],
        measurementTypes: ["number", "percentage"],
        objectiveIds: ["objective-1"],
        teamIds: ["team-1"],
      }),
    ).toBe(5);
  });

  it("ignores empty filter values", () => {
    expect(
      getActiveKeyResultFilterCount({
        leadIds: [],
        measurementTypes: [],
        objectiveIds: [],
        teamIds: [],
      }),
    ).toBe(0);
  });
});
