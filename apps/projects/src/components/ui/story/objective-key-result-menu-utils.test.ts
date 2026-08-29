/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { filterKeyResultsByName } from "./objective-key-result-menu-utils";

const keyResults = [
  { id: "accuracy", name: "Achieve 90% Accuracy" },
  { id: "testers", name: "Onboard Beta Testers" },
];

describe("filterKeyResultsByName", () => {
  it("matches key-result names without case sensitivity", () => {
    expect(filterKeyResultsByName(keyResults, "ACCURACY")).toEqual([
      keyResults[0],
    ]);
  });

  it("trims the query before matching", () => {
    expect(filterKeyResultsByName(keyResults, "  beta  ")).toEqual([
      keyResults[1],
    ]);
  });

  it("keeps the full list for an empty query", () => {
    expect(filterKeyResultsByName(keyResults, "   ")).toBe(keyResults);
  });
});
