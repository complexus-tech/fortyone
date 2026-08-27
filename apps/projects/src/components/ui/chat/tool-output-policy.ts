import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { isMutationToolCall, type MayaToolName } from "@/lib/ai/tool-policy";
import type { StoryPriority } from "@/modules/stories/types";
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

const MAX_APPROVAL_DETAIL_ITEMS = 60;
const MAX_APPROVAL_VALUE_CHARACTERS = 2_000;
const HIDDEN_APPROVAL_INPUT_KEYS = new Set([
  "confirmed",
  "idempotencyKey",
  "token",
]);
const INLINE_STORY_APPROVAL_KEYS = new Set([
  "assigneeId",
  "autoSchedulingEnabled",
  "description",
  "descriptionHTML",
  "endDate",
  "estimateValue",
  "estimatedDurationMinutes",
  "keyResultId",
  "labelIds",
  "minimumFocusBlockMinutes",
  "objectiveId",
  "parentId",
  "priority",
  "sprintId",
  "startDate",
  "statusId",
  "teamId",
  "title",
]);

const STORY_PRIORITIES = new Set<StoryPriority>([
  "No Priority",
  "Urgent",
  "High",
  "Medium",
  "Low",
]);

const MUTATION_APPROVAL_ACTIONS: Partial<Record<MayaToolName, string>> = {
  acceptAllIntegrationRequestsTool: "Accept all integration requests",
  acceptIntegrationRequestTool: "Accept this integration request",
  addStoryAssociation: "Add this story relationship",
  assignStoriesToUser: "Assign these stories",
  bulkDeleteStories: "Delete these stories",
  bulkUpdateStories: "Update these stories",
  createGitHubInstallSessionTool: "Continue to GitHub setup",
  createGitHubIssueSyncLinkTool: "Create this GitHub sync link",
  createKeyResultTool: "Create this key result",
  createMemory: "Save this memory",
  createObjectiveTool: "Create this objective",
  createTeamTool: "Create this team",
  declineAllIntegrationRequestsTool: "Decline all integration requests",
  declineIntegrationRequestTool: "Decline this integration request",
  deleteAttachment: "Delete this attachment",
  deleteGitHubIssueSyncLinkTool: "Delete this GitHub sync link",
  deleteKeyResultTool: "Delete this key result",
  deleteMemory: "Delete this memory",
  deleteObjectiveTool: "Delete this objective",
  deleteStory: "Delete this story",
  deleteStoryGitHubLinkTool: "Remove this story's GitHub link",
  deleteTeam: "Delete this team",
  duplicateStory: "Duplicate this story",
  joinTeam: "Join this team",
  leaveTeam: "Leave this team",
  applyMayaWorkPlanTool: "Apply the work plan shown above",
  postRequestGitHubCommentTool: "Post this GitHub comment",
  postStoryGitHubCommentTool: "Post this GitHub comment",
  removeStoryAssociation: "Remove this story relationship",
  restoreStory: "Restore this story",
  resyncGitHubRepositoriesTool: "Resync GitHub repositories",
  updateGitHubTeamSettingsTool: "Update GitHub team settings",
  updateGitHubWorkspaceSettingsTool: "Update GitHub workspace settings",
  updateIntegrationRequestTool: "Update this integration request",
  updateKeyResultTool: "Update this key result",
  updateMemory: "Update this memory",
  updateObjectiveTool: "Update this objective",
  updateSprintSettings: "Update sprint settings",
  updateStory: "Update this story",
  updateTeam: "Update this team",
};

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

const toSentenceCase = (value: string) => {
  const words = value.replaceAll("-", " ").trim();
  return words ? `${words.charAt(0).toUpperCase()}${words.slice(1)}` : "";
};

const toFieldLabel = (value: string) =>
  toSentenceCase(
    value
      .replace(/[a-z0-9][A-Z]/g, (pair) => `${pair[0]} ${pair[1]}`)
      .replace(/\bIds\b/g, "IDs")
      .replace(/\bId\b/g, "ID"),
  );

const formatApprovalValue = (value: unknown): string | undefined => {
  if (value === null) return "Clear this value";
  if (typeof value === "boolean") return value ? "Enabled" : "Disabled";
  if (typeof value === "number") return String(value);
  if (typeof value !== "string") return undefined;

  const normalized = value.trim();
  if (!normalized) return "Empty";
  if (normalized.length <= MAX_APPROVAL_VALUE_CHARACTERS) return normalized;

  return `${normalized.slice(0, MAX_APPROVAL_VALUE_CHARACTERS).trimEnd()}… (${normalized.length} characters total)`;
};

const asStoryPriority = (value: unknown): StoryPriority =>
  typeof value === "string" && STORY_PRIORITIES.has(value as StoryPriority)
    ? (value as StoryPriority)
    : "No Priority";

const getStoryApprovalDetail = (
  story: Record<string, unknown>,
  index: number,
) => {
  const title = formatApprovalValue(story.title) ?? `Story ${index + 1}`;
  const metadata: string[] = [];

  for (const [key, label] of [
    ["description", "description"],
    ["priority", "priority"],
    ["estimateValue", "complexity"],
    ["endDate", "delivery"],
    ["estimatedDurationMinutes", "time needed"],
    ["minimumFocusBlockMinutes", "minimum focus block"],
  ] as const) {
    if (story[key] === null || story[key] === undefined) continue;
    const value =
      (key === "estimatedDurationMinutes" ||
        key === "minimumFocusBlockMinutes") &&
      typeof story[key] === "number"
        ? `${story[key]} minutes`
        : formatApprovalValue(story[key]);
    if (value) metadata.push(`${label}: ${value}`);
  }

  if (!story.endDate && !story.sprintId) {
    metadata.push("delivery: not specified");
  }
  if (!story.estimatedDurationMinutes) {
    metadata.push("time needed: not specified");
  }

  if (typeof story.autoSchedulingEnabled === "boolean") {
    metadata.push(
      story.autoSchedulingEnabled
        ? "calendar scheduling: enabled"
        : "calendar scheduling: disabled",
    );
  } else {
    metadata.push("calendar scheduling: disabled (not specified)");
  }

  return {
    label: `Story ${index + 1}`,
    value: metadata.length > 0 ? `${title} — ${metadata.join(" · ")}` : title,
  };
};

const getStoryApprovalPreview = (
  story: Record<string, unknown>,
  index: number,
) => {
  const detail = getStoryApprovalDetail(story, index);

  return {
    id: `story-approval-${index + 1}`,
    priority: asStoryPriority(story.priority),
    statusId:
      typeof story.statusId === "string" && story.statusId.trim()
        ? story.statusId.trim()
        : undefined,
    summary: detail.value,
    title: formatApprovalValue(story.title) ?? `Story ${index + 1}`,
  };
};

const getStoryApprovalInputs = (input: Record<string, unknown>) => {
  const sharedValues = asToolOutputRecord(input.sharedValues);
  const storiesData = Array.isArray(input.storiesData)
    ? input.storiesData
    : [input];

  return storiesData.map((story) => {
    const storyValues = Object.fromEntries(
      Object.entries(asToolOutputRecord(story)).filter(
        ([, value]) => value !== null && value !== undefined,
      ),
    );

    return { ...sharedValues, ...storyValues };
  });
};

const getMutationApprovalDetails = (
  input: Record<string, unknown>,
  isStoryCreation: boolean,
  toolName: MayaToolName,
) => {
  const details: { label: string; value: string }[] = [];
  const addDetail = (label: string, value: unknown) => {
    if (details.length >= MAX_APPROVAL_DETAIL_ITEMS) return;
    const formatted = formatApprovalValue(value);
    if (formatted) details.push({ label, value: formatted });
  };

  if (Array.isArray(input.storiesData)) {
    for (const [index, story] of getStoryApprovalInputs(input).entries()) {
      if (details.length >= MAX_APPROVAL_DETAIL_ITEMS) break;
      details.push(getStoryApprovalDetail(story, index));
    }
  } else if (isStoryCreation && typeof input.title === "string") {
    details.push(getStoryApprovalDetail(input, 0));
  }

  const storyTitles = Array.isArray(input.storyTitles) ? input.storyTitles : [];
  if (
    Array.isArray(input.storyIds) &&
    storyTitles.length === input.storyIds.length
  ) {
    for (const [index, storyTitle] of storyTitles.entries()) {
      addDetail(`Story ${index + 1}`, storyTitle);
    }
  } else if (Array.isArray(input.storyIds)) {
    for (const [index] of input.storyIds.entries()) {
      addDetail(`Story ${index + 1}`, `Selected story ${index + 1}`);
    }
  }

  if (toolName === "deleteStory") {
    addDetail("Story", input.storyTitle);
  }

  const addObjectDetails = (value: Record<string, unknown>, prefix = "") => {
    for (const [key, item] of Object.entries(value)) {
      if (
        details.length >= MAX_APPROVAL_DETAIL_ITEMS ||
        HIDDEN_APPROVAL_INPUT_KEYS.has(key) ||
        key === "storiesData" ||
        key === "storyIds" ||
        key === "storyTitles" ||
        (toolName === "deleteStory" &&
          (key === "storyId" || key === "storyTitle")) ||
        (isStoryCreation && key === "sharedValues") ||
        (!prefix && isStoryCreation && INLINE_STORY_APPROVAL_KEYS.has(key))
      ) {
        continue;
      }

      const label = `${prefix}${toFieldLabel(key)}`;
      if (Array.isArray(item)) {
        const values = item.flatMap((entry) => {
          const scalar = formatApprovalValue(entry);
          if (scalar) return [scalar];

          const record = asToolOutputRecord(entry);
          const representative =
            formatApprovalValue(record.title) ??
            formatApprovalValue(record.name) ??
            formatApprovalValue(record.id);
          return representative ? [representative] : [];
        });
        if (values.length > 0) addDetail(label, values.join(", "));
        continue;
      }

      const record = asToolOutputRecord(item);
      if (Object.keys(record).length > 0) {
        addObjectDetails(record, `${label} · `);
        continue;
      }

      addDetail(label, item);
    }
  };

  addObjectDetails(input);
  return details;
};

const getMutationApprovalPrompt = ({
  input,
  toolName,
}: {
  input: Record<string, unknown>;
  toolName: MayaToolName;
}) => {
  const storyIds = Array.isArray(input.storyIds) ? input.storyIds : [];
  if (toolName === "bulkDeleteStories" && storyIds.length > 0) {
    return `Delete ${storyIds.length} stories?`;
  }
  if (toolName === "deleteStory" && typeof input.storyTitle === "string") {
    return `Delete “${input.storyTitle}”?`;
  }
  if (toolName === "bulkUpdateStories" && storyIds.length > 0) {
    return `Update ${storyIds.length} stories?`;
  }
  if (toolName === "assignStoriesToUser" && storyIds.length > 0) {
    return `Assign ${storyIds.length} stories?`;
  }

  const action = input.action;
  if (typeof action === "string") {
    const label = toSentenceCase(action);
    if (label) return `${label}?`;
  }

  const label = MUTATION_APPROVAL_ACTIONS[toolName];
  return label ? `${label}?` : "Confirm this change?";
};

export const getMutationApproval = (part: ToolMessagePart) => {
  if (!isMutationToolPart(part)) return undefined;
  if (
    part.state !== "approval-requested" &&
    part.state !== "approval-responded" &&
    part.state !== "output-denied"
  ) {
    return undefined;
  }
  if (!("approval" in part)) return undefined;

  const input = "input" in part ? asToolOutputRecord(part.input) : {};
  const toolName = part.type.slice("tool-".length) as MayaToolName;
  const isStoryCreation = isStoryCreationToolType(part.type);
  const storyInputs = getStoryApprovalInputs(input);
  const titles = storyInputs.flatMap((story) => {
    const title = asToolOutputRecord(story).title;
    return typeof title === "string" && title.trim() ? [title.trim()] : [];
  });

  const count = Math.max(storyInputs.length, 1);
  const title = titles.length === 1 ? titles[0] : undefined;
  let prompt = getMutationApprovalPrompt({ input, toolName });
  if (isStoryCreation) {
    prompt = title ? `Create “${title}”?` : `Create ${count} stories?`;
  }

  return {
    approved:
      "approved" in part.approval && typeof part.approval.approved === "boolean"
        ? part.approval.approved
        : undefined,
    cancelledMessage: isStoryCreation
      ? "Creation cancelled."
      : "Change cancelled.",
    count,
    description: isStoryCreation
      ? `Maya will create the prepared ${count === 1 ? "story" : "stories"} exactly as shown.`
      : "Maya will apply the prepared change exactly as requested.",
    details: getMutationApprovalDetails(input, isStoryCreation, toolName),
    id: part.approval.id,
    isStoryCreation,
    prompt,
    storyPreviews: isStoryCreation
      ? storyInputs.map(getStoryApprovalPreview)
      : [],
    title,
  };
};

export const asToolOutputRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};

export const getGitHubInstallSessionUrl = (output: unknown) => {
  const outputRecord = asToolOutputRecord(output);
  if (
    outputRecord.success !== true ||
    typeof outputRecord.installUrl !== "string"
  ) {
    return undefined;
  }

  try {
    const installUrl = new URL(outputRecord.installUrl);
    if (
      installUrl.protocol !== "https:" ||
      installUrl.hostname !== "github.com" ||
      installUrl.port ||
      installUrl.username ||
      installUrl.password ||
      !/^\/apps\/[^/]+\/installations\/new\/?$/.test(installUrl.pathname) ||
      !installUrl.searchParams.get("state")
    ) {
      return undefined;
    }
    return installUrl.toString();
  } catch {
    return undefined;
  }
};

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
  if (getMutationApproval(part)) return true;
  if (part.state !== "output-available") return false;

  if (
    part.type === "tool-mayaWorkPlanTool" ||
    part.type === "tool-applyMayaWorkPlanTool"
  ) {
    return isMayaWorkPlanOutput(part.output);
  }

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

  return (
    part.type === "tool-getSprintAnalyticsTool" ||
    isAnalyticsReportOutput(part.output)
  );
};
