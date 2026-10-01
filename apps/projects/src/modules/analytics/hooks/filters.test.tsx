/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { renderHook } from "@testing-library/react";
import { useAppliedFilters } from "./filters";

let mockFilters: Record<string, unknown> = {};

jest.mock("nuqs", () => ({
  parseAsArrayOf: jest.fn(),
  parseAsIsoDateTime: {},
  parseAsString: {},
  useQueryStates: () => [mockFilters],
}));

jest.mock("../components/filters/types", () => ({
  getDefaultDateRange: () => ({
    startDate: new Date(2026, 8, 1),
    endDate: new Date(2026, 9, 1),
  }),
}));

describe("useAppliedFilters", () => {
  beforeEach(() => {
    mockFilters = {};
  });

  it("queries the same default period displayed by the date control", () => {
    const { result } = renderHook(useAppliedFilters);

    expect(result.current).toMatchObject({
      startDate: "2026-09-01",
      endDate: "2026-10-01",
    });
  });

  it("preserves explicit dates and selected resource filters", () => {
    mockFilters = {
      startDate: new Date(2026, 7, 15),
      endDate: new Date(2026, 8, 15),
      teamIds: ["team-1"],
      objectiveIds: ["objective-1"],
      sprintIds: [],
    };
    const { result } = renderHook(useAppliedFilters);

    expect(result.current).toEqual({
      startDate: "2026-08-15",
      endDate: "2026-09-15",
      teamIds: ["team-1"],
      objectiveIds: ["objective-1"],
      sprintIds: undefined,
    });
  });
});
