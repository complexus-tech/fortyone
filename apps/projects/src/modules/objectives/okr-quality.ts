import { differenceInCalendarDays, isValid, parseISO } from "date-fns";
import type {
  OkrQualityAssessment,
  OkrQualityRequest,
} from "./schemas/okr-quality";

const normalizeName = (name: string) =>
  name.trim().toLocaleLowerCase().replace(/\s+/g, " ");

const getDurationInDays = (
  startDate: string | null,
  endDate: string | null,
) => {
  if (!startDate || !endDate) return null;
  const start = parseISO(startDate);
  const end = parseISO(endDate);
  if (!isValid(start) || !isValid(end)) return null;
  return differenceInCalendarDays(end, start);
};

const duplicateAssessment = (duplicateOf: string): OkrQualityAssessment => ({
  verdict: "duplicate",
  headline: "This looks like an existing goal",
  guidance: [
    "Update the existing item or make this outcome meaningfully different.",
  ],
  suggestedName: null,
  duplicateOf,
});

export const getLocalOkrQualityAssessment = (
  request: OkrQualityRequest,
): OkrQualityAssessment | null => {
  const comparableItems =
    request.kind === "objective"
      ? request.existingObjectives
      : request.existingKeyResults;
  const duplicate = comparableItems.find(
    (item) => normalizeName(item.name) === normalizeName(request.draft.name),
  );
  if (duplicate) return duplicateAssessment(duplicate.name);

  const duration = getDurationInDays(
    request.draft.startDate,
    request.draft.endDate,
  );
  if (duration !== null && duration < 0) {
    return {
      verdict: "needs_attention",
      headline: "The timeline runs backwards",
      guidance: ["Set the deadline after the start date."],
      suggestedName: null,
      duplicateOf: null,
    };
  }

  if (request.kind === "objective" && duration !== null && duration < 60) {
    return {
      verdict: "needs_attention",
      headline: "This may be too short-lived for an objective",
      guidance: [
        "Objectives usually describe a quarterly, six-month, or annual outcome. A goal under two months may be better framed as a key result or project.",
      ],
      suggestedName: null,
      duplicateOf: null,
    };
  }

  if (request.kind === "key_result") {
    if (
      request.draft.measurementType !== "boolean" &&
      request.draft.startValue === request.draft.targetValue
    ) {
      return {
        verdict: "needs_attention",
        headline: "The baseline and target are the same",
        guidance: [
          "A measurable key result needs a different target so progress can be tracked.",
        ],
        suggestedName: null,
        duplicateOf: null,
      };
    }
    if (request.draft.measurementType === "boolean") {
      return {
        verdict: "needs_attention",
        headline: "Can this result be measured instead?",
        guidance: [
          "Binary milestones often describe work completed. Prefer an outcome with a baseline and target when possible.",
        ],
        suggestedName: null,
        duplicateOf: null,
      };
    }
    if (
      request.objective.startDate &&
      request.draft.startDate &&
      request.draft.startDate < request.objective.startDate
    ) {
      return {
        verdict: "needs_attention",
        headline: "This starts before its objective",
        guidance: ["Keep the key-result timeline inside the objective period."],
        suggestedName: null,
        duplicateOf: null,
      };
    }
    if (
      request.objective.endDate &&
      request.draft.endDate &&
      request.draft.endDate > request.objective.endDate
    ) {
      return {
        verdict: "needs_attention",
        headline: "This ends after its objective",
        guidance: [
          "A key result should be achievable before the objective ends.",
        ],
        suggestedName: null,
        duplicateOf: null,
      };
    }
  }

  return null;
};

export const isReadyForAiQualityAssessment = (request: OkrQualityRequest) =>
  request.draft.name.trim().length >= 8 &&
  Boolean(request.draft.startDate && request.draft.endDate);
