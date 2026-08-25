import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { isMutationToolCall } from "@/lib/ai/tool-policy";
import {
  ENTITY_RESULT_TOOL_TYPES,
  getEntityResultsModel,
} from "./entity-results-data";
import { isStoryResultsOutput } from "./story-results-data";
import { isMayaWorkPlanOutput } from "./maya-work-plan-data";

export type ToolMessagePart = MayaUIMessage["parts"][number] & {
  output?: unknown;
  state: string;
};

const SUPPORTING_TOOL_TYPES = new Set([
  "tool-statuses",
  "tool-objectiveStatuses",
  "tool-resolveMember",
  "tool-focusBrief",
]);

const LEGACY_MEMBER_RESOLUTION_ACTIONS = new Set([
  "search-members",
  "get-member-details",
]);

const STORY_RESULT_TOOL_TYPES = new Set([
  "tool-listTeamStories",
  "tool-searchStories",
]);

const STORY_CREATION_TOOL_TYPES = new Set([
  "tool-createStory",
  "tool-bulkCreateStories",
]);

const isToolPart = (type: string): boolean => type.startsWith("tool-");

export const isToolMessagePart = (
  part: MayaUIMessage["parts"][number],
): part is ToolMessagePart => isToolPart(part.type) && "state" in part;

export const isSupportingToolType = (type: string) =>
  SUPPORTING_TOOL_TYPES.has(type);

export const isStoryResultToolType = (type: string) =>
  STORY_RESULT_TOOL_TYPES.has(type);

export const isEntityResultToolType = (type: string) =>
  ENTITY_RESULT_TOOL_TYPES.has(type);

export const isMutationToolPart = (part: ToolMessagePart) =>
  isMutationToolCall(
    part.type.slice("tool-".length),
    "input" in part ? part.input : undefined,
  );

export const isStoryCreationToolType = (type: string) =>
  STORY_CREATION_TOOL_TYPES.has(type);

export const asToolOutputRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};

export const isMemberResolverToolPart = (part: ToolMessagePart) => {
  if (part.type === "tool-resolveMember") return true;

  if (part.type === "tool-members") {
    const input = "input" in part ? asToolOutputRecord(part.input) : {};
    return (
      typeof input.action === "string" &&
      LEGACY_MEMBER_RESOLUTION_ACTIONS.has(input.action)
    );
  }

  if (part.type === "tool-listTeamMembers") {
    const input = "input" in part ? asToolOutputRecord(part.input) : {};
    return typeof input.searchQuery === "string" && Boolean(input.searchQuery);
  }

  return false;
};

export const getMutationMessage = (output: unknown) => {
  const outputRecord = asToolOutputRecord(output);
  const message = outputRecord.message;
  if (typeof message === "string" && message.trim()) return message.trim();

  if (outputRecord.success === false) {
    const error = outputRecord.error;
    if (typeof error === "string" && error.trim()) return error.trim();
  }

  if (outputRecord.success === true) return "Change completed successfully.";

  return undefined;
};

export const getStoryCreationMessage = getMutationMessage;

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

  if (isStoryCreationToolType(part.type)) {
    return getStoryCreationMessage(part.output) !== undefined;
  }

  if (isMutationToolPart(part)) {
    return getMutationMessage(part.output) !== undefined;
  }

  if (isEntityResultToolType(part.type)) {
    return getEntityResultsModel(part.type, part.output) !== null;
  }

  if (part.type === "tool-suggestions") {
    return getToolSuggestions(part.output).length > 0;
  }

  if (part.type === "tool-mayaWorkPlanTool") {
    return isMayaWorkPlanOutput(part.output);
  }

  return (
    part.type === "tool-getSprintAnalyticsTool" ||
    isAnalyticsReportOutput(part.output)
  );
};
