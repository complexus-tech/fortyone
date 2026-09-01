/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ImportMapping } from "../schema";
import {
  isImportMappingFieldLocked,
  prepareCompletedAIImportAnalysis,
} from "./import-wizard-safety";

const mapping: ImportMapping = {
  assigneeEmail: null,
  description: "description",
  endDate: null,
  priority: null,
  sourceId: "id",
  startDate: null,
  status: "status",
  title: "title",
};

describe("import wizard safety", () => {
  it("keeps every mapping editable before an import attempt", () => {
    for (const field of [
      "title",
      "description",
      "status",
      "priority",
      "assigneeEmail",
      "startDate",
      "endDate",
      "sourceId",
    ] satisfies (keyof ImportMapping)[]) {
      expect(isImportMappingFieldLocked(field, false)).toBe(false);
    }
  });

  it("locks only the source ID mapping after an import attempt", () => {
    expect(isImportMappingFieldLocked("sourceId", true)).toBe(true);
    expect(isImportMappingFieldLocked("title", true)).toBe(false);
    expect(isImportMappingFieldLocked("status", true)).toBe(false);
  });

  it("removes deterministic mapping controls from completed AI JSON", () => {
    const analysis = prepareCompletedAIImportAnalysis({
      mapping,
      sourceType: "json" as const,
    });

    expect(analysis.mapping).toBeNull();
  });

  it("preserves CSV mappings and already-unmapped JSON AI analysis", () => {
    const csvAnalysis = { mapping, sourceType: "jira_csv" as const };
    const jsonFallback = { mapping: null, sourceType: "json" as const };

    expect(prepareCompletedAIImportAnalysis(csvAnalysis)).toBe(csvAnalysis);
    expect(prepareCompletedAIImportAnalysis(jsonFallback)).toBe(jsonFallback);
  });
});
