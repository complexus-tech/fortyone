import {
  ESTIMATE_VALUES,
  formatEstimate,
  getEstimateOptions,
} from "./estimate";

describe("estimate formatting", () => {
  it("formats compact labels for each estimate scheme", () => {
    expect(formatEstimate("points", 1, "compact")).toBe("1");
    expect(formatEstimate("hours", 1, "compact")).toBe("0.5");
    expect(formatEstimate("tshirt", 8, "compact")).toBe("XL");
    expect(formatEstimate("ideal_days", 5, "compact")).toBe("3");
  });

  it("formats full labels for tooltips and menus", () => {
    expect(formatEstimate("points", 1, "full")).toBe("1 point");
    expect(formatEstimate("points", 8, "full")).toBe("8 points");
    expect(formatEstimate("hours", 1, "full")).toBe("0.5 hours");
    expect(formatEstimate("hours", 2, "full")).toBe("1 hour");
    expect(formatEstimate("ideal_days", 2, "full")).toBe("1 day");
    expect(formatEstimate("tshirt", 3, "full")).toBe("M");
  });

  it("returns no-estimate labels for nullish values", () => {
    expect(formatEstimate("points", null, "compact")).toBe("Estimate");
    expect(formatEstimate("points", null, "full")).toBe("No estimate");
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
