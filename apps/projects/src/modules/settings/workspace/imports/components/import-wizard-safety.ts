import type { ImportAnalysis, ImportMapping } from "../schema";

export const isImportMappingFieldLocked = (
  field: keyof ImportMapping,
  hasAttemptedImport: boolean,
) => hasAttemptedImport && field === "sourceId";

export const prepareCompletedAIImportAnalysis = <
  T extends Pick<ImportAnalysis, "mapping" | "sourceType">,
>(
  analysis: T,
): T => {
  if (analysis.sourceType !== "json" || analysis.mapping === null) {
    return analysis;
  }
  return { ...analysis, mapping: null };
};
