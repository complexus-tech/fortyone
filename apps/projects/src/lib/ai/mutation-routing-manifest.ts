import type { MAYA_TOOL_ACTIONS, MayaActionToolName } from "./tool-actions";
import type { MayaToolName } from "./tool-names";
import type { MayaToolDomain } from "./tool-routing";

export type StandaloneMutationRoutingEntry = {
  domain: MayaToolDomain;
  operationSource: "tool";
  operations: readonly string[];
  toolName: Exclude<MayaToolName, MayaActionToolName>;
};

export type ActionScopedMutationRoutingEntry = {
  [ToolName in MayaActionToolName]: {
    domain: MayaToolDomain;
    operationSource: "input-action";
    operations: readonly (typeof MAYA_TOOL_ACTIONS)[ToolName][number][];
    toolName: ToolName;
  };
}[MayaActionToolName];

export type MutationRoutingEntry =
  | StandaloneMutationRoutingEntry
  | ActionScopedMutationRoutingEntry;

/**
 * Canonical routing metadata for every tool that can mutate workspace state.
 * Standalone tools use normalized semantic operations; tools with an `action`
 * input list the exact mutating action values accepted by their schemas.
 */
export const MUTATION_ROUTING_MANIFEST = [
  {
    domain: "story",
    operationSource: "tool",
    operations: ["create"],
    toolName: "createStory",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["update"],
    toolName: "updateStory",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["delete"],
    toolName: "deleteStory",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["create"],
    toolName: "bulkCreateStories",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["update"],
    toolName: "bulkUpdateStories",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["delete"],
    toolName: "bulkDeleteStories",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["assign"],
    toolName: "assignStoriesToUser",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["duplicate"],
    toolName: "duplicateStory",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["restore"],
    toolName: "restoreStory",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["add-association"],
    toolName: "addStoryAssociation",
  },
  {
    domain: "story",
    operationSource: "tool",
    operations: ["remove-association"],
    toolName: "removeStoryAssociation",
  },
  {
    domain: "team",
    operationSource: "tool",
    operations: ["create"],
    toolName: "createTeamTool",
  },
  {
    domain: "team",
    operationSource: "tool",
    operations: ["update"],
    toolName: "updateTeam",
  },
  {
    domain: "team",
    operationSource: "tool",
    operations: ["join"],
    toolName: "joinTeam",
  },
  {
    domain: "team",
    operationSource: "tool",
    operations: ["leave"],
    toolName: "leaveTeam",
  },
  {
    domain: "team",
    operationSource: "tool",
    operations: ["delete"],
    toolName: "deleteTeam",
  },
  {
    domain: "objective",
    operationSource: "tool",
    operations: ["create-objective"],
    toolName: "createObjectiveTool",
  },
  {
    domain: "objective",
    operationSource: "tool",
    operations: ["update-objective"],
    toolName: "updateObjectiveTool",
  },
  {
    domain: "objective",
    operationSource: "tool",
    operations: ["delete-objective"],
    toolName: "deleteObjectiveTool",
  },
  {
    domain: "objective",
    operationSource: "tool",
    operations: ["create-key-result"],
    toolName: "createKeyResultTool",
  },
  {
    domain: "objective",
    operationSource: "tool",
    operations: ["update-key-result"],
    toolName: "updateKeyResultTool",
  },
  {
    domain: "objective",
    operationSource: "tool",
    operations: ["delete-key-result"],
    toolName: "deleteKeyResultTool",
  },
  {
    domain: "sprint",
    operationSource: "tool",
    operations: ["update-settings"],
    toolName: "updateSprintSettings",
  },
  {
    domain: "planning",
    operationSource: "tool",
    operations: ["apply"],
    toolName: "applyMayaWorkPlanTool",
  },
  {
    domain: "memory",
    operationSource: "tool",
    operations: ["create"],
    toolName: "createMemory",
  },
  {
    domain: "memory",
    operationSource: "tool",
    operations: ["update"],
    toolName: "updateMemory",
  },
  {
    domain: "memory",
    operationSource: "tool",
    operations: ["delete"],
    toolName: "deleteMemory",
  },
  {
    domain: "github",
    operationSource: "tool",
    operations: ["install"],
    toolName: "createGitHubInstallSessionTool",
  },
  {
    domain: "github",
    operationSource: "tool",
    operations: ["resync"],
    toolName: "resyncGitHubRepositoriesTool",
  },
  {
    domain: "github",
    operationSource: "tool",
    operations: ["create-link"],
    toolName: "createGitHubIssueSyncLinkTool",
  },
  {
    domain: "github",
    operationSource: "tool",
    operations: ["delete-link"],
    toolName: "deleteGitHubIssueSyncLinkTool",
  },
  {
    domain: "github",
    operationSource: "tool",
    operations: ["update-settings"],
    toolName: "updateGitHubWorkspaceSettingsTool",
  },
  {
    domain: "github",
    operationSource: "tool",
    operations: ["update-settings"],
    toolName: "updateGitHubTeamSettingsTool",
  },
  {
    domain: "github",
    operationSource: "tool",
    operations: ["post-comment"],
    toolName: "postStoryGitHubCommentTool",
  },
  {
    domain: "github",
    operationSource: "tool",
    operations: ["delete-link"],
    toolName: "deleteStoryGitHubLinkTool",
  },
  {
    domain: "integration-request",
    operationSource: "tool",
    operations: ["update"],
    toolName: "updateIntegrationRequestTool",
  },
  {
    domain: "integration-request",
    operationSource: "tool",
    operations: ["accept"],
    toolName: "acceptIntegrationRequestTool",
  },
  {
    domain: "integration-request",
    operationSource: "tool",
    operations: ["decline"],
    toolName: "declineIntegrationRequestTool",
  },
  {
    domain: "integration-request",
    operationSource: "tool",
    operations: ["accept"],
    toolName: "acceptAllIntegrationRequestsTool",
  },
  {
    domain: "integration-request",
    operationSource: "tool",
    operations: ["decline"],
    toolName: "declineAllIntegrationRequestsTool",
  },
  {
    domain: "integration-request",
    operationSource: "tool",
    operations: ["post-comment"],
    toolName: "postRequestGitHubCommentTool",
  },
  {
    domain: "attachment",
    operationSource: "tool",
    operations: ["delete"],
    toolName: "deleteAttachment",
  },
  {
    domain: "comment",
    operationSource: "input-action",
    operations: ["add-comment", "reply-to-comment"],
    toolName: "comments",
  },
  {
    domain: "label",
    operationSource: "input-action",
    operations: ["create-label", "edit-label", "delete-label"],
    toolName: "labels",
  },
  {
    domain: "link",
    operationSource: "input-action",
    operations: ["add-link", "update-link", "delete-link"],
    toolName: "links",
  },
  {
    domain: "notification",
    operationSource: "input-action",
    operations: [
      "mark-as-read",
      "mark-all-as-read",
      "mark-as-unread",
      "delete-notification",
      "delete-all-notifications",
      "delete-read-notifications",
      "update-notification-preferences",
    ],
    toolName: "notifications",
  },
  {
    domain: "objective",
    operationSource: "input-action",
    operations: [
      "create-objective-status",
      "update-objective-status",
      "delete-objective-status",
      "set-default-objective-status",
    ],
    toolName: "objectiveStatuses",
  },
  {
    domain: "status",
    operationSource: "input-action",
    operations: [
      "create-status",
      "update-status",
      "delete-status",
      "set-default-status",
    ],
    toolName: "statuses",
  },
  {
    domain: "story",
    operationSource: "input-action",
    operations: [
      "set-story-labels",
      "add-labels-to-story",
      "remove-labels-from-story",
    ],
    toolName: "storyLabels",
  },
] as const satisfies readonly MutationRoutingEntry[];

export type MutationRoute = (typeof MUTATION_ROUTING_MANIFEST)[number];
export type MutationOperation = MutationRoute["operations"][number];
export type StandaloneMutationRoute = Extract<
  MutationRoute,
  { operationSource: "tool" }
>;
export type ActionScopedMutationRoute = Extract<
  MutationRoute,
  { operationSource: "input-action" }
>;

const mutationRouteByToolName = new Map<string, MutationRoute>(
  MUTATION_ROUTING_MANIFEST.map((entry) => [entry.toolName, entry]),
);
const mutationOperationSet = new Set<string>(
  MUTATION_ROUTING_MANIFEST.flatMap(({ operations }) => operations),
);

export const getMutationRoute = (toolName: string) =>
  mutationRouteByToolName.get(toolName);

export const isMutationOperation = (
  operation: string,
): operation is MutationOperation => mutationOperationSet.has(operation);

export const getMutationRoutesByDomainOperation = (
  domain: MayaToolDomain,
  operation: MutationOperation,
) =>
  MUTATION_ROUTING_MANIFEST.filter(
    (entry) =>
      entry.domain === domain &&
      (entry.operations as readonly string[]).includes(operation),
  );
