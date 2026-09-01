import type { MayaToolName } from "./tool-names";
import {
  getMutationRoute,
  MUTATION_ROUTING_MANIFEST,
  type ActionScopedMutationRoute,
  type MutationRoute,
  type StandaloneMutationRoute,
} from "./mutation-routing-manifest";

export type { MayaToolName } from "./tool-names";
export type {
  MutationOperation,
  MutationRoute,
} from "./mutation-routing-manifest";
export {
  getMutationRoute,
  getMutationRoutesByDomainOperation,
  isMutationOperation,
  MUTATION_ROUTING_MANIFEST,
} from "./mutation-routing-manifest";

const isStandaloneMutationRoute = (
  entry: MutationRoute,
): entry is StandaloneMutationRoute => entry.operationSource === "tool";

const isActionScopedMutationRoute = (
  entry: MutationRoute,
): entry is ActionScopedMutationRoute =>
  entry.operationSource === "input-action";

const standaloneMutationRoutes = MUTATION_ROUTING_MANIFEST.filter(
  isStandaloneMutationRoute,
);
const actionScopedMutationRoutes = MUTATION_ROUTING_MANIFEST.filter(
  isActionScopedMutationRoute,
);

export const MUTATION_TOOL_NAMES = standaloneMutationRoutes.map(
  ({ toolName }) => toolName,
) as readonly StandaloneMutationRoute["toolName"][];

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

type MutatingActionToolName = ActionScopedMutationRoute["toolName"];

const buildMutatingToolActions = () => {
  const actions: Partial<Record<MutatingActionToolName, ReadonlySet<string>>> =
    {};
  for (const { operations, toolName } of actionScopedMutationRoutes) {
    actions[toolName] = new Set<string>(operations);
  }
  return actions as Readonly<
    Record<MutatingActionToolName, ReadonlySet<string>>
  >;
};

export const MUTATING_TOOL_ACTIONS = buildMutatingToolActions();

export const MUTATING_ACTION_TOOL_NAMES = actionScopedMutationRoutes.map(
  ({ toolName }) => toolName,
) as readonly MutatingActionToolName[];

const NON_TERMINAL_MUTATION_TOOLS = new Set(["createGitHubInstallSessionTool"]);

export const isMutationCapableToolName = (toolName: string) =>
  getMutationRoute(toolName) !== undefined;

export const isMutationToolCall = (toolName: string, input: unknown) => {
  const route = getMutationRoute(toolName);
  if (!route) return false;
  if (route.operationSource === "tool") return true;

  const action =
    input && typeof input === "object" && "action" in input
      ? input.action
      : undefined;
  return (
    typeof action === "string" &&
    (route.operations as readonly string[]).includes(action)
  );
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
