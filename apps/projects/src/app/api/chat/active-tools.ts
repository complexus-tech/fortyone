import type { UIMessage } from "ai";
import {
  isMutationCapableToolName,
  type MayaToolName,
} from "@/lib/ai/tool-policy";

const BASE_TOOLS = ["suggestions"] as const satisfies readonly MayaToolName[];

const STORY_READ_TOOLS = [
  "listTeamStories",
  "searchStories",
  "getStoryDetails",
  "statuses",
  "listTeams",
  "getTeamDetails",
  "listTeamMembers",
  "resolveMember",
  "listSprints",
  "listRunningSprints",
  "listTeamObjectivesTool",
  "listKeyResultsTool",
  "labels",
] as const satisfies readonly MayaToolName[];

const STORY_CREATE_TOOLS = [
  "createStory",
  "bulkCreateStories",
] as const satisfies readonly MayaToolName[];

const STORY_UPDATE_TOOLS = [
  "updateStory",
  "bulkUpdateStories",
  "assignStoriesToUser",
] as const satisfies readonly MayaToolName[];

const STORY_DELETE_TOOLS = [
  "deleteStory",
  "bulkDeleteStories",
] as const satisfies readonly MayaToolName[];

const TEAM_READ_TOOLS = [
  "members",
  "resolveMember",
  "listTeams",
  "listPublicTeams",
  "getTeamDetails",
  "listTeamMembers",
  "getTeamSettingsTool",
] as const satisfies readonly MayaToolName[];

const SPRINT_READ_TOOLS = [
  "listSprints",
  "listRunningSprints",
  "getSprintDetailsTool",
  "getSprintAnalyticsTool",
] as const satisfies readonly MayaToolName[];

const OBJECTIVE_READ_TOOLS = [
  "listObjectivesTool",
  "listTeamObjectivesTool",
  "getObjectiveDetailsTool",
  "getObjectiveActivitiesTool",
  "objectiveAnalyticsTool",
  "objectiveStatuses",
  "listKeyResultsTool",
  "getKeyResultActivitiesTool",
] as const satisfies readonly MayaToolName[];

const OBJECTIVE_WRITE_TOOLS = [
  "createObjectiveTool",
  "updateObjectiveTool",
  "deleteObjectiveTool",
  "createKeyResultTool",
  "updateKeyResultTool",
  "deleteKeyResultTool",
] as const satisfies readonly MayaToolName[];

const ANALYTICS_SUPPORT_TOOLS = [
  "listTeams",
  "listTeamMembers",
  "resolveMember",
  "listSprints",
  "listTeamStories",
  "listTeamObjectivesTool",
] as const satisfies readonly MayaToolName[];

const GITHUB_READ_TOOLS = [
  "getGitHubIntegrationTool",
  "getGitHubTeamSettingsTool",
  "getStoryGitHubLinksTool",
  "getStoryGitHubCommentsTool",
] as const satisfies readonly MayaToolName[];

const GITHUB_WRITE_TOOLS = [
  "createGitHubInstallSessionTool",
  "resyncGitHubRepositoriesTool",
  "createGitHubIssueSyncLinkTool",
  "deleteGitHubIssueSyncLinkTool",
  "updateGitHubWorkspaceSettingsTool",
  "updateGitHubTeamSettingsTool",
  "postStoryGitHubCommentTool",
  "deleteStoryGitHubLinkTool",
] as const satisfies readonly MayaToolName[];

const INTEGRATION_REQUEST_READ_TOOLS = [
  "listIntegrationRequestsTool",
  "getIntegrationRequestTool",
  "getRequestGitHubCommentsTool",
] as const satisfies readonly MayaToolName[];

const INTEGRATION_REQUEST_WRITE_TOOLS = [
  "updateIntegrationRequestTool",
  "acceptIntegrationRequestTool",
  "declineIntegrationRequestTool",
  "acceptAllIntegrationRequestsTool",
  "declineAllIntegrationRequestsTool",
  "postRequestGitHubCommentTool",
] as const satisfies readonly MayaToolName[];

const DEFAULT_DISCOVERY_TOOLS = [
  "focusBrief",
  "listTeamStories",
  "listTeams",
  "listObjectivesTool",
  "listSprints",
  "notifications",
] as const satisfies readonly MayaToolName[];

const ANALYTICS_PATH_TOOLS = [
  ...ANALYTICS_SUPPORT_TOOLS,
  "workspaceCommandCenterReportTool",
] as const satisfies readonly MayaToolName[];

const RECENT_TOOL_MESSAGES = 8;

const STORY_PATTERN =
  /\b(?:story|stories|task|tasks|backlog|assignee|priority|estimate)\b/;
const STORY_CREATE_PATTERN =
  /\b(?:create|add|draft|make|new|bulk create)\b[^\n]{0,80}\b(?:story|stories|task|tasks)\b|\b(?:story|stories|task|tasks)\b[^\n]{0,80}\b(?:create|add|draft|make|new)\b/;
const TEAM_PATTERN = /\b(?:team|teams|member|members|people)\b/;
const SPRINT_PATTERN = /\b(?:sprint|sprints|cycle|iteration)\b/;
const OBJECTIVE_PATTERN =
  /\b(?:objective|objectives|okr|okrs|key result|key results|goal|goals)\b/;
const ANALYTICS_PATTERN =
  /\b(?:analytics|report|reports|performance|trend|trends|pulse|workload|capacity|dashboard|command center)\b/;
const FOCUS_PATTERN =
  /\b(?:focus|prioritize|priorities|needs? attention|work on today|work on next)\b/;
const PLANNING_PATTERN =
  /\b(?:calendar|schedule|scheduling|focus block|time block|plan my work)\b/;
const GITHUB_PATTERN =
  /\b(?:github|repository|repositories|pull request|pull requests)\b/;
const INTEGRATION_REQUEST_PATTERN =
  /\b(?:integration request|integration requests)\b/;
const FEEDBACK_PATTERN = /\b(?:customer feedback|feedback)\b/;
const DOCUMENT_PATTERN = /\b(?:document|documents)\b/;
const MEMORY_PATTERN = /\b(?:memory|memories|remember|forget)\b/;
const NOTIFICATION_PATTERN = /\b(?:notification|notifications)\b/;
const ACTIVITY_PATTERN =
  /\b(?:activity|activities|what changed|recent changes|change history)\b/;
const COMMENT_PATTERN = /\b(?:comment|comments|reply|replies)\b/;
const LABEL_PATTERN = /\b(?:label|labels|tag|tags)\b/;
const LINK_PATTERN = /\b(?:link|links|url|urls)\b/;
const ATTACHMENT_PATTERN = /\b(?:attachment|attachments|file|files)\b/;
const NAVIGATION_PATTERN = /\b(?:navigate|open|take me|go to|show me where)\b/;
const THEME_PATTERN = /\b(?:theme|dark mode|light mode)\b/;
const SEARCH_PATTERN =
  /\b(?:search|find|look up|compare|comparison|duplicate check|review)\b/;
const CREATE_PATTERN = /\b(?:create|add|new|join|install|connect)\b/;
const UPDATE_PATTERN =
  /\b(?:update|edit|change|rename|move|assign|set|configure|resync)\b/;
const DELETE_PATTERN = /\b(?:delete|remove|decline|leave|disconnect|unlink)\b/;
const MUTATION_PATTERN =
  /\b(?:create|add|update|edit|change|rename|move|assign|set|delete|remove|restore|duplicate|join|leave|accept|decline|post|resync|install|connect|unlink|forget)\b/;

const PATH_DOMAINS = [
  { pattern: /\/stories(?:\/|$)|\/my-work(?:\/|$)/, tools: STORY_READ_TOOLS },
  { pattern: /\/sprints(?:\/|$)/, tools: SPRINT_READ_TOOLS },
  { pattern: /\/objectives(?:\/|$)/, tools: OBJECTIVE_READ_TOOLS },
  { pattern: /\/requests(?:\/|$)/, tools: INTEGRATION_REQUEST_READ_TOOLS },
  { pattern: /\/feedback(?:\/|$)/, tools: ["listCustomerFeedbackTool"] },
  { pattern: /\/(?:docs|documents)(?:\/|$)/, tools: ["listDocumentsTool"] },
  { pattern: /\/analytics(?:\/|$)/, tools: ANALYTICS_PATH_TOOLS },
  { pattern: /\/notifications(?:\/|$)/, tools: ["notifications"] },
  { pattern: /\/teams(?:\/|$)/, tools: TEAM_READ_TOOLS },
] as const;

const addTools = (
  selectedTools: Set<MayaToolName>,
  toolNames: readonly MayaToolName[],
) => {
  toolNames.forEach((toolName) => selectedTools.add(toolName));
};

const getLatestUserText = (messages: UIMessage[]) => {
  const message = messages.findLast((candidate) => candidate.role === "user");
  if (!message) return "";

  return message.parts
    .flatMap((part) => (part.type === "text" ? [part.text] : []))
    .join(" ")
    .toLowerCase();
};

const getPendingMutationTools = (messages: UIMessage[]) => {
  const pendingTools = new Set<MayaToolName>();

  for (const message of messages.slice(-RECENT_TOOL_MESSAGES)) {
    for (const part of message.parts) {
      if (!part.type.startsWith("tool-") || !("output" in part)) continue;

      const output = part.output;
      const needsConfirmation =
        output &&
        typeof output === "object" &&
        "needsConfirmation" in output &&
        output.needsConfirmation === true;
      const toolName = part.type.slice("tool-".length);

      if (needsConfirmation && isMutationCapableToolName(toolName)) {
        pendingTools.add(toolName as MayaToolName);
      }
    }
  }

  return pendingTools;
};

const addAnalyticsTools = (
  selectedTools: Set<MayaToolName>,
  intent: string,
) => {
  addTools(selectedTools, ANALYTICS_SUPPORT_TOOLS);

  if (/\bcommand center|dashboard\b/.test(intent)) {
    selectedTools.add("workspaceCommandCenterReportTool");
  } else if (/\bpulse\b/.test(intent)) {
    selectedTools.add("pulseReportTool");
  } else if (/\bsprint|cycle|iteration\b/.test(intent)) {
    selectedTools.add("sprintPerformanceReportTool");
  } else if (/\bobjective|okr|key result|goal\b/.test(intent)) {
    selectedTools.add("objectiveProgressReportTool");
  } else if (/\bteam|member|person|people\b/.test(intent)) {
    selectedTools.add("teamPerformanceReportTool");
  } else if (/\bstory|stories|task|tasks|backlog\b/.test(intent)) {
    selectedTools.add("storyPerformanceReportTool");
  } else if (/\btrend|trends|timeline\b/.test(intent)) {
    selectedTools.add("timelineTrendsReportTool");
  } else if (/\bworkload|capacity\b/.test(intent)) {
    selectedTools.add("workloadPlanningTool");
  } else if (/\bworkspace\b/.test(intent)) {
    selectedTools.add("workspacePerformanceReportTool");
  } else {
    selectedTools.add("workspaceCommandCenterReportTool");
  }
};

export const selectActiveTools = ({
  currentPath = "",
  messages,
}: {
  currentPath?: string;
  messages: UIMessage[];
}): MayaToolName[] => {
  const intent = getLatestUserText(messages);
  const selectedTools = new Set<MayaToolName>(BASE_TOOLS);
  const pendingTools = getPendingMutationTools(messages);
  addTools(selectedTools, Array.from(pendingTools));

  let matchedDomain = pendingTools.size > 0;
  const isStoryIntent = STORY_PATTERN.test(intent);
  const isStoryAction = isStoryIntent && MUTATION_PATTERN.test(intent);

  if (isStoryIntent) {
    matchedDomain = true;
    addTools(selectedTools, STORY_READ_TOOLS);

    if (STORY_CREATE_PATTERN.test(intent))
      addTools(selectedTools, STORY_CREATE_TOOLS);
    if (UPDATE_PATTERN.test(intent))
      addTools(selectedTools, STORY_UPDATE_TOOLS);
    if (DELETE_PATTERN.test(intent))
      addTools(selectedTools, STORY_DELETE_TOOLS);
    if (/\bduplicate\b/.test(intent)) selectedTools.add("duplicateStory");
    if (/\brestore\b/.test(intent)) selectedTools.add("restoreStory");
    if (/\bassociation|associate\b/.test(intent)) {
      selectedTools.add("addStoryAssociation");
      selectedTools.add("removeStoryAssociation");
    }
    if (/\bcomment|reply\b/.test(intent)) selectedTools.add("comments");
    if (/\blabel|tag\b/.test(intent)) selectedTools.add("storyLabels");
    if (/\blink|url\b/.test(intent)) selectedTools.add("links");
    if (/\battachment|file\b/.test(intent)) {
      selectedTools.add("listAttachments");
      if (DELETE_PATTERN.test(intent)) selectedTools.add("deleteAttachment");
    }
  }

  if (!isStoryAction && TEAM_PATTERN.test(intent)) {
    matchedDomain = true;
    addTools(selectedTools, TEAM_READ_TOOLS);
    if (CREATE_PATTERN.test(intent)) {
      selectedTools.add("createTeamTool");
      selectedTools.add("joinTeam");
    }
    if (UPDATE_PATTERN.test(intent)) selectedTools.add("updateTeam");
    if (DELETE_PATTERN.test(intent)) {
      selectedTools.add("deleteTeam");
      selectedTools.add("leaveTeam");
    }
  }

  if (!isStoryAction && SPRINT_PATTERN.test(intent)) {
    matchedDomain = true;
    addTools(selectedTools, SPRINT_READ_TOOLS);
    if (MUTATION_PATTERN.test(intent))
      selectedTools.add("updateSprintSettings");
  }

  if (!isStoryAction && OBJECTIVE_PATTERN.test(intent)) {
    matchedDomain = true;
    addTools(selectedTools, OBJECTIVE_READ_TOOLS);
    if (MUTATION_PATTERN.test(intent))
      addTools(selectedTools, OBJECTIVE_WRITE_TOOLS);
  }

  if (FOCUS_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("focusBrief");
    selectedTools.add("resolveMember");
  } else if (ANALYTICS_PATTERN.test(intent)) {
    matchedDomain = true;
    addAnalyticsTools(selectedTools, intent);
  }

  if (PLANNING_PATTERN.test(intent)) {
    matchedDomain = true;
    addTools(selectedTools, [
      "mayaWorkPlanTool",
      "workloadPlanningTool",
      "focusBrief",
      "listTeamStories",
      "listTeamMembers",
      "resolveMember",
    ]);
  }

  if (GITHUB_PATTERN.test(intent)) {
    matchedDomain = true;
    addTools(selectedTools, GITHUB_READ_TOOLS);
    if (MUTATION_PATTERN.test(intent))
      addTools(selectedTools, GITHUB_WRITE_TOOLS);
  }

  if (INTEGRATION_REQUEST_PATTERN.test(intent)) {
    matchedDomain = true;
    addTools(selectedTools, INTEGRATION_REQUEST_READ_TOOLS);
    if (MUTATION_PATTERN.test(intent)) {
      addTools(selectedTools, INTEGRATION_REQUEST_WRITE_TOOLS);
    }
  }

  if (FEEDBACK_PATTERN.test(intent)) {
    matchedDomain = true;
    addTools(selectedTools, [
      "listCustomerFeedbackTool",
      "getCustomerFeedbackTool",
    ]);
  }

  if (DOCUMENT_PATTERN.test(intent)) {
    matchedDomain = true;
    addTools(selectedTools, ["listDocumentsTool", "getDocumentDetailsTool"]);
  }

  if (MEMORY_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("listMemories");
    if (/\bremember|create|save|add\b/.test(intent))
      selectedTools.add("createMemory");
    if (/\bupdate|edit|change\b/.test(intent))
      selectedTools.add("updateMemory");
    if (/\bforget|delete|remove\b/.test(intent))
      selectedTools.add("deleteMemory");
  }

  if (NOTIFICATION_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("notifications");
  }

  if (ACTIVITY_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("activitySummaryTool");
  }

  if (!isStoryIntent && COMMENT_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("comments");
  }

  if (!isStoryIntent && LABEL_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("labels");
  }

  if (!isStoryIntent && LINK_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("links");
  }

  if (!isStoryIntent && ATTACHMENT_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("listAttachments");
    if (DELETE_PATTERN.test(intent)) selectedTools.add("deleteAttachment");
  }

  if (NAVIGATION_PATTERN.test(intent)) selectedTools.add("navigation");
  if (THEME_PATTERN.test(intent)) selectedTools.add("theme");
  if (SEARCH_PATTERN.test(intent)) selectedTools.add("search");

  if (!matchedDomain) {
    const path = currentPath.toLowerCase();
    const pathDomain = PATH_DOMAINS.find(({ pattern }) => pattern.test(path));
    if (pathDomain) addTools(selectedTools, pathDomain.tools);
    else addTools(selectedTools, DEFAULT_DISCOVERY_TOOLS);
  }

  return Array.from(selectedTools);
};
