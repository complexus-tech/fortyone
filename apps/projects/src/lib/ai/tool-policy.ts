import type { MayaToolName } from "./tool-names";

export type { MayaToolName } from "./tool-names";

export const MUTATION_TOOL_NAMES = [
  "createStory",
  "updateStory",
  "deleteStory",
  "bulkCreateStories",
  "bulkUpdateStories",
  "bulkDeleteStories",
  "assignStoriesToUser",
  "duplicateStory",
  "restoreStory",
  "addStoryAssociation",
  "removeStoryAssociation",
  "createTeamTool",
  "updateTeam",
  "joinTeam",
  "leaveTeam",
  "deleteTeam",
  "createObjectiveTool",
  "updateObjectiveTool",
  "deleteObjectiveTool",
  "createKeyResultTool",
  "updateKeyResultTool",
  "deleteKeyResultTool",
  "updateSprintSettings",
  "applyMayaWorkPlanTool",
  "createMemory",
  "updateMemory",
  "deleteMemory",
  "createGitHubInstallSessionTool",
  "resyncGitHubRepositoriesTool",
  "createGitHubIssueSyncLinkTool",
  "deleteGitHubIssueSyncLinkTool",
  "updateGitHubWorkspaceSettingsTool",
  "updateGitHubTeamSettingsTool",
  "postStoryGitHubCommentTool",
  "deleteStoryGitHubLinkTool",
  "updateIntegrationRequestTool",
  "acceptIntegrationRequestTool",
  "declineIntegrationRequestTool",
  "acceptAllIntegrationRequestsTool",
  "declineAllIntegrationRequestsTool",
  "postRequestGitHubCommentTool",
  "deleteAttachment",
] as const satisfies readonly MayaToolName[];

export const MUTATION_TOOL_NAME_SET: ReadonlySet<string> = new Set(
  MUTATION_TOOL_NAMES,
);

export const LEGACY_CONFIRMED_TOOL_NAMES = [
  "updateStory",
  "deleteStory",
  "bulkUpdateStories",
  "bulkDeleteStories",
  "resyncGitHubRepositoriesTool",
  "createGitHubIssueSyncLinkTool",
  "deleteGitHubIssueSyncLinkTool",
  "updateGitHubWorkspaceSettingsTool",
  "updateGitHubTeamSettingsTool",
  "postStoryGitHubCommentTool",
  "deleteStoryGitHubLinkTool",
  "updateIntegrationRequestTool",
  "acceptIntegrationRequestTool",
  "declineIntegrationRequestTool",
  "acceptAllIntegrationRequestsTool",
  "declineAllIntegrationRequestsTool",
  "postRequestGitHubCommentTool",
] as const satisfies readonly MayaToolName[];

const LEGACY_CONFIRMED_TOOL_NAME_SET: ReadonlySet<string> = new Set(
  LEGACY_CONFIRMED_TOOL_NAMES,
);

export const MUTATING_TOOL_ACTIONS = {
  comments: new Set(["add-comment", "reply-to-comment"]),
  labels: new Set(["create-label", "edit-label", "delete-label"]),
  links: new Set(["add-link", "update-link", "delete-link"]),
  notifications: new Set([
    "mark-as-read",
    "mark-all-as-read",
    "mark-as-unread",
    "delete-notification",
    "delete-all-notifications",
    "delete-read-notifications",
    "update-notification-preferences",
  ]),
  objectiveStatuses: new Set([
    "create-objective-status",
    "update-objective-status",
    "delete-objective-status",
    "set-default-objective-status",
  ]),
  statuses: new Set([
    "create-status",
    "update-status",
    "delete-status",
    "set-default-status",
  ]),
  storyLabels: new Set([
    "set-story-labels",
    "add-labels-to-story",
    "remove-labels-from-story",
  ]),
} as const satisfies Readonly<
  Partial<Record<MayaToolName, ReadonlySet<string>>>
>;

export const MUTATING_ACTION_TOOL_NAMES = Object.keys(
  MUTATING_TOOL_ACTIONS,
) as (keyof typeof MUTATING_TOOL_ACTIONS)[];

const NON_TERMINAL_MUTATION_TOOLS = new Set(["createGitHubInstallSessionTool"]);

export const isMutationCapableToolName = (toolName: string) =>
  MUTATION_TOOL_NAME_SET.has(toolName) || toolName in MUTATING_TOOL_ACTIONS;

export const isMutationToolCall = (toolName: string, input: unknown) => {
  if (MUTATION_TOOL_NAME_SET.has(toolName)) return true;

  const action =
    input && typeof input === "object" && "action" in input
      ? input.action
      : undefined;
  if (typeof action !== "string" || !(toolName in MUTATING_TOOL_ACTIONS)) {
    return false;
  }

  const mutatingActions = MUTATING_TOOL_ACTIONS[
    toolName as keyof typeof MUTATING_TOOL_ACTIONS
  ] as ReadonlySet<string>;
  return mutatingActions.has(action);
};

export const requiresMutationApproval = (toolName: string, input: unknown) =>
  isMutationToolCall(toolName, input) &&
  !NON_TERMINAL_MUTATION_TOOLS.has(toolName);

export const requiresLegacyConfirmedInput = (toolName: string) =>
  LEGACY_CONFIRMED_TOOL_NAME_SET.has(toolName);

export const toApprovedMutationInput = (
  toolName: string,
  input: unknown,
): unknown => {
  if (!requiresLegacyConfirmedInput(toolName)) return input;
  if (!input || typeof input !== "object" || Array.isArray(input)) {
    return input;
  }

  return {
    ...Object.fromEntries(
      Object.entries(input).filter(([key]) => key !== "confirmed"),
    ),
    confirmed: true,
  };
};

export const isTerminalMutationToolCall = (toolName: string, input: unknown) =>
  requiresMutationApproval(toolName, input);
