import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { isStoryResultsOutput } from "./story-results-data";

export type ToolMessagePart = MayaUIMessage["parts"][number] & {
  output?: unknown;
  state: string;
};

const SUPPORTING_TOOL_TYPES = new Set([
  "tool-statuses",
  "tool-objectiveStatuses",
]);

const STORY_RESULT_TOOL_TYPES = new Set([
  "tool-listTeamStories",
  "tool-searchStories",
]);

const isToolPart = (type: string): boolean => type.startsWith("tool-");

export const isToolMessagePart = (
  part: MayaUIMessage["parts"][number],
): part is ToolMessagePart => isToolPart(part.type) && "state" in part;

export const isSupportingToolType = (type: string) =>
  SUPPORTING_TOOL_TYPES.has(type);

export const isStoryResultToolType = (type: string) =>
  STORY_RESULT_TOOL_TYPES.has(type);

export const asToolOutputRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};

export const isAnalyticsReportOutput = (
  output: unknown,
): output is Record<string, unknown> => {
  if (!output || typeof output !== "object" || !("kind" in output)) {
    return false;
  }

  const kind = (output as { kind?: unknown }).kind;
  return typeof kind === "string" && kind.endsWith("-report");
};

export const getToolSuggestions = (output: unknown) => {
  const outputRecord = asToolOutputRecord(output);
  return Array.isArray(outputRecord.suggestions)
    ? outputRecord.suggestions.filter(
        (suggestion): suggestion is string => typeof suggestion === "string",
      )
    : [];
};

export const isRenderableToolPart = (part: ToolMessagePart) => {
  if (part.state !== "output-available") return false;

  if (isStoryResultToolType(part.type)) {
    return isStoryResultsOutput(part.output);
  }

  if (part.type === "tool-suggestions") {
    return getToolSuggestions(part.output).length > 0;
  }

  return (
    part.type === "tool-getSprintAnalyticsTool" ||
    isAnalyticsReportOutput(part.output)
  );
};
