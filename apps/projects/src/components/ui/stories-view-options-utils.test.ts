/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  DISPLAY_COLUMNS_VERSION,
  migrateDisplayColumns,
} from "./stories-view-options-utils";

describe("story display column migrations", () => {
  it("adds key results next to objectives for existing saved views", () => {
    expect(migrateDisplayColumns(["Status", "Objective", "Labels"])).toEqual([
      "Status",
      "Objective",
      "Key Result",
      "Labels",
    ]);
  });

  it("adds time needed next to visible complexity for saved views", () => {
    expect(migrateDisplayColumns(["Status", "Estimate", "Labels"], 2)).toEqual([
      "Status",
      "Estimate",
      "Time needed",
      "Labels",
    ]);
  });

  it("preserves a current view that intentionally hides key results", () => {
    expect(
      migrateDisplayColumns(
        ["Status", "Objective", "Labels"],
        DISPLAY_COLUMNS_VERSION,
      ),
    ).toEqual(["Status", "Objective", "Labels"]);
  });

  it("does not add key results when objectives are hidden", () => {
    expect(migrateDisplayColumns(["Status", "Labels"])).toEqual([
      "Status",
      "Labels",
    ]);
  });
});
