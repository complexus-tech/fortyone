/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  normalizeKeyResultsFilterValues,
  toggleKeyResultsFilterValue,
} from "./key-results-filter-values";

describe("key result filter values", () => {
  it("toggles a filter value without mutating the current selection", () => {
    const currentSelection = ["member-1"];

    expect(toggleKeyResultsFilterValue(currentSelection, "member-2")).toEqual([
      "member-1",
      "member-2",
    ]);
    expect(currentSelection).toEqual(["member-1"]);
    expect(toggleKeyResultsFilterValue(currentSelection, "member-1")).toEqual(
      [],
    );
  });

  it("removes empty arrays from the persisted filter state", () => {
    expect(normalizeKeyResultsFilterValues([])).toBeUndefined();
    expect(normalizeKeyResultsFilterValues(["team-1"])).toEqual(["team-1"]);
  });
});
