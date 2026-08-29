/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getLocalOkrQualityAssessment,
  isReadyForAiQualityAssessment,
} from "./okr-quality";
import type { OkrQualityRequest } from "./schemas/okr-quality";

describe("local OKR quality assessment", () => {
  it("flags an exact objective duplicate before an AI request", () => {
    const request: OkrQualityRequest = {
      kind: "objective",
      draft: {
        name: "Improve customer retention",
        summary: "",
        startDate: "2026-07-01",
        endDate: "2026-09-30",
      },
      existingObjectives: [
        { id: "objective-1", name: "  improve CUSTOMER retention " },
      ],
    };

    expect(getLocalOkrQualityAssessment(request)).toMatchObject({
      verdict: "duplicate",
      duplicateOf: "  improve CUSTOMER retention ",
    });
  });

  it("nudges short objectives toward a more appropriate scope", () => {
    const request: OkrQualityRequest = {
      kind: "objective",
      draft: {
        name: "Make onboarding effortless",
        summary: "",
        startDate: "2026-08-01",
        endDate: "2026-08-21",
      },
      existingObjectives: [],
    };

    expect(getLocalOkrQualityAssessment(request)?.headline).toContain(
      "too short-lived",
    );
  });

  it("rejects a key result with no measurable movement", () => {
    const request: OkrQualityRequest = {
      kind: "key_result",
      draft: {
        name: "Improve page speed",
        measurementType: "number",
        startValue: 10,
        targetValue: 10,
        startDate: "2026-07-01",
        endDate: "2026-09-30",
      },
      objective: {
        id: "objective-1",
        name: "Deliver a fast product experience",
        startDate: "2026-07-01",
        endDate: "2026-09-30",
      },
      existingKeyResults: [],
    };

    expect(getLocalOkrQualityAssessment(request)?.headline).toContain(
      "baseline and target",
    );
  });

  it("waits for dates before invoking AI coaching", () => {
    const request: OkrQualityRequest = {
      kind: "objective",
      draft: {
        name: "Make onboarding effortless",
        summary: "",
        startDate: null,
        endDate: null,
      },
      existingObjectives: [],
    };

    expect(isReadyForAiQualityAssessment(request)).toBe(false);
  });
});
