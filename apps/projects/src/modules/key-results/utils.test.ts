/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { formatKeyResultValue, getKeyResultProgress } from "./utils";

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
});
