/* global beforeEach, describe, expect, it -- Jest globals are provided by the projects test runner. */
import {
  getWorkspaceDraftKey,
  readWorkspaceDraft,
  saveWorkspaceDraft,
} from "./workspace-onboarding-model";

beforeEach(() => {
  sessionStorage.clear();
});

describe("workspace onboarding draft", () => {
  it("discards invalid or oversized stored data", () => {
    const key = getWorkspaceDraftKey("user");
    for (const value of [
      "not JSON",
      JSON.stringify({ version: 9 }),
      "x".repeat(4001),
    ]) {
      sessionStorage.setItem(key, value);
      expect(readWorkspaceDraft("user", "Ada")).toMatchObject({
        version: 1,
        step: 0,
        fullName: "Ada",
        name: "",
      });
    }
  });

  it("uses the saved server name and returns incomplete restored drafts to the right step", () => {
    const draft = readWorkspaceDraft("user", "");
    saveWorkspaceDraft("user", {
      ...draft,
      fullName: "Old name",
      step: 2,
      furthestStep: 2,
    });
    expect(readWorkspaceDraft("user", "Ada")).toMatchObject({
      fullName: "Ada",
      step: 0,
      furthestStep: 0,
    });
    saveWorkspaceDraft("user", {
      ...draft,
      fullName: "Old name",
      name: "Acme",
      slug: "acme",
      step: 2,
      furthestStep: 2,
    });
    expect(readWorkspaceDraft("user", "Ada")).toMatchObject({
      fullName: "Ada",
      step: 1,
    });
  });
});
