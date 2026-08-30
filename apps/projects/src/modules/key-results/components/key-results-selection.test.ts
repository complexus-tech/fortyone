/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  isKeyResultGroupSelected,
  setKeyResultGroupSelection,
  setKeyResultSelection,
} from "./key-results-selection";

describe("key result selection", () => {
  it("selects every row in a group without mutating unrelated selection", () => {
    const selectedKeyResultIds = new Set(["outside-group", "key-result-1"]);

    const nextSelected = setKeyResultGroupSelection(
      selectedKeyResultIds,
      ["key-result-1", "key-result-2"],
      true,
    );

    expect(nextSelected).toEqual(
      new Set(["outside-group", "key-result-1", "key-result-2"]),
    );
    expect(selectedKeyResultIds).toEqual(
      new Set(["outside-group", "key-result-1"]),
    );
    expect(
      isKeyResultGroupSelected(["key-result-1", "key-result-2"], nextSelected),
    ).toBe(true);
  });

  it("deselects only the requested rows and never selects an empty group", () => {
    const selectedKeyResultIds = new Set([
      "outside-group",
      "key-result-1",
      "key-result-2",
    ]);

    expect(
      setKeyResultGroupSelection(
        selectedKeyResultIds,
        ["key-result-1", "key-result-2"],
        false,
      ),
    ).toEqual(new Set(["outside-group"]));
    expect(isKeyResultGroupSelected([], selectedKeyResultIds)).toBe(false);
    expect(
      setKeyResultSelection(selectedKeyResultIds, "key-result-1", false),
    ).toEqual(new Set(["outside-group", "key-result-2"]));
  });
});
