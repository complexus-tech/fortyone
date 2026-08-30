/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { asSingleSprintBurndown, asWorkingDays } from "./model";

describe("analytics report model", () => {
  it("keeps only valid single-sprint burndown rows", () => {
    expect(
      asSingleSprintBurndown([
        { date: "2026-08-01", ideal: 10, remaining: 12 },
        { date: "not-a-date", ideal: 8, remaining: 9 },
        { date: "2026-08-03", ideal: "not-a-number", remaining: 4 },
        { date: "2026-08-04", ideal: 6, remaining: 3 },
      ]),
    ).toEqual([
      { date: "2026-08-01", ideal: 10, remaining: 12 },
      { date: "2026-08-04", ideal: 6, remaining: 3 },
    ]);
  });

  it("normalizes working days without inventing a schedule", () => {
    expect(asWorkingDays([0, 1, 2, 5, 7, 8, 1.5, "3"])).toEqual([1, 2, 5, 7]);
    expect(asWorkingDays([])).toBeUndefined();
    expect(asWorkingDays("weekdays")).toBeUndefined();
  });
});
