/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  formatTimeNeeded,
  normalizeTimeNeeded,
  normalizeTimeNeededPatch,
  parseTimeNeededInput,
} from "./time-needed";

describe("time-needed helpers", () => {
  it("formats minutes, whole hours, and mixed durations", () => {
    expect(formatTimeNeeded(15)).toBe("15m");
    expect(formatTimeNeeded(60)).toBe("1h");
    expect(formatTimeNeeded(90)).toBe("1h 30m");
    expect(formatTimeNeeded(90, "full")).toBe("1 hour 30 minutes");
  });

  it("keeps an unset duration distinct from complexity", () => {
    expect(formatTimeNeeded(null)).toBe("Time needed");
    expect(formatTimeNeeded(null, "full")).toBe("No time needed");
  });

  it("parses custom minutes and hours into whole minutes", () => {
    expect(parseTimeNeededInput("45", "minutes")).toBe(45);
    expect(parseTimeNeededInput("1.5", "hours")).toBe(90);
    expect(parseTimeNeededInput("0", "hours")).toBeNull();
    expect(parseTimeNeededInput("40.25", "hours")).toBeNull();
    expect(parseTimeNeededInput("2401", "minutes")).toBeNull();
    expect(parseTimeNeededInput("not a number", "minutes")).toBeNull();
  });

  it("drops invalid values and focus blocks longer than the duration", () => {
    expect(
      normalizeTimeNeeded({
        estimatedDurationMinutes: 60,
        minimumFocusBlockMinutes: 120,
      }),
    ).toEqual({
      estimatedDurationMinutes: 60,
      minimumFocusBlockMinutes: null,
    });
    expect(
      normalizeTimeNeeded({
        estimatedDurationMinutes: undefined,
        minimumFocusBlockMinutes: 30,
      }),
    ).toEqual({
      estimatedDurationMinutes: null,
      minimumFocusBlockMinutes: null,
    });
  });

  it("clears a dependent focus block when a duration patch clears time", () => {
    expect(normalizeTimeNeededPatch(null, undefined)).toEqual({
      estimatedDurationMinutes: null,
      minimumFocusBlockMinutes: null,
    });
    expect(normalizeTimeNeededPatch(60, undefined)).toEqual({
      estimatedDurationMinutes: 60,
      minimumFocusBlockMinutes: undefined,
    });
  });
});
