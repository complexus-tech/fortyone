/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { isSubstorySuggestionReadyToCreate } from "./substory-suggestion-state";

const completeSuggestion = {
  substories: [{ title: "Let users export their dashboard" }],
};

describe("isSubstorySuggestionReadyToCreate", () => {
  it("allows a complete, schema-valid response", () => {
    expect(
      isSubstorySuggestionReadyToCreate({
        error: undefined,
        isLoading: false,
        object: completeSuggestion,
      }),
    ).toBe(true);
  });

  it("rejects partial streaming data", () => {
    expect(
      isSubstorySuggestionReadyToCreate({
        error: undefined,
        isLoading: true,
        object: completeSuggestion,
      }),
    ).toBe(false);
    expect(
      isSubstorySuggestionReadyToCreate({
        error: undefined,
        isLoading: false,
        object: { substories: [{ title: "" }] },
      }),
    ).toBe(false);
  });

  it("rejects a response after streaming fails", () => {
    expect(
      isSubstorySuggestionReadyToCreate({
        error: new Error("provider disconnected"),
        isLoading: false,
        object: completeSuggestion,
      }),
    ).toBe(false);
  });
});
