import type { tools } from "./tools";

export type MayaToolName = keyof typeof tools;

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
  "mayaWorkPlanTool",
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

const MUTATING_ACTIONS: Readonly<Partial<Record<string, ReadonlySet<string>>>> =
  {
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
  };

const NON_TERMINAL_MUTATION_TOOLS = new Set(["createGitHubInstallSessionTool"]);

export const isMutationCapableToolName = (toolName: string) =>
  MUTATION_TOOL_NAME_SET.has(toolName) || toolName in MUTATING_ACTIONS;

export const isMutationToolCall = (toolName: string, input: unknown) => {
  if (MUTATION_TOOL_NAME_SET.has(toolName)) return true;

  const action =
    input && typeof input === "object" && "action" in input
      ? input.action
      : undefined;
  return (
    typeof action === "string" &&
    (MUTATING_ACTIONS[toolName]?.has(action) ?? false)
  );
};

export const isTerminalMutationToolCall = (toolName: string, input: unknown) =>
  isMutationToolCall(toolName, input) &&
  !NON_TERMINAL_MUTATION_TOOLS.has(toolName);
