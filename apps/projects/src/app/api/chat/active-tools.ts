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

type ToolDomain =
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

const TOOL_DOMAIN_PROVENANCE = [
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

const STORY_PATTERN =
  /\b(?:story|stories|task|tasks|ticket|tickets|issue|issues|backlog|assignee|priority|estimate)\b/;
const STORY_CREATE_PATTERN =
  /\b(?:create|add|draft|make|new|prepare|bulk create)\b[^\n]{0,80}\b(?:story|stories|task|tasks|ticket|tickets|issue|issues)\b|\b(?:story|stories|task|tasks|ticket|tickets|issue|issues)\b[^\n]{0,80}\b(?:create|add|draft|make|new|prepare)\b/;
const STORY_REFERENCE_PATTERN = /\b[a-z][a-z0-9]{1,9}-\d+\b/i;
const TEAM_PATTERN = /\b(?:team|teams|member|members|people)\b/;
const STATUS_PATTERN = /\b(?:status|statuses|workflow state|workflow states)\b/;
const SPRINT_PATTERN = /\b(?:sprint|sprints|cycle|iteration)\b/;
const OBJECTIVE_PATTERN =
  /\b(?:objective|objectives|okr|okrs|key result|key results|goal|goals)\b/;
const ANALYTICS_PATTERN =
  /\b(?:analyze|analyse|analysis|analytics|health|insight|insights|overview|report|reports|performance|trend|trends|pulse|workload|capacity|dashboard|command center)\b/;
const FOCUS_PATTERN =
  /\b(?:focus|prioritize|priorities|needs? attention|work on today|work on next)\b/;
const PLANNING_PATTERN =
  /\b(?:calendar|schedule|scheduling|focus block|time block|work plan|plan my work)\b/;
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
  /\b(?:update|edit|change|rename|move|assign|set|configure|resync|mark|complete|close|finish|reopen)\b/;
const DELETE_PATTERN =
  /\b(?:delete|deletes|deleted|deleting|deletion|deletions|remove|removes|removed|removing|removal|removals|decline|declines|declined|declining|leave|leaves|leaving|disconnect|disconnects|disconnected|disconnecting|unlink|unlinks|unlinked|unlinking)\b/;
const MUTATION_PATTERN =
  /\b(?:create|add|update|edit|change|rename|move|assign|set|delete|deletes|deleted|deleting|deletion|deletions|remove|removes|removed|removing|removal|removals|restore|duplicate|join|leave|accept|decline|post|resync|install|connect|unlink|forget)\b/;
const CONVERSATIONAL_REFERENCE_PATTERN =
  /\b(?:it|its|this|that|them|these|those|same|one|ones|above|previous|earlier)\b/;
const FOLLOW_THROUGH_ACTION_PATTERN =
  /\b(?:yes|yeah|yep|approve|approved|confirm|confirmed|proceed|go ahead|do it|do everything|make it so)\b/;
const STORY_PLANNING_CLARIFICATION_PATTERN =
  /\b(?:delivery|deliver|due|start date|work date|time needed|how (?:long|much time)|duration|calendar|auto[- ]?schedul|focus (?:time|block)|reserve(?:d|s|ing)?(?: calendar)? time)\b/;
const CLARIFICATION_LANGUAGE_PATTERN =
  /\?|\b(?:when|what|how|whether|would you|do you want|should|please (?:choose|provide|tell)|let me know)\b/;
const STORY_PLANNING_VALUE_PATTERN =
  /\b(?:\d+(?:\.\d+)?\s*(?:m|min|mins|minute|minutes|h|hr|hrs|hour|hours|day|days)|\d{4}-\d{2}-\d{2}|today|tomorrow|tonight|monday|tuesday|wednesday|thursday|friday|saturday|sunday|next week|next month|calendar|auto[- ]?schedul|scheduling|focus block|reserve time|skip planning|leave (?:it|them|the rest) unscheduled)\b/;

const normalizeCustomStoryTerm = (term: string | undefined) => {
  if (!term) return [];

  const normalized = term
    .normalize("NFKC")
    .trim()
    .toLowerCase()
    .replace(/\s+/g, " ");
  if (
    normalized.length < 2 ||
    normalized.length > 48 ||
    !/^[\p{L}\p{N}]+(?:[ -][\p{L}\p{N}]+){0,3}$/u.test(normalized)
  ) {
    return [];
  }

  const terms = new Set([normalized]);
  if (normalized.endsWith("ies") && normalized.length > 3) {
    terms.add(`${normalized.slice(0, -3)}y`);
  } else if (normalized.endsWith("s") && normalized.length > 2) {
    terms.add(normalized.slice(0, -1));
  }

  return Array.from(terms);
};

const includesWholePhrase = (intent: string, phrase: string) => {
  const normalizedIntent = intent
    .normalize("NFKC")
    .toLowerCase()
    .replace(/[^\p{L}\p{N}-]+/gu, " ")
    .trim();
  return ` ${normalizedIntent} `.includes(` ${phrase} `);
};

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

const getLatestAssistantContext = (messages: UIMessage[]) => {
  const latestUserMessageIndex = messages.findLastIndex(
    (message) => message.role === "user",
  );
  if (latestUserMessageIndex <= 0) {
    return { precedingUserText: "", text: "", toolNames: [] };
  }

  const latestAssistantMessage = messages[latestUserMessageIndex - 1];
  if (latestAssistantMessage.role !== "assistant") {
    return { precedingUserText: "", text: "", toolNames: [] };
  }

  const precedingUserMessage = messages
    .slice(0, latestUserMessageIndex - 1)
    .findLast((message) => message.role === "user");

  return {
    precedingUserText:
      precedingUserMessage?.parts
        .flatMap((part) => (part.type === "text" ? [part.text] : []))
        .join(" ")
        .toLowerCase() ?? "",
    text: latestAssistantMessage.parts
      .flatMap((part) => (part.type === "text" ? [part.text] : []))
      .join(" ")
      .toLowerCase(),
    toolNames: latestAssistantMessage.parts.flatMap((part) =>
      part.type.startsWith("tool-")
        ? [part.type.slice("tool-".length) as MayaToolName]
        : [],
    ),
  };
};

const isStoryCreationPlanningReply = ({
  assistantText,
  configuredStoryTerms,
  intent,
  precedingUserText,
}: {
  assistantText: string;
  configuredStoryTerms: string[];
  intent: string;
  precedingUserText: string;
}) => {
  if (
    !STORY_PLANNING_VALUE_PATTERN.test(intent) ||
    !STORY_PLANNING_CLARIFICATION_PATTERN.test(assistantText) ||
    !CLARIFICATION_LANGUAGE_PATTERN.test(assistantText)
  ) {
    return false;
  }

  const assistantHasStoryContext =
    STORY_PATTERN.test(assistantText) ||
    configuredStoryTerms.some((term) =>
      includesWholePhrase(assistantText, term),
    );
  const precedingUserRequestedStoryCreation =
    STORY_CREATE_PATTERN.test(precedingUserText) ||
    configuredStoryTerms.some(
      (term) =>
        includesWholePhrase(precedingUserText, term) &&
        CREATE_PATTERN.test(precedingUserText),
    );
  const assistantReferencesItsCreationProposal =
    CREATE_PATTERN.test(assistantText) &&
    CONVERSATIONAL_REFERENCE_PATTERN.test(assistantText);

  return (
    assistantHasStoryContext ||
    precedingUserRequestedStoryCreation ||
    assistantReferencesItsCreationProposal
  );
};

const inferFollowUpDomain = ({
  assistantText,
  assistantToolNames,
  configuredStoryTerms,
  isFollowUp,
}: {
  assistantText: string;
  assistantToolNames: MayaToolName[];
  configuredStoryTerms: string[];
  isFollowUp: boolean;
}): ToolDomain | undefined => {
  if (!isFollowUp) return undefined;

  for (const toolName of assistantToolNames.toReversed()) {
    const matchedGroup = TOOL_DOMAIN_PROVENANCE.find(({ tools: toolNames }) =>
      toolNames.has(toolName),
    );
    if (matchedGroup) return matchedGroup.domain;
  }

  if (PLANNING_PATTERN.test(assistantText)) {
    return "planning";
  }

  if (
    configuredStoryTerms.some((term) =>
      includesWholePhrase(assistantText, term),
    )
  ) {
    return "story";
  }

  const textDomains = [
    { domain: "story", pattern: STORY_PATTERN },
    { domain: "integration-request", pattern: INTEGRATION_REQUEST_PATTERN },
    { domain: "github", pattern: GITHUB_PATTERN },
    { domain: "objective", pattern: OBJECTIVE_PATTERN },
    { domain: "sprint", pattern: SPRINT_PATTERN },
    { domain: "status", pattern: STATUS_PATTERN },
    { domain: "team", pattern: TEAM_PATTERN },
    { domain: "memory", pattern: MEMORY_PATTERN },
    { domain: "notification", pattern: NOTIFICATION_PATTERN },
    { domain: "comment", pattern: COMMENT_PATTERN },
    { domain: "label", pattern: LABEL_PATTERN },
    { domain: "link", pattern: LINK_PATTERN },
    { domain: "attachment", pattern: ATTACHMENT_PATTERN },
    { domain: "feedback", pattern: FEEDBACK_PATTERN },
    { domain: "document", pattern: DOCUMENT_PATTERN },
  ] as const satisfies readonly { domain: ToolDomain; pattern: RegExp }[];

  return textDomains.find(({ pattern }) => pattern.test(assistantText))?.domain;
};

const getPendingMutationTools = (messages: UIMessage[]) => {
  const pendingTools = new Set<MayaToolName>();
  const latestAssistantMessage = messages.findLast(
    (message) => message.role === "assistant",
  );
  if (!latestAssistantMessage) return pendingTools;

  for (const part of latestAssistantMessage.parts) {
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

  return pendingTools;
};

const addAnalyticsTools = (
  selectedTools: Set<MayaToolName>,
  intent: string,
) => {
  addTools(selectedTools, ANALYTICS_SUPPORT_TOOLS);
  const isBroadWorkspaceRequest =
    /\bworkspace\b/.test(intent) &&
    /\b(?:analyze|analyse|analysis|health|insight|insights|overview|report|reports)\b/.test(
      intent,
    );

  if (/\bcommand center|dashboard\b/.test(intent)) {
    selectedTools.add("workspaceCommandCenterReportTool");
  } else if (isBroadWorkspaceRequest) {
    selectedTools.add("workspaceCommandCenterReportTool");
    selectedTools.add("workspacePerformanceReportTool");
  } else if (/\bpulse\b/.test(intent)) {
    selectedTools.add("pulseReportTool");
  } else if (/\bsprint|cycle|iteration\b/.test(intent)) {
    selectedTools.add("sprintPerformanceReportTool");
  } else if (/\bobjective|okr|key result|goal\b/.test(intent)) {
    selectedTools.add("objectiveProgressReportTool");
  } else if (/\bteam|member|person|people\b/.test(intent)) {
    selectedTools.add("teamPerformanceReportTool");
  } else if (
    /\b(?:story|stories|task|tasks|ticket|tickets|issue|issues|backlog)\b/.test(
      intent,
    )
  ) {
    selectedTools.add("storyPerformanceReportTool");
  } else if (/\btrend|trends|timeline\b/.test(intent)) {
    selectedTools.add("timelineTrendsReportTool");
  } else if (/\bworkload|capacity\b/.test(intent)) {
    selectedTools.add("workloadPlanningTool");
  } else if (/\bworkspace\b/.test(intent)) {
    selectedTools.add("workspaceCommandCenterReportTool");
    selectedTools.add("workspacePerformanceReportTool");
  } else {
    selectedTools.add("workspaceCommandCenterReportTool");
  }
};

export const selectActiveTools = ({
  currentPath = "",
  messages,
  storyTerminology,
}: {
  currentPath?: string;
  messages: UIMessage[];
  storyTerminology?: string;
}): MayaToolName[] => {
  const intent = getLatestUserText(messages);
  const configuredStoryTerms = normalizeCustomStoryTerm(storyTerminology);
  const hasConfiguredStoryTerm = configuredStoryTerms.some((term) =>
    includesWholePhrase(intent, term),
  );
  const hasStoryReference = STORY_REFERENCE_PATTERN.test(intent);
  const isStoryDetailPath = /\/story\/[^/?#]+/.test(currentPath.toLowerCase());
  const usesStoryDetailContext =
    isStoryDetailPath &&
    (ACTIVITY_PATTERN.test(intent) ||
      LABEL_PATTERN.test(intent) ||
      COMMENT_PATTERN.test(intent) ||
      LINK_PATTERN.test(intent) ||
      ATTACHMENT_PATTERN.test(intent) ||
      CONVERSATIONAL_REFERENCE_PATTERN.test(intent));
  const latestAssistantContext = getLatestAssistantContext(messages);
  const selectedTools = new Set<MayaToolName>(BASE_TOOLS);
  const pendingTools = getPendingMutationTools(messages);
  addTools(selectedTools, Array.from(pendingTools));

  // Legacy confirmation turns must keep the model constrained to the exact
  // prepared mutation. Native AI SDK approvals bypass this model-selection
  // path and are validated separately against persisted chat state.
  if (pendingTools.size > 0) return Array.from(selectedTools);

  let matchedDomain = false;
  const isFollowUp =
    CONVERSATIONAL_REFERENCE_PATTERN.test(intent) ||
    FOLLOW_THROUGH_ACTION_PATTERN.test(intent);
  const inferredDomain = inferFollowUpDomain({
    assistantText: latestAssistantContext.text,
    assistantToolNames: latestAssistantContext.toolNames,
    configuredStoryTerms,
    isFollowUp,
  });
  const isGenericFollowThrough = FOLLOW_THROUGH_ACTION_PATTERN.test(intent);
  const isStoryPlanningReply = isStoryCreationPlanningReply({
    assistantText: latestAssistantContext.text,
    configuredStoryTerms,
    intent,
    precedingUserText: latestAssistantContext.precedingUserText,
  });
  const actionIntent = isGenericFollowThrough
    ? `${intent} ${latestAssistantContext.text}`
    : intent;
  const actionHasConfiguredStoryTerm = configuredStoryTerms.some((term) =>
    includesWholePhrase(actionIntent, term),
  );
  const isStoryIntent =
    STORY_PATTERN.test(intent) ||
    hasConfiguredStoryTerm ||
    inferredDomain === "story" ||
    isStoryPlanningReply ||
    hasStoryReference ||
    usesStoryDetailContext;
  const isStoryFollowThroughAction =
    pendingTools.size === 0 &&
    inferredDomain === "story" &&
    isGenericFollowThrough;
  const isStoryAction =
    isStoryIntent &&
    (STORY_CREATE_PATTERN.test(actionIntent) ||
      MUTATION_PATTERN.test(actionIntent) ||
      isStoryPlanningReply ||
      isStoryFollowThroughAction);

  if (isStoryIntent) {
    matchedDomain = true;
    addTools(selectedTools, STORY_READ_TOOLS);

    if (isStoryPlanningReply) {
      addTools(selectedTools, STORY_CREATE_TOOLS);
    } else {
      if (
        !DELETE_PATTERN.test(actionIntent) &&
        (STORY_CREATE_PATTERN.test(actionIntent) ||
          (actionHasConfiguredStoryTerm && CREATE_PATTERN.test(actionIntent)))
      )
        addTools(selectedTools, STORY_CREATE_TOOLS);
      if (UPDATE_PATTERN.test(actionIntent))
        addTools(selectedTools, STORY_UPDATE_TOOLS);
      if (DELETE_PATTERN.test(actionIntent))
        addTools(selectedTools, STORY_DELETE_TOOLS);
      if (/\bduplicate\b/.test(actionIntent))
        selectedTools.add("duplicateStory");
      if (/\brestore\b/.test(actionIntent)) selectedTools.add("restoreStory");
      if (/\bassociation|associate\b/.test(actionIntent)) {
        selectedTools.add("addStoryAssociation");
        selectedTools.add("removeStoryAssociation");
      }
      if (/\bcomment|reply\b/.test(actionIntent)) selectedTools.add("comments");
      if (/\blabel|tag\b/.test(actionIntent)) selectedTools.add("storyLabels");
      if (ACTIVITY_PATTERN.test(actionIntent))
        selectedTools.add("storyActivities");
      if (/\blink|url\b/.test(actionIntent)) selectedTools.add("links");
      if (/\battachment|file\b/.test(actionIntent)) {
        selectedTools.add("listAttachments");
        if (DELETE_PATTERN.test(actionIntent))
          selectedTools.add("deleteAttachment");
      }
    }
  }

  if (
    !isStoryAction &&
    (TEAM_PATTERN.test(intent) || inferredDomain === "team")
  ) {
    matchedDomain = true;
    addTools(selectedTools, TEAM_READ_TOOLS);
    const isTeamStatusIntent = STATUS_PATTERN.test(actionIntent);
    if (CREATE_PATTERN.test(actionIntent) && !isTeamStatusIntent) {
      selectedTools.add("createTeamTool");
      selectedTools.add("joinTeam");
    }
    if (UPDATE_PATTERN.test(actionIntent) && !isTeamStatusIntent)
      selectedTools.add("updateTeam");
    if (DELETE_PATTERN.test(actionIntent) && !isTeamStatusIntent) {
      selectedTools.add("deleteTeam");
      selectedTools.add("leaveTeam");
    }
  }

  if (
    !isStoryAction &&
    !OBJECTIVE_PATTERN.test(intent) &&
    (STATUS_PATTERN.test(intent) || inferredDomain === "status")
  ) {
    matchedDomain = true;
    selectedTools.add("statuses");
    selectedTools.add("listTeams");
  }

  if (
    !isStoryAction &&
    (SPRINT_PATTERN.test(intent) || inferredDomain === "sprint")
  ) {
    matchedDomain = true;
    addTools(selectedTools, SPRINT_READ_TOOLS);
    if (MUTATION_PATTERN.test(actionIntent))
      selectedTools.add("updateSprintSettings");
  }

  if (
    !isStoryAction &&
    (OBJECTIVE_PATTERN.test(intent) || inferredDomain === "objective")
  ) {
    matchedDomain = true;
    addTools(selectedTools, OBJECTIVE_READ_TOOLS);
    if (MUTATION_PATTERN.test(actionIntent))
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

  if (
    !isStoryPlanningReply &&
    (PLANNING_PATTERN.test(intent) || inferredDomain === "planning")
  ) {
    matchedDomain = true;
    addTools(selectedTools, [
      "mayaWorkPlanTool",
      "applyMayaWorkPlanTool",
      "workloadPlanningTool",
      "focusBrief",
      "listTeamStories",
      "listTeamMembers",
      "resolveMember",
    ]);
  }

  if (GITHUB_PATTERN.test(intent) || inferredDomain === "github") {
    matchedDomain = true;
    addTools(selectedTools, GITHUB_READ_TOOLS);
    if (MUTATION_PATTERN.test(actionIntent))
      addTools(selectedTools, GITHUB_WRITE_TOOLS);
  }

  if (
    INTEGRATION_REQUEST_PATTERN.test(intent) ||
    inferredDomain === "integration-request"
  ) {
    matchedDomain = true;
    addTools(selectedTools, INTEGRATION_REQUEST_READ_TOOLS);
    if (MUTATION_PATTERN.test(actionIntent)) {
      addTools(selectedTools, INTEGRATION_REQUEST_WRITE_TOOLS);
    }
  }

  if (FEEDBACK_PATTERN.test(intent) || inferredDomain === "feedback") {
    matchedDomain = true;
    addTools(selectedTools, [
      "listCustomerFeedbackTool",
      "getCustomerFeedbackTool",
    ]);
  }

  if (DOCUMENT_PATTERN.test(intent) || inferredDomain === "document") {
    matchedDomain = true;
    addTools(selectedTools, ["listDocumentsTool", "getDocumentDetailsTool"]);
  }

  if (MEMORY_PATTERN.test(intent) || inferredDomain === "memory") {
    matchedDomain = true;
    selectedTools.add("listMemories");
    if (/\bremember|create|save|add\b/.test(actionIntent))
      selectedTools.add("createMemory");
    if (/\bupdate|edit|change\b/.test(actionIntent))
      selectedTools.add("updateMemory");
    if (/\bforget|delete|remove\b/.test(actionIntent))
      selectedTools.add("deleteMemory");
  }

  if (NOTIFICATION_PATTERN.test(intent) || inferredDomain === "notification") {
    matchedDomain = true;
    selectedTools.add("notifications");
  }

  if (
    ACTIVITY_PATTERN.test(intent) &&
    !hasStoryReference &&
    !usesStoryDetailContext
  ) {
    matchedDomain = true;
    selectedTools.add("activitySummaryTool");
  }

  if (
    !isStoryIntent &&
    (COMMENT_PATTERN.test(intent) || inferredDomain === "comment")
  ) {
    matchedDomain = true;
    selectedTools.add("comments");
  }

  if (
    !isStoryIntent &&
    (LABEL_PATTERN.test(intent) || inferredDomain === "label")
  ) {
    matchedDomain = true;
    selectedTools.add("labels");
  }

  if (
    !isStoryIntent &&
    (LINK_PATTERN.test(intent) || inferredDomain === "link")
  ) {
    matchedDomain = true;
    selectedTools.add("links");
  }

  if (
    !isStoryIntent &&
    (ATTACHMENT_PATTERN.test(intent) || inferredDomain === "attachment")
  ) {
    matchedDomain = true;
    selectedTools.add("listAttachments");
    if (DELETE_PATTERN.test(actionIntent))
      selectedTools.add("deleteAttachment");
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
