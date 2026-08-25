import type { UIMessage } from "ai";
import type { tools } from "@/lib/ai/tools";

export type MayaToolName = keyof typeof tools;

const CORE_TOOLS = [
  "navigation",
  "theme",
  "suggestions",
  "search",
] as const satisfies readonly MayaToolName[];

const STORY_CREATION_TOOLS = [
  "createStory",
  "bulkCreateStories",
  "listTeamStories",
  "searchStories",
  "listTeams",
  "listPublicTeams",
  "getTeamDetails",
  "listTeamMembers",
  "members",
  "resolveMember",
  "statuses",
  "listSprints",
  "listRunningSprints",
  "listTeamObjectivesTool",
  "listKeyResultsTool",
  "labels",
] as const satisfies readonly MayaToolName[];

const STORY_TOOLS = [
  ...STORY_CREATION_TOOLS,
  "getStoryDetails",
  "updateStory",
  "deleteStory",
  "bulkUpdateStories",
  "bulkDeleteStories",
  "assignStoriesToUser",
  "duplicateStory",
  "restoreStory",
  "addStoryAssociation",
  "removeStoryAssociation",
  "comments",
  "storyActivities",
  "storyLabels",
  "links",
  "listAttachments",
  "deleteAttachment",
] as const satisfies readonly MayaToolName[];

const TEAM_TOOLS = [
  "members",
  "resolveMember",
  "listTeams",
  "listPublicTeams",
  "getTeamDetails",
  "listTeamMembers",
  "createTeamTool",
  "updateTeam",
  "joinTeam",
  "deleteTeam",
  "leaveTeam",
  "getTeamSettingsTool",
] as const satisfies readonly MayaToolName[];

const SPRINT_TOOLS = [
  "listSprints",
  "listRunningSprints",
  "getSprintDetailsTool",
  "getSprintAnalyticsTool",
  "updateSprintSettings",
  "listTeamStories",
] as const satisfies readonly MayaToolName[];

const OBJECTIVE_TOOLS = [
  "listObjectivesTool",
  "listTeamObjectivesTool",
  "createObjectiveTool",
  "updateObjectiveTool",
  "deleteObjectiveTool",
  "objectiveAnalyticsTool",
  "getObjectiveDetailsTool",
  "getObjectiveActivitiesTool",
  "objectiveStatuses",
  "listKeyResultsTool",
  "createKeyResultTool",
  "updateKeyResultTool",
  "deleteKeyResultTool",
  "getKeyResultActivitiesTool",
] as const satisfies readonly MayaToolName[];

const ANALYTICS_TOOLS = [
  "workspacePerformanceReportTool",
  "workspaceCommandCenterReportTool",
  "pulseReportTool",
  "storyPerformanceReportTool",
  "objectiveProgressReportTool",
  "teamPerformanceReportTool",
  "sprintPerformanceReportTool",
  "timelineTrendsReportTool",
  "workloadPlanningTool",
  "focusBrief",
  "activitySummaryTool",
  "listTeamStories",
  "listTeamObjectivesTool",
  "listSprints",
  "listTeamMembers",
] as const satisfies readonly MayaToolName[];

const PLANNING_TOOLS = [
  "mayaWorkPlanTool",
  "workloadPlanningTool",
  "focusBrief",
  "listTeamStories",
  "searchStories",
  "listTeamMembers",
  "members",
  "resolveMember",
] as const satisfies readonly MayaToolName[];

const GITHUB_TOOLS = [
  "getGitHubIntegrationTool",
  "createGitHubInstallSessionTool",
  "resyncGitHubRepositoriesTool",
  "createGitHubIssueSyncLinkTool",
  "deleteGitHubIssueSyncLinkTool",
  "updateGitHubWorkspaceSettingsTool",
  "getGitHubTeamSettingsTool",
  "updateGitHubTeamSettingsTool",
  "getStoryGitHubLinksTool",
  "getStoryGitHubCommentsTool",
  "postStoryGitHubCommentTool",
  "deleteStoryGitHubLinkTool",
] as const satisfies readonly MayaToolName[];

const INTEGRATION_REQUEST_TOOLS = [
  "listIntegrationRequestsTool",
  "getIntegrationRequestTool",
  "updateIntegrationRequestTool",
  "acceptIntegrationRequestTool",
  "declineIntegrationRequestTool",
  "acceptAllIntegrationRequestsTool",
  "declineAllIntegrationRequestsTool",
  "getRequestGitHubCommentsTool",
  "postRequestGitHubCommentTool",
] as const satisfies readonly MayaToolName[];

const FEEDBACK_TOOLS = [
  "listCustomerFeedbackTool",
  "getCustomerFeedbackTool",
] as const satisfies readonly MayaToolName[];

const DOCUMENT_TOOLS = [
  "listDocumentsTool",
  "getDocumentDetailsTool",
] as const satisfies readonly MayaToolName[];

const MEMORY_TOOLS = [
  "listMemories",
  "createMemory",
  "updateMemory",
  "deleteMemory",
] as const satisfies readonly MayaToolName[];

const NOTIFICATION_TOOLS = [
  "notifications",
] as const satisfies readonly MayaToolName[];

const DEFAULT_DISCOVERY_TOOLS = [
  "listTeamStories",
  "listTeams",
  "listObjectivesTool",
  "listSprints",
  "notifications",
  "focusBrief",
] as const satisfies readonly MayaToolName[];

const RECENT_ROUTING_MESSAGES = 8;

const STORY_CREATION_PATTERNS = [
  /\b(?:create|add|draft|make|new|bulk)\b[^\n]{0,60}\b(?:story|stories|task|tasks)\b/,
  /\b(?:story|stories|task|tasks)\b[^\n]{0,60}\b(?:create|add|draft|make|new|bulk)\b/,
] as const;
const STORY_PATTERNS = [
  /\b(?:story|stories|task|tasks|backlog|assignee|priority|estimate)\b/,
  /\/stories(?:\/|$)/,
  /\/my-work(?:\/|$)/,
] as const;
const TEAM_PATTERNS = [
  /\b(?:team|teams|member|members|people|workspace)\b/,
  /\/settings\/teams(?:\/|$)/,
] as const;
const SPRINT_PATTERNS = [
  /\b(?:sprint|sprints|cycle|iteration)\b/,
  /\/sprints(?:\/|$)/,
] as const;
const OBJECTIVE_PATTERNS = [
  /\b(?:objective|objectives|okr|okrs|key result|key results|goal|goals)\b/,
  /\/objectives(?:\/|$)/,
] as const;
const ANALYTICS_PATTERNS = [
  /\b(?:analytics|report|reports|performance|trend|trends|pulse|workload|summary|brief|attention|priorities)\b/,
  /\/analytics(?:\/|$)/,
] as const;
const PLANNING_PATTERNS = [
  /\b(?:calendar|schedule|scheduling|focus block|time block|plan my work)\b/,
  /\/calendar(?:\/|$)/,
] as const;
const GITHUB_PATTERNS = [
  /\b(?:github|repository|repositories|pull request|pull requests)\b/,
  /\/github(?:\/|$)/,
] as const;
const INTEGRATION_REQUEST_PATTERNS = [
  /\b(?:integration request|integration requests)\b/,
  /\/requests(?:\/|$)/,
] as const;
const FEEDBACK_PATTERNS = [
  /\b(?:customer feedback|feedback)\b/,
  /\/feedback(?:\/|$)/,
] as const;
const DOCUMENT_PATTERNS = [
  /\b(?:document|documents)\b/,
  /\/documents(?:\/|$)/,
] as const;
const MEMORY_PATTERNS = [/\b(?:memory|memories|remember|forget)\b/] as const;
const NOTIFICATION_PATTERNS = [/\b(?:notification|notifications)\b/] as const;

const addTools = (
  selectedTools: Set<MayaToolName>,
  toolNames: readonly MayaToolName[],
) => {
  toolNames.forEach((toolName) => selectedTools.add(toolName));
};

const getRoutingContext = (messages: UIMessage[], currentPath: string) => {
  const values = [currentPath];

  for (const message of messages.slice(-RECENT_ROUTING_MESSAGES)) {
    for (const part of message.parts) {
      values.push(part.type);
      if (part.type === "text") values.push(part.text);
    }
  }

  return values.join(" ").toLowerCase();
};

const includesAny = (context: string, patterns: readonly RegExp[]) =>
  patterns.some((pattern) => pattern.test(context));

export const selectActiveTools = ({
  currentPath,
  messages,
}: {
  currentPath: string;
  messages: UIMessage[];
}): MayaToolName[] => {
  const context = getRoutingContext(messages, currentPath);
  const selectedTools = new Set<MayaToolName>(CORE_TOOLS);

  const addDomain = (toolNames: readonly MayaToolName[]) => {
    addTools(selectedTools, toolNames);
  };

  const isStoryCreation =
    context.includes("tool-bulkcreatestories") ||
    context.includes("tool-createstory") ||
    includesAny(context, STORY_CREATION_PATTERNS);

  if (isStoryCreation) {
    addDomain(STORY_CREATION_TOOLS);
  } else if (includesAny(context, STORY_PATTERNS)) {
    addDomain(STORY_TOOLS);
  }

  if (!isStoryCreation && includesAny(context, TEAM_PATTERNS)) {
    addDomain(TEAM_TOOLS);
  }

  if (!isStoryCreation && includesAny(context, SPRINT_PATTERNS)) {
    addDomain(SPRINT_TOOLS);
  }

  if (!isStoryCreation && includesAny(context, OBJECTIVE_PATTERNS)) {
    addDomain(OBJECTIVE_TOOLS);
  }

  if (includesAny(context, ANALYTICS_PATTERNS)) {
    addDomain(ANALYTICS_TOOLS);
  }

  if (includesAny(context, PLANNING_PATTERNS)) {
    addDomain(PLANNING_TOOLS);
  }

  if (includesAny(context, GITHUB_PATTERNS)) {
    addDomain(GITHUB_TOOLS);
  }

  if (includesAny(context, INTEGRATION_REQUEST_PATTERNS)) {
    addDomain(INTEGRATION_REQUEST_TOOLS);
  }

  if (includesAny(context, FEEDBACK_PATTERNS)) {
    addDomain(FEEDBACK_TOOLS);
  }

  if (includesAny(context, DOCUMENT_PATTERNS)) {
    addDomain(DOCUMENT_TOOLS);
  }

  if (includesAny(context, MEMORY_PATTERNS)) {
    addDomain(MEMORY_TOOLS);
  }

  if (includesAny(context, NOTIFICATION_PATTERNS)) {
    addDomain(NOTIFICATION_TOOLS);
  }

  if (selectedTools.size === CORE_TOOLS.length) {
    addTools(selectedTools, DEFAULT_DISCOVERY_TOOLS);
  }

  return Array.from(selectedTools);
};
