/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { DEFAULT_STORIES_FILTER } from "../stories-filter-types";
import {
  getEditorContentClassName,
  getNames,
  getOperatorConfig,
  normalizeArrayFilter,
  removeStoriesFilterField,
  shouldFetchNextPage,
} from "./filter-model";

describe("stories filter bar model", () => {
  it("resolves default and selected operator labels by field semantics", () => {
    expect(
      getOperatorConfig(DEFAULT_STORIES_FILTER, "statusIds").operator,
    ).toBe("is any of");
    expect(
      getOperatorConfig(
        {
          ...DEFAULT_STORIES_FILTER,
          operators: { startDate: "isOnOrAfter" },
        },
        "startDate",
      ).operator,
    ).toBe("is on or after");
    expect(
      getOperatorConfig(DEFAULT_STORIES_FILTER, "hasNoAssignee")
        .operatorOptions,
    ).toEqual([
      { label: "is", value: "isEmpty" },
      { label: "is not", value: "isNotEmpty" },
    ]);
  });

  it("removes operator-backed values and their stale operator override", () => {
    expect(
      removeStoriesFilterField(
        {
          ...DEFAULT_STORIES_FILTER,
          operators: { statusIds: "isNotAnyOf", teamIds: "isAnyOf" },
          statusIds: ["status-1"],
        },
        "statusIds",
      ),
    ).toMatchObject({
      operators: { statusIds: undefined, teamIds: "isAnyOf" },
      statusIds: null,
    });
  });

  it("uses boolean false for shortcut fields and null for value fields", () => {
    expect(
      removeStoriesFilterField(
        { ...DEFAULT_STORIES_FILTER, assignedToMe: true },
        "assignedToMe",
      ).assignedToMe,
    ).toBe(false);
    expect(
      removeStoriesFilterField(
        { ...DEFAULT_STORIES_FILTER, keyResultId: "kr-1" },
        "keyResultId",
      ).keyResultId,
    ).toBeNull();
  });

  it("normalizes empty selections without changing populated selections", () => {
    expect(normalizeArrayFilter([])).toBeNull();
    expect(normalizeArrayFilter([1, 3])).toEqual([1, 3]);
  });

  it("preserves missing identifiers when lookup data has not loaded", () => {
    expect(
      getNames(["team-1", "team-2"], new Map([["team-1", "Platform"]])),
    ).toBe("Platform, team-2");
  });

  it("keeps editor sizing tied to filter interaction needs", () => {
    expect(getEditorContentClassName("objectiveId")).toContain("w-96");
    expect(getEditorContentClassName("startDate")).toContain("w-auto");
    expect(getEditorContentClassName("statusIds")).toContain("w-64");
  });

  it("loads the next page only near the scroll boundary", () => {
    expect(
      shouldFetchNextPage(
        { clientHeight: 300, scrollHeight: 1_000, scrollTop: 630 },
        true,
        false,
      ),
    ).toBe(true);
    expect(
      shouldFetchNextPage(
        { clientHeight: 300, scrollHeight: 1_000, scrollTop: 500 },
        true,
        false,
      ),
    ).toBe(false);
  });
});
