/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  DEFAULT_ESTIMATE_SCHEME,
  ESTIMATE_VALUES,
  formatEstimate,
  getEstimateOptions,
} from "./estimate";

describe("estimate formatting", () => {
  it("formats compact labels for each complexity scheme", () => {
    expect(formatEstimate("points", 1, "compact")).toBe("1");
    expect(formatEstimate("tshirt", 8, "compact")).toBe("XL");
  });

  it("formats full labels for tooltips and menus", () => {
    expect(formatEstimate("points", 1, "full")).toBe("1 point");
    expect(formatEstimate("points", 8, "full")).toBe("8 points");
    expect(formatEstimate("tshirt", 3, "full")).toBe("M");
  });

  it("returns no-estimate labels for nullish values", () => {
    expect(formatEstimate("points", null, "compact")).toBe("Complexity");
    expect(formatEstimate("points", null, "full")).toBe("No complexity");
  });

  it("defaults unknown schemes to t-shirt sizes", () => {
    expect(DEFAULT_ESTIMATE_SCHEME).toBe("tshirt");
    expect(formatEstimate(undefined, 2, "full")).toBe("S");
  });

  it("returns scheme-specific selectable options", () => {
    expect(getEstimateOptions("tshirt")).toEqual([
      { label: "XS", value: ESTIMATE_VALUES[0] },
      { label: "S", value: ESTIMATE_VALUES[1] },
      { label: "M", value: ESTIMATE_VALUES[2] },
      { label: "L", value: ESTIMATE_VALUES[3] },
      { label: "XL", value: ESTIMATE_VALUES[4] },
    ]);
  });
});
