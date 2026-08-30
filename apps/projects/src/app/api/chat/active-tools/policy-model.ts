import "server-only";

import type { MayaToolName } from "@/lib/ai/tool-policy";

export const BASE_TOOLS = [
  "suggestions",
] as const satisfies readonly MayaToolName[];

export const STORY_READ_TOOLS = [
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

export const STORY_CREATE_TOOLS = [
  "createStory",
  "bulkCreateStories",
] as const satisfies readonly MayaToolName[];

export const STORY_UPDATE_TOOLS = [
  "updateStory",
  "bulkUpdateStories",
  "assignStoriesToUser",
] as const satisfies readonly MayaToolName[];

export const STORY_DELETE_TOOLS = [
  "deleteStory",
  "bulkDeleteStories",
] as const satisfies readonly MayaToolName[];

const STORY_PROVENANCE_TOOLS = new Set<MayaToolName>([
  "listTeamStories",
  "searchStories",
  "getStoryDetails",
  "storyActivities",
  ...STORY_CREATE_TOOLS,
  ...STORY_UPDATE_TOOLS,
  ...STORY_DELETE_TOOLS,
  "duplicateStory",
  "restoreStory",
  "addStoryAssociation",
  "removeStoryAssociation",
  "storyLabels",
  "getStoryGitHubLinksTool",
  "getStoryGitHubCommentsTool",
  "postStoryGitHubCommentTool",
  "deleteStoryGitHubLinkTool",
]);

export const TEAM_READ_TOOLS = [
  "members",
  "resolveMember",
  "listTeams",
  "listPublicTeams",
  "getTeamDetails",
  "listTeamMembers",
  "getTeamSettingsTool",
] as const satisfies readonly MayaToolName[];

export const SPRINT_READ_TOOLS = [
  "listSprints",
  "listRunningSprints",
  "getSprintDetailsTool",
  "getSprintAnalyticsTool",
] as const satisfies readonly MayaToolName[];

export const OBJECTIVE_READ_TOOLS = [
  "listObjectivesTool",
  "listTeamObjectivesTool",
  "getObjectiveDetailsTool",
  "getObjectiveActivitiesTool",
  "objectiveAnalyticsTool",
  "objectiveStatuses",
  "listKeyResultsTool",
  "getKeyResultActivitiesTool",
] as const satisfies readonly MayaToolName[];

export const OBJECTIVE_WRITE_TOOLS = [
  "createObjectiveTool",
  "updateObjectiveTool",
  "deleteObjectiveTool",
  "createKeyResultTool",
  "updateKeyResultTool",
  "deleteKeyResultTool",
] as const satisfies readonly MayaToolName[];

export const ANALYTICS_SUPPORT_TOOLS = [
  "listTeams",
  "listTeamMembers",
  "resolveMember",
  "listSprints",
  "listTeamStories",
  "listTeamObjectivesTool",
] as const satisfies readonly MayaToolName[];

export const GITHUB_READ_TOOLS = [
  "getGitHubIntegrationTool",
  "getGitHubTeamSettingsTool",
  "getStoryGitHubLinksTool",
  "getStoryGitHubCommentsTool",
] as const satisfies readonly MayaToolName[];

export const GITHUB_WRITE_TOOLS = [
  "createGitHubInstallSessionTool",
  "resyncGitHubRepositoriesTool",
  "createGitHubIssueSyncLinkTool",
  "deleteGitHubIssueSyncLinkTool",
  "updateGitHubWorkspaceSettingsTool",
  "updateGitHubTeamSettingsTool",
  "postStoryGitHubCommentTool",
  "deleteStoryGitHubLinkTool",
] as const satisfies readonly MayaToolName[];

export const INTEGRATION_REQUEST_READ_TOOLS = [
  "listIntegrationRequestsTool",
  "getIntegrationRequestTool",
  "getRequestGitHubCommentsTool",
] as const satisfies readonly MayaToolName[];

export const INTEGRATION_REQUEST_WRITE_TOOLS = [
  "updateIntegrationRequestTool",
  "acceptIntegrationRequestTool",
  "declineIntegrationRequestTool",
  "acceptAllIntegrationRequestsTool",
  "declineAllIntegrationRequestsTool",
  "postRequestGitHubCommentTool",
] as const satisfies readonly MayaToolName[];

export type ToolDomain =
  | "attachment"
  | "comment"
  | "document"
  | "feedback"
  | "github"
  | "integration-request"
  | "label"
  | "link"
  | "memory"
  | "notification"
  | "objective"
  | "planning"
  | "sprint"
  | "status"
  | "story"
  | "team";

export const TOOL_DOMAIN_PROVENANCE = [
  {
    domain: "integration-request",
    tools: new Set<MayaToolName>([
      ...INTEGRATION_REQUEST_READ_TOOLS,
      ...INTEGRATION_REQUEST_WRITE_TOOLS,
    ]),
  },
  {
    domain: "github",
    tools: new Set<MayaToolName>([...GITHUB_READ_TOOLS, ...GITHUB_WRITE_TOOLS]),
  },
  {
    domain: "planning",
    tools: new Set<MayaToolName>(["mayaWorkPlanTool", "applyMayaWorkPlanTool"]),
  },
  {
    domain: "memory",
    tools: new Set<MayaToolName>([
      "listMemories",
      "createMemory",
      "updateMemory",
      "deleteMemory",
    ]),
  },
  {
    domain: "notification",
    tools: new Set<MayaToolName>(["notifications"]),
  },
  {
    domain: "comment",
    tools: new Set<MayaToolName>(["comments"]),
  },
  {
    domain: "label",
    tools: new Set<MayaToolName>(["labels"]),
  },
  {
    domain: "link",
    tools: new Set<MayaToolName>(["links"]),
  },
  {
    domain: "attachment",
    tools: new Set<MayaToolName>(["listAttachments", "deleteAttachment"]),
  },
  {
    domain: "objective",
    tools: new Set<MayaToolName>([
      ...OBJECTIVE_READ_TOOLS,
      ...OBJECTIVE_WRITE_TOOLS,
    ]),
  },
  {
    domain: "sprint",
    tools: new Set<MayaToolName>([
      ...SPRINT_READ_TOOLS,
      "updateSprintSettings",
    ]),
  },
  {
    domain: "status",
    tools: new Set<MayaToolName>(["statuses"]),
  },
  {
    domain: "team",
    tools: new Set<MayaToolName>([
      "members",
      "resolveMember",
      "listTeams",
      "listPublicTeams",
      "getTeamDetails",
      "listTeamMembers",
      "getTeamSettingsTool",
      "createTeamTool",
      "updateTeam",
      "joinTeam",
      "deleteTeam",
      "leaveTeam",
    ]),
  },
  {
    domain: "story",
    tools: STORY_PROVENANCE_TOOLS,
  },
  {
    domain: "feedback",
    tools: new Set<MayaToolName>([
      "listCustomerFeedbackTool",
      "getCustomerFeedbackTool",
    ]),
  },
  {
    domain: "document",
    tools: new Set<MayaToolName>([
      "listDocumentsTool",
      "getDocumentDetailsTool",
    ]),
  },
] as const satisfies readonly {
  domain: ToolDomain;
  tools: ReadonlySet<MayaToolName>;
}[];

export const DEFAULT_DISCOVERY_TOOLS = [
  "focusBrief",
  "listTeamStories",
  "listTeams",
  "listObjectivesTool",
  "listSprints",
  "notifications",
] as const satisfies readonly MayaToolName[];

export const ANALYTICS_PATH_TOOLS = [
  ...ANALYTICS_SUPPORT_TOOLS,
  "workspaceCommandCenterReportTool",
] as const satisfies readonly MayaToolName[];

export const STORY_PATTERN =
  /\b(?:story|stories|task|tasks|ticket|tickets|issue|issues|backlog|assignee|priority|estimate)\b/;
export const STORY_CREATE_PATTERN =
  /\b(?:create|add|draft|make|new|prepare|bulk create)\b[^\n]{0,80}\b(?:story|stories|task|tasks|ticket|tickets|issue|issues)\b|\b(?:story|stories|task|tasks|ticket|tickets|issue|issues)\b[^\n]{0,80}\b(?:create|add|draft|make|new|prepare)\b/;
export const STORY_REFERENCE_PATTERN = /\b[a-z][a-z0-9]{1,9}-\d+\b/i;
export const TEAM_PATTERN = /\b(?:team|teams|member|members|people)\b/;
export const STATUS_PATTERN =
  /\b(?:status|statuses|workflow state|workflow states)\b/;
export const SPRINT_PATTERN = /\b(?:sprint|sprints|cycle|iteration)\b/;
export const OBJECTIVE_PATTERN =
  /\b(?:objective|objectives|okr|okrs|key result|key results|goal|goals)\b/;
export const ANALYTICS_PATTERN =
  /\b(?:analyze|analyse|analysis|analytics|health|insight|insights|overview|report|reports|performance|trend|trends|pulse|workload|capacity|dashboard|command center)\b/;
export const FOCUS_PATTERN =
  /\b(?:focus|prioritize|priorities|needs? attention|work on today|work on next)\b/;
export const PLANNING_PATTERN =
  /\b(?:calendar|schedule|scheduling|focus block|time block|work plan|plan my work)\b/;
export const GITHUB_PATTERN =
  /\b(?:github|repository|repositories|pull request|pull requests)\b/;
export const INTEGRATION_REQUEST_PATTERN =
  /\b(?:integration request|integration requests)\b/;
export const FEEDBACK_PATTERN = /\b(?:customer feedback|feedback)\b/;
export const DOCUMENT_PATTERN = /\b(?:document|documents)\b/;
export const MEMORY_PATTERN = /\b(?:memory|memories|remember|forget)\b/;
export const NOTIFICATION_PATTERN = /\b(?:notification|notifications)\b/;
export const ACTIVITY_PATTERN =
  /\b(?:activity|activities|what changed|recent changes|change history)\b/;
export const COMMENT_PATTERN = /\b(?:comment|comments|reply|replies)\b/;
export const LABEL_PATTERN = /\b(?:label|labels|tag|tags)\b/;
export const LINK_PATTERN = /\b(?:link|links|url|urls)\b/;
export const ATTACHMENT_PATTERN = /\b(?:attachment|attachments|file|files)\b/;
export const NAVIGATION_PATTERN =
  /\b(?:navigate|open|take me|go to|show me where)\b/;
export const THEME_PATTERN = /\b(?:theme|dark mode|light mode)\b/;
export const SEARCH_PATTERN =
  /\b(?:search|find|look up|compare|comparison|duplicate check|review)\b/;
export const CREATE_PATTERN = /\b(?:create|add|new|join|install|connect)\b/;
export const UPDATE_PATTERN =
  /\b(?:update|edit|change|rename|move|assign|set|configure|resync|mark|complete|close|finish|reopen)\b/;
export const DELETE_PATTERN =
  /\b(?:delete|deletes|deleted|deleting|deletion|deletions|remove|removes|removed|removing|removal|removals|decline|declines|declined|declining|leave|leaves|leaving|disconnect|disconnects|disconnected|disconnecting|unlink|unlinks|unlinked|unlinking)\b/;
export const MUTATION_PATTERN =
  /\b(?:create|add|update|edit|change|rename|move|assign|set|delete|deletes|deleted|deleting|deletion|deletions|remove|removes|removed|removing|removal|removals|restore|duplicate|join|leave|accept|decline|post|resync|install|connect|unlink|forget)\b/;
export const CONVERSATIONAL_REFERENCE_PATTERN =
  /\b(?:it|its|this|that|them|these|those|same|one|ones|above|previous|earlier)\b/;
export const FOLLOW_THROUGH_ACTION_PATTERN =
  /\b(?:yes|yeah|yep|approve|approved|confirm|confirmed|proceed|go ahead|do it|do everything|make it so)\b/;
export const STORY_PLANNING_CLARIFICATION_PATTERN =
  /\b(?:delivery|deliver|due|start date|work date|time needed|how (?:long|much time)|duration|calendar|auto[- ]?schedul|focus (?:time|block)|reserve(?:d|s|ing)?(?: calendar)? time)\b/;
export const CLARIFICATION_LANGUAGE_PATTERN =
  /\?|\b(?:when|what|how|whether|would you|do you want|should|please (?:choose|provide|tell)|let me know)\b/;
export const STORY_PLANNING_VALUE_PATTERN =
  /\b(?:\d+(?:\.\d+)?\s*(?:m|min|mins|minute|minutes|h|hr|hrs|hour|hours|day|days)|\d{4}-\d{2}-\d{2}|today|tomorrow|tonight|monday|tuesday|wednesday|thursday|friday|saturday|sunday|next week|next month|calendar|auto[- ]?schedul|scheduling|focus block|reserve time|skip planning|leave (?:it|them|the rest) unscheduled)\b/;
